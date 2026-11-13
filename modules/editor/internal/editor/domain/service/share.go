package service

import (
	"editor/internal/editor/domain/model"
	"errors"
	"time"
	"uuid"

	"github.com/distributed-programming-2026/lib/event"

	editorevents "editor/api/events"
)

func NewModuleShareService(
	repo model.ModuleShareRepository,
	dispatcher event.Dispatcher,
) *ModuleShareService {
	return &ModuleShareService{
		repo:       repo,
		dispatcher: dispatcher,
	}
}

type ModuleShareService struct {
	repo       model.ModuleShareRepository
	dispatcher event.Dispatcher
}

func (s *ModuleShareService) Create(moduleID, groupID, ownerID uuid.UUID) (model.ModuleShare, error) {
	if moduleID == uuid.Nil() || groupID == uuid.Nil() || ownerID == uuid.Nil() {
		return model.ModuleShare{}, model.ErrInvalidShare
	}
	existing, err := s.repo.FindByOwnerGroup(ownerID, groupID)
	if err == nil && existing.Status != model.ShareFailed && existing.Status != model.ShareRevoked {
		return model.ModuleShare{}, model.ErrShareConflict
	} else if err != nil && !errors.Is(err, model.ErrShareNotFound) {
		return model.ModuleShare{}, err
	}
	id, err := s.repo.NextID()
	if err != nil {
		return model.ModuleShare{}, err
	}

	now := time.Now().UTC()
	share := model.ModuleShare{
		ID:          id,
		ModuleID:    moduleID,
		GroupID:     groupID,
		OwnerUserID: ownerID,
		Status:      model.SharePending,
		Attempt:     1,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err = s.repo.Store(share); err != nil {
		return model.ModuleShare{}, err
	}
	return share, s.dispatchRequested(share)
}

func (s *ModuleShareService) Get(id uuid.UUID) (model.ModuleShare, error) {
	return s.repo.Find(id)
}

func (s *ModuleShareService) List(ownerID uuid.UUID) ([]model.ModuleShare, error) {
	return s.repo.ListByOwner(ownerID)
}

func (s *ModuleShareService) Activate(id uuid.UUID) error {
	share, err := s.repo.Find(id)
	if err != nil {
		return err
	}
	if share.Status == model.ShareActive {
		return nil
	}
	if share.Status != model.SharePending {
		return s.dispatchCancelled(share, "late_grant")
	}

	share.Status, share.UpdatedAt = model.ShareActive, time.Now().UTC()
	if err = s.repo.Store(share); err != nil {
		return err
	}

	return s.dispatcher.Dispatch(&editorevents.ModuleShared{
		ShareId:        share.ID.String(),
		ModuleId:       share.ModuleID.String(),
		GroupId:        share.GroupID.String(),
		OccurredAtUnix: share.UpdatedAt.Unix(),
	})
}

func (s *ModuleShareService) Reject(id uuid.UUID, reason string) error {
	share, err := s.repo.Find(id)
	if err != nil {
		return err
	}
	if share.Status != model.SharePending {
		return nil
	}

	share.Status, share.FailureReason, share.UpdatedAt = model.ShareFailed, reason, time.Now().UTC()
	if err = s.repo.Store(share); err != nil {
		return err
	}

	return s.dispatchFailed(share)
}

func (s *ModuleShareService) Cancel(id uuid.UUID, reason string) (model.ModuleShare, error) {
	share, err := s.repo.Find(id)
	if err != nil {
		return model.ModuleShare{}, err
	}
	if share.Status == model.ShareFailed || share.Status == model.ShareRevoked {
		return share, nil
	}
	share.Status, share.FailureReason, share.UpdatedAt = model.ShareCompensating, reason, time.Now().UTC()
	if err = s.repo.Store(share); err != nil {
		return model.ModuleShare{}, err
	}
	return share, s.dispatchCancelled(share, reason)
}

func (s *ModuleShareService) Revoked(id uuid.UUID, reason string) error {
	share, err := s.repo.Find(id)
	if err != nil {
		return err
	}
	if share.Status == model.ShareFailed || share.Status == model.ShareRevoked {
		return nil
	}
	if share.FailureReason == "revoked_by_owner" || reason == "group_deleted" {
		share.Status = model.ShareRevoked
	} else {
		share.Status = model.ShareFailed
	}
	if share.FailureReason == "" {
		share.FailureReason = reason
	}
	share.UpdatedAt = time.Now().UTC()
	if err = s.repo.Store(share); err != nil {
		return err
	}
	if share.Status == model.ShareFailed {
		return s.dispatchFailed(share)
	}
	return nil
}

func (s *ModuleShareService) Recover(now time.Time, attemptTimeout time.Duration, maxAttempts int) error {
	values, err := s.repo.ListPendingBefore(now.Add(-attemptTimeout))
	if err != nil {
		return err
	}
	for _, share := range values {
		if share.Attempt >= maxAttempts {
			if _, err = s.Cancel(share.ID, "activation_timeout"); err != nil {
				return err
			}
			continue
		}
		share.Attempt++
		share.UpdatedAt = now
		if err = s.repo.Store(share); err != nil {
			return err
		}
		if err = s.dispatchRequested(share); err != nil {
			return err
		}
	}
	compensating, err := s.repo.ListCompensatingBefore(now.Add(-attemptTimeout))
	if err != nil {
		return err
	}
	for _, share := range compensating {
		share.UpdatedAt = now
		if err = s.repo.Store(share); err != nil {
			return err
		}
		if err = s.dispatchCancelled(share, share.FailureReason); err != nil {
			return err
		}
	}
	return nil
}

func (s *ModuleShareService) dispatchRequested(v model.ModuleShare) error {
	//nolint:gosec
	attempt := int32(v.Attempt)
	return s.dispatcher.Dispatch(&editorevents.ShareRequested{
		ShareId:        v.ID.String(),
		ModuleId:       v.ModuleID.String(),
		GroupId:        v.GroupID.String(),
		OwnerUserId:    v.OwnerUserID.String(),
		Attempt:        attempt,
		OccurredAtUnix: time.Now().UTC().Unix(),
	})
}

func (s *ModuleShareService) dispatchCancelled(v model.ModuleShare, reason string) error {
	return s.dispatcher.Dispatch(&editorevents.ShareCancelled{ShareId: v.ID.String(), ModuleId: v.ModuleID.String(), GroupId: v.GroupID.String(), Reason: reason, OccurredAtUnix: time.Now().UTC().Unix()})
}

func (s *ModuleShareService) dispatchFailed(v model.ModuleShare) error {
	return s.dispatcher.Dispatch(&editorevents.ModuleShareFailed{ShareId: v.ID.String(), ModuleId: v.ModuleID.String(), GroupId: v.GroupID.String(), Reason: v.FailureReason, OccurredAtUnix: time.Now().UTC().Unix()})
}
