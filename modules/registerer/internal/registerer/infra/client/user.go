package client

import (
	"context"
	"time"
	"uuid"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"registerer/internal/registerer/app"
	userinternalapi "user/api/server/internalapi"
)

type userClient struct {
	client  userinternalapi.UserServiceClient
	timeout time.Duration
}

func NewUser(client userinternalapi.UserServiceClient, timeout time.Duration) app.UserAdapter {
	return &userClient{
		client:  client,
		timeout: timeout,
	}
}

func (c *userClient) PrepareRegistration(
	ctx context.Context,
	transactionID,
	userID uuid.UUID,
	spec app.ProfileSpec,
) error {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	_, err := c.client.PrepareUserRegistration(ctx, &userinternalapi.PrepareUserRegistrationRequest{
		TransactionId: transactionID.String(), UserId: userID.String(),
		Fields: &userinternalapi.UserFields{
			DisplayName: spec.DisplayName,
			FirstName:   spec.FirstName,
			LastName:    spec.LastName,
			Email:       spec.Email,
		},
	})
	if status.Code(err) == codes.AlreadyExists {
		return app.ErrEmailTaken
	}
	return err
}

func (c *userClient) CommitRegistration(ctx context.Context, transactionID uuid.UUID) error {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	_, err := c.client.CommitUserRegistration(
		ctx,
		&userinternalapi.RegistrationTransactionRequest{
			TransactionId: transactionID.String(),
		})
	return err
}

func (c *userClient) RollbackRegistration(ctx context.Context, transactionID uuid.UUID) error {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	_, err := c.client.RollbackUserRegistration(
		ctx,
		&userinternalapi.RegistrationTransactionRequest{
			TransactionId: transactionID.String(),
		})
	return err
}
