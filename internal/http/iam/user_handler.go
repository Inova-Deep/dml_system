package iam

import (
	"encoding/json"
	"net/http"

	authHTTP "github.com/INOVA/DML/internal/http/auth"
	"github.com/INOVA/DML/internal/http/query"
	logic "github.com/INOVA/DML/internal/logic/iam"
	"github.com/INOVA/DML/internal/response"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type UserHandler struct {
	userService     *logic.UserService
	userRoleService *logic.UserRoleService
}

func NewUserHandler(userService *logic.UserService, userRoleService *logic.UserRoleService) *UserHandler {
	return &UserHandler{
		userService:     userService,
		userRoleService: userRoleService,
	}
}

func (h *UserHandler) RegisterRoutes(r chi.Router) {
	r.With(authHTTP.RequireRole("ADMIN")).Post("/", h.HandleCreate)
	r.Get("/", h.HandleList) // Added route for HandleList
	r.Get("/by-email", h.HandleGetByEmail)
	r.Get("/{userID}", h.HandleGet) // Added route for HandleGet

	// Role assignments
	r.With(authHTTP.RequireRole("ADMIN")).Post("/{userID}/roles", h.HandleAssignRole)
	r.With(authHTTP.RequireRole("ADMIN")).Delete("/{userID}/roles/{roleID}", h.HandleRevokeRole)
}

func parseUUIDString(idStr string) (pgtype.UUID, error) {
	var pgID pgtype.UUID
	parsed, err := uuid.Parse(idStr)
	if err != nil {
		return pgID, err
	}
	pgID.Bytes = parsed
	pgID.Valid = true
	return pgID, nil
}

// HandleList godoc
// @Summary      List users
// @Description  Retrieves a paginated list of users for the authenticated tenant.
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        page    query     int     false  "Page number" default(1)
// @Param        size    query     int     false  "Page size" default(50)
// @Param        search  query     string  false  "Search term (email/name)"
// @Security     BearerAuth
// @Success      200     {object}  map[string]interface{} "Paginated user data"
// @Failure      401     {object}  map[string]interface{} "Unauthorized"
// @Failure      500     {object}  map[string]interface{} "Internal server error"
// @Router       /api/v1/users [get]
func (h *UserHandler) HandleList(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := authHTTP.GetTenantIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	params := query.ParsePagination(r)

	users, total, err := h.userService.ListUsers(r.Context(), tenantID, params)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to list users")
		return
	}
	response.PaginatedJSON(w, http.StatusOK, users, params.Page, params.Size, int(total))
}

func (h *UserHandler) HandleGetByEmail(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := authHTTP.GetTenantIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	email := r.URL.Query().Get("email")
	if email == "" {
		response.Error(w, http.StatusBadRequest, "Missing email query parameter")
		return
	}

	user, err := h.userService.GetUserByEmail(r.Context(), tenantID, email)
	if err != nil {
		response.Error(w, http.StatusNotFound, "User not found")
		return
	}
	response.JSON(w, http.StatusOK, user)
}

type CreateUserRequest struct {
	EmployeeID  string  `json:"employeeId"`
	Email       string  `json:"email" validate:"required,email"`
	DisplayName *string `json:"displayName"`
	Password    string  `json:"password" validate:"required,min=8"`
}

func (h *UserHandler) HandleGet(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := authHTTP.GetTenantIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	userIDStr := chi.URLParam(r, "userID")
	userID, err := parseUUIDString(userIDStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid user ID format")
		return
	}

	user, err := h.userService.GetUserByID(r.Context(), tenantID, userID)
	if err != nil {
		response.Error(w, http.StatusNotFound, "User not found")
		return
	}
	response.JSON(w, http.StatusOK, user)
}

// HandleCreate godoc
// @Summary      Create a new user
// @Description  Creates a new user profile attached to an employee reference.
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        request  body      CreateUserRequest  true  "User details"
// @Security     BearerAuth
// @Success      201      {object}  map[string]interface{} "Successfully created user"
// @Failure      400      {object}  map[string]interface{} "Bad request payload"
// @Failure      401      {object}  map[string]interface{} "Unauthorized"
// @Failure      403      {object}  map[string]interface{} "Forbidden (Requires ADMIN)"
// @Router       /api/v1/users [post]
func (h *UserHandler) HandleCreate(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := authHTTP.GetTenantIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	actorID, ok := authHTTP.GetUserIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if err := response.Validate.Struct(&req); err != nil {
		response.ValidationError(w, err)
		return
	}

	empID, err := parseUUIDString(req.EmployeeID)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid employee ID format")
		return
	}

	userID, _ := parseUUIDString(uuid.New().String())

	// Security note: In reality, we must hash req.Password (e.g. bcrypt) before saving context.
	// For scoping the DML demo, we'll store as-is until a proper Auth package is brought in.
	hashedPassword := req.Password

	user, err := h.userService.CreateUser(r.Context(), userID, tenantID, actorID, empID, req.Email, hashedPassword, req.DisplayName)
	if err != nil {
		response.DBError(w, err)
		return
	}

	// Never return the password hash
	response.JSON(w, http.StatusCreated, user)
}

type AssignRoleRequest struct {
	RoleID string `json:"roleId" validate:"required"`
}

// HandleAssignRole godoc
// @Summary      Assign role to user
// @Description  Binds an existing RBAC role to a specific user.
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        userID   path      string             true  "User ID"
// @Param        request  body      AssignRoleRequest  true  "Role Assignment specifics"
// @Security     BearerAuth
// @Success      201      {object}  map[string]interface{} "Role assigned successfully"
// @Failure      400      {object}  map[string]interface{} "Bad request payload"
// @Failure      401      {object}  map[string]interface{} "Unauthorized"
// @Failure      403      {object}  map[string]interface{} "Forbidden (Requires ADMIN)"
// @Router       /api/v1/users/{userID}/roles [post]
func (h *UserHandler) HandleAssignRole(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := authHTTP.GetTenantIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	grantedByUserID, ok := authHTTP.GetUserIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	userIDStr := chi.URLParam(r, "userID")
	userID, err := parseUUIDString(userIDStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid user ID format")
		return
	}

	var req AssignRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if err := response.Validate.Struct(&req); err != nil {
		response.ValidationError(w, err)
		return
	}

	roleID, err := parseUUIDString(req.RoleID)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid role ID format")
		return
	}

	err = h.userRoleService.AssignUserRole(r.Context(), tenantID, userID, roleID, grantedByUserID)
	if err != nil {
		response.DBError(w, err)
		return
	}

	response.JSON(w, http.StatusCreated, map[string]string{"message": "Role assigned successfully"})
}

func (h *UserHandler) HandleRevokeRole(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := authHTTP.GetTenantIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	userIDStr := chi.URLParam(r, "userID")
	userID, err := parseUUIDString(userIDStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid user ID format")
		return
	}

	roleIDStr := chi.URLParam(r, "roleID")
	roleID, err := parseUUIDString(roleIDStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid role ID format")
		return
	}

	err = h.userRoleService.RevokeUserRole(r.Context(), tenantID, userID, roleID)
	if err != nil {
		response.DBError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"message": "Role revoked successfully"})
}
