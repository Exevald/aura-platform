package app

import (
	"context"
	"time"
	"uuid"

	accessevents "access/api/events"
	"access/internal/access/domain/model"
	"access/internal/access/domain/service"
)

var (
	EditorModuleID    = uuid.MustParse("b8547c76-51fd-4f7c-9c68-8f30c84f61cb")
	ShortlinkModuleID = uuid.MustParse("d5b2367d-af92-4fa8-bc65-93b41f37a397")
	CalendarModuleID  = uuid.MustParse("8997f5f2-b127-4b9f-b5d1-80d0726afbe1")
	PlannerModuleID   = uuid.MustParse("bd46e533-1971-4697-8324-2bd29a158627")
)

type Module struct {
	ID   uuid.UUID
	Code string
	Name string
}

func Modules() []Module {
	return []Module{
		{ID: EditorModuleID, Code: "editor", Name: "Редактор документов"},
		{ID: ShortlinkModuleID, Code: "shortlink", Name: "Короткие ссылки"},
		{ID: CalendarModuleID, Code: "calendar", Name: "Календарь"},
		{ID: PlannerModuleID, Code: "planner", Name: "Планировщик"},
	}
}

type ModuleShareSpec struct {
	ShareID     uuid.UUID
	ModuleID    uuid.UUID
	GroupID     uuid.UUID
	OwnerUserID uuid.UUID
}

func (s *moduleAccessService) groupModuleAccessService(
	ctx context.Context,
	provider RepositoryProvider,
) *service.GroupModuleAccessService {
	return service.NewGroupModuleAccessService(
		provider.ModuleAccess(ctx),
		provider.ModuleShareOperation(ctx),
		s.dispatcherFactory.NewDispatcher(ctx),
	)
}

func (s *moduleAccessService) GrantGroupModuleAccess(
	ctx context.Context,
	spec ModuleShareSpec,
	members []uuid.UUID,
) error {
	return s.unit.Execute(ctx, func(p RepositoryProvider) error {
		return s.groupModuleAccessService(ctx, p).Grant(model.ModuleShare{
			ShareID:     spec.ShareID,
			ModuleID:    spec.ModuleID,
			GroupID:     spec.GroupID,
			OwnerUserID: spec.OwnerUserID,
		}, members)
	})
}

func (s *moduleAccessService) RejectGroupModuleAccess(ctx context.Context, spec ModuleShareSpec, reason string) error {
	return s.unit.Execute(ctx, func(_ RepositoryProvider) error {
		return s.dispatcherFactory.NewDispatcher(ctx).Dispatch(newRejected(spec, reason))
	})
}

func (s *moduleAccessService) CancelGroupModuleAccess(
	ctx context.Context,
	shareID,
	moduleID,
	groupID uuid.UUID,
	reason string,
) error {
	return s.unit.Execute(ctx, func(p RepositoryProvider) error {
		return s.groupModuleAccessService(ctx, p).Cancel(shareID, moduleID, groupID, reason)
	})
}

func (s *moduleAccessService) ChangeGroupMembership(ctx context.Context, groupID, userID uuid.UUID, added bool) error {
	return s.unit.Execute(ctx, func(p RepositoryProvider) error {
		return s.groupModuleAccessService(ctx, p).Membership(groupID, userID, added)
	})
}

func (s *moduleAccessService) DeleteGroupModuleAccess(ctx context.Context, groupID uuid.UUID) error {
	return s.unit.Execute(ctx, func(p RepositoryProvider) error {
		return s.groupModuleAccessService(ctx, p).DeleteGroup(groupID)
	})
}

func (s *moduleAccessService) ListGroupModuleAccess(
	ctx context.Context,
	ownerID uuid.UUID,
) (out []model.ModuleShare, err error) {
	err = s.unit.Execute(ctx, func(p RepositoryProvider) error {
		out, err = p.ModuleShareOperation(ctx).ListByOwner(ownerID)
		return err
	})
	return
}

