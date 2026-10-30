package app

import (
	"context"
	"time"
	"uuid"

	"github.com/distributed-programming-2026/go-sdk/pkg/uow"
	domainevent "github.com/distributed-programming-2026/lib/event"

	"session/internal/session/domain"
	"session/internal/session/infra/mysql"
)

type DispatcherFactory interface {
	NewDispatcher(ctx context.Context) domainevent.Dispatcher
}

type SessionService interface {
	Create(ctx context.Context, userID uuid.UUID, persistent bool) (uuid.UUID, time.Time, error)
	Validate(ctx context.Context, id uuid.UUID) (domain.Session, error)
	Revoke(ctx context.Context, id uuid.UUID) error
	RevokeAll(ctx context.Context, userID uuid.UUID) error
}

func NewSessionService(
	unit uow.UnitOfWorkWithRepositoryProvider[*mysql.RepositoryProvider],
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

type sessionService struct {
	unit               uow.UnitOfWorkWithRepositoryProvider[*mysql.RepositoryProvider]
	dispatcherFactory  DispatcherFactory
	clock              domain.Clock
	defaultLifetime    time.Duration
	persistentLifetime time.Duration
}

func (s *sessionService) Create(
	ctx context.Context,
	userID uuid.UUID,
	persistent bool,
) (id uuid.UUID, expiresAt time.Time, err error) {
	err = s.unit.ExecuteWithRepositoryProvider(ctx, func(provider *mysql.RepositoryProvider) error {
		id, expiresAt, err = s.domainService(ctx, provider).Create(userID, persistent)
		return err
	})
	return id, expiresAt, err
}

func (s *sessionService) Validate(
	ctx context.Context,
	id uuid.UUID,
) (result domain.Session, err error) {
	err = s.unit.ExecuteWithRepositoryProvider(ctx, func(provider *mysql.RepositoryProvider) error {
		result, err = s.domainService(ctx, provider).Validate(id)
		return err
	})
	return result, err
}

func (s *sessionService) Revoke(ctx context.Context, id uuid.UUID) error {
	return s.unit.ExecuteWithRepositoryProvider(ctx, func(provider *mysql.RepositoryProvider) error {
		return s.domainService(ctx, provider).Revoke(id)
	})
}

func (s *sessionService) RevokeAll(ctx context.Context, userID uuid.UUID) error {
	return s.unit.ExecuteWithRepositoryProvider(ctx, func(provider *mysql.RepositoryProvider) error {
		return s.domainService(ctx, provider).RevokeAll(userID)
	})
}

func (s *sessionService) domainService(
	ctx context.Context,
	provider *mysql.RepositoryProvider,
) domain.SessionService {
	return domain.NewSessionService(
		provider.Session(ctx),
		s.dispatcherFactory.NewDispatcher(ctx),
		s.clock,
		s.defaultLifetime,
		s.persistentLifetime,
	)
}
