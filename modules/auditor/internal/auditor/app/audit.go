package app

import (
	"context"
	"time"
	"uuid"

	"github.com/distributed-programming-2026/go-sdk/pkg/uow"

	"auditor/internal/auditor/domain"
	"auditor/internal/auditor/infra/mysql"
)

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
	unit uow.UnitOfWorkWithRepositoryProvider[*mysql.RepositoryProvider]
}

func NewAuditService(unit uow.UnitOfWorkWithRepositoryProvider[*mysql.RepositoryProvider]) *AuditService {
	return &AuditService{
		unit: unit,
	}
}

func (s *AuditService) Record(ctx context.Context, event AuditEvent) (id uuid.UUID, err error) {
	err = s.unit.ExecuteWithRepositoryProvider(ctx, func(provider *mysql.RepositoryProvider) error {
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
		})
		return err
	})
	return id, err
}

func (s *AuditService) Get(ctx context.Context, id uuid.UUID) (record domain.AuditRecord, err error) {
	err = s.unit.ExecuteWithRepositoryProvider(ctx, func(provider *mysql.RepositoryProvider) error {
		record, err = domain.NewAuditService(provider.Audit(ctx)).Find(id)
		return err
	})
	return record, err
}
