package hr

import (
	"context"
	"fmt"

	"github.com/INOVA/DML/internal/db"
	"github.com/INOVA/DML/internal/domain"
	"github.com/INOVA/DML/internal/logic/audit"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/crypto/bcrypt"
)

type OnboardingService struct {
	db       *db.DB
	auditSvc *audit.AuditService
}

func NewOnboardingService(database *db.DB, audit *audit.AuditService) *OnboardingService {
	return &OnboardingService{
		db:       database,
		auditSvc: audit,
	}
}

type OnboardingResult struct {
	EmployeeID string `json:"employeeId"`
	UserID     string `json:"userId"`
	Email      string `json:"email"`
}

func (s *OnboardingService) ExecuteOnboarding(
	ctx context.Context,
	tenantID pgtype.UUID,
	actorID pgtype.UUID,
	empNo, first, last string,
	display *string,
	email, passwordHash string,
	initialRoleID pgtype.UUID,
	busID, deptID, jobID, mgrID pgtype.UUID,
) (OnboardingResult, error) {

	tx, err := s.db.Pool.Begin(ctx)
	if err != nil {
		return OnboardingResult{}, fmt.Errorf("failed to begin onboarding transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	qtx := domain.New(tx)

	// 1. Create Employee
	empIDBytes := uuid.New()
	var newEmpID pgtype.UUID
	newEmpID.Bytes = empIDBytes
	newEmpID.Valid = true

	var pgDisplay pgtype.Text
	if display != nil && *display != "" {
		pgDisplay.String = *display
		pgDisplay.Valid = true
	}

	var pgEmail pgtype.Text
	pgEmail.String = email
	pgEmail.Valid = true

	_, err = qtx.CreateEmployee(ctx, domain.CreateEmployeeParams{
		ID:             newEmpID,
		TenantID:       tenantID,
		EmployeeNo:     empNo,
		FirstName:      first,
		LastName:       last,
		DisplayName:    pgDisplay,
		WorkEmail:      pgEmail,
		BusinessUnitID: busID,
		DepartmentID:   deptID,
		JobTitleID:     jobID,
		ManagerID:      mgrID,
	})
	if err != nil {
		return OnboardingResult{}, fmt.Errorf("failed creating employee record: %w", err)
	}

	// 2. Create User Identity
	userIDBytes := uuid.New()
	var newUserID pgtype.UUID
	newUserID.Bytes = userIDBytes
	newUserID.Valid = true

	var pgPass pgtype.Text

	bytes, hashErr := bcrypt.GenerateFromPassword([]byte(passwordHash), 14)
	if hashErr != nil {
		return OnboardingResult{}, fmt.Errorf("failed hashing user auth bounds securely: %w", hashErr)
	}

	pgPass.String = string(bytes)
	pgPass.Valid = true

	_, err = qtx.CreateUser(ctx, domain.CreateUserParams{
		ID:           newUserID,
		TenantID:     tenantID,
		EmployeeID:   newEmpID,
		Email:        email,
		DisplayName:  pgDisplay,
		PasswordHash: pgPass,
	})
	if err != nil {
		return OnboardingResult{}, fmt.Errorf("failed creating identity record: %w", err)
	}

	// 3. Assign Default Role
	// Ensure the role actually belongs to this tenant natively
	_, err = qtx.GetRole(ctx, domain.GetRoleParams{
		TenantID: tenantID,
		ID:       initialRoleID,
	})
	if err != nil {
		return OnboardingResult{}, fmt.Errorf("failed resolving target role context: %w", err)
	}

	err = qtx.AssignUserRole(ctx, domain.AssignUserRoleParams{
		TenantID:        tenantID,
		UserID:          newUserID,
		RoleID:          initialRoleID,
		BusinessUnitID:  busID,
		DepartmentID:    deptID,
		GrantedByUserID: actorID,
	})
	if err != nil {
		return OnboardingResult{}, fmt.Errorf("failed mapping underlying user roles: %w", err)
	}

	// Commit Transaction safely
	if err := tx.Commit(ctx); err != nil {
		return OnboardingResult{}, fmt.Errorf("failed committing onboarding transaction bounds: %w", err)
	}

	// Asynchronous Audit Logging safely triggered upon transaction completion bounds securely
	if s.auditSvc != nil {
		s.auditSvc.Log(tenantID, actorID, "ONBOARD", "Users", newUserID.Bytes, map[string]interface{}{
			"action":         "Complete Onboarding Flow",
			"employee_no":    empNo,
			"target_role_id": initialRoleID.Bytes,
		})
	}

	return OnboardingResult{
		EmployeeID: empIDBytes.String(),
		UserID:     userIDBytes.String(),
		Email:      email,
	}, nil
}
