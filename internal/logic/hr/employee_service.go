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

type BusinessUnitSummary struct {
	ID   pgtype.UUID `json:"id"`
	Code *string     `json:"code"`
	Name *string     `json:"name"`
}

type DepartmentSummary struct {
	ID   pgtype.UUID `json:"id"`
	Code *string     `json:"code"`
	Name *string     `json:"name"`
}

type JobTitleSummary struct {
	ID    pgtype.UUID `json:"id"`
	Code  *string     `json:"code"`
	Name  *string     `json:"name"`
	Grade *string     `json:"grade"`
}

type ManagerSummary struct {
	ID          pgtype.UUID `json:"id"`
	EmployeeNo  *string     `json:"employeeNo"`
	FirstName   *string     `json:"firstName"`
	LastName    *string     `json:"lastName"`
	DisplayName *string     `json:"displayName"`
}

type EmployeeWithDetails struct {
	ID           pgtype.UUID          `json:"id"`
	TenantID     pgtype.UUID          `json:"tenantId"`
	EmployeeNo   string               `json:"employeeNo"`
	FirstName    string               `json:"firstName"`
	LastName     string               `json:"lastName"`
	DisplayName  pgtype.Text          `json:"displayName"`
	WorkEmail    pgtype.Text          `json:"workEmail"`
	Status       string               `json:"status"`
	IsActive     bool                 `json:"isActive"`
	CreatedAt    pgtype.Timestamptz   `json:"createdAt"`
	UpdatedAt    pgtype.Timestamptz   `json:"updatedAt"`
	BusinessUnit *BusinessUnitSummary `json:"businessUnit"`
	Department   *DepartmentSummary   `json:"department"`
	JobTitle     *JobTitleSummary     `json:"jobTitle"`
	Manager      *ManagerSummary      `json:"manager"`
}

func mapRowToEmployeeWithDetails(row domain.GetEmployeeWithDetailsRow) EmployeeWithDetails {
	emp := EmployeeWithDetails{
		ID:          row.ID,
		TenantID:    row.TenantID,
		EmployeeNo:  row.EmployeeNo,
		FirstName:   row.FirstName,
		LastName:    row.LastName,
		DisplayName: row.DisplayName,
		WorkEmail:   row.WorkEmail,
		Status:      row.Status,
		IsActive:    row.IsActive,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}

	if row.BusinessUnitID.Valid {
		emp.BusinessUnit = &BusinessUnitSummary{
			ID: row.BusinessUnitID,
		}
		if row.BusinessUnitCode.Valid {
			emp.BusinessUnit.Code = &row.BusinessUnitCode.String
		}
		if row.BusinessUnitName.Valid {
			emp.BusinessUnit.Name = &row.BusinessUnitName.String
		}
	}

	if row.DepartmentID.Valid {
		emp.Department = &DepartmentSummary{
			ID: row.DepartmentID,
		}
		if row.DepartmentCode.Valid {
			emp.Department.Code = &row.DepartmentCode.String
		}
		if row.DepartmentName.Valid {
			emp.Department.Name = &row.DepartmentName.String
		}
	}

	if row.JobTitleID.Valid {
		emp.JobTitle = &JobTitleSummary{
			ID: row.JobTitleID,
		}
		if row.JobTitleCode.Valid {
			emp.JobTitle.Code = &row.JobTitleCode.String
		}
		if row.JobTitleName.Valid {
			emp.JobTitle.Name = &row.JobTitleName.String
		}
		if row.JobTitleGrade.Valid {
			emp.JobTitle.Grade = &row.JobTitleGrade.String
		}
	}

	if row.ManagerID.Valid {
		emp.Manager = &ManagerSummary{
			ID: row.ManagerID,
		}
		if row.ManagerEmployeeNo.Valid {
			emp.Manager.EmployeeNo = &row.ManagerEmployeeNo.String
		}
		if row.ManagerFirstName.Valid {
			emp.Manager.FirstName = &row.ManagerFirstName.String
		}
		if row.ManagerLastName.Valid {
			emp.Manager.LastName = &row.ManagerLastName.String
		}
		if row.ManagerDisplayName.Valid {
			emp.Manager.DisplayName = &row.ManagerDisplayName.String
		}
	}

	return emp
}

func mapListRowToEmployeeWithDetails(row domain.ListEmployeesWithDetailsRow) EmployeeWithDetails {
	emp := EmployeeWithDetails{
		ID:          row.ID,
		TenantID:    row.TenantID,
		EmployeeNo:  row.EmployeeNo,
		FirstName:   row.FirstName,
		LastName:    row.LastName,
		DisplayName: row.DisplayName,
		WorkEmail:   row.WorkEmail,
		Status:      row.Status,
		IsActive:    row.IsActive,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}

	if row.BusinessUnitID.Valid {
		emp.BusinessUnit = &BusinessUnitSummary{
			ID: row.BusinessUnitID,
		}
		if row.BusinessUnitCode.Valid {
			emp.BusinessUnit.Code = &row.BusinessUnitCode.String
		}
		if row.BusinessUnitName.Valid {
			emp.BusinessUnit.Name = &row.BusinessUnitName.String
		}
	}

	if row.DepartmentID.Valid {
		emp.Department = &DepartmentSummary{
			ID: row.DepartmentID,
		}
		if row.DepartmentCode.Valid {
			emp.Department.Code = &row.DepartmentCode.String
		}
		if row.DepartmentName.Valid {
			emp.Department.Name = &row.DepartmentName.String
		}
	}

	if row.JobTitleID.Valid {
		emp.JobTitle = &JobTitleSummary{
			ID: row.JobTitleID,
		}
		if row.JobTitleCode.Valid {
			emp.JobTitle.Code = &row.JobTitleCode.String
		}
		if row.JobTitleName.Valid {
			emp.JobTitle.Name = &row.JobTitleName.String
		}
		if row.JobTitleGrade.Valid {
			emp.JobTitle.Grade = &row.JobTitleGrade.String
		}
	}

	if row.ManagerID.Valid {
		emp.Manager = &ManagerSummary{
			ID: row.ManagerID,
		}
		if row.ManagerEmployeeNo.Valid {
			emp.Manager.EmployeeNo = &row.ManagerEmployeeNo.String
		}
		if row.ManagerFirstName.Valid {
			emp.Manager.FirstName = &row.ManagerFirstName.String
		}
		if row.ManagerLastName.Valid {
			emp.Manager.LastName = &row.ManagerLastName.String
		}
		if row.ManagerDisplayName.Valid {
			emp.Manager.DisplayName = &row.ManagerDisplayName.String
		}
	}

	return emp
}

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

func (s *EmployeeService) GetEmployeeWithDetails(ctx context.Context, tenantID, id pgtype.UUID) (EmployeeWithDetails, error) {
	row, err := s.queries.GetEmployeeWithDetails(ctx, domain.GetEmployeeWithDetailsParams{
		TenantID: tenantID,
		ID:       id,
	})
	if err != nil {
		return EmployeeWithDetails{}, err
	}
	return mapRowToEmployeeWithDetails(row), nil
}

func (s *EmployeeService) ListEmployeesWithDetails(ctx context.Context, tenantID pgtype.UUID, params query.PaginationParams) ([]EmployeeWithDetails, int64, error) {
	rows, err := s.queries.ListEmployeesWithDetails(ctx, domain.ListEmployeesWithDetailsParams{
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

	emps := make([]EmployeeWithDetails, len(rows))
	for i, row := range rows {
		emps[i] = mapListRowToEmployeeWithDetails(row)
	}

	return emps, total, nil
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
