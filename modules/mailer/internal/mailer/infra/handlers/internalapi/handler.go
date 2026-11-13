package internalapi

import (
	"context"
	"log/slog"
	api "mailer/api/server/internalapi"
	"mailer/internal/mailer/app"
	"mailer/internal/mailer/domain"
	"math"
	"uuid"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type handler struct {
	svc    *app.Service
	logger *slog.Logger
	api.UnimplementedMailerServiceServer
}

func NewHandler(s *app.Service, l *slog.Logger) api.MailerServiceServer {
	return &handler{svc: s, logger: l}
}

func (h *handler) SendMail(ctx context.Context, r *api.SendMailRequest) (*api.SendMailResponse, error) {
	v, created, e := h.svc.Send(ctx, r.IdempotencyKey, r.Recipient, r.Subject, r.Body)
	if e == domain.ErrInvalid {
		return nil, status.Error(codes.InvalidArgument, e.Error())
	}
	if e != nil {
		return nil, status.Error(codes.Unavailable, "mail delivery failed")
	}
	return &api.SendMailResponse{MailId: v.ID.String(), Status: v.Status, Created: created}, nil
}

func (h *handler) GetMail(ctx context.Context, r *api.GetMailRequest) (*api.MailResponse, error) {
	id, e := uuid.Parse(r.MailId)
	if e != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid mail id")
	}
	v, e := h.svc.Get(ctx, id)
	if e == domain.ErrNotFound {
		return nil, status.Error(codes.NotFound, e.Error())
	}
	if e != nil {
		return nil, status.Error(codes.Internal, "mailer failed")
	}
	if v.Attempts < 0 || v.Attempts > math.MaxInt32 {
		return nil, status.Error(codes.Internal, "invalid mail attempt count")
	}

	return &api.MailResponse{
		MailId:    v.ID.String(),
		Recipient: v.Recipient,
		Subject:   v.Subject,
		Status:    v.Status,
		Attempts:  int32(v.Attempts),
		LastError: v.LastError,
	}, nil
}
