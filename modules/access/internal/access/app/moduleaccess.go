package app

import (
	"access/internal/access/domain/model"
	"access/internal/access/domain/service"
	"context"
	"uuid"

	"github.com/distributed-programming-2026/lib/event"
)

type RepositoryProvider interface {
	ModuleAccess(context.Context) model.ModuleAccessRepository
	ModuleShareOperation(context.Context) model.ModuleShareRepository
}

type UnitOfWork interface {
	Execute(context.Context, func(RepositoryProvider) error) error
}

type DispatcherFactory interface {
	NewDispatcher(ctx context.Context) event.Dispatcher
}

type ModuleAccessSpec struct {
	UserID   uuid.UUID
	ModuleID uuid.UUID
	Effect   model.Effect
}

type ModuleAccessService interface {
	Create(ctx context.Context, spec ModuleAccessSpec) (id uuid.UUID, err error)
	Update(ctx context.Context, id uuid.UUID, spec ModuleAccessSpec) error
	Delete(ctx context.Context, id uuid.UUID) error
	CheckModuleAccess(ctx context.Context, userID, moduleID uuid.UUID) (bool, error)
	GrantModuleAccess(ctx context.Context, userID, moduleID uuid.UUID) error
	RevokeModuleAccess(ctx context.Context, userID, moduleID uuid.UUID) error
	GrantGroupModuleAccess(ctx context.Context, spec ModuleShareSpec, members []uuid.UUID) error
	RejectGroupModuleAccess(ctx context.Context, spec ModuleShareSpec, reason string) error
	CancelGroupModuleAccess(ctx context.Context, shareID, moduleID, groupID uuid.UUID, reason string) error
	ChangeGroupMembership(ctx context.Context, groupID, userID uuid.UUID, added bool) error
	DeleteGroupModuleAccess(ctx context.Context, groupID uuid.UUID) error
	ListGroupModuleAccess(ctx context.Context, ownerID uuid.UUID) ([]model.ModuleShare, error)
	SetGroupModules(ctx context.Context, ownerID, groupID uuid.UUID, moduleIDs, members []uuid.UUID) ([]model.ModuleShare, error)
}

type moduleAccessService struct {
	unit              UnitOfWork
	dispatcherFactory DispatcherFactory
}

func NewModuleAccessService(
	unit UnitOfWork,
	dispatcherFactory DispatcherFactory,
) ModuleAccessService {
	return &moduleAccessService{
		unit:              unit,
		dispatcherFactory: dispatcherFactory,
	}
}

func (s *moduleAccessService) GrantModuleAccess(ctx context.Context, userID, moduleID uuid.UUID) error {
	return s.unit.Execute(ctx, func(provider RepositoryProvider) error {
		_, err := s.domainService(ctx, provider).CreateWithSource(userID, moduleID, model.EffectAllow, directSource(moduleID))
		return err
	})
}

func (s *moduleAccessService) RevokeModuleAccess(ctx context.Context, userID, moduleID uuid.UUID) error {
	return s.unit.Execute(ctx, func(provider RepositoryProvider) error {
		return provider.ModuleAccess(ctx).DeleteBySourceAndUser(directSource(moduleID), userID)
	})
}

func (s *moduleAccessService) Create(ctx context.Context, spec ModuleAccessSpec) (id uuid.UUID, err error) {
	err = s.unit.Execute(ctx, func(provider RepositoryProvider) error {
		id, err = s.domainService(ctx, provider).Create(spec.UserID, spec.ModuleID, spec.Effect)
		return err
	})
	return id, err
}

func (s *moduleAccessService) Update(ctx context.Context, id uuid.UUID, spec ModuleAccessSpec) error {
	return s.unit.Execute(ctx, func(provider RepositoryProvider) error {
		return s.domainService(ctx, provider).Update(id, spec.UserID, spec.ModuleID, spec.Effect)
	})
}

func (s *moduleAccessService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.unit.Execute(ctx, func(provider RepositoryProvider) error {
		return s.domainService(ctx, provider).Delete(id)
	})
}

func (s *moduleAccessService) CheckModuleAccess(
	ctx context.Context,
	userID,
	moduleID uuid.UUID,
) (allowed bool, err error) {
	err = s.unit.Execute(ctx, func(provider RepositoryProvider) error {
		allowed, err = s.domainService(ctx, provider).CheckModuleAccess(userID, moduleID)
		return err
	})
	return allowed, err
}

func (s *moduleAccessService) domainService(
	ctx context.Context,
	provider RepositoryProvider,
) service.ModuleAccessService {
	return service.NewModuleAccessService(
		provider.ModuleAccess(ctx),
		s.dispatcherFactory.NewDispatcher(ctx),
	)
}

func directSource(moduleID uuid.UUID) string {
	return "direct:" + moduleID.String()
}
