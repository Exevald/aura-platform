package internalapi

import (
	"context"
	"errors"
	"log/slog"
	"shortlink/internal/shortlink/domain"
	"uuid"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	shortlinkinternalapi "shortlink/api/server/internalapi"
	"shortlink/internal/shortlink/app"
)

func NewHandler(svc *app.Service, logger *slog.Logger) shortlinkinternalapi.ShortlinkServiceServer {
	return &handler{
		svc:    svc,
		logger: logger,
	}
}

type handler struct {
	svc    *app.Service
	logger *slog.Logger
	shortlinkinternalapi.UnimplementedShortlinkServiceServer
}

func (h handler) EnsureLink(ctx context.Context, req *shortlinkinternalapi.EnsureLinkRequest) (*shortlinkinternalapi.EnsureLinkResponse, error) {
	owner, err := uuid.Parse(req.OwnerUserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid owner")
	}
	link, err := h.svc.Ensure(ctx, owner, req.IdempotencyKey, req.TargetUrl)
	if errors.Is(err, domain.ErrInvalidTarget) {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if err != nil {
		return nil, status.Error(codes.Internal, "ensure link failed")
	}

	return &shortlinkinternalapi.EnsureLinkResponse{
		LinkId:    link.ID.String(),
		ShortPath: "/s/" + link.Code,
	}, nil
}

func (h handler) DeleteOwnedLink(ctx context.Context, req *shortlinkinternalapi.DeleteOwnedLinkRequest) (*shortlinkinternalapi.DeleteOwnedLinkResponse, error) {
	owner, err := uuid.Parse(req.OwnerUserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid owner")
	}
	var id uuid.UUID
	if req.IdempotencyKey != "" {
		id = domain.LinkIDForKey(owner, req.IdempotencyKey)
	} else {
		id, err = uuid.Parse(req.LinkId)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid link")
		}
	}
	err = h.svc.Delete(ctx, owner, id)
	if err != nil && !errors.Is(err, domain.ErrNotFound) {
		return nil, status.Error(codes.Internal, "delete link failed")
	}
	return &shortlinkinternalapi.DeleteOwnedLinkResponse{}, nil
}
