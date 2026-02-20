package audit

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/INOVA/DML/internal/db"
	"github.com/INOVA/DML/internal/domain"
	"github.com/INOVA/DML/internal/http/query"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// AuditEvent represents the internal payload sent to the logging channel
type AuditEvent struct {
	TenantID   pgtype.UUID
	ActorID    pgtype.UUID
	Action     string
	EntityType string
	EntityID   uuid.UUID
	Changes    interface{} // Will be serialized to JSONB
}

type AuditService struct {
	queries *domain.Queries
	events  chan AuditEvent
}

// NewAuditService creates a new audit service and starts the background worker pool
func NewAuditService(database *db.DB) *AuditService {
	svc := &AuditService{
		queries: domain.New(database.Pool),
		events:  make(chan AuditEvent, 1000), // Buffered channel to prevent blocking the HTTP handlers
	}
	go svc.worker() // Start background processing
	return svc
}

// Log pushes an event to the background channel instantly mapping the HTTP thread execution speed natively.
func (s *AuditService) Log(tenantID, actorID pgtype.UUID, action, entityType string, entityID uuid.UUID, changes interface{}) {
	s.events <- AuditEvent{
		TenantID:   tenantID,
		ActorID:    actorID,
		Action:     action,
		EntityType: entityType,
		EntityID:   entityID,
		Changes:    changes,
	}
}

// worker processes the channel stream securely committing records to Postgres natively decoupled from requests
func (s *AuditService) worker() {
	ctx := context.Background()
	for event := range s.events {
		var pgChanges []byte
		if event.Changes != nil {
			var err error
			pgChanges, err = json.Marshal(event.Changes)
			if err != nil {
				log.Printf("audit worker failed mapping json changes: %v", err)
				continue
			}
		}

		var entityIDBytes pgtype.UUID
		entityIDBytes.Bytes = event.EntityID
		entityIDBytes.Valid = true

		var eventIDBytes pgtype.UUID
		eventIDBytes.Bytes = uuid.New()
		eventIDBytes.Valid = true

		_, err := s.queries.InsertAuditLog(ctx, domain.InsertAuditLogParams{
			ID:         eventIDBytes,
			TenantID:   event.TenantID,
			ActorID:    event.ActorID,
			Action:     event.Action,
			EntityType: event.EntityType,
			EntityID:   entityIDBytes,
			Changes:    pgChanges,
		})

		if err != nil {
			log.Printf("audit worker failed persisting record natively: %v", err)
		}
	}
}

func (s *AuditService) ListLogs(ctx context.Context, tenantID pgtype.UUID, entityType string, action string, params query.PaginationParams) ([]domain.AuditLog, int64, error) {
	logs, err := s.queries.ListAuditLogs(ctx, domain.ListAuditLogsParams{
		TenantID:   tenantID,
		EntityType: entityType,
		Action:     action,
		Limit:      params.Limit(),
		Offset:     params.Offset(),
	})
	if err != nil {
		return nil, 0, fmt.Errorf("listing audit logs: %w", err)
	}

	total, err := s.queries.CountAuditLogs(ctx, domain.CountAuditLogsParams{
		TenantID:   tenantID,
		EntityType: entityType,
		Action:     action,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("counting audit logs: %w", err)
	}

	return logs, total, nil
}
