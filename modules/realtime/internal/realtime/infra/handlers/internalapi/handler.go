package internalapi

import (
	"context"
	"encoding/json"
	"log/slog"
	"math"
	api "realtime/api/server/internalapi"
	"realtime/internal/realtime/app"
	"uuid"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type handler struct {
	svc    *app.Service
	logger *slog.Logger
	api.UnimplementedRealtimeServiceServer
}

func NewHandler(s *app.Service, l *slog.Logger) api.RealtimeServiceServer {
	return &handler{svc: s, logger: l}
}

func (h *handler) Publish(_ context.Context, r *api.PublishRequest) (*api.PublishResponse, error) {
	user, workspace := uuid.Nil(), uuid.Nil()
	var e error
	if r.UserId != "" {
		user, e = uuid.Parse(r.UserId)
		if e != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid user")
		}
	}
	if r.WorkspaceId != "" {
		workspace, e = uuid.Parse(r.WorkspaceId)
		if e != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid workspace")
		}
	}
	if r.Type == "" || !json.Valid(r.PayloadJson) {
		return nil, status.Error(codes.InvalidArgument, "invalid event")
	}
	n := h.svc.Publish(user, workspace, r.Type, r.PayloadJson)
	if n > math.MaxInt32 {
		return nil, status.Error(codes.Internal, "subscriber count overflow")
	}
	// #nosec G115 -- the subscriber count is checked against MaxInt32 above.
	return &api.PublishResponse{Delivered: int32(n)}, nil
}
