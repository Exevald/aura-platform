package integration_tests

import (
	"context"
	"testing"
	"time"
	"uuid"

	"github.com/distributed-programming-2026/go-sdk/pkg/event"
	rabbit "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/proto"

	accessapi "access/api/server/internalapi"
	usergroupevents "usergroup/api/events"
)

const membershipAdded = "added"

type groupEventScenario struct {
	t       *testing.T
	ctx     context.Context
	module  uuid.UUID
	client  accessapi.ModuleAccessServiceClient
	channel *rabbit.Channel
}

func TestUserGroupEventsUpdateModuleAccess(t *testing.T) {
	owner, member, secondMember := uuid.New(), uuid.New(), uuid.New()
	group, otherGroup, module := uuid.New(), uuid.New(), uuid.New()
	share, otherShare := uuid.New(), uuid.New()
	seedGroupAccess(t, group, module, owner, share, member)
	seedGroupAccess(t, otherGroup, module, owner, otherShare, member)
	s := newGroupEventScenario(t, module)

	s.verifyAddAndDuplicate(group, owner, secondMember, share)
	s.verifyMalformedEvent(group, owner, secondMember, share)
	s.verifyOwnerAndMemberRemoval(group, owner, member, share, otherShare)
	s.verifyGroupDeletion(group, otherGroup, owner, member, secondMember, share, otherShare)
}

func newGroupEventScenario(t *testing.T, module uuid.UUID) *groupEventScenario {
	t.Helper()
	connection, err := grpc.NewClient(testEnv.conf.GRPC.Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, connection.Close()) })
	broker, err := rabbit.Dial(testEnv.conf.AMQP.URL)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, broker.Close()) })
	channel, err := broker.Channel()
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, channel.Close()) })
	return &groupEventScenario{t: t, ctx: t.Context(), module: module,
		client: accessapi.NewModuleAccessServiceClient(connection), channel: channel}
}

func (s *groupEventScenario) verifyAddAndDuplicate(group, owner, user, share uuid.UUID) {
	s.t.Helper()
	message := &usergroupevents.GroupMembershipChanged{Type: membershipAdded, GroupId: group.String(), UserId: user.String(), OwnerUserId: owner.String()}
	s.publish(message)
	s.waitCount(user, share, 1)
	require.True(s.t, s.allowed(user))
	duplicateID := s.publish(message)
	s.waitProcessed(duplicateID)
	require.Equal(s.t, 1, s.accessCount(user, share))
}

func (s *groupEventScenario) verifyMalformedEvent(group, owner, user, share uuid.UUID) {
	s.t.Helper()
	s.publish(&usergroupevents.GroupMembershipChanged{Type: membershipAdded, GroupId: group.String(), UserId: "invalid-uuid", OwnerUserId: owner.String()})
	require.Eventually(s.t, func() bool {
		queue, err := s.channel.QueueDeclarePassive(testEnv.conf.AMQP.Queue+".dead", true, false, false, false, nil)
		return err == nil && queue.Messages > 0
	}, 5*time.Second, 50*time.Millisecond)
	require.Equal(s.t, 1, s.accessCount(user, share))
}

func (s *groupEventScenario) verifyOwnerAndMemberRemoval(group, owner, member, share, otherShare uuid.UUID) {
	s.t.Helper()
	id := s.publish(&usergroupevents.GroupMembershipChanged{Type: "removed", GroupId: group.String(), UserId: owner.String(), OwnerUserId: owner.String()})
	s.waitProcessed(id)
	require.Equal(s.t, 1, s.accessCount(owner, share))
	s.publish(&usergroupevents.GroupMembershipChanged{Type: "removed", GroupId: group.String(), UserId: member.String(), OwnerUserId: owner.String()})
	s.waitCount(member, share, 0)
	require.Equal(s.t, 1, s.accessCount(member, otherShare))
	require.True(s.t, s.allowed(member))
}

