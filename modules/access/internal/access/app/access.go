package app

import (
	"context"
	"uuid"

	"github.com/distributed-programming-2026/go-sdk/pkg/uow"
	"github.com/distributed-programming-2026/lib/event"

	"access/internal/access/domain"
	"access/internal/access/infra/mysql"
)

type DispatcherFactory interface {
	NewDispatcher(ctx context.Context) event.Dispatcher
}

type ModuleAccessSpec struct {
	UserID   uuid.UUID
	ModuleID uuid.UUID
	Effect   domain.Effect
}

type ModuleAccessService interface {
	Create(ctx context.Context, spec ModuleAccessSpec) (id uuid.UUID, err error)
	Update(ctx context.Context, id uuid.UUID, spec ModuleAccessSpec) error
	Delete(ctx context.Context, id uuid.UUID) error
	CheckModuleAccess(ctx context.Context, userID, moduleID uuid.UUID) (bool, error)
	GrantModuleAccess(ctx context.Context, userID, moduleID uuid.UUID) error
	RevokeModuleAccess(ctx context.Context, userID, moduleID uuid.UUID) error
}

type moduleAccessService struct {
	unit              uow.UnitOfWorkWithRepositoryProvider[*mysql.RepositoryProvider]
	dispatcherFactory DispatcherFactory
}

func (s *moduleAccessService) GrantModuleAccess(ctx context.Context, userID, moduleID uuid.UUID) error {
	allowed, err := s.CheckModuleAccess(ctx, userID, moduleID)
	if err != nil {
		return err
	}
	if allowed {
		return nil
	}
	_, err = s.Create(ctx, ModuleAccessSpec{UserID: userID, ModuleID: moduleID, Effect: domain.EffectAllow})
	return err
}

func (s *moduleAccessService) RevokeModuleAccess(ctx context.Context, userID, moduleID uuid.UUID) error {
	return s.unit.ExecuteWithRepositoryProvider(ctx, func(provider *mysql.RepositoryProvider) error {
		items, err := provider.ModuleAccess(ctx).FindByUserAndModule(userID, moduleID)
		if err != nil {
			return err
		}
		for _, item := range items {
			if err = s.domainService(ctx, provider).Delete(item.ID); err != nil {
				return err
			}
		}
		return nil
	})
}

func NewModuleAccessService(
	unit uow.UnitOfWorkWithRepositoryProvider[*mysql.RepositoryProvider],
	dispatcherFactory DispatcherFactory,
) ModuleAccessService {
	return &moduleAccessService{
		unit:              unit,
		dispatcherFactory: dispatcherFactory,
	}
}

func (s *moduleAccessService) Create(ctx context.Context, spec ModuleAccessSpec) (id uuid.UUID, err error) {
	err = s.unit.ExecuteWithRepositoryProvider(ctx, func(provider *mysql.RepositoryProvider) error {
		id, err = s.domainService(ctx, provider).Create(spec.UserID, spec.ModuleID, spec.Effect)
		return err
	})
	return id, err
}

func (s *moduleAccessService) Update(ctx context.Context, id uuid.UUID, spec ModuleAccessSpec) error {
	return s.unit.ExecuteWithRepositoryProvider(ctx, func(provider *mysql.RepositoryProvider) error {
		return s.domainService(ctx, provider).Update(id, spec.UserID, spec.ModuleID, spec.Effect)
	})
}

func (s *moduleAccessService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.unit.ExecuteWithRepositoryProvider(ctx, func(provider *mysql.RepositoryProvider) error {
		return s.domainService(ctx, provider).Delete(id)
	})
}

func (s *moduleAccessService) CheckModuleAccess(
	ctx context.Context,
	userID,
	moduleID uuid.UUID,
) (allowed bool, err error) {
	err = s.unit.ExecuteWithRepositoryProvider(ctx, func(provider *mysql.RepositoryProvider) error {
		allowed, err = s.domainService(ctx, provider).CheckModuleAccess(userID, moduleID)
		return err
	})
	return allowed, err
}

func (s *moduleAccessService) domainService(
	ctx context.Context,
	provider *mysql.RepositoryProvider,
) domain.ModuleAccessService {
	return domain.NewModuleAccessService(
		provider.ModuleAccess(ctx),
		s.dispatcherFactory.NewDispatcher(ctx),
	)
}
