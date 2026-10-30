package client

import (
	"context"
	"time"
	"uuid"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"auth/internal/auth/app"
	userinternalapi "user/api/server/internalapi"
)

func NewUser(client userinternalapi.UserServiceClient, timeout time.Duration) app.UserClient {
	return &userClient{client: client, timeout: timeout}
}

type userClient struct {
	client  userinternalapi.UserServiceClient
	timeout time.Duration
}

func (c *userClient) CreateUser(ctx context.Context, spec app.ProfileSpec) (uuid.UUID, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	response, err := c.client.CreateUser(ctx, &userinternalapi.CreateUserRequest{
		Fields: &userinternalapi.UserFields{
			DisplayName: spec.DisplayName,
			FirstName:   spec.FirstName,
			LastName:    spec.LastName,
			Email:       spec.Email,
		},
	})
	if err != nil {
		if status.Code(err) == codes.AlreadyExists {
			return uuid.Nil(), app.ErrEmailTaken
		}
		return uuid.Nil(), err
	}
	return uuid.Parse(response.GetUserId())
}

func (c *userClient) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	_, err := c.client.DeleteUser(ctx, &userinternalapi.DeleteUserRequest{UserId: userID.String()})
	return err
}
