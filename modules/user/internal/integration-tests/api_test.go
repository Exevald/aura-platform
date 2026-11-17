package integration_tests

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	userinternalapi "user/api/server/internalapi"
)

func TestUserLifecycle(t *testing.T) {
	const firstName = "Steve"
	connection, err := grpc.NewClient(
		testEnv.conf.GRPC.Address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, connection.Close()) })
	client := userinternalapi.NewUserServiceClient(connection)

	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	userID := createTestUser(ctx, t, client, firstName)

	verifyUserUpdate(ctx, t, client, userID, firstName)
	verifyUserDeletion(ctx, t, client, userID)
}

func createTestUser(ctx context.Context, t *testing.T, client userinternalapi.UserServiceClient, firstName string) string {
	t.Helper()
	created, err := client.CreateUser(ctx, &userinternalapi.CreateUserRequest{
		Fields: &userinternalapi.UserFields{
			DisplayName: firstName,
			FirstName:   firstName,
			LastName:    "Jobs",
			Email:       "steve@example.com",
		}})
	require.NoError(t, err)
	require.NotEmpty(t, created.GetUserId())

	user, err := client.GetUser(ctx, &userinternalapi.GetUserRequest{UserId: created.GetUserId()})
	require.NoError(t, err)
	require.Equal(t, firstName, user.GetFields().GetDisplayName())
	return created.GetUserId()
}

func verifyUserUpdate(ctx context.Context, t *testing.T, client userinternalapi.UserServiceClient, userID, firstName string) {
	t.Helper()
	_, err := client.UpdateUser(ctx, &userinternalapi.UpdateUserRequest{
		UserId: userID,
		Fields: &userinternalapi.UserFields{
			DisplayName: "Nickname",
			FirstName:   firstName,
			LastName:    "Jobs",
			Email:       "steve@example.com",
		},
	})
	require.NoError(t, err)

	var displayName string
	err = testEnv.connector.TransactionalClient().GetContext(
		ctx,
		"test.user",
		&displayName,
		"SELECT display_name FROM user WHERE id = UUID_TO_BIN(?)",
		userID,
	)
	require.NoError(t, err)
	require.Equal(t, "Nickname", displayName)
}

func verifyUserDeletion(ctx context.Context, t *testing.T, client userinternalapi.UserServiceClient, userID string) {
	t.Helper()
	_, err := client.DeleteUser(ctx, &userinternalapi.DeleteUserRequest{UserId: userID})
	require.NoError(t, err)
	_, err = client.GetUser(ctx, &userinternalapi.GetUserRequest{UserId: userID})
	require.Error(t, err)
}
