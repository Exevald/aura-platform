package domain

import (
	"testing"
	"time"

	"uuid"
)

type auditRepoStub struct{ records []AuditRecord }

func TestRecordIsIdempotentByEventID(t *testing.T) {
	repo := &auditRepoStub{}
	service := NewAuditService(repo)
	record := AuditRecord{
		EventID:   "event-1",
		EventType: "UserRegistered",
		Producer:  "auth",
	}
	now := time.Date(2026, time.November, 13, 12, 0, 0, 0, time.UTC)
	first, err := service.Record(record, now)
	if err != nil {
		t.Fatal(err)
	}
	second, err := service.Record(record, now)
	if err != nil {
		t.Fatal(err)
	}
	if first != second || len(repo.records) != 1 {
		t.Fatal("duplicate event created more than one record")
	}
	if !repo.records[0].ReceivedAt.Equal(now) {
		t.Fatal("record time did not match supplied time")
	}
}

func (r *auditRepoStub) NextID() (uuid.UUID, error) {
	return uuid.New(), nil
}

func (r *auditRepoStub) Store(v AuditRecord) error {
	r.records = append(r.records, v)
	return nil
}

func (r *auditRepoStub) Find(id uuid.UUID) (AuditRecord, error) {
	for _, v := range r.records {
		if v.ID == id {
			return v, nil
		}
	}
	return AuditRecord{}, ErrAuditRecordNotFound
}

func (r *auditRepoStub) FindByEventID(id string) (AuditRecord, error) {
	for _, v := range r.records {
		if v.EventID == id {
			return v, nil
		}
	}
	return AuditRecord{}, ErrAuditRecordNotFound
}
