package internalapi

import (
	api "appreg/api/server/internalapi"
	"appreg/internal/appreg/app"
	"context"
	"log/slog"
	"uuid"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type handler struct {
	svc    *app.Service
	logger *slog.Logger
	api.UnimplementedAppregServiceServer
}

func NewHandler(s *app.Service, l *slog.Logger) api.AppregServiceServer {
	return &handler{svc: s, logger: l}
}

func (h *handler) ListApps(_ context.Context, r *api.ListAppsRequest) (*api.ListAppsResponse, error) {
	values := h.svc.List(r.DefaultEnabledOnly)
	out := &api.ListAppsResponse{Apps: make([]*api.App, 0, len(values))}
	for _, v := range values {
		out.Apps = append(out.Apps, toProto(v))
	}
	return out, nil
}

func (h *handler) GetApp(_ context.Context, r *api.GetAppRequest) (*api.App, error) {
	id, err := uuid.Parse(r.AppId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid app id")
	}

	v, err := h.svc.Get(id)
	if err == app.ErrNotFound {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	if err != nil {
		return nil, status.Error(codes.Internal, "app registry failed")
	}

	return toProto(v), nil
}

func toProto(v app.App) *api.App {
	return &api.App{
		Id:             v.ID.String(),
		Code:           v.Code,
		Name:           v.Name,
		FrontendRoute:  v.FrontendRoute,
		BackendTarget:  v.BackendTarget,
		DefaultEnabled: v.DefaultEnabled,
	}
}
