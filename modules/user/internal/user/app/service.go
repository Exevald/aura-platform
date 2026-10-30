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
	DisplayName string
	FirstName   string
	LastName    string
	Email       string
}

type UserService interface {
	CreateUser(ctx context.Context, spec UserFieldsSpec) (uuid.UUID, error)
	GetUser(ctx context.Context, id uuid.UUID) (domain.User, error)
	FindUserByEmail(ctx context.Context, email string) (domain.User, error)
	UpdateUser(ctx context.Context, id uuid.UUID, spec UserFieldsSpec) error
	DeleteUser(ctx context.Context, id uuid.UUID) error
}

func NewUserService(
	unitOfWork uow.UnitOfWorkWithRepositoryProvider[*mysql.RepositoryProvider],
	dispatcherFactory DispatcherFactory,
) UserService {
	return &userService{
		unitOfWork:        unitOfWork,
		dispatcherFactory: dispatcherFactory,
	}
}

type userService struct {
	unitOfWork        uow.UnitOfWorkWithRepositoryProvider[*mysql.RepositoryProvider]
	dispatcherFactory DispatcherFactory
}

func (u *userService) FindUserByEmail(ctx context.Context, email string) (result domain.User, err error) {
	err = u.unitOfWork.ExecuteWithRepositoryProvider(ctx, func(provider *mysql.RepositoryProvider) error {
		result, err = domain.NewUserService(
			provider.UserRepository(ctx),
			u.dispatcherFactory.NewDispatcher(ctx),
		).FindUserByEmail(email)
		return err
	})
	return result, err
}

func (u *userService) CreateUser(ctx context.Context, spec UserFieldsSpec) (id uuid.UUID, err error) {
	err = u.unitOfWork.ExecuteWithRepositoryProvider(ctx, func(provider *mysql.RepositoryProvider) error {
		domainService := domain.NewUserService(
			provider.UserRepository(ctx),
			u.dispatcherFactory.NewDispatcher(ctx),
		)
		id, err = domainService.CreateUser(spec.DisplayName, spec.FirstName, spec.LastName, spec.Email)
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

func (u *userService) GetUser(ctx context.Context, id uuid.UUID) (result domain.User, err error) {
	err = u.unitOfWork.ExecuteWithRepositoryProvider(ctx, func(provider *mysql.RepositoryProvider) error {
		domainService := domain.NewUserService(
			provider.UserRepository(ctx),
			u.dispatcherFactory.NewDispatcher(ctx),
		)
		result, err = domainService.GetUser(id)
		return err
	})
	return result, err
}

func (u *userService) UpdateUser(ctx context.Context, id uuid.UUID, spec UserFieldsSpec) error {
	err := u.unitOfWork.ExecuteWithRepositoryProvider(ctx, func(provider *mysql.RepositoryProvider) error {
		domainService := domain.NewUserService(
			provider.UserRepository(ctx),
			u.dispatcherFactory.NewDispatcher(ctx),
		)
		err := domainService.UpdateUser(id, spec.DisplayName, spec.FirstName, spec.LastName, spec.Email)
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
	err := u.unitOfWork.ExecuteWithRepositoryProvider(ctx, func(provider *mysql.RepositoryProvider) error {
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
