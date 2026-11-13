package main

import (
	"log/slog"

	interceptorgrpc "github.com/distributed-programming-2026/gateway/api/client/interceptorgrpc"
	"github.com/distributed-programming-2026/lib/runtime/grpc"
	"github.com/distributed-programming-2026/lib/runtime/http"
	grpcserver "google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	accessinternalapi "access/api/server/internalapi"
	"access/internal/access/app"
	"access/internal/access/config"
	clientinfra "access/internal/access/infra/client"
	interceptorhandler "access/internal/access/infra/handlers/interceptor"
	internalhandler "access/internal/access/infra/handlers/internalapi"
	publichandler "access/internal/access/infra/handlers/public"
)

func newHTTPServer(logger *slog.Logger, conf config.Env, svc app.ModuleAccessService, groups clientinfra.UserGroup, registry clientinfra.AppRegistry) (*http.Server, error) {
	register, err := publichandler.Register(svc, groups, logger, registry)
	if err != nil {
		return nil, err
	}
	return http.NewServer(http.Config{
		Address:      conf.HTTP.Address,
		GraceTimeout: conf.App.GraceTimeout,
	}, register)
}

func newGRPCServer(conf config.Env, svc app.ModuleAccessService, registration app.RegistrationParticipant) (*grpc.Server, error) {
	return grpc.NewServer(grpc.Config{
		Address:      conf.GRPC.Address,
		GraceTimeout: conf.App.GraceTimeout,
	}, func(server *grpcserver.Server) error {
		accessinternalapi.RegisterModuleAccessServiceServer(server, internalhandler.New(svc, registration))
		interceptorgrpc.RegisterInterceptorServer(server, interceptorhandler.New(svc))
		reflection.Register(server)
		return nil
	})
}
