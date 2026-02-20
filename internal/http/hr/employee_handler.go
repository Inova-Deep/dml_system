package hr

import (
	"encoding/json"
	"net/http"

	authHTTP "github.com/INOVA/DML/internal/http/auth"
	"github.com/INOVA/DML/internal/http/query"
	logic "github.com/INOVA/DML/internal/logic/hr"
	"github.com/INOVA/DML/internal/response"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type EmployeeHandler struct {
	service *logic.EmployeeService
}

func NewEmployeeHandler(service *logic.EmployeeService) *EmployeeHandler {
	return &EmployeeHandler{service: service}
}

func (h *EmployeeHandler) RegisterRoutes(r chi.Router) {
	r.Get("/", h.HandleList)
	r.Post("/", h.HandleCreate)
	r.Get("/{id}", h.HandleGet)
	r.Get("/{id}/hierarchy", h.HandleGetHierarchy)
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

func parseOptionalUUID(idStr *string) pgtype.UUID {
	if idStr == nil || *idStr == "" {
		return pgtype.UUID{Valid: false}
	}
	parsed, err := uuid.Parse(*idStr)
	if err != nil {
		return pgtype.UUID{Valid: false}
	}
	var pgID pgtype.UUID
	pgID.Bytes = parsed
	pgID.Valid = true
	return pgID
}

// @Summary Get Employee Hierarchy
// @Description Fetches an employee and a recursively mapped tree of their direct and indirect reporting subordinates.
// @Tags Employees
// @Produce json
// @Security BearerAuth
// @Param id path string true "Employee UUID"
// @Success 200 {array} map[string]interface{}
// @Router /api/v1/employees/{id}/hierarchy [get]
func (h *EmployeeHandler) HandleGetHierarchy(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := authHTTP.GetTenantIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	idStr := chi.URLParam(r, "id")
	empID, err := parseUUIDString(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid employee ID format")
		return
	}

	hierarchy, err := h.service.GetEmployeeHierarchy(r.Context(), tenantID, empID)
	if err != nil {
		response.Error(w, http.StatusNotFound, "Hierarchy not found")
		return
	}
	response.JSON(w, http.StatusOK, hierarchy)
}

// @Summary List Employees
// @Description Get a paginated list of employees.
// @Tags Employees
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number"
// @Param pageSize query int false "Items per page"
// @Param search query string false "Search fuzzy match"
// @Success 200 {object} map[string]interface{} "Paginated Employee data"
// @Router /api/v1/employees [get]
func (h *EmployeeHandler) HandleList(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := authHTTP.GetTenantIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	params := query.ParsePagination(r)

	emps, total, err := h.service.ListEmployees(r.Context(), tenantID, params)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to list employees")
		return
	}
	response.PaginatedJSON(w, http.StatusOK, emps, params.Page, params.Size, int(total))
}

// @Summary Get Employee by ID
// @Description Fetch a single employee record.
// @Tags Employees
// @Produce json
// @Security BearerAuth
// @Param id path string true "Employee UUID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/employees/{id} [get]
func (h *EmployeeHandler) HandleGet(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := authHTTP.GetTenantIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	idStr := chi.URLParam(r, "id")
	empID, err := parseUUIDString(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid employee ID format")
		return
	}

	emp, err := h.service.GetEmployee(r.Context(), tenantID, empID)
	if err != nil {
		response.Error(w, http.StatusNotFound, "Employee not found")
		return
	}
	response.JSON(w, http.StatusOK, emp)
}

type CreateEmployeeRequest struct {
	EmployeeNo     string  `json:"employeeNo" validate:"required"`
	FirstName      string  `json:"firstName" validate:"required"`
	LastName       string  `json:"lastName" validate:"required"`
	DisplayName    *string `json:"displayName"`
	WorkEmail      *string `json:"workEmail" validate:"omitempty,email"`
	BusinessUnitID *string `json:"businessUnitId" validate:"omitempty,uuid"`
	DepartmentID   *string `json:"departmentId" validate:"omitempty,uuid"`
	JobTitleID     *string `json:"jobTitleId" validate:"omitempty,uuid"`
	ManagerID      *string `json:"managerId" validate:"omitempty,uuid"`
}

// @Summary Create an Employee
// @Description Adds a new employee to the human resources schema. Supports assigning to Organizational Units (Business Unit, Department, Job Title) and a Manager for tracking hierarchical reporting lines.
// @Tags Employees
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateEmployeeRequest true "Employee Payload"
// @Success 201 {object} map[string]interface{}
// @Router /api/v1/employees [post]
func (h *EmployeeHandler) HandleCreate(w http.ResponseWriter, r *http.Request) {
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

	var req CreateEmployeeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if err := response.Validate.Struct(&req); err != nil {
		response.ValidationError(w, err)
		return
	}

	empID, _ := parseUUIDString(uuid.New().String())
	busID := parseOptionalUUID(req.BusinessUnitID)
	deptID := parseOptionalUUID(req.DepartmentID)
	jobID := parseOptionalUUID(req.JobTitleID)
	mgrID := parseOptionalUUID(req.ManagerID)

	emp, err := h.service.CreateEmployee(r.Context(), empID, tenantID, actorID, req.EmployeeNo, req.FirstName, req.LastName, req.DisplayName, req.WorkEmail, busID, deptID, jobID, mgrID)
	if err != nil {
		response.DBError(w, err)
		return
	}

	response.JSON(w, http.StatusCreated, emp)
}
