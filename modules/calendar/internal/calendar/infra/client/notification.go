package client

import (
	"calendar/internal/calendar/app"
	"calendar/internal/calendar/domain"
	"context"
	api "notification/api/server/internalapi"
	"strconv"
	"time"
)

type notification struct {
	client  api.NotificationServiceClient
	timeout time.Duration
}

func NewNotification(c api.NotificationServiceClient, timeout time.Duration) app.NotificationSender {
	return &notification{client: c, timeout: timeout}
}

func (c *notification) SendReminder(ctx context.Context, r domain.ReminderDelivery) error {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	_, err := c.client.CreateNotification(ctx, &api.CreateNotificationRequest{
		IdempotencyKey: "calendar:" + r.ID.String() + ":" + strconv.FormatInt(r.DeliverAt.Unix(), 10),
		UserId:         r.UserID.String(), WorkspaceId: r.WorkspaceID.String(), Kind: "calendar.reminder",
		Title: r.Title, Body: r.Body, EmailRequested: r.Email, RecipientEmail: r.RecipientEmail,
	})
	return err
}
