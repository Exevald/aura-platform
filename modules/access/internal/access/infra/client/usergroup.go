package client

import (
	"context"
	"time"
	"uuid"

	usergroupapi "usergroup/api/server/internalapi"
)

type UserGroup interface {
	Members(context.Context, uuid.UUID, uuid.UUID) ([]uuid.UUID, error)
}

type userGroup struct {
	client  usergroupapi.UsergroupServiceClient
	timeout time.Duration
}

func NewUserGroup(client usergroupapi.UsergroupServiceClient, timeout time.Duration) UserGroup {
	return &userGroup{client: client, timeout: timeout}
}

func (c *userGroup) Members(ctx context.Context, groupID, ownerID uuid.UUID) ([]uuid.UUID, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	response, err := c.client.GetOwnedGroupMembers(
		ctx,
		&usergroupapi.GetOwnedGroupMembersRequest{
			GroupId:     groupID.String(),
			OwnerUserId: ownerID.String(),
		})
	if err != nil {
		return nil, err
	}

	result := make([]uuid.UUID, 0, len(response.GetUserIds()))
	for _, raw := range response.GetUserIds() {
		id, parseErr := uuid.Parse(raw)
		if parseErr != nil {
			return nil, parseErr
		}
		result = append(result, id)
	}
	return result, nil
}
