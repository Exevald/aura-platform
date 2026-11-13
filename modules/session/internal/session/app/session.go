package app

import (
	"context"
	"time"
	"uuid"

	domainevent "github.com/distributed-programming-2026/lib/event"

	"session/internal/session/domain"
)

type RepositoryProvider interface {
	Session(context.Context) domain.SessionRepository
}

type UnitOfWork interface {
	Execute(context.Context, func(RepositoryProvider) error) error
}

type DispatcherFactory interface {
	NewDispatcher(ctx context.Context) domainevent.Dispatcher
}

type SessionService interface {
	Create(ctx context.Context, userID uuid.UUID, persistent bool) (uuid.UUID, time.Time, error)
	Validate(ctx context.Context, id uuid.UUID) (domain.Session, error)
	Revoke(ctx context.Context, id uuid.UUID) error
	RevokeAll(ctx context.Context, userID uuid.UUID) error
}

type sessionService struct {
	unit               UnitOfWork
	dispatcherFactory  DispatcherFactory
	clock              domain.Clock
	defaultLifetime    time.Duration
	persistentLifetime time.Duration
}

func NewSessionService(
	unit UnitOfWork,
	dispatcherFactory DispatcherFactory,
	defaultLifetime time.Duration,
	persistentLifetime time.Duration,
) SessionService {
	return &sessionService{
		unit:               unit,
		dispatcherFactory:  dispatcherFactory,
		clock:              domain.SystemClock{},
		defaultLifetime:    defaultLifetime,
		persistentLifetime: persistentLifetime,
	}
}

func (s *sessionService) Create(
	ctx context.Context,
	userID uuid.UUID,
	persistent bool,
) (id uuid.UUID, expiresAt time.Time, err error) {
	err = s.unit.Execute(ctx, func(provider RepositoryProvider) error {
		id, expiresAt, err = s.domainService(ctx, provider).Create(userID, persistent)
		return err
	})
	return id, expiresAt, err
}

func (s *sessionService) Validate(
	ctx context.Context,
	id uuid.UUID,
) (result domain.Session, err error) {
	err = s.unit.Execute(ctx, func(provider RepositoryProvider) error {
		result, err = s.domainService(ctx, provider).Validate(id)
		return err
	})
	return result, err
}

func (s *sessionService) Revoke(ctx context.Context, id uuid.UUID) error {
	return s.unit.Execute(ctx, func(provider RepositoryProvider) error {
		return s.domainService(ctx, provider).Revoke(id)
	})
}

func (s *sessionService) RevokeAll(ctx context.Context, userID uuid.UUID) error {
	return s.unit.Execute(ctx, func(provider RepositoryProvider) error {
		return s.domainService(ctx, provider).RevokeAll(userID)
	})
}

func (s *sessionService) domainService(
	ctx context.Context,
	provider RepositoryProvider,
) domain.SessionService {
	return domain.NewSessionService(
		provider.Session(ctx),
		s.dispatcherFactory.NewDispatcher(ctx),
		s.clock,
		s.defaultLifetime,
		s.persistentLifetime,
	)
}
