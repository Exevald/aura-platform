package integration_tests

import (
	"context"
	"net/http"
	internalapi "notification/api/server/internalapi"
	publicapi "notification/api/server/publicapi"
	"testing"
	"time"
	"uuid"

	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestNotificationInboxAndPreferences(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	user, workspace := uuid.New(), uuid.New()
	conn, e := grpc.NewClient(testEnv.conf.GRPC.Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, e)
	defer conn.Close()
	created, e := internalapi.NewNotificationServiceClient(conn).CreateNotification(ctx, &internalapi.CreateNotificationRequest{IdempotencyKey: "test-1", UserId: user.String(), WorkspaceId: workspace.String(), Kind: "reminder", Title: "Встреча", Body: "Через 15 минут"})
	require.NoError(t, e)
	require.True(t, created.Created)
	client, e := publicapi.NewClientWithResponses("http://" + testEnv.conf.HTTP.Address)
	require.NoError(t, e)
	list, e := client.ListNotificationsWithResponse(ctx, &publicapi.ListNotificationsParams{XAuthUser: openapi_types.UUID(user)})
	require.NoError(t, e)
	require.Equal(t, http.StatusOK, list.StatusCode())
	require.Len(t, *list.JSON200, 1)
	prefs, e := client.PutPreferencesWithResponse(ctx, &publicapi.PutPreferencesParams{XAuthUser: openapi_types.UUID(user)}, publicapi.PutPreferencesJSONRequestBody{InApp: true, Email: false})
	require.NoError(t, e)
	require.False(t, prefs.JSON200.Email)
}
