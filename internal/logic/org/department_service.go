package org

import (
	"context"

	"github.com/INOVA/DML/internal/db"
	"github.com/INOVA/DML/internal/domain"
	"github.com/INOVA/DML/internal/http/query"
	"github.com/jackc/pgx/v5/pgtype"
)

type DepartmentService struct {
	queries *domain.Queries
}

func NewDepartmentService(database *db.DB) *DepartmentService {
	return &DepartmentService{
		queries: domain.New(database.Pool),
	}
}

func (s *DepartmentService) CreateDepartment(ctx context.Context, id, tenantID pgtype.UUID, parentID *pgtype.UUID, code, name string) (domain.Department, error) {
	var pgCode pgtype.Text
	if code != "" {
		pgCode.String = code
		pgCode.Valid = true
	}

	var pgParentID pgtype.UUID
	if parentID != nil && parentID.Valid {
		pgParentID = *parentID
	}

	return s.queries.CreateDepartment(ctx, domain.CreateDepartmentParams{
		ID:                 id,
		TenantID:           tenantID,
		ParentDepartmentID: pgParentID,
		Code:               pgCode,
		Name:               name,
	})
}

func (s *DepartmentService) ListDepartments(ctx context.Context, tenantID pgtype.UUID, params query.PaginationParams) ([]domain.Department, int64, error) {
	deps, err := s.queries.ListDepartments(ctx, domain.ListDepartmentsParams{
		TenantID: tenantID,
		Search:   params.Search,
		Limit:    params.Limit(),
		Offset:   params.Offset(),
	})
	if err != nil {
		return nil, 0, err
	}

	total, err := s.queries.CountDepartments(ctx, domain.CountDepartmentsParams{
		TenantID: tenantID,
		Search:   params.Search,
	})
	if err != nil {
		return nil, 0, err
	}

	return deps, total, nil
}

func (s *DepartmentService) GetDepartment(ctx context.Context, tenantID, id pgtype.UUID) (domain.Department, error) {
	return s.queries.GetDepartment(ctx, domain.GetDepartmentParams{
		TenantID: tenantID,
		ID:       id,
	})
}
