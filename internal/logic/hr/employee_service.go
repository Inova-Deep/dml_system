package hr

import (
	"context"
	"fmt"

	"github.com/INOVA/DML/internal/db"
	"github.com/INOVA/DML/internal/domain"
	"github.com/INOVA/DML/internal/http/query"
	"github.com/INOVA/DML/internal/logic/audit"
	"github.com/jackc/pgx/v5/pgtype"
)

type EmployeeService struct {
	queries  *domain.Queries
	auditSvc *audit.AuditService
}

func NewEmployeeService(database *db.DB, auditSvc *audit.AuditService) *EmployeeService {
	return &EmployeeService{
		queries:  domain.New(database.Pool),
		auditSvc: auditSvc,
	}
}

func (s *EmployeeService) CreateEmployee(ctx context.Context, id, tenantID, actorID pgtype.UUID, empNo, first, last string, display, email *string, busID, deptID, jobID, mgrID pgtype.UUID) (domain.Employee, error) {
	var pgDisplay pgtype.Text
	if display != nil && *display != "" {
		pgDisplay.String = *display
		pgDisplay.Valid = true
	}

	var pgEmail pgtype.Text
	if email != nil && *email != "" {
		pgEmail.String = *email
		pgEmail.Valid = true
	}

	// Structural enforcement tracking
	if mgrID.Valid {
		_, err := s.queries.GetEmployee(ctx, domain.GetEmployeeParams{
			TenantID: tenantID,
			ID:       mgrID,
		})
		if err != nil {
			return domain.Employee{}, fmt.Errorf("designated manager does not exist or is inaccessible: %w", err)
		}
	}

	emp, err := s.queries.CreateEmployee(ctx, domain.CreateEmployeeParams{
		ID:             id,
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

	if err == nil && s.auditSvc != nil {
		s.auditSvc.Log(tenantID, actorID, "CREATE", "Employees", id.Bytes, map[string]interface{}{
			"employee_no": empNo,
			"first_name":  first,
			"last_name":   last,
			"work_email":  email,
		})
	}

	return emp, err
}

func (s *EmployeeService) ListEmployees(ctx context.Context, tenantID pgtype.UUID, params query.PaginationParams) ([]domain.Employee, int64, error) {
	emps, err := s.queries.ListEmployees(ctx, domain.ListEmployeesParams{
		TenantID: tenantID,
		Search:   params.Search,
		Limit:    params.Limit(),
		Offset:   params.Offset(),
	})
	if err != nil {
		return nil, 0, err
	}

	total, err := s.queries.CountEmployees(ctx, domain.CountEmployeesParams{
		TenantID: tenantID,
		Search:   params.Search,
	})
	if err != nil {
		return nil, 0, err
	}

	return emps, total, nil
}

func (s *EmployeeService) GetEmployee(ctx context.Context, tenantID, id pgtype.UUID) (domain.Employee, error) {
	return s.queries.GetEmployee(ctx, domain.GetEmployeeParams{
		TenantID: tenantID,
		ID:       id,
	})
}

func (s *EmployeeService) GetEmployeeHierarchy(ctx context.Context, tenantID, employeeID pgtype.UUID) ([]domain.Employee, error) {
	rows, err := s.queries.GetEmployeeHierarchy(ctx, domain.GetEmployeeHierarchyParams{
		TenantID: tenantID,
		ID:       employeeID,
	})
	if err != nil {
		return nil, err
	}

	emps := make([]domain.Employee, len(rows))
	for i, r := range rows {
		emps[i] = domain.Employee{
			ID:             r.ID,
			TenantID:       r.TenantID,
			EmployeeNo:     r.EmployeeNo,
			FirstName:      r.FirstName,
			LastName:       r.LastName,
			DisplayName:    r.DisplayName,
			WorkEmail:      r.WorkEmail,
			Status:         r.Status,
			IsActive:       r.IsActive,
			CreatedAt:      r.CreatedAt,
			UpdatedAt:      r.UpdatedAt,
			BusinessUnitID: r.BusinessUnitID,
			DepartmentID:   r.DepartmentID,
			JobTitleID:     r.JobTitleID,
			ManagerID:      r.ManagerID,
		}
	}
	return emps, nil
}
