package hr

import (
	"encoding/json"
	"net/http"

	authHTTP "github.com/INOVA/DML/internal/http/auth"
	logic "github.com/INOVA/DML/internal/logic/hr"
	"github.com/INOVA/DML/internal/response"
	"github.com/go-chi/chi/v5"
)

type OnboardingHandler struct {
	service *logic.OnboardingService
}

func NewOnboardingHandler(service *logic.OnboardingService) *OnboardingHandler {
	return &OnboardingHandler{service: service}
}

func (h *OnboardingHandler) RegisterRoutes(r chi.Router) {
	r.With(authHTTP.RequireRole("ADMIN")).Post("/", h.HandleOnboard)
}

type OnboardRequest struct {
	EmployeeNo     string  `json:"employeeNo" validate:"required"`
	FirstName      string  `json:"firstName" validate:"required"`
	LastName       string  `json:"lastName" validate:"required"`
	DisplayName    *string `json:"displayName"`
	WorkEmail      string  `json:"workEmail" validate:"required,email"`
	BusinessUnitID *string `json:"businessUnitId" validate:"omitempty,uuid"`
	DepartmentID   *string `json:"departmentId" validate:"omitempty,uuid"`
	JobTitleID     *string `json:"jobTitleId" validate:"omitempty,uuid"`
	ManagerID      *string `json:"managerId" validate:"omitempty,uuid"`

	// User / Auth Info
	Password      string `json:"password" validate:"required,min=8"`
	InitialRoleID string `json:"initialRoleId" validate:"required,uuid"`
}

// @Summary Onboard new Staff Member
// @Description Natively constructs the Employee profile, creates the Identity provider User account securely, assigns the primary RBAC Role, and safely tracks an Audit stream atomically using Postgres Transactions securely bound.
// @Tags Onboarding
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body OnboardRequest true "Comprehensive Onboarding Details"
// @Success 201 {object} map[string]interface{} "Successfully completed sequence returning the active Identity UUID map structure"
// @Router /api/v1/onboard [post]
func (h *OnboardingHandler) HandleOnboard(w http.ResponseWriter, r *http.Request) {
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

	var req OnboardRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if err := response.Validate.Struct(&req); err != nil {
		response.ValidationError(w, err)
		return
	}

	busID := parseOptionalUUID(req.BusinessUnitID)
	deptID := parseOptionalUUID(req.DepartmentID)
	jobID := parseOptionalUUID(req.JobTitleID)
	mgrID := parseOptionalUUID(req.ManagerID)

	roleID, _ := parseUUIDString(req.InitialRoleID)

	res, err := h.service.ExecuteOnboarding(
		r.Context(),
		tenantID,
		actorID,
		req.EmployeeNo,
		req.FirstName,
		req.LastName,
		req.DisplayName,
		req.WorkEmail,
		req.Password,
		roleID,
		busID,
		deptID,
		jobID,
		mgrID,
	)

	if err != nil {
		response.DBError(w, err)
		return
	}

	response.JSON(w, http.StatusCreated, res)
}
