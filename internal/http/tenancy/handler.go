package tenancy

import (
	"encoding/json"
	"net/http"

	logic "github.com/INOVA/DML/internal/logic/tenancy"
	"github.com/INOVA/DML/internal/response"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type Handler struct {
	service *logic.Service
}

func NewHandler(service *logic.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Get("/", h.HandleList)
	r.Post("/", h.HandleCreate)
	r.Get("/{id}", h.HandleGet)
}

func (h *Handler) HandleList(w http.ResponseWriter, r *http.Request) {
	tenants, err := h.service.ListTenants(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to list tenants")
		return
	}
	response.JSON(w, http.StatusOK, tenants)
}

func (h *Handler) HandleGet(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")

	parsedUUID, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid tenant ID format")
		return
	}

	var pgID pgtype.UUID
	pgID.Bytes = parsedUUID
	pgID.Valid = true

	tenant, err := h.service.GetTenant(r.Context(), pgID)
	if err != nil {
		// Realistically should check for sql.ErrNoRows or pgx equivalent here for 404
		response.Error(w, http.StatusNotFound, "Tenant not found")
		return
	}
	response.JSON(w, http.StatusOK, tenant)
}

type CreateRequest struct {
	Code string `json:"code" validate:"required"`
	Name string `json:"name" validate:"required"`
}

func (h *Handler) HandleCreate(w http.ResponseWriter, r *http.Request) {
	var req CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if err := response.Validate.Struct(&req); err != nil {
		response.ValidationError(w, err)
		return
	}

	newUUID := uuid.New()
	var pgID pgtype.UUID
	pgID.Bytes = newUUID
	pgID.Valid = true

	tenant, err := h.service.CreateTenant(r.Context(), pgID, req.Code, req.Name)
	if err != nil {
		response.DBError(w, err)
		return
	}

	response.JSON(w, http.StatusCreated, tenant)
}
