package events

import (
	"context"
	"testing"
	"uuid"

	sdkevent "github.com/distributed-programming-2026/go-sdk/pkg/event"
	"github.com/stretchr/testify/require"

	"access/internal/access/app"
	usergroupevents "usergroup/api/events"
)

type recordingModuleAccess struct {
	app.ModuleAccessService
	changes int
}

func (s *recordingModuleAccess) ChangeGroupMembership(context.Context, uuid.UUID, uuid.UUID, bool) error {
	s.changes++
	return nil
}

func TestOwnerMembershipChangesDoNotRevokeAutomaticAccess(t *testing.T) {
	ownerID := uuid.New()
	payload := &usergroupevents.GroupMembershipChanged{
		Type:        "removed",
		GroupId:     uuid.New().String(),
		UserId:      ownerID.String(),
		OwnerUserId: ownerID.String(),
	}
	envelope, err := sdkevent.New(t.Context(), "usergroup", "GroupMembershipChanged", payload)
	require.NoError(t, err)
	svc := &recordingModuleAccess{}

	require.NoError(t, NewHandler(svc).Handle(t.Context(), envelope))
	require.Zero(t, svc.changes)
}

func TestRegularMembershipChangesStillUpdateAccess(t *testing.T) {
	payload := &usergroupevents.GroupMembershipChanged{
		Type:        "added",
		GroupId:     uuid.New().String(),
		UserId:      uuid.New().String(),
		OwnerUserId: uuid.New().String(),
	}
	envelope, err := sdkevent.New(t.Context(), "usergroup", "GroupMembershipChanged", payload)
	require.NoError(t, err)
	svc := &recordingModuleAccess{}

	require.NoError(t, NewHandler(svc).Handle(t.Context(), envelope))
	require.Equal(t, 1, svc.changes)
}
