package client

import (
	"context"
	"time"
	"uuid"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	authinternalapi "auth/api/server/internalapi"
	"registerer/internal/registerer/app"
)

type authClient struct {
	client  authinternalapi.AuthServiceClient
	timeout time.Duration
}

func NewAuth(client authinternalapi.AuthServiceClient, timeout time.Duration) app.AuthAdapter {
	return &authClient{client: client, timeout: timeout}
}

func (c *authClient) PrepareRegistration(ctx context.Context, transactionID, userID uuid.UUID, login, password string) (uuid.UUID, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	response, err := c.client.PrepareCredentialRegistration(ctx, &authinternalapi.PrepareCredentialRegistrationRequest{
		TransactionId: transactionID.String(), UserId: userID.String(), Login: login, Password: password,
	})
	if err != nil {
		if status.Code(err) == codes.AlreadyExists {
			return uuid.Nil(), app.ErrLoginTaken
		}
		if status.Code(err) == codes.InvalidArgument {
			return uuid.Nil(), app.ErrInvalidRegistration
		}
		return uuid.Nil(), err
	}
	return uuid.Parse(response.GetCredentialId())
}

func (c *authClient) CommitRegistration(ctx context.Context, transactionID uuid.UUID) error {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	_, err := c.client.CommitCredentialRegistration(ctx, &authinternalapi.RegistrationTransactionRequest{TransactionId: transactionID.String()})
	return err
}

func (c *authClient) RollbackRegistration(ctx context.Context, transactionID uuid.UUID) error {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	_, err := c.client.RollbackCredentialRegistration(ctx, &authinternalapi.RegistrationTransactionRequest{TransactionId: transactionID.String()})
	return err
}
