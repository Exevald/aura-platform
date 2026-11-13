package client

import (
	api "access/api/server/internalapi"
	"context"
	"registerer/internal/registerer/app"
	"time"
	"uuid"
)

type accessClient struct {
	client   api.ModuleAccessServiceClient
	timeout  time.Duration
	moduleID uuid.UUID
}

func NewAccess(c api.ModuleAccessServiceClient, t time.Duration, m uuid.UUID) app.AccessAdapter {
	return &accessClient{c, t, m}
}

func (c *accessClient) PrepareRegistration(ctx context.Context, tx, user uuid.UUID) error {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	_, e := c.client.PrepareModuleAccessRegistration(ctx, &api.PrepareModuleAccessRegistrationRequest{TransactionId: tx.String(), UserId: user.String(), ModuleId: c.moduleID.String()})
	return e
}

func (c *accessClient) CommitRegistration(ctx context.Context, tx uuid.UUID) error {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	_, e := c.client.CommitModuleAccessRegistration(ctx, &api.RegistrationTransactionRequest{TransactionId: tx.String()})
	return e
}

func (c *accessClient) RollbackRegistration(ctx context.Context, tx uuid.UUID) error {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	_, e := c.client.RollbackModuleAccessRegistration(ctx, &api.RegistrationTransactionRequest{TransactionId: tx.String()})
	return e
}
