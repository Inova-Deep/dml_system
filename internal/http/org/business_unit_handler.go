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

type BusinessUnitHandler struct {
	service *logic.BusinessUnitService
}

func NewBusinessUnitHandler(service *logic.BusinessUnitService) *BusinessUnitHandler {
	return &BusinessUnitHandler{service: service}
}

func (h *BusinessUnitHandler) RegisterRoutes(r chi.Router) {
	// Note: In an enterprise app, the TenantID should be pulled from a JWT context middleware.
	// For this scaffolding phase, we'll accept tenant_id as a header.
	r.Get("/", h.HandleList)
	r.Post("/", h.HandleCreate)
	r.Get("/{id}", h.HandleGet)
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
// @Summary      List business units
// @Description  Retrieves a paginated list of business units for the authenticated tenant.
// @Tags         Organization
// @Accept       json
// @Produce      json
// @Param        page    query     int     false  "Page number" default(1)
// @Param        size    query     int     false  "Page size" default(50)
// @Param        search  query     string  false  "Search term (name/code)"
// @Security     BearerAuth
// @Success      200     {object}  map[string]interface{} "Paginated business unit data"
// @Failure      401     {object}  map[string]interface{} "Unauthorized"
// @Failure      500     {object}  map[string]interface{} "Internal server error"
// @Router       /api/v1/business-units [get]
func (h *BusinessUnitHandler) HandleList(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := authHTTP.GetTenantIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	params := query.ParsePagination(r)

	units, total, err := h.service.ListBusinessUnits(r.Context(), tenantID, params)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to list business units")
		return
	}
	response.PaginatedJSON(w, http.StatusOK, units, params.Page, params.Size, int(total))
}

// HandleGet godoc
// @Summary      Get a business unit
// @Description  Retrieves a specific business unit by its ID.
// @Tags         Organization
// @Accept       json
// @Produce      json
// @Param        id      path      string  true  "Business Unit ID"
// @Security     BearerAuth
// @Success      200     {object}  map[string]interface{} "Business unit data"
// @Failure      400     {object}  map[string]interface{} "Invalid ID format"
// @Failure      401     {object}  map[string]interface{} "Unauthorized"
// @Failure      404     {object}  map[string]interface{} "Not found"
// @Router       /api/v1/business-units/{id} [get]
func (h *BusinessUnitHandler) HandleGet(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := authHTTP.GetTenantIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	idStr := chi.URLParam(r, "id")
	buID, err := parseUUIDString(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid business unit ID format")
		return
	}

	unit, err := h.service.GetBusinessUnit(r.Context(), tenantID, buID)
	if err != nil {
		response.Error(w, http.StatusNotFound, "Business unit not found")
		return
	}
	response.JSON(w, http.StatusOK, unit)
}

type CreateBURequest struct {
	Code string `json:"code" validate:"required"`
	Name string `json:"name" validate:"required"`
}

func (h *BusinessUnitHandler) HandleCreate(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := authHTTP.GetTenantIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req CreateBURequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if err := response.Validate.Struct(&req); err != nil {
		response.ValidationError(w, err)
		return
	}

	buID, _ := parseUUIDString(uuid.New().String())

	unit, err := h.service.CreateBusinessUnit(r.Context(), buID, tenantID, req.Code, req.Name)
	if err != nil {
		response.DBError(w, err)
		return
	}

	response.JSON(w, http.StatusCreated, unit)
}
