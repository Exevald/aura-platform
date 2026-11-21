package adapter

import (
	"context"
	"fmt"
	port "publication/internal/publication/app/adapter"
	"uuid"

	api "workspace/api/server/internalapi"
)

func NewWorkspaceService(client api.WorkspaceServiceClient) port.WorkspaceService {
	return workspaceAdapter{client}
}

type workspaceAdapter struct {
	client api.WorkspaceServiceClient
}

func (s workspaceAdapter) Members(ctx context.Context, id, owner uuid.UUID) ([]uuid.UUID, error) {
	r, err := s.client.ListMembers(ctx, &api.ListMembersRequest{WorkspaceId: id.String(), ActorUserId: owner.String()})
	if err != nil {
		return nil, err
	}
	members := make([]uuid.UUID, 0, len(r.UserIds))
	for _, value := range r.UserIds {
		member, err2 := uuid.Parse(value)
		if err2 != nil {
			return nil, fmt.Errorf("invalid workspace member ID %q: %w", value, err2)
		}
		members = append(members, member)
	}
	return members, nil
}
