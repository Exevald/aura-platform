package internalapi

import (
	"context"
	"log/slog"

	"github.com/distributed-programming-2026/go-sdk/pkg/logging"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	userinternalapi "user/api/server/internalapi"
	"user/internal/user/app"
)

func NewHandler(svc *app.Service, logger *slog.Logger) userinternalapi.UserServiceServer {
	return &handler{
		svc:    svc,
		logger: logger,
	}
}

type handler struct {
	svc    *app.Service
	logger *slog.Logger
	userinternalapi.UnimplementedUserServiceServer
}

func (h handler) Echo(ctx context.Context, req *userinternalapi.EchoRequest) (*userinternalapi.EchoResponse, error) {
	id, err := h.svc.Echo(ctx, req.Body)
	if err != nil {
		h.logger.ErrorContext(ctx, "echo failed", logging.Error(err))
		return nil, status.Error(codes.Internal, "echo failed")
	}

	return &userinternalapi.EchoResponse{
		OperationID: id.String(),
	}, nil
}
