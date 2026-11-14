package service

import (
	"access/internal/access/domain/model"
	"testing"
	"uuid"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	accessevents "access/api/events"
)

type mockModuleAccessRepo struct {
	values []model.ModuleAccess
}

func (*mockModuleAccessRepo) NextID() (uuid.UUID, error) {
	return uuid.New(), nil
}

func (r *mockModuleAccessRepo) Store(v model.ModuleAccess) error {
	for i, existing := range r.values {
		if existing.SourceID == v.SourceID && existing.UserID == v.UserID && existing.ModuleID == v.ModuleID {
			r.values[i] = v
			return nil
		}
	}
	r.values = append(r.values, v)
	return nil
}

func (r *mockModuleAccessRepo) Find(id uuid.UUID) (model.ModuleAccess, error) {
	for _, value := range r.values {
		if value.ID == id {
			return value, nil
		}
	}
	return model.ModuleAccess{}, model.ErrModuleAccessNotFound
}

func (r *mockModuleAccessRepo) FindByUserAndModule(userID, moduleID uuid.UUID) (out []model.ModuleAccess, err error) {
	for _, value := range r.values {
		if value.UserID == userID && value.ModuleID == moduleID {
			out = append(out, value)
		}
	}
	return
}

func (r *mockModuleAccessRepo) Delete(id uuid.UUID) error {
	out := r.values[:0]
	for _, value := range r.values {
		if value.ID != id {
			out = append(out, value)
		}
	}
	r.values = out
	return nil
}

func (r *mockModuleAccessRepo) DeleteBySource(source string) error {
	out := r.values[:0]
	for _, value := range r.values {
		if value.SourceID != source {
			out = append(out, value)
		}
	}
	r.values = out
	return nil
}

func (r *mockModuleAccessRepo) DeleteBySourceAndUser(source string, userID uuid.UUID) error {
	out := r.values[:0]
	for _, value := range r.values {
		if value.SourceID != source || value.UserID != userID {
			out = append(out, value)
		}
	}
	r.values = out
	return nil
}

type mockOperationRepo struct {
	values map[uuid.UUID]model.ModuleShare
}

func (r *mockOperationRepo) ListByGroup(groupID uuid.UUID) (out []model.ModuleShare, err error) {
	for _, value := range r.values {
		if value.GroupID == groupID && value.Status == model.ModuleShareGranted {
			out = append(out, value)
		}
	}
	return
}

func (r *mockOperationRepo) ListByOwner(ownerID uuid.UUID) (out []model.ModuleShare, err error) {
	for _, value := range r.values {
		if value.OwnerUserID == ownerID && value.Status == model.ModuleShareGranted {
			out = append(out, value)
		}
	}
	return
}

func (r *mockOperationRepo) Find(id uuid.UUID) (model.ModuleShare, error) {
	value, ok := r.values[id]
	if !ok {
		return model.ModuleShare{}, model.ErrModuleShareOperationNotFound
	}
	return value, nil
}

func (r *mockOperationRepo) Store(value model.ModuleShare) error {
	if r.values == nil {
		r.values = map[uuid.UUID]model.ModuleShare{}
	}
	r.values[value.ShareID] = value
	return nil
}

type moduleAccessDispatcher struct {
	values []proto.Message
}

func (d *moduleAccessDispatcher) Dispatch(value proto.Message) error {
	d.values = append(d.values, value)
	return nil
}

func TestGroupGrantIsIdempotentAndCompensatesOnlyItsSource(t *testing.T) {
	accesses := &mockModuleAccessRepo{}
	operations := &mockOperationRepo{}
	events := &moduleAccessDispatcher{}
	service := NewGroupModuleAccessService(accesses, operations, events)
	user, module, group, share := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	_, err := NewModuleAccessService(accesses, events).CreateWithSource(
		user,
		module,
		model.EffectAllow,
		"direct:test",
	)
	require.NoError(t, err)
	op := model.ModuleShare{ShareID: share, ModuleID: module, GroupID: group, OwnerUserID: uuid.New()}
	require.NoError(t, service.Grant(op, []uuid.UUID{user}))
	require.NoError(t, service.Grant(op, []uuid.UUID{user}))
	require.Len(t, accesses.values, 2)
	require.NoError(t, service.Cancel(share, module, group, "failed"))
	require.Len(t, accesses.values, 1)
	require.Equal(t, "direct:test", accesses.values[0].SourceID)
	require.IsType(t, &accessevents.GroupModuleAccessRevoked{}, events.values[len(events.values)-1])
}

