package app

import (
	"context"
	"encoding/json"
	"notification/internal/notification/domain"
	"time"
	"uuid"

	domainevent "github.com/distributed-programming-2026/lib/event"
)

type RepositoryProvider interface {
	Notification(context.Context) domain.Repository
}

type UnitOfWork interface {
	Execute(context.Context, func(RepositoryProvider) error) error
}

type DispatcherFactory interface {
	NewDispatcher(context.Context) domainevent.Dispatcher
}

type RealtimePublisher interface {
	Publish(context.Context, domain.Notification, []byte) error
}

type MailerSender interface {
	Send(context.Context, domain.Notification) error
}

type Service struct {
	unit     UnitOfWork
	events   DispatcherFactory
	realtime RealtimePublisher
	mailer   MailerSender
}

func NewService(u UnitOfWork, e DispatcherFactory, realtime RealtimePublisher, mailer MailerSender) *Service {
	return &Service{unit: u, events: e, realtime: realtime, mailer: mailer}
}

func (s *Service) run(ctx context.Context, fn func(*domain.Service) error) error {
	return s.unit.Execute(ctx, func(p RepositoryProvider) error {
		return fn(domain.NewService(p.Notification(ctx), s.events.NewDispatcher(ctx)))
	})
}

func (s *Service) Create(ctx context.Context, v domain.Notification) (out domain.Notification, created bool, err error) {
	err = s.run(ctx, func(d *domain.Service) error {
		out, created, err = d.Create(v, time.Now().UTC())
		return err
	})
	if err != nil || !created {
		return
	}
	delivery := out
	go func() {
		deliveryCtx := context.WithoutCancel(ctx)
		if s.realtime != nil && delivery.UserID != uuid.Nil() {
			payload, _ := json.Marshal(delivery)
			_ = s.realtime.Publish(deliveryCtx, delivery, payload)
		}
		if s.mailer != nil && delivery.EmailRequested && delivery.RecipientEmail != "" {
			_ = s.mailer.Send(deliveryCtx, delivery)
		}
	}()
	return
}

func (s *Service) List(ctx context.Context, u uuid.UUID) (out []domain.Notification, err error) {
	err = s.run(ctx, func(d *domain.Service) error {
		out, err = d.List(u)
		return err
	})
	return
}

func (s *Service) MarkRead(ctx context.Context, u, id uuid.UUID) error {
	return s.run(ctx, func(d *domain.Service) error {
		return d.MarkRead(u, id)
	})
}

func (s *Service) Preferences(ctx context.Context, u uuid.UUID) (out domain.Preferences, err error) {
	err = s.run(ctx, func(d *domain.Service) error {
		out, err = d.Preferences(u)
		return err
	})
	return
}

func (s *Service) PutPreferences(ctx context.Context, v domain.Preferences) (out domain.Preferences, err error) {
	err = s.run(ctx, func(d *domain.Service) error {
		out, err = d.PutPreferences(v)
		return err
	})
	return
}
