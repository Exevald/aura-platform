package client

import (
	"context"
	"time"
	"uuid"

	"bff/internal/bff/app"
	userinternalapi "user/api/server/internalapi"
)

type User struct {
	client  userinternalapi.UserServiceClient
	timeout time.Duration
}

func NewUser(client userinternalapi.UserServiceClient, timeout time.Duration) *User {
	return &User{client: client, timeout: timeout}
}

func (u *User) Get(ctx context.Context, id uuid.UUID) (app.Profile, error) {
	ctx, cancel := context.WithTimeout(ctx, u.timeout)
	defer cancel()
	response, err := u.client.GetUser(ctx, &userinternalapi.GetUserRequest{UserId: id.String()})
	if err != nil {
		return app.Profile{}, err
	}
	return profile(response), nil
}

func (u *User) Update(ctx context.Context, p app.Profile) error {
	ctx, cancel := context.WithTimeout(ctx, u.timeout)
	defer cancel()
	_, err := u.client.UpdateUser(ctx, &userinternalapi.UpdateUserRequest{UserId: p.UserID.String(), Fields: &userinternalapi.UserFields{
		DisplayName: p.DisplayName, FirstName: p.FirstName, LastName: p.LastName, Email: p.Email,
	}})
	return err
}

func profile(response *userinternalapi.GetUserResponse) app.Profile {
	f := response.GetFields()
	id, _ := uuid.Parse(response.GetUserId())
	return app.Profile{UserID: id, DisplayName: f.GetDisplayName(), FirstName: f.GetFirstName(), LastName: f.GetLastName(), Email: f.GetEmail()}
}
