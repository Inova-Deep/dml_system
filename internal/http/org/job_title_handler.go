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
)

type JobTitleHandler struct {
	service *logic.JobTitleService
}

func NewJobTitleHandler(service *logic.JobTitleService) *JobTitleHandler {
	return &JobTitleHandler{service: service}
}

func (h *JobTitleHandler) RegisterRoutes(r chi.Router) {
	r.Get("/", h.HandleList)
	r.Post("/", h.HandleCreate)
	r.Get("/{id}", h.HandleGet)
}

// HandleList godoc
// @Summary      List job titles
// @Description  Retrieves a paginated list of job titles for the authenticated tenant.
// @Tags         Organization
// @Accept       json
// @Produce      json
// @Param        page    query     int     false  "Page number" default(1)
// @Param        size    query     int     false  "Page size" default(50)
// @Param        search  query     string  false  "Search term (name/code)"
// @Security     BearerAuth
// @Success      200     {object}  map[string]interface{} "Paginated job title data"
// @Failure      401     {object}  map[string]interface{} "Unauthorized"
// @Failure      500     {object}  map[string]interface{} "Internal server error"
// @Router       /api/v1/job-titles [get]
func (h *JobTitleHandler) HandleList(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := authHTTP.GetTenantIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	params := query.ParsePagination(r)

	jobs, total, err := h.service.ListJobTitles(r.Context(), tenantID, params)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to list job titles")
		return
	}
	response.PaginatedJSON(w, http.StatusOK, jobs, params.Page, params.Size, int(total))
}

// HandleGet godoc
// @Summary      Get Job Title
// @Description  Get job title by ID
// @Tags         organization
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Job Title ID"
// @Security BearerAuth
// @Success      200  {object}  map[string]interface{} "Job title data"
// @Failure      400  {object}  map[string]interface{} "Invalid ID format"
// @Failure      401  {object}  map[string]interface{} "Unauthorized"
// @Failure      404  {object}  map[string]interface{} "Not found"
// @Router       /api/v1/job-titles/{id} [get]
func (h *JobTitleHandler) HandleGet(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := authHTTP.GetTenantIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	idStr := chi.URLParam(r, "id")
	jobID, err := parseUUIDString(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid job title ID format")
		return
	}

	job, err := h.service.GetJobTitle(r.Context(), tenantID, jobID)
	if err != nil {
		response.Error(w, http.StatusNotFound, "Job title not found")
		return
	}
	response.JSON(w, http.StatusOK, job)
}

type CreateJobTitleRequest struct {
	Code  string `json:"code" validate:"required"`
	Name  string `json:"name" validate:"required"`
	Grade string `json:"grade"`
}

func (h *JobTitleHandler) HandleCreate(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := authHTTP.GetTenantIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req CreateJobTitleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if err := response.Validate.Struct(&req); err != nil {
		response.ValidationError(w, err)
		return
	}

	jobID, _ := parseUUIDString(uuid.New().String())

	job, err := h.service.CreateJobTitle(r.Context(), jobID, tenantID, req.Code, req.Name, req.Grade)
	if err != nil {
		response.DBError(w, err)
		return
	}

	response.JSON(w, http.StatusCreated, job)
}
