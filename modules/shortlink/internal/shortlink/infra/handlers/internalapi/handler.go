package internalapi

import (
	"context"
	"log/slog"

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

func (h handler) Echo(ctx context.Context, req *shortlinkinternalapi.EchoRequest) (*shortlinkinternalapi.EchoResponse, error) {
	h.logger.DebugContext(ctx, "legacy echo endpoint called", slog.String("body", req.Body))
	return nil, status.Error(codes.Unimplemented, "echo is not supported")
}
