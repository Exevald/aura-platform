package domain

import (
	"strings"
	"time"
	"uuid"

	"github.com/distributed-programming-2026/lib/event"

	usergroupevents "usergroup/api/events"
)

type GroupService struct {
	repo       GroupRepository
	dispatcher event.Dispatcher
}

func NewGroupService(repo GroupRepository, dispatcher event.Dispatcher) *GroupService {
	return &GroupService{
		repo:       repo,
		dispatcher: dispatcher,
	}
}

func (s *GroupService) Create(owner, workspace uuid.UUID, name string) (Group, error) {
	if owner == uuid.Nil() || workspace == uuid.Nil() || !validName(name) {
		return Group{}, ErrInvalidGroup
	}
	id, err := s.repo.NextID()
	if err != nil {
		return Group{}, err
	}
	now := time.Now().UTC()
	group := Group{
		ID:          id,
		OwnerUserID: owner,
		WorkspaceID: workspace,
		Name:        strings.TrimSpace(name),
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	return group, s.repo.Store(group)
}

func (s *GroupService) GetInWorkspace(owner, workspace, id uuid.UUID) (Group, error) {
	group, err := s.Get(owner, id)
	if err != nil || group.WorkspaceID != workspace {
		return Group{}, ErrGroupNotFound
	}
	return group, nil
}

func (s *GroupService) Get(owner, id uuid.UUID) (Group, error) {
	group, err := s.repo.Find(id)
	if err != nil {
		return Group{}, err
	}
	if group.OwnerUserID != owner {
		return Group{}, ErrGroupNotFound
	}

	return group, nil
}

func (s *GroupService) List(owner, workspace uuid.UUID) ([]Group, error) {
	if owner == uuid.Nil() || workspace == uuid.Nil() {
		return nil, ErrInvalidGroup
	}
	return s.repo.ListByOwner(owner, workspace)
}

func (s *GroupService) Update(owner, id uuid.UUID, name string) (Group, error) {
	if !validName(name) {
		return Group{}, ErrInvalidGroup
	}
	group, err := s.Get(owner, id)
	if err != nil {
		return Group{}, err
	}
	group.Name, group.UpdatedAt = strings.TrimSpace(name), time.Now().UTC()

	return group, s.repo.Store(group)
}

func (s *GroupService) Delete(owner, id uuid.UUID) error {
	group, err := s.Get(owner, id)
	if err != nil {
		return err
	}
	if err = s.repo.Delete(id); err != nil {
		return err
	}
	return s.dispatcher.Dispatch(&usergroupevents.GroupDeleted{
		GroupId:        group.ID.String(),
		OwnerUserId:    owner.String(),
		OccurredAtUnix: time.Now().UTC().Unix(),
	})
}
