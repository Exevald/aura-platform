package client

import (
	"context"
	"time"
	"uuid"

	"github.com/pkg/errors"

	"editor/internal/editor/app"
	userinternalapi "user/api/server/internalapi"
)

type User struct {
	client  userinternalapi.UserServiceClient
	timeout time.Duration
}

func NewUser(client userinternalapi.UserServiceClient, timeout time.Duration) *User {
	return &User{client: client, timeout: timeout}
}

func (u *User) FindByEmail(ctx context.Context, email string) (app.User, error) {
	ctx, cancel := context.WithTimeout(ctx, u.timeout)
	defer cancel()
	user, err := u.client.FindUserByEmail(ctx, &userinternalapi.FindUserByEmailRequest{Email: email})
	if err != nil {
		return app.User{}, err
	}

	return convertUserModel(user)
}

func (u *User) Get(ctx context.Context, id uuid.UUID) (app.User, error) {
	ctx, cancel := context.WithTimeout(ctx, u.timeout)
	defer cancel()
	user, err := u.client.GetUser(ctx, &userinternalapi.GetUserRequest{UserId: id.String()})
	if err != nil {
		return app.User{}, err
	}

	return convertUserModel(user)
}

func convertUserModel(v *userinternalapi.GetUserResponse) (app.User, error) {
	id, err := uuid.Parse(v.GetUserId())
	return app.User{
		ID:          id,
		Email:       v.GetFields().GetEmail(),
		DisplayName: v.GetFields().GetDisplayName(),
	}, errors.WithStack(err)
}
