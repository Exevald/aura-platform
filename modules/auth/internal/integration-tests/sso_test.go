package integration_tests

import (
	"context"
	"net"
	"net/http"
	"net/http/cookiejar"
	"testing"
	"time"
	"uuid"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	authinternalapi "auth/api/server/internalapi"
	"auth/api/server/publicapi"
)

func TestSSOTicketIsBoundToOneProductAndCannotBeReplayed(t *testing.T) {
	login := "sso-" + uuid.New().String()
	password := "correct horse battery staple"
	registerSSOUser(t, login, password)
	httpClient := newSSOHTTPClient(t)

	verifySSOTicketConsumption(t, httpClient, login, password)
}

func registerSSOUser(t *testing.T, login, password string) {
	t.Helper()
	connection, err := grpc.NewClient(testEnv.conf.GRPC.Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, connection.Close()) })
	participant := authinternalapi.NewAuthServiceClient(connection)
	transactionID := uuid.New().String()
	prepared, err := participant.PrepareCredentialRegistration(t.Context(), &authinternalapi.PrepareCredentialRegistrationRequest{
		TransactionId: transactionID, UserId: uuid.New().String(), Login: login, Password: password,
	})
	require.NoError(t, err)
	require.NotEmpty(t, prepared.GetCredentialId())
	_, err = participant.CommitCredentialRegistration(t.Context(), &authinternalapi.RegistrationTransactionRequest{TransactionId: transactionID})
	require.NoError(t, err)
}

func newSSOHTTPClient(t *testing.T) *http.Client {
	t.Helper()
	jar, err := cookiejar.New(nil)
	require.NoError(t, err)
	return &http.Client{
		Timeout: 10 * time.Second,
		Jar:     jar,
		Transport: &http.Transport{DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
			return (&net.Dialer{}).DialContext(ctx, "tcp", testEnv.conf.HTTP.Address)
		}},
	}
}

func verifySSOTicketConsumption(t *testing.T, httpClient *http.Client, login, password string) {
	t.Helper()
	clientForHost := func(host string) *publicapi.ClientWithResponses {
		t.Helper()
		client, clientErr := publicapi.NewClientWithResponses("http://"+host, publicapi.WithHTTPClient(httpClient))
		require.NoError(t, clientErr)
		return client
	}
	central := clientForHost("aura-platform.auth")
	loggedIn, err := central.LoginWithResponse(t.Context(), publicapi.LoginJSONRequestBody{
		Login: login, Password: password, RememberMe: true,
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, loggedIn.StatusCode())
	issue, err := central.IssueSsoTicketWithResponse(t.Context(), publicapi.IssueSsoTicketJSONRequestBody{
		TargetHost: "aura-platform.documents",
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, issue.StatusCode(), string(issue.Body))
	require.NotNil(t, issue.JSON200)
	consume := func(host string) *publicapi.ConsumeSsoTicketResponse {
		t.Helper()
		response, consumeErr := clientForHost(host).ConsumeSsoTicketWithResponse(t.Context(), publicapi.ConsumeSsoTicketJSONRequestBody{
			Code: issue.JSON200.Code,
		})
		require.NoError(t, consumeErr)
		return response
	}
	require.Equal(t, http.StatusBadRequest, consume("aura-platform.calendar").StatusCode())
	require.Equal(t, http.StatusNoContent, consume("aura-platform.documents").StatusCode())
	require.Equal(t, http.StatusBadRequest, consume("aura-platform.documents").StatusCode())
}
