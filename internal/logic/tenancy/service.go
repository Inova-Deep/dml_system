package tenancy

import (
	"context"

	"github.com/INOVA/DML/internal/db"
	"github.com/INOVA/DML/internal/domain"
	"github.com/jackc/pgx/v5/pgtype"
)

type Service struct {
	queries *domain.Queries
}

func NewService(database *db.DB) *Service {
	// Initialize SQLC queries wrapper with our connection pool
	return &Service{
		queries: domain.New(database.Pool),
	}
}

// CreateTenant creates a new tenant
func (s *Service) CreateTenant(ctx context.Context, id pgtype.UUID, code, name string) (domain.Tenant, error) {
	return s.queries.CreateTenant(ctx, domain.CreateTenantParams{
		ID:   id,
		Code: code,
		Name: name,
	})
}

// ListTenants retrieves all tenants
func (s *Service) ListTenants(ctx context.Context) ([]domain.Tenant, error) {
	return s.queries.ListTenants(ctx)
}

// GetTenant retrieves a single tenant
func (s *Service) GetTenant(ctx context.Context, id pgtype.UUID) (domain.Tenant, error) {
	return s.queries.GetTenant(ctx, id)
}
