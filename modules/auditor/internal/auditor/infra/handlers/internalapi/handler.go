package internalapi

import (
	"context"
	"log/slog"

	"github.com/distributed-programming-2026/go-sdk/pkg/logging"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	auditorinternalapi "auditor/api/server/internalapi"
	"auditor/internal/auditor/app"
)

func NewHandler(svc *app.Service, logger *slog.Logger) auditorinternalapi.AuditorServiceServer {
	return &handler{
		svc:    svc,
		logger: logger,
	}
}

type handler struct {
	svc    *app.Service
	logger *slog.Logger
	auditorinternalapi.UnimplementedAuditorServiceServer
}

func (h handler) Echo(ctx context.Context, req *auditorinternalapi.EchoRequest) (*auditorinternalapi.EchoResponse, error) {
	id, err := h.svc.Echo(ctx, req.Body)
	if err != nil {
		h.logger.ErrorContext(ctx, "echo failed", logging.Error(err))
		return nil, status.Error(codes.Internal, "echo failed")
	}

	return &auditorinternalapi.EchoResponse{
		OperationID: id.String(),
	}, nil
}
