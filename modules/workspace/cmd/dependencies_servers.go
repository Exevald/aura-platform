package main

import (
	"log/slog"

	grpcserver "google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/distributed-programming-2026/lib/runtime/grpc"
	"github.com/distributed-programming-2026/lib/runtime/http"

	internalapi "workspace/api/server/internalapi"
	"workspace/internal/workspace/app"
	"workspace/internal/workspace/config"
	interceptorhandler "workspace/internal/workspace/infra/handlers/interceptor"
	internalhandler "workspace/internal/workspace/infra/handlers/internalapi"
	publichandler "workspace/internal/workspace/infra/handlers/public"

	interceptorgrpc "github.com/distributed-programming-2026/gateway/api/client/interceptorgrpc"
)

func newHTTPServer(logger *slog.Logger, conf config.Env, svc *app.Service) (*http.Server, error) {
	register, err := publichandler.Register(svc, logger)
	if err != nil {
		return nil, err
	}

	return http.NewServer(http.Config{
		Address:      conf.HTTP.Address,
		GraceTimeout: conf.App.GraceTimeout,
	}, register)
}

func newGRPCServer(logger *slog.Logger, conf config.Env, svc *app.Service) (*grpc.Server, error) {
	return grpc.NewServer(grpc.Config{
		Address:      conf.GRPC.Address,
		GraceTimeout: conf.App.GraceTimeout,
	}, func(server *grpcserver.Server) error {
		internalapi.RegisterWorkspaceServiceServer(server, internalhandler.NewHandler(svc, logger))
		interceptorgrpc.RegisterInterceptorServer(server, interceptorhandler.New(svc))
		reflection.Register(server)
		return nil
	})
}
