package integration_tests

import (
	"net/http"
	"net/http/cookiejar"
	"testing"
	"time"
	"uuid"

	rabbit "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	authinternalapi "auth/api/server/internalapi"
	"auth/api/server/publicapi"
)

func TestCredentialParticipantLoginLogoutAndPasswordChange(t *testing.T) {
	const (
		login = "ada"
	)
	deliveries := subscribeDomainEvents(t)
	jar, err := cookiejar.New(nil)
	require.NoError(t, err)
	httpClient := &http.Client{Timeout: 10 * time.Second, Jar: jar}
	client, err := publicapi.NewClientWithResponses(
		"http://"+testEnv.conf.HTTP.Address,
		publicapi.WithHTTPClient(httpClient),
	)
	require.NoError(t, err)
	password := "p-Tc!9Hns8aw76" //nolint:gosec // Regression value from the reported registration request.
	nextPassword := "another correct horse password"
	connection, err := grpc.NewClient(testEnv.conf.GRPC.Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, connection.Close()) })
	participant := authinternalapi.NewAuthServiceClient(connection)
	userID := uuid.New()
	_, err = participant.PrepareCredentialRegistration(t.Context(), &authinternalapi.PrepareCredentialRegistrationRequest{
		TransactionId: uuid.New().String(), UserId: userID.String(), Login: "short-password", Password: "1234567",
	})
	require.Equal(t, codes.InvalidArgument, status.Code(err))
	transactionID := uuid.New()
	prepared, err := participant.PrepareCredentialRegistration(t.Context(), &authinternalapi.PrepareCredentialRegistrationRequest{
		TransactionId: transactionID.String(), UserId: userID.String(), Login: login, Password: password,
	})
	require.NoError(t, err)
	require.NotEmpty(t, prepared.GetCredentialId())
	_, err = participant.CommitCredentialRegistration(t.Context(), &authinternalapi.RegistrationTransactionRequest{TransactionId: transactionID.String()})
	require.NoError(t, err)

	var passwordHash string
	err = testEnv.connector.TransactionalClient().GetContext(
		t.Context(),
		"test.credential",
		&passwordHash,
		"SELECT password_hash FROM credential WHERE login = ?",
		login,
	)
	require.NoError(t, err)
	require.NotEqual(t, password, passwordHash)
	require.Contains(t, passwordHash, "$argon2id$")

	_, err = participant.PrepareCredentialRegistration(t.Context(), &authinternalapi.PrepareCredentialRegistrationRequest{
		TransactionId: uuid.New().String(), UserId: uuid.New().String(), Login: login, Password: password,
	})
	require.Equal(t, codes.AlreadyExists, status.Code(err))

	loggedOut, err := client.LogoutWithResponse(t.Context())
	require.NoError(t, err)
	require.Equal(t, http.StatusNoContent, loggedOut.StatusCode())

	invalid, err := client.LoginWithResponse(t.Context(), publicapi.LoginJSONRequestBody{
		Login: login, Password: "wrong password", RememberMe: true,
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusUnauthorized, invalid.StatusCode())

	loggedIn, err := client.LoginWithResponse(t.Context(), publicapi.LoginJSONRequestBody{
		Login: login, Password: password, RememberMe: true,
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, loggedIn.StatusCode())
	require.NotNil(t, loggedIn.JSON200)
	require.WithinDuration(t, time.Now().Add(30*24*time.Hour), *loggedIn.JSON200.ExpiresAt, time.Minute)
	require.Len(t, loggedIn.HTTPResponse.Cookies(), 1)
	require.Greater(t, loggedIn.HTTPResponse.Cookies()[0].MaxAge, 29*24*60*60)

	failedChange, err := client.ChangePasswordWithResponse(t.Context(), publicapi.ChangePasswordJSONRequestBody{
		CurrentPassword: "wrong password",
		NewPassword:     nextPassword,
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusBadRequest, failedChange.StatusCode())

	changed, err := client.ChangePasswordWithResponse(t.Context(), publicapi.ChangePasswordJSONRequestBody{
		CurrentPassword: password,
		NewPassword:     nextPassword,
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusNoContent, changed.StatusCode())
	require.Len(t, changed.HTTPResponse.Cookies(), 1)
	require.Less(t, changed.HTTPResponse.Cookies()[0].MaxAge, 0)

	oldPassword, err := client.LoginWithResponse(t.Context(), publicapi.LoginJSONRequestBody{
		Login: login, Password: password, RememberMe: false,
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusUnauthorized, oldPassword.StatusCode())

	newPassword, err := client.LoginWithResponse(t.Context(), publicapi.LoginJSONRequestBody{
		Login: login, Password: nextPassword, RememberMe: false,
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, newPassword.StatusCode())
	requireEventTypes(t, deliveries,
		"UserLoggedIn",
		"PasswordChanged",
		"SessionCreated",
		"SessionRevoked",
	)
}

func subscribeDomainEvents(t *testing.T) <-chan rabbit.Delivery {
	connection, err := rabbit.Dial(testEnv.conf.AMQP.URL)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, connection.Close()) })
	channel, err := connection.Channel()
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, channel.Close()) })
	queue, err := channel.QueueDeclare("", false, true, true, false, nil)
	require.NoError(t, err)
	require.NoError(t, channel.QueueBind(queue.Name, "#", "domain-event", false, nil))
	deliveries, err := channel.Consume(queue.Name, "", true, true, false, false, nil)
	require.NoError(t, err)
	return deliveries
}

func requireEventTypes(t *testing.T, deliveries <-chan rabbit.Delivery, expected ...string) {
	want := make(map[string]struct{}, len(expected))
	for _, eventType := range expected {
		want[eventType] = struct{}{}
	}
	timer := time.NewTimer(5 * time.Second)
	defer timer.Stop()
	for len(want) > 0 {
		select {
		case delivery := <-deliveries:
			delete(want, delivery.Type)
		case <-timer.C:
			require.Empty(t, want, "expected domain events were not published")
		}
	}
}
