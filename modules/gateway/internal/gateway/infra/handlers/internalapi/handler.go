package internalapi

import (
	"context"
	"log/slog"

	"github.com/distributed-programming-2026/go-sdk/pkg/logging"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	gatewayinternalapi "gateway/api/server/internalapi"
	"gateway/internal/gateway/app"
)

func NewHandler(svc *app.Service, logger *slog.Logger) gatewayinternalapi.GatewayServiceServer {
	return &handler{
		svc:    svc,
		logger: logger,
	}
}

type handler struct {
	svc    *app.Service
	logger *slog.Logger
	gatewayinternalapi.UnimplementedGatewayServiceServer
}

func (h handler) Echo(ctx context.Context, req *gatewayinternalapi.EchoRequest) (*gatewayinternalapi.EchoResponse, error) {
	id, err := h.svc.Echo(ctx, req.Body)
	if err != nil {
		h.logger.ErrorContext(ctx, "echo failed", logging.Error(err))
		return nil, status.Error(codes.Internal, "echo failed")
	}

	return &gatewayinternalapi.EchoResponse{
		OperationID: id.String(),
	}, nil
}
