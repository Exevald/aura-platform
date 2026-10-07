package main

import (
	"log/slog"

	"github.com/distributed-programming-2026/lib/runtime/grpc"
	"github.com/distributed-programming-2026/lib/runtime/http"
	grpcserver "google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	internalapi "user/api/server/internalapi"
	"user/internal/user/app"
	internalhandler "user/internal/user/infra/handlers/internalapi"
	publichandler "user/internal/user/infra/handlers/public"
)

func newHTTPServer(logger *slog.Logger, conf Env, svc *app.Service) (*http.Server, error) {
	register, err := publichandler.Register(svc, logger)
	if err != nil {
		return nil, err
	}

	return http.NewServer(http.Config{
		Address:      conf.HTTP.Address,
		GraceTimeout: conf.App.GraceTimeout,
	}, register)
}

func newGRPCServer(logger *slog.Logger, conf Env, svc *app.Service) (*grpc.Server, error) {
	return grpc.NewServer(grpc.Config{
		Address:      conf.GRPC.Address,
		GraceTimeout: conf.App.GraceTimeout,
	}, func(server *grpcserver.Server) error {
		internalapi.RegisterUserServiceServer(server, internalhandler.NewHandler(svc, logger))
		reflection.Register(server)
		return nil
	})
}
