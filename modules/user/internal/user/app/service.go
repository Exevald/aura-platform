package app

import (
	"context"
	"uuid"

	"github.com/distributed-programming-2026/go-sdk/pkg/uow"
	domainevent "github.com/distributed-programming-2026/lib/event"

	"user/internal/user/domain"
	"user/internal/user/infra/mysql"
)

type DispatcherFactory interface {
	NewDispatcher(ctx context.Context) domainevent.Dispatcher
}

type UserFieldsSpec struct {
	Login     string
	FirstName string
	LastName  string
	Email     string
}

type UserService interface {
	CreateUser(ctx context.Context, spec UserFieldsSpec) (uuid.UUID, error)
	UpdateUser(ctx context.Context, id uuid.UUID, spec UserFieldsSpec) error
	DeleteUser(ctx context.Context, id uuid.UUID) error
}

func NewUserService(
	uow uow.UnitOfWorkWithRepositoryProvider[*mysql.RepositoryProvider],
	dispatcherFactory DispatcherFactory,
) UserService {
	return &userService{
		uow:               uow,
		dispatcherFactory: dispatcherFactory,
	}
}

type userService struct {
	uow               uow.UnitOfWorkWithRepositoryProvider[*mysql.RepositoryProvider]
	dispatcherFactory DispatcherFactory
}

func (u *userService) CreateUser(ctx context.Context, spec UserFieldsSpec) (id uuid.UUID, err error) {
	err = u.uow.ExecuteWithRepositoryProvider(ctx, func(provider *mysql.RepositoryProvider) error {
		domainService := domain.NewUserService(
			provider.UserRepository(ctx),
			u.dispatcherFactory.NewDispatcher(ctx),
		)
		id, err = domainService.CreateUser(spec.Login, spec.FirstName, spec.LastName, spec.Email)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return uuid.Nil(), err
	}
	return id, nil
}

func (u *userService) UpdateUser(ctx context.Context, id uuid.UUID, spec UserFieldsSpec) error {
	err := u.uow.ExecuteWithRepositoryProvider(ctx, func(provider *mysql.RepositoryProvider) error {
		domainService := domain.NewUserService(
			provider.UserRepository(ctx),
			u.dispatcherFactory.NewDispatcher(ctx),
		)
		err := domainService.UpdateUser(id, spec.Login, spec.FirstName, spec.LastName, spec.Email)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (u *userService) DeleteUser(ctx context.Context, id uuid.UUID) error {
	err := u.uow.ExecuteWithRepositoryProvider(ctx, func(provider *mysql.RepositoryProvider) error {
		domainService := domain.NewUserService(
			provider.UserRepository(ctx),
			u.dispatcherFactory.NewDispatcher(ctx),
		)
		err := domainService.RemoveUser(id)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

type Service struct {
	unit              uow.UnitOfWorkWithRepositoryProvider[*mysql.RepositoryProvider]
	dispatcherFactory DispatcherFactory
}

func NewService(
	unit uow.UnitOfWorkWithRepositoryProvider[*mysql.RepositoryProvider],
	dispatcherFactory DispatcherFactory,
) *Service {
	return &Service{
		unit:              unit,
		dispatcherFactory: dispatcherFactory,
	}
}

func (s *Service) Echo(ctx context.Context, body string) (uuid.UUID, error) {
	var id uuid.UUID
	err := s.unit.ExecuteWithRepositoryProvider(ctx, func(provider *mysql.RepositoryProvider) error {
		svc := domain.NewEchoService(
			provider.Echo(ctx),
			s.dispatcherFactory.NewDispatcher(ctx),
		)

		var err error
		id, err = svc.Echo(body)
		return err
	})
	if err != nil {
		return uuid.Nil(), err
	}
	return id, nil
}
