package repository

import (
	"context"
	"database/sql"
	stderrors "errors"
	"time"
	"uuid"

	"github.com/distributed-programming-2026/go-sdk/pkg/mysql"
	"github.com/pkg/errors"

	"auditor/internal/auditor/domain"
)

type auditRepository struct {
	ctx    context.Context
	client mysql.ClientContext
}

type auditRow struct {
	ID            binaryUUID `db:"id"`
	EventID       string     `db:"event_id"`
	EventType     string     `db:"event_type"`
	Producer      string     `db:"producer"`
	CorrelationID string     `db:"correlation_id"`
	ActorID       string     `db:"actor_id"`
	ResourceType  string     `db:"resource_type"`
	ResourceID    string     `db:"resource_id"`
	Payload       []byte     `db:"payload"`
	OccurredAt    time.Time  `db:"occurred_at"`
	ReceivedAt    time.Time  `db:"received_at"`
}

func NewAuditRepository(
	ctx context.Context,
	client mysql.ClientContext,
) domain.AuditRepository {
	return &auditRepository{
		ctx:    ctx,
		client: client,
	}
}

func (r *auditRepository) NextID() (uuid.UUID, error) {
	return uuid.New(), nil
}

func (r *auditRepository) Store(record domain.AuditRecord) error {
	const query = `
		INSERT INTO audit_record (
		    id,
		    event_id,
		    event_type,
		    producer,
		    correlation_id,
		    actor_id,
		    resource_type,
		    resource_id,
		    payload,
		    occurred_at,
		    received_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := r.client.ExecContext(
		r.ctx,
		"store-audit-record",
		query,
		newBinaryUUID(record.ID),
		record.EventID,
		record.EventType,
		record.Producer,
		record.CorrelationID,
		record.ActorID,
		record.ResourceType,
		record.ResourceID,
		record.Payload,
		record.OccurredAt,
		record.ReceivedAt,
	)
	return errors.WithStack(err)
}

func (r *auditRepository) Find(id uuid.UUID) (domain.AuditRecord, error) {
	const query = `
		SELECT
		    id,
		    event_id,
		    event_type,
		    producer,
		    correlation_id,
		    actor_id,
		    resource_type,
		    resource_id,
		    payload,
		    occurred_at,
		    received_at
		FROM audit_record
		WHERE id = ?
	`
	return r.find("find-audit-record", query, newBinaryUUID(id))
}

func (r *auditRepository) FindByEventID(id string) (domain.AuditRecord, error) {
	const query = `
		SELECT
		    id,
		    event_id,
		    event_type,
		    producer,
		    correlation_id,
		    actor_id,
		    resource_type,
		    resource_id,
		    payload,
		    occurred_at,
		    received_at
		FROM audit_record
		WHERE event_id = ?
	`
	return r.find("find-audit-record-by-event", query, id)
}

func (r *auditRepository) find(name, query string, arg any) (domain.AuditRecord, error) {
	var row auditRow
	err := r.client.GetContext(r.ctx, name, &row, query, arg)
	if stderrors.Is(err, sql.ErrNoRows) {
		return domain.AuditRecord{}, errors.WithStack(domain.ErrAuditRecordNotFound)
	}
	if err != nil {
		return domain.AuditRecord{}, errors.WithStack(err)
	}
	return row.toDomainModel(), nil
}

func (r auditRow) toDomainModel() domain.AuditRecord {
	return domain.AuditRecord{
		ID:            r.ID.UUID(),
		EventID:       r.EventID,
		EventType:     r.EventType,
		Producer:      r.Producer,
		CorrelationID: r.CorrelationID,
		ActorID:       r.ActorID,
		ResourceType:  r.ResourceType,
		ResourceID:    r.ResourceID,
		Payload:       r.Payload,
		OccurredAt:    r.OccurredAt,
		ReceivedAt:    r.ReceivedAt,
	}
}
