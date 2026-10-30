package internalapi

import (
	"context"
	"uuid"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	accessinternalapi "access/api/server/internalapi"
	"access/internal/access/app"
)

type Handler struct {
	accessinternalapi.UnimplementedAccessServiceServer
	svc app.ModuleAccessService
}

func New(svc app.ModuleAccessService) *Handler { return &Handler{svc: svc} }
func ids(user, module string) (userID, moduleID uuid.UUID, err error) {
	u, e := uuid.Parse(user)
	if e != nil {
		return uuid.Nil(), uuid.Nil(), e
	}
	m, e := uuid.Parse(module)
	return u, m, e
}
func (h *Handler) GrantModuleAccess(ctx context.Context, r *accessinternalapi.GrantModuleAccessRequest) (*accessinternalapi.GrantModuleAccessResponse, error) {
	u, m, e := ids(r.GetUserId(), r.GetModuleId())
	if e != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid ids")
	}
	if e = h.svc.GrantModuleAccess(ctx, u, m); e != nil {
		return nil, status.Error(codes.Internal, "grant failed")
	}
	return &accessinternalapi.GrantModuleAccessResponse{}, nil
}
func (h *Handler) RevokeModuleAccess(ctx context.Context, r *accessinternalapi.RevokeModuleAccessRequest) (*accessinternalapi.RevokeModuleAccessResponse, error) {
	u, m, e := ids(r.GetUserId(), r.GetModuleId())
	if e != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid ids")
	}
	if e = h.svc.RevokeModuleAccess(ctx, u, m); e != nil {
		return nil, status.Error(codes.Internal, "revoke failed")
	}
	return &accessinternalapi.RevokeModuleAccessResponse{}, nil
}
func (h *Handler) CheckModuleAccess(ctx context.Context, r *accessinternalapi.CheckModuleAccessRequest) (*accessinternalapi.CheckModuleAccessResponse, error) {
	u, m, e := ids(r.GetUserId(), r.GetModuleId())
	if e != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid ids")
	}
	allowed, e := h.svc.CheckModuleAccess(ctx, u, m)
	if e != nil {
		return nil, status.Error(codes.Internal, "check failed")
	}
	return &accessinternalapi.CheckModuleAccessResponse{Allowed: allowed}, nil
}
