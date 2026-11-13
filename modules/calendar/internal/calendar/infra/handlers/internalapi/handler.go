package internalapi

import (
	api "calendar/api/server/internalapi"
	"calendar/internal/calendar/app"
	"calendar/internal/calendar/domain"
	"context"
	"log/slog"
	"time"
	"uuid"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type handler struct {
	svc    *app.Service
	logger *slog.Logger
	api.UnimplementedCalendarServiceServer
}

func NewHandler(s *app.Service, l *slog.Logger) api.CalendarServiceServer {
	return &handler{svc: s, logger: l}
}

func (h *handler) UpsertTaskEvent(ctx context.Context, r *api.UpsertTaskEventRequest) (*api.TaskEventResponse, error) {
	w, e := uuid.Parse(r.WorkspaceId)
	if e != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid workspace")
	}
	a, e := uuid.Parse(r.ActorUserId)
	if e != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid actor")
	}
	c, e := uuid.Parse(r.CalendarId)
	if e != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid calendar")
	}
	if r.StartsAt == nil {
		return nil, status.Error(codes.InvalidArgument, "starts_at is required")
	}
	id := uuid.New()
	if r.EventId != "" {
		id, e = uuid.Parse(r.EventId)
		if e != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid event")
		}
	}
	start := r.StartsAt.AsTime()
	end := start.Add(time.Hour)
	if r.EndsAt != nil {
		end = r.EndsAt.AsTime()
	}
	v, e := h.svc.UpsertTaskEvent(ctx, w, a, id, c, domain.Event{Title: r.Title, Description: r.Description, StartsAt: start, EndsAt: end, Timezone: r.Timezone})
	if e == domain.ErrInvalid {
		return nil, status.Error(codes.InvalidArgument, e.Error())
	}
	if e == domain.ErrNotFound {
		return nil, status.Error(codes.NotFound, e.Error())
	}
	if e != nil {
		return nil, status.Error(codes.Internal, "calendar upsert failed")
	}
	return &api.TaskEventResponse{EventId: v.ID.String(), Version: v.Version}, nil
}

func (h *handler) DeleteTaskEvent(ctx context.Context, r *api.DeleteTaskEventRequest) (*api.DeleteTaskEventResponse, error) {
	w, e := uuid.Parse(r.WorkspaceId)
	if e != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid workspace")
	}
	a, e := uuid.Parse(r.ActorUserId)
	if e != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid actor")
	}
	id, e := uuid.Parse(r.EventId)
	if e != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid event")
	}
	e = h.svc.DeleteEvent(ctx, w, a, id)
	if e == domain.ErrNotFound {
		return &api.DeleteTaskEventResponse{}, nil
	}
	if e != nil {
		return nil, status.Error(codes.Internal, "calendar delete failed")
	}
	return &api.DeleteTaskEventResponse{}, nil
}
