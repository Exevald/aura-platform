package domain

import (
	"errors"
	"time"
	"uuid"
)

type AuditService struct {
	repo AuditRepository
}

func NewAuditService(repo AuditRepository) *AuditService {
	return &AuditService{
		repo: repo,
	}
}

func (s *AuditService) Record(record AuditRecord) (uuid.UUID, error) {
	if record.EventID == "" || record.EventType == "" || record.Producer == "" {
		return uuid.Nil(), ErrInvalidAuditRecord
	}
	existing, err := s.repo.FindByEventID(record.EventID)
	if err == nil {
		return existing.ID, nil
	}
	if !errors.Is(err, ErrAuditRecordNotFound) {
		return uuid.Nil(), err
	}
	record.ID, err = s.repo.NextID()
	if err != nil {
		return uuid.Nil(), err
	}
	record.ReceivedAt = time.Now().UTC()
	if record.OccurredAt.IsZero() {
		record.OccurredAt = record.ReceivedAt
	}
	if err = s.repo.Store(record); err != nil {
		return uuid.Nil(), err
	}
	return record.ID, nil
}

func (s *AuditService) Find(id uuid.UUID) (AuditRecord, error) {
	return s.repo.Find(id)
}
