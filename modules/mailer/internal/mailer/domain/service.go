package domain

import (
	events "mailer/api/events"
	"net/mail"
	"strings"
	"time"
	"uuid"

	"github.com/distributed-programming-2026/lib/event"
)

type Service struct {
	repo   Repository
	events event.Dispatcher
}

func NewService(r Repository, e event.Dispatcher) *Service { return &Service{r, e} }

func (s *Service) Create(key, to, subject, body string, now time.Time) (Job, bool, error) {
	if strings.TrimSpace(key) == "" || strings.TrimSpace(subject) == "" {
		return Job{}, false, ErrInvalid
	}
	if _, e := mail.ParseAddress(to); e != nil {
		return Job{}, false, ErrInvalid
	}
	id, err := s.repo.NextID()
	if err != nil {
		return Job{}, false, err
	}

	now = now.UTC()
	return s.repo.Create(Job{
		ID:             id,
		IdempotencyKey: key,
		Recipient:      to,
		Subject:        subject,
		Body:           body,
		Status:         "pending",
		CreatedAt:      now,
		UpdatedAt:      now,
	})
}

func (s *Service) Get(id uuid.UUID) (Job, error) {
	return s.repo.Find(id)
}

func (s *Service) Complete(v Job, sendErr error, now time.Time) (Job, error) {
	v.Attempts++
	v.UpdatedAt = now.UTC()
	if sendErr == nil {
		v.Status, v.LastError = "sent", ""
	} else {
		v.Status, v.LastError = "failed", sendErr.Error()
	}
	if e := s.repo.StoreResult(v); e != nil {
		return Job{}, e
	}
	if sendErr == nil {
		return v, s.events.Dispatch(&events.MailDelivered{MailId: v.ID.String(), Recipient: v.Recipient, OccurredAtUnix: v.UpdatedAt.Unix()})
	}
	return v, s.events.Dispatch(&events.MailDeliveryFailed{MailId: v.ID.String(), Recipient: v.Recipient, Error: v.LastError, OccurredAtUnix: v.UpdatedAt.Unix()})
}
