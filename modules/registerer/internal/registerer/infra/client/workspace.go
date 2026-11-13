package client

import (
	"context"
	"registerer/internal/registerer/app"
	"time"
	"uuid"
	api "workspace/api/server/internalapi"
)

type workspaceClient struct {
	client  api.WorkspaceServiceClient
	timeout time.Duration
}

func NewWorkspace(c api.WorkspaceServiceClient, t time.Duration) app.WorkspaceAdapter {
	return &workspaceClient{c, t}
}

func (c *workspaceClient) EnsurePersonal(ctx context.Context, id, user uuid.UUID, name string) error {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	_, err := c.client.EnsurePersonalWorkspace(ctx, &api.EnsurePersonalWorkspaceRequest{WorkspaceId: id.String(), UserId: user.String(), Name: name})
	return err
}
