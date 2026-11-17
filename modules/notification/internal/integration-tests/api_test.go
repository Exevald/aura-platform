package integration_tests

import (
	"context"
	"net/http"
	"testing"
	"time"
	"uuid"

	openapitypes "github.com/oapi-codegen/runtime/types"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	internalapi "notification/api/server/internalapi"
	publicapi "notification/api/server/publicapi"
)

func TestNotificationInboxAndPreferences(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	user, workspace := uuid.New(), uuid.New()
	connection, err := grpc.NewClient(testEnv.conf.GRPC.Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer connection.Close()
	createTestNotification(ctx, t, internalapi.NewNotificationServiceClient(connection), user, workspace)
	client, err := publicapi.NewClientWithResponses("http://" + testEnv.conf.HTTP.Address)
	require.NoError(t, err)
	verifyNotificationInbox(ctx, t, client, user)
	verifyNotificationPreferences(ctx, t, client, user)
}

func createTestNotification(ctx context.Context, t *testing.T, client internalapi.NotificationServiceClient, user, workspace uuid.UUID) {
	t.Helper()
	response, err := client.CreateNotification(ctx, &internalapi.CreateNotificationRequest{
		IdempotencyKey: "test-1", UserId: user.String(), WorkspaceId: workspace.String(),
		Kind: "reminder", Title: "Встреча", Body: "Через 15 минут",
	})
	require.NoError(t, err)
	require.True(t, response.Created)
}

func verifyNotificationInbox(ctx context.Context, t *testing.T, client *publicapi.ClientWithResponses, user uuid.UUID) {
	t.Helper()
	response, err := client.ListNotificationsWithResponse(ctx,
		&publicapi.ListNotificationsParams{XAuthUser: openapitypes.UUID(user)})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, response.StatusCode())
	require.Len(t, *response.JSON200, 1)
}

func verifyNotificationPreferences(ctx context.Context, t *testing.T, client *publicapi.ClientWithResponses, user uuid.UUID) {
	t.Helper()
	response, err := client.PutPreferencesWithResponse(ctx,
		&publicapi.PutPreferencesParams{XAuthUser: openapitypes.UUID(user)},
		publicapi.PutPreferencesJSONRequestBody{InApp: true, Email: false})
	require.NoError(t, err)
	require.False(t, response.JSON200.Email)
}
