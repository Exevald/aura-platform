package domain

import (
	"testing"
	"uuid"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	usergroupevents "usergroup/api/events"
)

type groupMembershipRepositoryFake struct {
	values map[uuid.UUID]map[uuid.UUID]GroupMembership
}

type groupMembershipDispatcherFake struct {
	messages []proto.Message
}

func TestGroupMembershipServiceChangesMembershipIdempotently(t *testing.T) {
	repository := &groupMembershipRepositoryFake{}
	dispatcher := &groupMembershipDispatcherFake{}
	service := NewGroupMembershipService(repository, dispatcher)
	group := Group{ID: uuid.New(), OwnerUserID: uuid.New()}
	userID := uuid.New()

	require.NoError(t, service.Add(group, userID))
	require.NoError(t, service.Add(group, userID))

	memberships, err := service.List(group.ID)
	require.NoError(t, err)
	require.Len(t, memberships, 1)
	require.Equal(t, userID, memberships[0].UserID)
	require.False(t, memberships[0].CreatedAt.IsZero())
	require.Len(t, dispatcher.messages, 1)
	added := dispatcher.messages[0].(*usergroupevents.GroupMembershipChanged)
	require.Equal(t, "added", added.GetType())
	require.Equal(t, group.OwnerUserID.String(), added.GetOwnerUserId())

	require.NoError(t, service.Remove(group, userID))
	require.NoError(t, service.Remove(group, userID))
	require.Empty(t, repository.values[group.ID])
	require.Len(t, dispatcher.messages, 2)
	removed := dispatcher.messages[1].(*usergroupevents.GroupMembershipChanged)
	require.Equal(t, "removed", removed.GetType())
}

func TestGroupMembershipServiceRejectsEmptyUser(t *testing.T) {
	service := NewGroupMembershipService(
		&groupMembershipRepositoryFake{},
		&groupMembershipDispatcherFake{},
	)

	err := service.Add(Group{ID: uuid.New(), OwnerUserID: uuid.New()}, uuid.Nil())

	require.ErrorIs(t, err, ErrInvalidGroup)
}

func (r *groupMembershipRepositoryFake) ListByGroup(groupID uuid.UUID) ([]GroupMembership, error) {
	result := make([]GroupMembership, 0, len(r.values[groupID]))
	for _, membership := range r.values[groupID] {
		result = append(result, membership)
	}
	return result, nil
}

func (r *groupMembershipRepositoryFake) Store(membership GroupMembership) (bool, error) {
	if r.values == nil {
		r.values = make(map[uuid.UUID]map[uuid.UUID]GroupMembership)
	}
	if r.values[membership.GroupID] == nil {
		r.values[membership.GroupID] = make(map[uuid.UUID]GroupMembership)
	}
	if _, exists := r.values[membership.GroupID][membership.UserID]; exists {
		return false, nil
	}
	r.values[membership.GroupID][membership.UserID] = membership
	return true, nil
}

func (r *groupMembershipRepositoryFake) Delete(groupID, userID uuid.UUID) (bool, error) {
	if _, exists := r.values[groupID][userID]; !exists {
		return false, nil
	}
	delete(r.values[groupID], userID)
	return true, nil
}

func (d *groupMembershipDispatcherFake) Dispatch(message proto.Message) error {
	d.messages = append(d.messages, message)
	return nil
}
