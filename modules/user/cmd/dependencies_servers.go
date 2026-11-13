package main

import (
	"github.com/distributed-programming-2026/lib/runtime/grpc"
	"github.com/distributed-programming-2026/lib/runtime/http"
	grpcserver "google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	userinternalapi "user/api/server/internalapi"
	"user/internal/user/app"
	"user/internal/user/config"
	internalhandler "user/internal/user/infra/handlers/internalapi"
)

func newHTTPServer(conf config.Env) (*http.Server, error) {
	return http.NewServer(http.Config{
		Address:      conf.HTTP.Address,
		GraceTimeout: conf.App.GraceTimeout,
	})
}

func newGRPCServer(conf config.Env, svc app.UserService, registration app.RegistrationParticipant) (*grpc.Server, error) {
	return grpc.NewServer(grpc.Config{
		Address:      conf.GRPC.Address,
		GraceTimeout: conf.App.GraceTimeout,
	}, func(server *grpcserver.Server) error {
		userinternalapi.RegisterUserServiceServer(server, internalhandler.NewHandler(svc, registration))
		reflection.Register(server)
		return nil
	})
}
