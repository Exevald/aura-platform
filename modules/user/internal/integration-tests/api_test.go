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

	_, err = client.UpdateUser(ctx, &userinternalapi.UpdateUserRequest{
		UserId: created.GetUserId(),
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
		created.GetUserId(),
	)
	require.NoError(t, err)
	require.Equal(t, "Nickname", displayName)

	_, err = client.DeleteUser(ctx, &userinternalapi.DeleteUserRequest{UserId: created.GetUserId()})
	require.NoError(t, err)
	_, err = client.GetUser(ctx, &userinternalapi.GetUserRequest{UserId: created.GetUserId()})
	require.Error(t, err)
}