func TestMembershipChangesUpdateOnlyActiveGroupSources(t *testing.T) {
	accesses := &mockModuleAccessRepo{}
	operations := &mockOperationRepo{}
	events := &moduleAccessDispatcher{}
	service := NewGroupModuleAccessService(accesses, operations, events)
	moduleID, groupID, shareID := uuid.New(), uuid.New(), uuid.New()
	require.NoError(t, service.Grant(model.ModuleShare{
		ShareID:     shareID,
		ModuleID:    moduleID,
		GroupID:     groupID,
		OwnerUserID: uuid.New(),
	}, []uuid.UUID{uuid.New()}))
	memberID := uuid.New()
	require.NoError(t, service.Membership(groupID, memberID, true))
	values, err := accesses.FindByUserAndModule(memberID, moduleID)
	require.NoError(t, err)
	require.Len(t, values, 1)
	require.Equal(t, "group:"+shareID.String(), values[0].SourceID)
	require.NoError(t, service.Membership(groupID, memberID, false))
	values, err = accesses.FindByUserAndModule(memberID, moduleID)
	require.NoError(t, err)
	require.Empty(t, values)
}

func TestCancelledShareRejectsLateGrant(t *testing.T) {
	accesses := &mockModuleAccessRepo{}
	operations := &mockOperationRepo{}
	events := &moduleAccessDispatcher{}
	service := NewGroupModuleAccessService(accesses, operations, events)
	moduleID, groupID, shareID := uuid.New(), uuid.New(), uuid.New()
	require.NoError(t, service.Cancel(shareID, moduleID, groupID, "timeout"))
	require.NoError(t, service.Grant(model.ModuleShare{
		ShareID:     shareID,
		ModuleID:    moduleID,
		GroupID:     groupID,
		OwnerUserID: uuid.New(),
	}, []uuid.UUID{uuid.New()}))
	require.Empty(t, accesses.values)
	require.IsType(t, &accessevents.GroupModuleAccessRevoked{}, events.values[len(events.values)-1])
}

func TestRevokingOneOfTwoGroupsKeepsModuleAccess(t *testing.T) {
	accesses := &mockModuleAccessRepo{}
	operations := &mockOperationRepo{}
	events := &moduleAccessDispatcher{}
	groups := NewGroupModuleAccessService(accesses, operations, events)
	moduleID, userID := uuid.New(), uuid.New()
	first := model.ModuleShare{ShareID: uuid.New(), ModuleID: moduleID, GroupID: uuid.New(), OwnerUserID: uuid.New()}
	second := model.ModuleShare{ShareID: uuid.New(), ModuleID: moduleID, GroupID: uuid.New(), OwnerUserID: uuid.New()}
	require.NoError(t, groups.Grant(first, []uuid.UUID{userID}))
	require.NoError(t, groups.Grant(second, []uuid.UUID{userID}))

	allowed, err := NewModuleAccessService(accesses, events).CheckModuleAccess(userID, moduleID)
	require.NoError(t, err)
	require.True(t, allowed)
	require.NoError(t, groups.Cancel(first.ShareID, first.ModuleID, first.GroupID, "revoked_by_owner"))
	allowed, err = NewModuleAccessService(accesses, events).CheckModuleAccess(userID, moduleID)
	require.NoError(t, err)
	require.True(t, allowed)
	require.NoError(t, groups.Cancel(second.ShareID, second.ModuleID, second.GroupID, "revoked_by_owner"))
	allowed, err = NewModuleAccessService(accesses, events).CheckModuleAccess(userID, moduleID)
	require.NoError(t, err)
	require.False(t, allowed)
}
