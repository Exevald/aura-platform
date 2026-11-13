package domain

import (
	events "notification/api/events"
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

func (s *Service) Create(v Notification, now time.Time) (Notification, bool, error) {
	if (v.UserID == uuid.Nil() && (!v.EmailRequested || strings.TrimSpace(v.RecipientEmail) == "")) || strings.TrimSpace(v.IdempotencyKey) == "" || strings.TrimSpace(v.Title) == "" {
		return Notification{}, false, ErrInvalid
	}
	id, err := s.repo.NextID()
	if err != nil {
		return Notification{}, false, err
	}
	v.ID = id
	v.Kind = strings.TrimSpace(v.Kind)
	v.Title = strings.TrimSpace(v.Title)
	if preferences, err := s.repo.GetPreferences(v.UserID); err == nil {
		v.EmailRequested = v.EmailRequested && preferences.Email
	}
	v.CreatedAt = now.UTC()
	created, err := s.repo.Store(v)
	if err != nil || !created {
		return v, created, err
	}
	err = s.events.Dispatch(&events.NotificationCreated{NotificationId: v.ID.String(), UserId: v.UserID.String(), WorkspaceId: v.WorkspaceID.String(), Kind: v.Kind, Title: v.Title, Body: v.Body, EmailRequested: v.EmailRequested, OccurredAtUnix: v.CreatedAt.Unix(), RecipientEmail: v.RecipientEmail})
	return v, true, err
}

func (s *Service) List(user uuid.UUID) ([]Notification, error) {
	if user == uuid.Nil() {
		return nil, ErrInvalid
	}
	return s.repo.List(user)
}

func (s *Service) MarkRead(user, id uuid.UUID) error {
	return s.repo.MarkRead(user, id)
}

func (s *Service) Preferences(user uuid.UUID) (Preferences, error) {
	v, e := s.repo.GetPreferences(user)
	if e == ErrNotFound {
		return Preferences{UserID: user, InApp: true, Email: true}, nil
	}
	return v, e
}

func (s *Service) PutPreferences(v Preferences) (Preferences, error) {
	if v.UserID == uuid.Nil() {
		return Preferences{}, ErrInvalid
	}
	return v, s.repo.StorePreferences(v)
}
