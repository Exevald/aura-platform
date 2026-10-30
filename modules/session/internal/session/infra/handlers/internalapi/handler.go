package internalapi

import (
	"context"
	"errors"
	"uuid"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	sessioninternalapi "session/api/server/internalapi"
	"session/internal/session/app"
	"session/internal/session/domain"
)

func NewHandler(svc app.SessionService) sessioninternalapi.SessionServiceServer {
	return &handler{svc: svc}
}

type handler struct {
	svc app.SessionService
	sessioninternalapi.UnimplementedSessionServiceServer
}

func (h *handler) CreateSession(
	ctx context.Context,
	req *sessioninternalapi.CreateSessionRequest,
) (*sessioninternalapi.CreateSessionResponse, error) {
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user id")
	}
	id, expiresAt, err := h.svc.Create(ctx, userID, req.GetRememberMe())
	if err != nil {
		return nil, mapError(err)
	}

	return &sessioninternalapi.CreateSessionResponse{
		SessionId: id.String(),
		ExpiresAt: timestamppb.New(expiresAt),
	}, nil
}

func (h *handler) ValidateSession(
	ctx context.Context,
	req *sessioninternalapi.ValidateSessionRequest,
) (*sessioninternalapi.ValidateSessionResponse, error) {
	id, err := uuid.Parse(req.GetSessionId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid session id")
	}
	session, err := h.svc.Validate(ctx, id)
	if err != nil {
		return nil, mapError(err)
	}

	return &sessioninternalapi.ValidateSessionResponse{
		UserId:     session.UserID.String(),
		ExpiresAt:  timestamppb.New(session.ExpiresAt),
		Persistent: session.Persistent,
	}, nil
}

func (h *handler) RevokeSession(
	ctx context.Context,
	req *sessioninternalapi.RevokeSessionRequest,
) (*sessioninternalapi.RevokeSessionResponse, error) {
	id, err := uuid.Parse(req.GetSessionId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid session id")
	}
	if err = h.svc.Revoke(ctx, id); err != nil {
		return nil, mapError(err)
	}

	return &sessioninternalapi.RevokeSessionResponse{}, nil
}

func (h *handler) RevokeUserSessions(
	ctx context.Context,
	req *sessioninternalapi.RevokeUserSessionsRequest,
) (*sessioninternalapi.RevokeUserSessionsResponse, error) {
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user id")
	}
	if err = h.svc.RevokeAll(ctx, userID); err != nil {
		return nil, mapError(err)
	}

	return &sessioninternalapi.RevokeUserSessionsResponse{}, nil
}

func mapError(err error) error {
	switch {
	case errors.Is(err, domain.ErrSessionNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, domain.ErrInvalidSession):
		return status.Error(codes.Unauthenticated, err.Error())
	default:
		return status.Error(codes.Internal, "session operation failed")
	}
}
