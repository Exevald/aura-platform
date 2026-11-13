package client

import (
	"context"
	"time"
	"uuid"

	userinternalapi "user/api/server/internalapi"
	"usergroup/internal/usergroup/app"
)

type userClient struct {
	client  userinternalapi.UserServiceClient
	timeout time.Duration
}

func NewUser(client userinternalapi.UserServiceClient, timeout time.Duration) app.UserClient {
	return &userClient{client: client, timeout: timeout}
}

func (c *userClient) FindByEmail(ctx context.Context, email string) (app.User, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	value, err := c.client.FindUserByEmail(ctx, &userinternalapi.FindUserByEmailRequest{Email: email})
	if err != nil {
		return app.User{}, err
	}
	return convert(value)
}

func (c *userClient) Get(ctx context.Context, id uuid.UUID) (app.User, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	value, err := c.client.GetUser(ctx, &userinternalapi.GetUserRequest{UserId: id.String()})
	if err != nil {
		return app.User{}, err
	}

	return convert(value)
}

func convert(value *userinternalapi.GetUserResponse) (app.User, error) {
	id, err := uuid.Parse(value.GetUserId())
	if err != nil {
		return app.User{}, err
	}

	return app.User{
		ID:          id,
		Email:       value.GetFields().GetEmail(),
		DisplayName: value.GetFields().GetDisplayName(),
	}, nil
}
