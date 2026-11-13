package app

import (
	"context"
	"mailer/internal/mailer/domain"
	"time"
	"uuid"

	domainevent "github.com/distributed-programming-2026/lib/event"
)

type RepositoryProvider interface {
	Mail(context.Context) domain.Repository
}

type UnitOfWork interface {
	Execute(context.Context, func(RepositoryProvider) error) error
}

type DispatcherFactory interface {
	NewDispatcher(context.Context) domainevent.Dispatcher
}

type Sender interface {
	Send(context.Context, string, string, string) error
}

type Service struct {
	unit   UnitOfWork
	events DispatcherFactory
	sender Sender
}

func NewService(u UnitOfWork, e DispatcherFactory, senders ...Sender) *Service {
	s := &Service{unit: u, events: e}
	if len(senders) > 0 {
		s.sender = senders[0]
	}
	return s
}

func (s *Service) domain(ctx context.Context, p RepositoryProvider) *domain.Service {
	return domain.NewService(p.Mail(ctx), s.events.NewDispatcher(ctx))
}

func (s *Service) Send(ctx context.Context, key, to, subject, body string) (out domain.Job, created bool, err error) {
	err = s.unit.Execute(ctx, func(p RepositoryProvider) error {
		out, created, err = s.domain(ctx, p).Create(key, to, subject, body, time.Now().UTC())
		return err
	})
	if err != nil || !created || s.sender == nil {
		return
	}
	sendErr := s.sender.Send(ctx, to, subject, body)
	err = s.unit.Execute(context.WithoutCancel(ctx), func(p RepositoryProvider) error {
		out, err = s.domain(context.WithoutCancel(ctx), p).Complete(out, sendErr, time.Now().UTC())
		return err
	})
	if err == nil && sendErr != nil {
		err = sendErr
	}
	return
}

func (s *Service) Get(ctx context.Context, id uuid.UUID) (out domain.Job, err error) {
	err = s.unit.Execute(ctx, func(p RepositoryProvider) error { out, err = s.domain(ctx, p).Get(id); return err })
	return
}
