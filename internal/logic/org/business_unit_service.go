package org

import (
	"context"

	"github.com/INOVA/DML/internal/db"
	"github.com/INOVA/DML/internal/domain"
	"github.com/INOVA/DML/internal/http/query"
	"github.com/jackc/pgx/v5/pgtype"
)

type BusinessUnitService struct {
	queries *domain.Queries
}

func NewBusinessUnitService(database *db.DB) *BusinessUnitService {
	return &BusinessUnitService{
		queries: domain.New(database.Pool),
	}
}

func (s *BusinessUnitService) CreateBusinessUnit(ctx context.Context, id, tenantID pgtype.UUID, code, name string) (domain.BusinessUnit, error) {
	// The code column in the DB allows nulls conceptually, but typically we require it or let it default.
	// We'll map the string to pgtype.Text based on how SQLC generated it.

	// Assuming sqlc generated code as pgtype.Text since it lacked `not null` in the schema
	var pgCode pgtype.Text
	if code != "" {
		pgCode.String = code
		pgCode.Valid = true
	}

	return s.queries.CreateBusinessUnit(ctx, domain.CreateBusinessUnitParams{
		ID:       id,
		TenantID: tenantID,
		Code:     pgCode,
		Name:     name,
	})
}

func (s *BusinessUnitService) ListBusinessUnits(ctx context.Context, tenantID pgtype.UUID, params query.PaginationParams) ([]domain.BusinessUnit, int64, error) {
	bus, err := s.queries.ListBusinessUnits(ctx, domain.ListBusinessUnitsParams{
		TenantID: tenantID,
		Search:   params.Search,
		Limit:    params.Limit(),
		Offset:   params.Offset(),
	})
	if err != nil {
		return nil, 0, err
	}

	total, err := s.queries.CountBusinessUnits(ctx, domain.CountBusinessUnitsParams{
		TenantID: tenantID,
		Search:   params.Search,
	})
	if err != nil {
		return nil, 0, err
	}

	return bus, total, nil
}

func (s *BusinessUnitService) GetBusinessUnit(ctx context.Context, tenantID, id pgtype.UUID) (domain.BusinessUnit, error) {
	return s.queries.GetBusinessUnit(ctx, domain.GetBusinessUnitParams{
		TenantID: tenantID,
		ID:       id,
	})
}
