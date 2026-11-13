package integration_tests

import (
	"encoding/json"
	"net/http"
	api "realtime/api/server/internalapi"
	"testing"
	"time"
	"uuid"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestTargetedWebSocketDelivery(t *testing.T) {
	user, workspace := uuid.New(), uuid.New()
	headers := http.Header{"X-Auth-User": []string{user.String()}, "X-Workspace-ID": []string{workspace.String()}}
	socket, response, e := websocket.DefaultDialer.Dial("ws://"+testEnv.conf.HTTP.Address+"/api/realtime/v1/ws", headers)
	if response != nil {
		defer response.Body.Close()
	}
	require.NoError(t, e)
	defer socket.Close()
	conn, e := grpc.NewClient(testEnv.conf.GRPC.Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, e)
	defer conn.Close()
	payload := []byte(`{"title":"Напоминание"}`)
	result, e := api.NewRealtimeServiceClient(conn).Publish(t.Context(), &api.PublishRequest{UserId: user.String(), WorkspaceId: workspace.String(), Type: "NotificationCreated", PayloadJson: payload})
	require.NoError(t, e)
	require.EqualValues(t, 1, result.Delivered)
	require.NoError(t, socket.SetReadDeadline(time.Now().Add(5*time.Second)))
	var message struct {
		Type    string          `json:"type"`
		Payload json.RawMessage `json:"payload"`
	}
	require.NoError(t, socket.ReadJSON(&message))
	require.Equal(t, "NotificationCreated", message.Type)
	require.JSONEq(t, string(payload), string(message.Payload))
}
