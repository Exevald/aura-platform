package main

import (
	"log/slog"

	grpcserver "google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/distributed-programming-2026/lib/runtime/grpc"
	"github.com/distributed-programming-2026/lib/runtime/http"

	authinternalapi "auth/api/server/internalapi"
	"auth/internal/auth/app"
	"auth/internal/auth/config"
	internalhandler "auth/internal/auth/infra/handlers/internalapi"
	publichandler "auth/internal/auth/infra/handlers/public"
)

func newHTTPServer(logger *slog.Logger, conf config.Env, svc app.AuthService, sso *app.SSOService) (*http.Server, error) {
	register, err := publichandler.Register(svc, sso, logger, conf.Cookie)
	if err != nil {
		return nil, err
	}
	return http.NewServer(http.Config{
		Address:      conf.HTTP.Address,
		GraceTimeout: conf.App.GraceTimeout,
	}, register)
}

func newGRPCServer(conf config.Env, registration app.RegistrationParticipant) (*grpc.Server, error) {
	return grpc.NewServer(grpc.Config{
		Address:      conf.GRPC.Address,
		GraceTimeout: conf.App.GraceTimeout,
	}, func(server *grpcserver.Server) error {
		authinternalapi.RegisterAuthServiceServer(server, internalhandler.New(registration))
		reflection.Register(server)
		return nil
	})
}
