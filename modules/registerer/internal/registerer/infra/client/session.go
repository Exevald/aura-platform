package client

import (
	"context"
	"time"
	"uuid"

	"registerer/internal/registerer/app"
	sessioninternalapi "session/api/server/internalapi"
)

type sessionClient struct {
	client  sessioninternalapi.SessionServiceClient
	timeout time.Duration
}

func NewSession(client sessioninternalapi.SessionServiceClient, timeout time.Duration) app.SessionAdapter {
	return &sessionClient{client: client, timeout: timeout}
}

func (c *sessionClient) CreateSession(
	ctx context.Context,
	userID uuid.UUID,
	rememberMe bool,
) (uuid.UUID, time.Time, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	response, err := c.client.CreateSession(
		ctx,
		&sessioninternalapi.CreateSessionRequest{
			UserId:     userID.String(),
			RememberMe: rememberMe,
		},
	)
	if err != nil {
		return uuid.Nil(), time.Time{}, err
	}
	id, err := uuid.Parse(response.GetSessionId())
	if err != nil {
		return uuid.Nil(), time.Time{}, err
	}
	return id, response.GetExpiresAt().AsTime(), nil
}