func (s *groupEventScenario) verifyGroupDeletion(group, otherGroup, owner, member, secondMember, share, otherShare uuid.UUID) {
	s.t.Helper()
	s.publish(&usergroupevents.GroupDeleted{GroupId: otherGroup.String(), OwnerUserId: owner.String()})
	s.waitCount(member, otherShare, 0)
	require.False(s.t, s.allowed(member))
	var status string
	require.NoError(s.t, testEnv.connector.TransactionalClient().GetContext(s.ctx, "test.share-status", &status,
		"SELECT status FROM module_share_operation WHERE share_id = ?", otherShare[:]))
	require.Equal(s.t, "cancelled", status)
	s.publish(&usergroupevents.GroupDeleted{GroupId: group.String(), OwnerUserId: owner.String()})
	s.waitCount(secondMember, share, 0)
	require.False(s.t, s.allowed(secondMember))
}

func (s *groupEventScenario) publish(message proto.Message) string {
	s.t.Helper()
	envelope, err := event.New(s.ctx, "usergroup", string(message.ProtoReflect().Descriptor().Name()), message)
	require.NoError(s.t, err)
	body, err := envelope.Marshal()
	require.NoError(s.t, err)
	require.NoError(s.t, s.channel.PublishWithContext(s.ctx, "domain-event", "usergroup."+envelope.Type(), false, false, rabbit.Publishing{
		ContentType: "application/vnd.distributed-programming.event+protobuf",
		Type:        envelope.Type(), MessageId: envelope.ID(), Body: body,
	}))
	return envelope.ID()
}

func (s *groupEventScenario) waitProcessed(id string) {
	s.t.Helper()
	require.Eventually(s.t, func() bool {
		var count int
		err := testEnv.connector.TransactionalClient().GetContext(s.ctx, "test.inbox-count", &count,
			"SELECT COUNT(*) FROM event_inbox WHERE consumer = ? AND event_id = ?", testEnv.conf.AMQP.Queue, id)
		return err == nil && count == 1
	}, 5*time.Second, 50*time.Millisecond)
}

func (s *groupEventScenario) accessCount(user, share uuid.UUID) int {
	s.t.Helper()
	var count int
	require.NoError(s.t, testEnv.connector.TransactionalClient().GetContext(s.ctx, "test.access-count", &count,
		"SELECT COUNT(*) FROM module_access WHERE user_id = ? AND module_id = ? AND source_id = ?",
		user[:], s.module[:], "group:"+share.String()))
	return count
}

func (s *groupEventScenario) waitCount(user, share uuid.UUID, expected int) {
	s.t.Helper()
	require.EventuallyWithT(s.t, func(c *assert.CollectT) {
		assert.Equal(c, expected, s.accessCount(user, share))
	}, 5*time.Second, 50*time.Millisecond)
}

func (s *groupEventScenario) allowed(user uuid.UUID) bool {
	s.t.Helper()
	response, err := s.client.CheckModuleAccess(s.ctx,
		&accessapi.CheckModuleAccessRequest{UserId: user.String(), ModuleId: s.module.String()})
	require.NoError(s.t, err)
	return response.GetAllowed()
}

func seedGroupAccess(t *testing.T, group, module, owner, share, member uuid.UUID) {
	t.Helper()
	ctx := t.Context()
	db := testEnv.connector.TransactionalClient()
	_, err := db.ExecContext(ctx, "test.seed-share", "INSERT INTO module_share_operation (share_id, module_id, group_id, owner_user_id, status, created_at, updated_at) VALUES (?, ?, ?, ?, 'granted', NOW(6), NOW(6))", share[:], module[:], group[:], owner[:])
	require.NoError(t, err)
	for _, user := range []uuid.UUID{owner, member} {
		id := uuid.New()
		_, err = db.ExecContext(ctx, "test.seed-access", "INSERT INTO module_access (id, user_id, module_id, effect, source_id, created_at, updated_at) VALUES (?, ?, ?, 'allow', ?, NOW(6), NOW(6))", id[:], user[:], module[:], "group:"+share.String())
		require.NoError(t, err)
	}
}
