package internalapi

import (
	"context"
	"errors"
	"log/slog"
	"uuid"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	usergroupinternalapi "usergroup/api/server/internalapi"
	"usergroup/internal/usergroup/app"
	"usergroup/internal/usergroup/domain"
)

type handler struct {
	usergroupinternalapi.UnimplementedUsergroupServiceServer
	svc    *app.Service
	logger *slog.Logger
}

func NewHandler(svc *app.Service, logger *slog.Logger) usergroupinternalapi.UsergroupServiceServer {
	return &handler{svc: svc, logger: logger}
}

func (h *handler) GetOwnedGroupMembers(ctx context.Context, req *usergroupinternalapi.GetOwnedGroupMembersRequest) (*usergroupinternalapi.GetOwnedGroupMembersResponse, error) {
	groupID, groupErr := uuid.Parse(req.GetGroupId())
	ownerID, ownerErr := uuid.Parse(req.GetOwnerUserId())
	if groupErr != nil || ownerErr != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid ids")
	}
	members, err := h.svc.OwnedMembers(ctx, ownerID, groupID)
	if errors.Is(err, domain.ErrGroupNotFound) {
		return nil, status.Error(codes.NotFound, "group not found")
	}
	if err != nil {
		return nil, status.Error(codes.Internal, "get group members")
	}
	result := make([]string, 0, len(members))
	for _, id := range members {
		result = append(result, id.String())
	}
	return &usergroupinternalapi.GetOwnedGroupMembersResponse{UserIds: result}, nil
}
