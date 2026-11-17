package integration_tests

import (
	"io"
	api "mailer/api/server/internalapi"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestSMTPDeliveryAndIdempotency(t *testing.T) {
	conn, e := grpc.NewClient(testEnv.conf.GRPC.Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, e)
	defer conn.Close()
	client := api.NewMailerServiceClient(conn)
	request := &api.SendMailRequest{IdempotencyKey: "invite:42", Recipient: "guest@example.com", Subject: "Приглашение в Aura", Body: "Добро пожаловать"}
	first := sendTestMail(t, client, request)
	verifyMailIdempotency(t, client, request, first)
	verifyMailpitDelivery(t)
}

func sendTestMail(t *testing.T, client api.MailerServiceClient, request *api.SendMailRequest) *api.SendMailResponse {
	t.Helper()
	first, e := client.SendMail(t.Context(), request)
	require.NoError(t, e)
	require.True(t, first.Created)
	require.Equal(t, "sent", first.Status)
	return first
}

func verifyMailIdempotency(t *testing.T, client api.MailerServiceClient, request *api.SendMailRequest, first *api.SendMailResponse) {
	t.Helper()
	second, e := client.SendMail(t.Context(), request)
	require.NoError(t, e)
	require.False(t, second.Created)
	require.Equal(t, first.MailId, second.MailId)
}

func verifyMailpitDelivery(t *testing.T) {
	t.Helper()
	require.Eventually(t, func() bool {
		response, e := http.Get(testEnv.mailpitAPI + "/api/v1/messages")
		if e != nil {
			return false
		}
		defer response.Body.Close()
		body, _ := io.ReadAll(response.Body)
		return strings.Contains(string(body), "guest@example.com") && strings.Contains(string(body), "Приглашение в Aura")
	}, 5*time.Second, 100*time.Millisecond)
}
