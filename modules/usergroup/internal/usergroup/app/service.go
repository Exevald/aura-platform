package app

import (
	"context"
	"uuid"

	domainevent "github.com/distributed-programming-2026/lib/event"

	"usergroup/internal/usergroup/domain"
)

type RepositoryProvider interface {
	Group(context.Context) domain.GroupRepository
	GroupMembership(context.Context) domain.GroupMembershipRepository
}

type UnitOfWork interface {
	Execute(context.Context, func(RepositoryProvider) error) error
}

type DispatcherFactory interface {
	NewDispatcher(context.Context) domainevent.Dispatcher
}

type User struct {
	ID                 uuid.UUID
	Email, DisplayName string
}

type UserClient interface {
	FindByEmail(context.Context, string) (User, error)
	Get(context.Context, uuid.UUID) (User, error)
}

type Member struct {
	UserID             uuid.UUID
	Email, DisplayName string
}

type Group struct {
	domain.Group
	Members []Member
}

type Service struct {
	unit              UnitOfWork
	dispatcherFactory DispatcherFactory
	users             UserClient
}

func NewService(
	unit UnitOfWork,
	dispatcherFactory DispatcherFactory,
	users UserClient,
) *Service {
	return &Service{
		unit:              unit,
		dispatcherFactory: dispatcherFactory,
		users:             users,
	}
}

func (s *Service) Create(ctx context.Context, owner, workspace uuid.UUID, name string) (out Group, err error) {
	err = s.executeUnitOfWork(ctx, func(groups *domain.GroupService, _ *domain.GroupMembershipService) error {
		var createErr error
		out.Group, createErr = groups.Create(owner, workspace, name)
		return createErr
	})
	return
}

func (s *Service) Get(ctx context.Context, owner, workspace, id uuid.UUID) (out Group, err error) {
	var userIDs []uuid.UUID
	err = s.executeUnitOfWork(ctx, func(groups *domain.GroupService, memberships *domain.GroupMembershipService) error {
		var getErr error
		out.Group, getErr = groups.GetInWorkspace(owner, workspace, id)
		if getErr != nil {
			return getErr
		}
		userIDs, getErr = membershipUserIDs(memberships.List(id))
		return getErr
	})
	if err == nil {
		out.Members, err = s.members(ctx, userIDs)
	}
	return
}

func (s *Service) List(ctx context.Context, owner, workspace uuid.UUID) (out []Group, err error) {
	var groups []domain.Group
	userIDsByGroup := make(map[uuid.UUID][]uuid.UUID)
	err = s.executeUnitOfWork(ctx, func(groupService *domain.GroupService, memberships *domain.GroupMembershipService) error {
		var listErr error
		groups, listErr = groupService.List(owner, workspace)
		if listErr != nil {
			return listErr
		}
		for _, group := range groups {
			userIDsByGroup[group.ID], listErr = membershipUserIDs(memberships.List(group.ID))
			if listErr != nil {
				return listErr
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	out = make([]Group, 0, len(groups))
	for _, group := range groups {
		members, memberErr := s.members(ctx, userIDsByGroup[group.ID])
		if memberErr != nil {
			return nil, memberErr
		}
		out = append(out, Group{Group: group, Members: members})
	}
	return out, nil
}

func (s *Service) Update(ctx context.Context, owner, workspace, id uuid.UUID, name string) (out Group, err error) {
	var userIDs []uuid.UUID
	err = s.executeUnitOfWork(ctx, func(groups *domain.GroupService, memberships *domain.GroupMembershipService) error {
		var updateErr error
		if _, updateErr = groups.GetInWorkspace(owner, workspace, id); updateErr != nil {
			return updateErr
		}
		out.Group, updateErr = groups.Update(owner, id, name)
		if updateErr != nil {
			return updateErr
		}
		userIDs, updateErr = membershipUserIDs(memberships.List(id))
		return updateErr
	})
	if err == nil {
		out.Members, err = s.members(ctx, userIDs)
	}
	return
}

func (s *Service) AddMember(ctx context.Context, owner, workspace, id uuid.UUID, email string) (out Group, err error) {
	user, err := s.users.FindByEmail(ctx, email)
	if err != nil {
		return Group{}, err
	}
	var userIDs []uuid.UUID
	err = s.executeUnitOfWork(ctx, func(groups *domain.GroupService, memberships *domain.GroupMembershipService) error {
		var addErr error
		out.Group, addErr = groups.GetInWorkspace(owner, workspace, id)
		if addErr != nil {
			return addErr
		}
		if addErr = memberships.Add(out.Group, user.ID); addErr != nil {
			return addErr
		}
		userIDs, addErr = membershipUserIDs(memberships.List(id))
		return addErr
	})
	if err == nil {
		out.Members, err = s.members(ctx, userIDs)
	}
	return
}

func (s *Service) RemoveMember(ctx context.Context, owner, workspace, id, userID uuid.UUID) error {
	return s.executeUnitOfWork(ctx, func(groups *domain.GroupService, memberships *domain.GroupMembershipService) error {
		group, err := groups.GetInWorkspace(owner, workspace, id)
		if err != nil {
			return err
		}
		return memberships.Remove(group, userID)
	})
}

func (s *Service) Delete(ctx context.Context, owner, workspace, id uuid.UUID) error {
	return s.executeUnitOfWork(ctx, func(groups *domain.GroupService, _ *domain.GroupMembershipService) error {
		if _, err := groups.GetInWorkspace(owner, workspace, id); err != nil {
			return err
		}
		return groups.Delete(owner, id)
	})
}

func (s *Service) OwnedMembers(ctx context.Context, owner, id uuid.UUID) (out []uuid.UUID, err error) {
	err = s.executeUnitOfWork(ctx, func(groups *domain.GroupService, memberships *domain.GroupMembershipService) error {
		if _, getErr := groups.Get(owner, id); getErr != nil {
			return getErr
		}
		out, err = membershipUserIDs(memberships.List(id))
		return err
	})
	return
}

func (s *Service) executeUnitOfWork(
	ctx context.Context,
	operation func(*domain.GroupService, *domain.GroupMembershipService) error,
) error {
	return s.unit.Execute(ctx, func(provider RepositoryProvider) error {
		dispatcher := s.dispatcherFactory.NewDispatcher(ctx)
		return operation(
			domain.NewGroupService(provider.Group(ctx), dispatcher),
			domain.NewGroupMembershipService(provider.GroupMembership(ctx), dispatcher),
		)
	})
}

func (s *Service) members(ctx context.Context, ids []uuid.UUID) ([]Member, error) {
	result := make([]Member, 0, len(ids))
	for _, id := range ids {
		user, err := s.users.Get(ctx, id)
		if err != nil {
			return nil, err
		}
		result = append(result, Member{UserID: id, Email: user.Email, DisplayName: user.DisplayName})
	}
	return result, nil
}

func membershipUserIDs(memberships []domain.GroupMembership, err error) ([]uuid.UUID, error) {
	if err != nil {
		return nil, err
	}
	result := make([]uuid.UUID, 0, len(memberships))
	for _, membership := range memberships {
		result = append(result, membership.UserID)
	}
	return result, nil
}
