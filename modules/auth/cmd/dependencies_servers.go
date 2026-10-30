package main

import (
	"log/slog"

	grpcserver "google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/distributed-programming-2026/lib/runtime/grpc"
	"github.com/distributed-programming-2026/lib/runtime/http"

	"auth/internal/auth/app"
	"auth/internal/auth/config"
	publichandler "auth/internal/auth/infra/handlers/public"
)

func newHTTPServer(logger *slog.Logger, conf config.Env, svc app.AuthService) (*http.Server, error) {
	register, err := publichandler.Register(svc, logger, conf.Cookie)
	if err != nil {
		return nil, err
	}
	return http.NewServer(http.Config{
		Address:      conf.HTTP.Address,
		GraceTimeout: conf.App.GraceTimeout,
	}, register)
}

func newGRPCServer(conf config.Env) (*grpc.Server, error) {
	return grpc.NewServer(grpc.Config{
		Address:      conf.GRPC.Address,
		GraceTimeout: conf.App.GraceTimeout,
	}, func(server *grpcserver.Server) error {
		reflection.Register(server)
		return nil
	})
}
