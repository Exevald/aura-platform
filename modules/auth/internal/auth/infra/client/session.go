package client

import (
	"context"
	"time"
	"uuid"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"auth/internal/auth/app"
	sessioninternalapi "session/api/server/internalapi"
)

type sessionClient struct {
	client  sessioninternalapi.SessionServiceClient
	timeout time.Duration
}

func NewSession(client sessioninternalapi.SessionServiceClient, timeout time.Duration) app.SessionClient {
	return &sessionClient{client: client, timeout: timeout}
}

func (c *sessionClient) CreateSession(ctx context.Context, userID uuid.UUID, rememberMe bool) (uuid.UUID, time.Time, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	response, err := c.client.CreateSession(ctx, &sessioninternalapi.CreateSessionRequest{
		UserId:     userID.String(),
		RememberMe: rememberMe,
	})
	if err != nil {
		return uuid.Nil(), time.Time{}, mapSessionError(err)
	}
	id, err := uuid.Parse(response.GetSessionId())
	if err != nil {
		return uuid.Nil(), time.Time{}, err
	}
	return id, response.GetExpiresAt().AsTime(), nil
}

func (c *sessionClient) ValidateSession(ctx context.Context, sessionID uuid.UUID) (uuid.UUID, time.Time, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	response, err := c.client.ValidateSession(ctx, &sessioninternalapi.ValidateSessionRequest{SessionId: sessionID.String()})
	if err != nil {
		return uuid.Nil(), time.Time{}, mapSessionError(err)
	}
	userID, err := uuid.Parse(response.GetUserId())
	if err != nil {
		return uuid.Nil(), time.Time{}, err
	}
	return userID, response.GetExpiresAt().AsTime(), nil
}

func (c *sessionClient) RevokeSession(ctx context.Context, sessionID uuid.UUID) error {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	_, err := c.client.RevokeSession(ctx, &sessioninternalapi.RevokeSessionRequest{SessionId: sessionID.String()})
	return mapSessionError(err)
}

func (c *sessionClient) RevokeUserSessions(ctx context.Context, userID uuid.UUID) error {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	_, err := c.client.RevokeUserSessions(ctx, &sessioninternalapi.RevokeUserSessionsRequest{UserId: userID.String()})
	return mapSessionError(err)
}

func mapSessionError(err error) error {
	if err == nil {
		return nil
	}
	switch status.Code(err) {
	case codes.NotFound, codes.Unauthenticated:
		return app.ErrSessionNotFound
	default:
		return err
	}
}
