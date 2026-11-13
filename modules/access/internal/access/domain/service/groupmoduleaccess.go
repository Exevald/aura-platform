package service

import (
	"access/internal/access/domain/model"
	"errors"
	"time"
	"uuid"

	"github.com/distributed-programming-2026/lib/event"

	accessevents "access/api/events"
)

type GroupModuleAccessService struct {
	accessRepository    model.ModuleAccessRepository
	operationRepository model.ModuleShareRepository
	dispatcher          event.Dispatcher
}

func NewGroupModuleAccessService(
	accessRepository model.ModuleAccessRepository,
	shareRepository model.ModuleShareRepository,
	dispatcher event.Dispatcher,
) *GroupModuleAccessService {
	return &GroupModuleAccessService{
		accessRepository:    accessRepository,
		operationRepository: shareRepository,
		dispatcher:          dispatcher,
	}
}

func (s *GroupModuleAccessService) Grant(op model.ModuleShare, members []uuid.UUID) error {
	existing, err := s.operationRepository.Find(op.ShareID)
	if err == nil {
		if existing.Status == model.ModuleShareCancelled {
			return s.dispatchRevoked(existing, "cancelled")
		}
		return s.dispatchGranted(existing, len(members))
	}
	if !errors.Is(err, model.ErrModuleShareOperationNotFound) {
		return err
	}
	if op.ShareID == uuid.Nil() || op.ModuleID == uuid.Nil() || op.GroupID == uuid.Nil() || op.OwnerUserID == uuid.Nil() || len(members) == 0 {
		return model.ErrInvalidGroupModuleAccess
	}
	now := time.Now().UTC()
	op.Status, op.CreatedAt, op.UpdatedAt = model.ModuleShareGranted, now, now
	if err = s.operationRepository.Store(op); err != nil {
		return err
	}
	accessService := NewModuleAccessService(s.accessRepository, s.dispatcher)
	for _, member := range members {
		_, err = accessService.CreateWithSource(
			member,
			op.ModuleID,
			model.EffectAllow,
			groupSource(op.ShareID),
		)
		if err != nil {
			return err
		}
	}

	return s.dispatchGranted(op, len(members))
}

func (s *GroupModuleAccessService) Cancel(shareID, moduleID, groupID uuid.UUID, reason string) error {
	op, err := s.operationRepository.Find(shareID)
	if errors.Is(err, model.ErrModuleShareOperationNotFound) {
		now := time.Now().UTC()
		op = model.ModuleShare{
			ShareID:   shareID,
			ModuleID:  moduleID,
			GroupID:   groupID,
			Status:    model.ModuleShareCancelled,
			CreatedAt: now,
			UpdatedAt: now,
		}
	} else if err != nil {
		return err
	}
	op.Status, op.UpdatedAt = model.ModuleShareCancelled, time.Now().UTC()
	if err = s.operationRepository.Store(op); err != nil {
		return err
	}
	if err = s.accessRepository.DeleteBySource(groupSource(shareID)); err != nil {
		return err
	}

	return s.dispatchRevoked(op, reason)
}

func (s *GroupModuleAccessService) Membership(groupID, userID uuid.UUID, added bool) error {
	operations, err := s.operationRepository.ListByGroup(groupID)
	if err != nil {
		return err
	}
	accessService := NewModuleAccessService(s.accessRepository, s.dispatcher)
	for _, op := range operations {
		if added {
			_, err = accessService.CreateWithSource(userID, op.ModuleID, model.EffectAllow, groupSource(op.ShareID))
		} else {
			err = s.accessRepository.DeleteBySourceAndUser(groupSource(op.ShareID), userID)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *GroupModuleAccessService) DeleteGroup(groupID uuid.UUID) error {
	operations, err := s.operationRepository.ListByGroup(groupID)
	if err != nil {
		return err
	}
	for _, op := range operations {
		err = s.Cancel(op.ShareID, op.ModuleID, op.GroupID, "group_deleted")
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *GroupModuleAccessService) dispatchGranted(op model.ModuleShare, count int) error {
	//nolint:gosec
	memberCount := int32(count)
	return s.dispatcher.Dispatch(&accessevents.GroupModuleAccessGranted{
		ShareId:        op.ShareID.String(),
		ModuleId:       op.ModuleID.String(),
		GroupId:        op.GroupID.String(),
		MemberCount:    memberCount,
		OccurredAtUnix: time.Now().UTC().Unix(),
	})
}

func (s *GroupModuleAccessService) dispatchRevoked(op model.ModuleShare, reason string) error {
	return s.dispatcher.Dispatch(&accessevents.GroupModuleAccessRevoked{
		ShareId:        op.ShareID.String(),
		ModuleId:       op.ModuleID.String(),
		GroupId:        op.GroupID.String(),
		Reason:         reason,
		OccurredAtUnix: time.Now().UTC().Unix(),
	})
}

func groupSource(shareID uuid.UUID) string {
	return "group:" + shareID.String()
}
