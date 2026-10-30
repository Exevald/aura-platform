package integration_tests

import (
	"net/http"
	"net/http/cookiejar"
	"testing"
	"time"

	rabbit "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/require"

	"auth/api/server/publicapi"
)

func TestRegistrationLoginLogoutAndPasswordChange(t *testing.T) {
	const (
		login     = "ada"
		firstName = "Ada"
		lastName  = "Lovelace"
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
	tooShort, err := client.RegisterWithResponse(t.Context(), publicapi.RegisterJSONRequestBody{
		Login:       "short-password",
		Password:    "1234567",
		DisplayName: firstName,
		FirstName:   firstName,
		LastName:    lastName,
		Email:       "short-password@example.com",
		RememberMe:  false,
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusBadRequest, tooShort.StatusCode())
	require.NotNil(t, tooShort.JSON400)
	require.Equal(t, "validation failed", tooShort.JSON400.Message)
	require.NotNil(t, tooShort.JSON400.Fields)
	require.Contains(t, *tooShort.JSON400.Fields, "password")

	registered, err := client.RegisterWithResponse(t.Context(), publicapi.RegisterJSONRequestBody{
		Login:       login,
		Password:    password,
		DisplayName: firstName,
		FirstName:   firstName,
		LastName:    lastName,
		Email:       "ada@example.com",
		RememberMe:  false,
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, registered.StatusCode())
	require.NotNil(t, registered.JSON201)
	require.True(t, registered.JSON201.Authenticated)
	require.NotNil(t, registered.JSON201.ExpiresAt)
	require.Len(t, registered.HTTPResponse.Cookies(), 1)
	require.Equal(t, 0, registered.HTTPResponse.Cookies()[0].MaxAge)

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

	duplicate, err := client.RegisterWithResponse(t.Context(), publicapi.RegisterJSONRequestBody{
		Login:       login,
		Password:    password,
		DisplayName: "Another Ada",
		FirstName:   firstName,
		LastName:    lastName,
		Email:       "other@example.com",
		RememberMe:  false,
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusConflict, duplicate.StatusCode())
	duplicateEmail, err := client.RegisterWithResponse(t.Context(), publicapi.RegisterJSONRequestBody{
		Login:       "another-login",
		Password:    password,
		DisplayName: "Another Ada",
		FirstName:   firstName,
		LastName:    lastName,
		Email:       "ADA@EXAMPLE.COM",
		RememberMe:  false,
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusConflict, duplicateEmail.StatusCode())

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
		"UserCreated",
		"UserRegistered",
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
