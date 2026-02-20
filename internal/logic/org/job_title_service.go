package org

import (
	"context"

	"github.com/INOVA/DML/internal/db"
	"github.com/INOVA/DML/internal/domain"
	"github.com/INOVA/DML/internal/http/query"
	"github.com/jackc/pgx/v5/pgtype"
)

type JobTitleService struct {
	queries *domain.Queries
}

func NewJobTitleService(database *db.DB) *JobTitleService {
	return &JobTitleService{
		queries: domain.New(database.Pool),
	}
}

func (s *JobTitleService) CreateJobTitle(ctx context.Context, id, tenantID pgtype.UUID, code, name, grade string) (domain.JobTitle, error) {
	var pgCode pgtype.Text
	if code != "" {
		pgCode.String = code
		pgCode.Valid = true
	}

	var pgGrade pgtype.Text
	if grade != "" {
		pgGrade.String = grade
		pgGrade.Valid = true
	}

	return s.queries.CreateJobTitle(ctx, domain.CreateJobTitleParams{
		ID:       id,
		TenantID: tenantID,
		Code:     pgCode,
		Name:     name,
		Grade:    pgGrade,
	})
}

func (s *JobTitleService) ListJobTitles(ctx context.Context, tenantID pgtype.UUID, params query.PaginationParams) ([]domain.JobTitle, int64, error) {
	titles, err := s.queries.ListJobTitles(ctx, domain.ListJobTitlesParams{
		TenantID: tenantID,
		Search:   params.Search,
		Limit:    params.Limit(),
		Offset:   params.Offset(),
	})
	if err != nil {
		return nil, 0, err
	}

	total, err := s.queries.CountJobTitles(ctx, domain.CountJobTitlesParams{
		TenantID: tenantID,
		Search:   params.Search,
	})
	if err != nil {
		return nil, 0, err
	}

	return titles, total, nil
}

func (s *JobTitleService) GetJobTitle(ctx context.Context, tenantID, id pgtype.UUID) (domain.JobTitle, error) {
	return s.queries.GetJobTitle(ctx, domain.GetJobTitleParams{
		TenantID: tenantID,
		ID:       id,
	})
}
