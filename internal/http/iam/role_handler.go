package iam

import (
	"encoding/json"
	"net/http"

	authHTTP "github.com/INOVA/DML/internal/http/auth"
	logic "github.com/INOVA/DML/internal/logic/iam"
	"github.com/INOVA/DML/internal/response"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type RoleHandler struct {
	service *logic.RoleService
}

func NewRoleHandler(service *logic.RoleService) *RoleHandler {
	return &RoleHandler{service: service}
}

func (h *RoleHandler) RegisterRoutes(r chi.Router) {
	r.Get("/", h.HandleList)
	r.With(authHTTP.RequireRole("ADMIN")).Post("/", h.HandleCreate)
	r.Get("/{id}", h.HandleGet)
}

func (h *RoleHandler) HandleList(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := authHTTP.GetTenantIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	roles, err := h.service.ListRoles(r.Context(), tenantID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to list roles")
		return
	}
	response.JSON(w, http.StatusOK, roles)
}

func (h *RoleHandler) HandleGet(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := authHTTP.GetTenantIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	idStr := chi.URLParam(r, "id")
	roleID, err := parseUUIDString(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid role ID format")
		return
	}

	role, err := h.service.GetRole(r.Context(), tenantID, roleID)
	if err != nil {
		response.Error(w, http.StatusNotFound, "Role not found")
		return
	}
	response.JSON(w, http.StatusOK, role)
}

type CreateRoleRequest struct {
	Code        string  `json:"code" validate:"required"`
	Name        string  `json:"name" validate:"required"`
	Description *string `json:"description"`
}

func (h *RoleHandler) HandleCreate(w http.ResponseWriter, r *http.Request) {
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

	var req CreateRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if err := response.Validate.Struct(&req); err != nil {
		response.ValidationError(w, err)
		return
	}

	roleID, _ := parseUUIDString(uuid.New().String())

	role, err := h.service.CreateRole(r.Context(), roleID, tenantID, actorID, req.Code, req.Name, req.Description)
	if err != nil {
		response.DBError(w, err)
		return
	}

	response.JSON(w, http.StatusCreated, role)
}
