package iam

import (
	"context"

	"github.com/INOVA/DML/internal/db"
	"github.com/INOVA/DML/internal/domain"
	"github.com/INOVA/DML/internal/http/query"
	"github.com/INOVA/DML/internal/logic/audit"
	"github.com/jackc/pgx/v5/pgtype"
)

type UserService struct {
	queries  *domain.Queries
	auditSvc *audit.AuditService
}

func NewUserService(database *db.DB, auditSvc *audit.AuditService) *UserService {
	return &UserService{
		queries:  domain.New(database.Pool),
		auditSvc: auditSvc,
	}
}

func (s *UserService) ListUsers(ctx context.Context, tenantID pgtype.UUID, params query.PaginationParams) ([]domain.User, int64, error) {
	users, err := s.queries.ListUsers(ctx, domain.ListUsersParams{
		TenantID: tenantID,
		Search:   params.Search,
		Limit:    params.Limit(),
		Offset:   params.Offset(),
	})
	if err != nil {
		return nil, 0, err
	}

	total, err := s.queries.CountUsers(ctx, domain.CountUsersParams{
		TenantID: tenantID,
		Search:   params.Search,
	})
	if err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

func (s *UserService) GetUserByID(ctx context.Context, tenantID, id pgtype.UUID) (domain.User, error) {
	return s.queries.GetUser(ctx, domain.GetUserParams{
		TenantID: tenantID,
		ID:       id,
	})
}

func (s *UserService) CreateUser(ctx context.Context, id, tenantID, actorID, employeeID pgtype.UUID, email, passwordHash string, display *string) (domain.User, error) {
	var pgDisplay pgtype.Text
	if display != nil && *display != "" {
		pgDisplay.String = *display
		pgDisplay.Valid = true
	}

	var pgPass pgtype.Text
	if passwordHash != "" {
		pgPass.String = passwordHash
		pgPass.Valid = true
	}

	user, err := s.queries.CreateUser(ctx, domain.CreateUserParams{
		ID:           id,
		TenantID:     tenantID,
		EmployeeID:   employeeID,
		Email:        email,
		DisplayName:  pgDisplay,
		PasswordHash: pgPass,
	})

	if err == nil && s.auditSvc != nil {
		s.auditSvc.Log(tenantID, actorID, "CREATE", "Users", id.Bytes, map[string]interface{}{
			"email":        email,
			"display_name": display,
		})
	}

	return user, err
}

func (s *UserService) GetUserByEmail(ctx context.Context, tenantID pgtype.UUID, email string) (domain.User, error) {
	return s.queries.GetUserByEmail(ctx, domain.GetUserByEmailParams{
		TenantID: tenantID,
		Email:    email,
	})
}
