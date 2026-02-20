package iam

import (
	"context"

	"github.com/INOVA/DML/internal/db"
	"github.com/INOVA/DML/internal/domain"
	"github.com/jackc/pgx/v5/pgtype"
)

type UserRoleService struct {
	queries *domain.Queries
}

func NewUserRoleService(database *db.DB) *UserRoleService {
	return &UserRoleService{
		queries: domain.New(database.Pool),
	}
}

func (s *UserRoleService) AssignUserRole(ctx context.Context, tenantID, userID, roleID, grantedByUserID pgtype.UUID) error {
	return s.queries.AssignUserRole(ctx, domain.AssignUserRoleParams{
		TenantID:        tenantID,
		UserID:          userID,
		RoleID:          roleID,
		GrantedByUserID: grantedByUserID,
	})
}

func (s *UserRoleService) RevokeUserRole(ctx context.Context, tenantID, userID, roleID pgtype.UUID) error {
	return s.queries.RevokeUserRole(ctx, domain.RevokeUserRoleParams{
		TenantID: tenantID,
		UserID:   userID,
		RoleID:   roleID,
	})
}
