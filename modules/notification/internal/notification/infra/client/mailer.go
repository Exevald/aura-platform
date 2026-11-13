package client

import (
	"context"
	api "mailer/api/server/internalapi"
	"notification/internal/notification/app"
	"notification/internal/notification/domain"
	"time"
)

type mailer struct {
	client  api.MailerServiceClient
	timeout time.Duration
}

func NewMailer(c api.MailerServiceClient, timeout time.Duration) app.MailerSender {
	return &mailer{client: c, timeout: timeout}
}

func (c *mailer) Send(ctx context.Context, n domain.Notification) error {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	_, err := c.client.SendMail(ctx, &api.SendMailRequest{IdempotencyKey: "notification:" + n.ID.String(), Recipient: n.RecipientEmail, Subject: n.Title, Body: n.Body})
	return err
}
