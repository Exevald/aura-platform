package main

import (
	interceptorgrpc "github.com/distributed-programming-2026/gateway/api/client/interceptorgrpc"
	"github.com/distributed-programming-2026/lib/runtime/grpc"
	"github.com/distributed-programming-2026/lib/runtime/http"
	grpcserver "google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	sessioninternalapi "session/api/server/internalapi"
	"session/internal/session/app"
	"session/internal/session/config"
	interceptorhandler "session/internal/session/infra/handlers/interceptor"
	internalhandler "session/internal/session/infra/handlers/internalapi"
)

func newHTTPServer(conf config.Env) (*http.Server, error) {
	return http.NewServer(http.Config{
		Address:      conf.HTTP.Address,
		GraceTimeout: conf.App.GraceTimeout,
	})
}

func newGRPCServer(conf config.Env, svc app.SessionService) (*grpc.Server, error) {
	return grpc.NewServer(grpc.Config{
		Address:      conf.GRPC.Address,
		GraceTimeout: conf.App.GraceTimeout,
	}, func(server *grpcserver.Server) error {
		sessioninternalapi.RegisterSessionServiceServer(server, internalhandler.NewHandler(svc))
		interceptorgrpc.RegisterInterceptorServer(server, interceptorhandler.NewSessionInterceptor(svc, conf.CookieName))
		reflection.Register(server)
		return nil
	})
}
