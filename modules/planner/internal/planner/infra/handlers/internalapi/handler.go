package internalapi

import (
	"context"
	"log/slog"
	api "planner/api/server/internalapi"
	"planner/internal/planner/app"
	"planner/internal/planner/domain"
	"uuid"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type handler struct {
	svc    *app.Service
	logger *slog.Logger
	api.UnimplementedPlannerServiceServer
}

func NewHandler(s *app.Service, l *slog.Logger) api.PlannerServiceServer {
	return &handler{svc: s, logger: l}
}

func (h *handler) GetTask(ctx context.Context, r *api.GetTaskRequest) (*api.TaskResponse, error) {
	w, e := uuid.Parse(r.WorkspaceId)
	if e != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid workspace")
	}
	id, e := uuid.Parse(r.TaskId)
	if e != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid task")
	}
	v, e := h.svc.GetTask(ctx, w, id)
	if e == domain.ErrNotFound {
		return nil, status.Error(codes.NotFound, e.Error())
	}
	if e != nil {
		return nil, status.Error(codes.Internal, "planner failed")
	}
	return &api.TaskResponse{TaskId: v.ID.String(), ProjectId: v.ProjectID.String(), Title: v.Title, Status: v.Status, Version: v.Version}, nil
}
