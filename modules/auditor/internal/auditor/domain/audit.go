package domain

import (
	"errors"
	"time"
	"uuid"
)

var (
	ErrAuditRecordNotFound = errors.New("audit record not found")
	ErrAuditRecordExists   = errors.New("audit record already exists")
	ErrInvalidAuditRecord  = errors.New("invalid audit record")
)

type AuditRecord struct {
	ID            uuid.UUID
	EventID       string
	EventType     string
	Producer      string
	CorrelationID string
	ActorID       string
	ResourceType  string
	ResourceID    string
	Payload       []byte
	OccurredAt    time.Time
	ReceivedAt    time.Time
}

type AuditRepository interface {
	NextID() (uuid.UUID, error)
	Store(AuditRecord) error
	Find(uuid.UUID) (AuditRecord, error)
	FindByEventID(string) (AuditRecord, error)
}