func (s *moduleAccessService) SetGroupModules(
	ctx context.Context,
	ownerID,
	groupID uuid.UUID,
	moduleIDs,
	members []uuid.UUID,
) (out []model.ModuleShare, err error) {
	desired, desiredErr := desiredModules(moduleIDs)
	if desiredErr != nil {
		return nil, desiredErr
	}
	err = s.unit.Execute(ctx, func(p RepositoryProvider) error {
		operations, listErr := p.ModuleShareOperation(ctx).ListByGroup(groupID)
		if listErr != nil {
			return listErr
		}
		active := activeModules(operations, ownerID)
		if addsModulesToEmptyGroup(desired, active, members) {
			return model.ErrEmptyGroup
		}
		out, err = s.replaceGroupModules(ctx, p, ownerID, groupID, desired, active, members)
		return err
	})
	return out, err
}

func desiredModules(moduleIDs []uuid.UUID) (map[uuid.UUID]struct{}, error) {
	desired := make(map[uuid.UUID]struct{}, len(moduleIDs))
	for _, moduleID := range moduleIDs {
		if !knownModule(moduleID) {
			return nil, model.ErrUnknownModule
		}
		desired[moduleID] = struct{}{}
	}
	return desired, nil
}

func activeModules(operations []model.ModuleShare, ownerID uuid.UUID) map[uuid.UUID]model.ModuleShare {
	active := make(map[uuid.UUID]model.ModuleShare, len(operations))
	for _, operation := range operations {
		if operation.OwnerUserID == ownerID {
			active[operation.ModuleID] = operation
		}
	}
	return active
}

func addsModulesToEmptyGroup(
	desired map[uuid.UUID]struct{},
	active map[uuid.UUID]model.ModuleShare,
	members []uuid.UUID,
) bool {
	if len(members) != 0 {
		return false
	}
	for moduleID := range desired {
		if _, exists := active[moduleID]; !exists {
			return true
		}
	}
	return false
}

func (s *moduleAccessService) replaceGroupModules(
	ctx context.Context,
	p RepositoryProvider,
	ownerID,
	groupID uuid.UUID,
	desired map[uuid.UUID]struct{},
	active map[uuid.UUID]model.ModuleShare,
	members []uuid.UUID,
) ([]model.ModuleShare, error) {
	groupService := s.groupModuleAccessService(ctx, p)
	for moduleID, operation := range active {
		if _, keep := desired[moduleID]; !keep {
			err := groupService.Cancel(
				operation.ShareID,
				operation.ModuleID,
				operation.GroupID,
				"revoked_by_owner",
			)
			if err != nil {
				return nil, err
			}
		}
	}
	result := make([]model.ModuleShare, 0, len(desired))
	for moduleID := range desired {
		if operation, exists := active[moduleID]; exists {
			result = append(result, operation)
			continue
		}
		operation := model.ModuleShare{
			ShareID:     uuid.New(),
			ModuleID:    moduleID,
			GroupID:     groupID,
			OwnerUserID: ownerID,
		}
		if err := groupService.Grant(operation, members); err != nil {
			return nil, err
		}
		stored, err := p.ModuleShareOperation(ctx).Find(operation.ShareID)
		if err != nil {
			return nil, err
		}
		result = append(result, stored)
	}
	return result, nil
}

func knownModule(id uuid.UUID) bool {
	return id == EditorModuleID || id == ShortlinkModuleID || id == CalendarModuleID || id == PlannerModuleID
}

func newRejected(spec ModuleShareSpec, reason string) *accessevents.GroupModuleAccessRejected {
	return &accessevents.GroupModuleAccessRejected{
		ShareId:        spec.ShareID.String(),
		ModuleId:       spec.ModuleID.String(),
		GroupId:        spec.GroupID.String(),
		Reason:         reason,
		OccurredAtUnix: time.Now().UTC().Unix(),
	}
}
