package internalapi

import (
	"context"
	"log/slog"
	"uuid"
	api "workspace/api/server/internalapi"
	"workspace/internal/workspace/app"
	"workspace/internal/workspace/domain"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type handler struct {
	svc    *app.Service
	logger *slog.Logger
	api.UnimplementedWorkspaceServiceServer
}

func NewHandler(s *app.Service, l *slog.Logger) api.WorkspaceServiceServer {
	return &handler{svc: s, logger: l}
}

func (h *handler) EnsurePersonalWorkspace(ctx context.Context, r *api.EnsurePersonalWorkspaceRequest) (*api.WorkspaceResponse, error) {
	id, e := uuid.Parse(r.WorkspaceId)
	if e != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid workspace id")
	}
	u, e := uuid.Parse(r.UserId)
	if e != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user id")
	}
	w, e := h.svc.EnsurePersonal(ctx, id, u, r.Name)
	if e != nil {
		return nil, status.Error(codes.Internal, "ensure personal workspace failed")
	}
	return &api.WorkspaceResponse{WorkspaceId: w.ID.String(), Kind: string(w.Kind), Name: w.Name}, nil
}

func (h *handler) CheckMembership(ctx context.Context, r *api.CheckMembershipRequest) (*api.CheckMembershipResponse, error) {
	w, e := uuid.Parse(r.WorkspaceId)
	if e != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid workspace id")
	}
	u, e := uuid.Parse(r.UserId)
	if e != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user id")
	}
	m, value, e := h.svc.Check(ctx, u, w)
	if e == domain.ErrNotFound {
		return &api.CheckMembershipResponse{Allowed: false}, nil
	}
	if e != nil {
		return nil, status.Error(codes.Internal, "membership check failed")
	}
	return &api.CheckMembershipResponse{Allowed: true, Role: string(m.Role), Kind: string(value.Kind)}, nil
}
