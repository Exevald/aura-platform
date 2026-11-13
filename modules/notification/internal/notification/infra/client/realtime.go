package client

import (
	"context"
	"notification/internal/notification/app"
	"notification/internal/notification/domain"
	api "realtime/api/server/internalapi"
	"time"
)

type realtime struct {
	client  api.RealtimeServiceClient
	timeout time.Duration
}

func NewRealtime(c api.RealtimeServiceClient, timeout time.Duration) app.RealtimePublisher {
	return &realtime{client: c, timeout: timeout}
}

func (c *realtime) Publish(ctx context.Context, n domain.Notification, payload []byte) error {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	_, err := c.client.Publish(ctx, &api.PublishRequest{UserId: n.UserID.String(), WorkspaceId: n.WorkspaceID.String(), Type: "notification.created", PayloadJson: payload})
	return err
}
