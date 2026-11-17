package integration_tests

import (
	"context"
	"testing"
	"time"
	"uuid"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	sessioninternalapi "session/api/server/internalapi"
)

func TestSessionLifecycle(t *testing.T) {
	connection, err := grpc.NewClient(
		testEnv.conf.GRPC.Address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, connection.Close()) })
	client := sessioninternalapi.NewSessionServiceClient(connection)
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	userID := uuid.New()
	regularID, persistentID := createTestSessions(ctx, t, client, userID)

	verifyPersistentSession(ctx, t, client, userID, persistentID)
	verifySessionRevocation(ctx, t, client, userID, regularID, persistentID)
}

func createTestSessions(ctx context.Context, t *testing.T, client sessioninternalapi.SessionServiceClient, userID uuid.UUID) (regularID, persistentID string) {
	t.Helper()
	regular, err := client.CreateSession(ctx, &sessioninternalapi.CreateSessionRequest{
		UserId: userID.String(),
	})
	require.NoError(t, err)
	require.WithinDuration(t, time.Now().Add(8*time.Hour), regular.GetExpiresAt().AsTime(), time.Minute)

	persistent, err := client.CreateSession(ctx, &sessioninternalapi.CreateSessionRequest{
		UserId: userID.String(), RememberMe: true,
	})
	require.NoError(t, err)
	require.WithinDuration(t, time.Now().Add(30*24*time.Hour), persistent.GetExpiresAt().AsTime(), time.Minute)
	return regular.GetSessionId(), persistent.GetSessionId()
}

func verifyPersistentSession(ctx context.Context, t *testing.T, client sessioninternalapi.SessionServiceClient, userID uuid.UUID, persistentID string) {
	t.Helper()
	validated, err := client.ValidateSession(
		ctx,
		&sessioninternalapi.ValidateSessionRequest{SessionId: persistentID},
	)
	require.NoError(t, err)
	require.Equal(t, userID.String(), validated.GetUserId())
	require.True(t, validated.GetPersistent())
}

func verifySessionRevocation(ctx context.Context, t *testing.T, client sessioninternalapi.SessionServiceClient, userID uuid.UUID, regularID, persistentID string) {
	t.Helper()
	_, err := client.RevokeUserSessions(ctx, &sessioninternalapi.RevokeUserSessionsRequest{
		UserId: userID.String(),
	})
	require.NoError(t, err)
	_, err = client.ValidateSession(ctx, &sessioninternalapi.ValidateSessionRequest{
		SessionId: regularID,
	})
	require.Equal(t, codes.Unauthenticated, status.Code(err))
	_, err = client.ValidateSession(ctx, &sessioninternalapi.ValidateSessionRequest{
		SessionId: persistentID,
	})
	require.Equal(t, codes.Unauthenticated, status.Code(err))
}
