package app

import (
	"context"
	"time"
	"uuid"

	"auditor/internal/auditor/domain"
)

type RepositoryProvider interface {
	Audit(context.Context) domain.AuditRepository
}

type UnitOfWork interface {
	Execute(context.Context, func(RepositoryProvider) error) error
}

type AuditEvent struct {
	EventID       string
	EventType     string
	Producer      string
	CorrelationID string
	ActorID       string
	ResourceType  string
	ResourceID    string
	Payload       []byte
	OccurredAt    time.Time
}

type AuditService struct {
	unit UnitOfWork
}

func NewAuditService(unit UnitOfWork) *AuditService {
	return &AuditService{
		unit: unit,
	}
}

func (s *AuditService) Record(ctx context.Context, event AuditEvent) (id uuid.UUID, err error) {
	err = s.unit.Execute(ctx, func(provider RepositoryProvider) error {
		id, err = domain.NewAuditService(provider.Audit(ctx)).Record(domain.AuditRecord{
			EventID:       event.EventID,
			EventType:     event.EventType,
			Producer:      event.Producer,
			CorrelationID: event.CorrelationID,
			ActorID:       event.ActorID,
			ResourceType:  event.ResourceType,
			ResourceID:    event.ResourceID,
			Payload:       event.Payload,
			OccurredAt:    event.OccurredAt,
		}, time.Now().UTC())
		return err
	})
	return id, err
}

func (s *AuditService) Get(ctx context.Context, id uuid.UUID) (record domain.AuditRecord, err error) {
	err = s.unit.Execute(ctx, func(provider RepositoryProvider) error {
		record, err = domain.NewAuditService(provider.Audit(ctx)).Find(id)
		return err
	})
	return record, err
}
