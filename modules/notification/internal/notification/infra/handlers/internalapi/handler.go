package internalapi

import (
	"context"
	"log/slog"
	api "notification/api/server/internalapi"
	"notification/internal/notification/app"
	"notification/internal/notification/domain"
	"uuid"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type handler struct {
	svc    *app.Service
	logger *slog.Logger
	api.UnimplementedNotificationServiceServer
}

func NewHandler(s *app.Service, l *slog.Logger) api.NotificationServiceServer {
	return &handler{svc: s, logger: l}
}

func (h *handler) CreateNotification(ctx context.Context, r *api.CreateNotificationRequest) (*api.CreateNotificationResponse, error) {
	u, e := uuid.Parse(r.UserId)
	if e != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user id")
	}
	w := uuid.Nil()
	if r.WorkspaceId != "" {
		w, e = uuid.Parse(r.WorkspaceId)
		if e != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid workspace id")
		}
	}
	v, created, e := h.svc.Create(ctx, domain.Notification{UserID: u, WorkspaceID: w, IdempotencyKey: r.IdempotencyKey, Kind: r.Kind, Title: r.Title, Body: r.Body, EmailRequested: r.EmailRequested, RecipientEmail: r.RecipientEmail})
	if e == domain.ErrInvalid {
		return nil, status.Error(codes.InvalidArgument, e.Error())
	}
	if e != nil {
		return nil, status.Error(codes.Internal, "create notification failed")
	}
	return &api.CreateNotificationResponse{NotificationId: v.ID.String(), Created: created}, nil
}
