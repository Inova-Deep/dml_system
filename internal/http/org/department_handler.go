package org

import (
	"encoding/json"
	"net/http"

	authHTTP "github.com/INOVA/DML/internal/http/auth"
	"github.com/INOVA/DML/internal/http/query"
	logic "github.com/INOVA/DML/internal/logic/org"
	"github.com/INOVA/DML/internal/response"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type DepartmentHandler struct {
	service *logic.DepartmentService
}

func NewDepartmentHandler(service *logic.DepartmentService) *DepartmentHandler {
	return &DepartmentHandler{service: service}
}

func (h *DepartmentHandler) RegisterRoutes(r chi.Router) {
	r.Get("/", h.HandleList)
	r.Post("/", h.HandleCreate)
	r.Get("/{id}", h.HandleGet)
}

// HandleList godoc
// @Summary      List departments
// @Description  Retrieves a paginated list of departments for the authenticated tenant.
// @Tags         Organization
// @Accept       json
// @Produce      json
// @Param        page    query     int     false  "Page number" default(1)
// @Param        size    query     int     false  "Page size" default(50)
// @Param        search  query     string  false  "Search term (name/code)"
// @Security     BearerAuth
// @Success      200     {object}  map[string]interface{} "Paginated department data"
// @Failure      401     {object}  map[string]interface{} "Unauthorized"
// @Failure      500     {object}  map[string]interface{} "Internal server error"
// @Router       /api/v1/departments [get]
func (h *DepartmentHandler) HandleList(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := authHTTP.GetTenantIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	params := query.ParsePagination(r)

	depts, total, err := h.service.ListDepartments(r.Context(), tenantID, params)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to list departments")
		return
	}
	response.PaginatedJSON(w, http.StatusOK, depts, params.Page, params.Size, int(total))
}

// HandleGet godoc
// @Summary      Get a department
// @Description  Retrieves a specific department by its ID.
// @Tags         Organization
// @Accept       json
// @Produce      json
// @Param        id      path      string  true  "Department ID"
// @Security     BearerAuth
// @Success      200     {object}  map[string]interface{} "Department data"
// @Failure      400     {object}  map[string]interface{} "Invalid ID format"
// @Failure      401     {object}  map[string]interface{} "Unauthorized"
// @Failure      404     {object}  map[string]interface{} "Not found"
// @Router       /api/v1/departments/{id} [get]
func (h *DepartmentHandler) HandleGet(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := authHTTP.GetTenantIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	idStr := chi.URLParam(r, "id")
	deptID, err := parseUUIDString(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid department ID format")
		return
	}

	dept, err := h.service.GetDepartment(r.Context(), tenantID, deptID)
	if err != nil {
		response.Error(w, http.StatusNotFound, "Department not found")
		return
	}
	response.JSON(w, http.StatusOK, dept)
}

type CreateDeptRequest struct {
	Code               string  `json:"code" validate:"required"`
	Name               string  `json:"name" validate:"required"`
	ParentDepartmentID *string `json:"parentDepartmentId"`
}

func (h *DepartmentHandler) HandleCreate(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := authHTTP.GetTenantIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req CreateDeptRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if err := response.Validate.Struct(&req); err != nil {
		response.ValidationError(w, err)
		return
	}

	deptID, _ := parseUUIDString(uuid.New().String())

	var pgParentID *pgtype.UUID
	if req.ParentDepartmentID != nil {
		parsed, err := parseUUIDString(*req.ParentDepartmentID)
		if err == nil {
			pgParentID = &parsed
		}
	}

	dept, err := h.service.CreateDepartment(r.Context(), deptID, tenantID, pgParentID, req.Code, req.Name)
	if err != nil {
		response.DBError(w, err)
		return
	}

	response.JSON(w, http.StatusCreated, dept)
}
