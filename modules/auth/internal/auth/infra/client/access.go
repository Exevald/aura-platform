package client

import (
	accessinternalapi "access/api/server/internalapi"
	"context"
	"time"
	"uuid"
)

type Access struct {
	client   accessinternalapi.AccessServiceClient
	timeout  time.Duration
	moduleID string
}

func NewAccess(client accessinternalapi.AccessServiceClient, timeout time.Duration, moduleID string) *Access {
	return &Access{client: client, timeout: timeout, moduleID: moduleID}
}
func (a *Access) Grant(ctx context.Context, user uuid.UUID) error {
	ctx, cancel := context.WithTimeout(ctx, a.timeout)
	defer cancel()
	_, err := a.client.GrantModuleAccess(ctx, &accessinternalapi.GrantModuleAccessRequest{UserId: user.String(), ModuleId: a.moduleID})
	return err
}
func (a *Access) Revoke(ctx context.Context, user uuid.UUID) error {
	ctx, cancel := context.WithTimeout(ctx, a.timeout)
	defer cancel()
	_, err := a.client.RevokeModuleAccess(ctx, &accessinternalapi.RevokeModuleAccessRequest{UserId: user.String(), ModuleId: a.moduleID})
	return err
}
