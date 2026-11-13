package domain

import (
	"time"
	"uuid"

	"github.com/distributed-programming-2026/lib/event"

	usergroupevents "usergroup/api/events"
)

type GroupMembershipService struct {
	repo       GroupMembershipRepository
	dispatcher event.Dispatcher
}

func NewGroupMembershipService(
	repo GroupMembershipRepository,
	dispatcher event.Dispatcher,
) *GroupMembershipService {
	return &GroupMembershipService{
		repo:       repo,
		dispatcher: dispatcher,
	}
}

func (s *GroupMembershipService) List(groupID uuid.UUID) ([]GroupMembership, error) {
	return s.repo.ListByGroup(groupID)
}

func (s *GroupMembershipService) Add(group Group, userID uuid.UUID) error {
	if userID == uuid.Nil() {
		return ErrInvalidGroup
	}
	changed, err := s.repo.Store(GroupMembership{
		GroupID:   group.ID,
		UserID:    userID,
		CreatedAt: time.Now().UTC(),
	})
	if err != nil || !changed {
		return err
	}
	return s.dispatchMembership("added", group, userID)
}

func (s *GroupMembershipService) Remove(group Group, userID uuid.UUID) error {
	changed, err := s.repo.Delete(group.ID, userID)
	if err != nil || !changed {
		return err
	}
	return s.dispatchMembership("removed", group, userID)
}

func (s *GroupMembershipService) dispatchMembership(kind string, group Group, userID uuid.UUID) error {
	return s.dispatcher.Dispatch(&usergroupevents.GroupMembershipChanged{
		Type:           kind,
		GroupId:        group.ID.String(),
		UserId:         userID.String(),
		OwnerUserId:    group.OwnerUserID.String(),
		OccurredAtUnix: time.Now().UTC().Unix(),
	})
}
