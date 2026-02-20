package audit

import (
	"net/http"

	authHTTP "github.com/INOVA/DML/internal/http/auth"
	"github.com/INOVA/DML/internal/http/query"
	logic "github.com/INOVA/DML/internal/logic/audit"
	"github.com/INOVA/DML/internal/response"
	"github.com/go-chi/chi/v5"
)

type AuditHandler struct {
	service *logic.AuditService
}

func NewAuditHandler(service *logic.AuditService) *AuditHandler {
	return &AuditHandler{service: service}
}

func (h *AuditHandler) RegisterRoutes(r chi.Router) {
	// Exclusively guarded by the ADMIN Requirement
	r.With(authHTTP.RequireRole("ADMIN")).Get("/", h.HandleList)
}

// @Summary List Audit Logs
// @Description Securely retrieves the compliance trail bounding global changes happening across the system.
// @Tags Audit
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number"
// @Param pageSize query int false "Items per page"
// @Param entityType query string false "Filter by entity type (User, Employee, Role)"
// @Param action query string false "Filter by action type (CREATE, UPDATE, DELETE)"
// @Success 200 {object} map[string]interface{} "Paginated log data"
// @Router /api/v1/audit-logs [get]
func (h *AuditHandler) HandleList(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := authHTTP.GetTenantIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	entityType := r.URL.Query().Get("entityType")
	action := r.URL.Query().Get("action")
	params := query.ParsePagination(r)

	logs, total, err := h.service.ListLogs(r.Context(), tenantID, entityType, action, params)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to list audit logs natively")
		return
	}

	response.PaginatedJSON(w, http.StatusOK, logs, params.Page, params.Size, int(total))
}
