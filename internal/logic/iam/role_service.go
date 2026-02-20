package iam

import (
	"context"

	"github.com/INOVA/DML/internal/db"
	"github.com/INOVA/DML/internal/domain"
	"github.com/INOVA/DML/internal/logic/audit"
	"github.com/jackc/pgx/v5/pgtype"
)

type RoleService struct {
	queries  *domain.Queries
	auditSvc *audit.AuditService
}

func NewRoleService(database *db.DB, auditSvc *audit.AuditService) *RoleService {
	return &RoleService{
		queries:  domain.New(database.Pool),
		auditSvc: auditSvc,
	}
}

func (s *RoleService) CreateRole(ctx context.Context, id, tenantID, actorID pgtype.UUID, code, name string, description *string) (domain.RbacRole, error) {
	var pgDesc pgtype.Text
	if description != nil && *description != "" {
		pgDesc.String = *description
		pgDesc.Valid = true
	}

	role, err := s.queries.CreateRole(ctx, domain.CreateRoleParams{
		ID:          id,
		TenantID:    tenantID,
		Code:        code,
		Name:        name,
		Description: pgDesc,
	})

	if err == nil && s.auditSvc != nil {
		s.auditSvc.Log(tenantID, actorID, "CREATE", "Roles", id.Bytes, map[string]interface{}{
			"code":        code,
			"name":        name,
			"description": description,
		})
	}

	return role, err
}

func (s *RoleService) ListRoles(ctx context.Context, tenantID pgtype.UUID) ([]domain.RbacRole, error) {
	return s.queries.ListRoles(ctx, tenantID)
}

func (s *RoleService) GetRole(ctx context.Context, tenantID, id pgtype.UUID) (domain.RbacRole, error) {
	return s.queries.GetRole(ctx, domain.GetRoleParams{
		TenantID: tenantID,
		ID:       id,
	})
}
