package adapter

import (
	"context"
	"uuid"

	api "notification/api/server/internalapi"
	port "publication/internal/publication/app/adapter"
)

func NewNotificationService(client api.NotificationServiceClient) port.NotificationService {
	return notificationAdapter{client}
}

type notificationAdapter struct {
	client api.NotificationServiceClient
}

func (s notificationAdapter) Create(ctx context.Context, key string, user, workspace uuid.UUID, title, body string) error {
	_, err := s.client.CreateNotification(ctx, &api.CreateNotificationRequest{
		IdempotencyKey: key,
		UserId:         user.String(),
		WorkspaceId:    workspace.String(),
		Kind:           "publication",
		Title:          title,
		Body:           body,
	})
	return err
}
