package main

import (
	interceptorgrpc "github.com/distributed-programming-2026/gateway/api/client/interceptorgrpc"
	"github.com/distributed-programming-2026/lib/runtime/grpc"
	"github.com/distributed-programming-2026/lib/runtime/http"
	grpcserver "google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	accessinternalapi "access/api/server/internalapi"
	"access/internal/access/app"
	"access/internal/access/config"
	interceptorhandler "access/internal/access/infra/handlers/interceptor"
	internalhandler "access/internal/access/infra/handlers/internalapi"
)

func newHTTPServer(conf config.Env) (*http.Server, error) {
	return http.NewServer(http.Config{
		Address:      conf.HTTP.Address,
		GraceTimeout: conf.App.GraceTimeout,
	})
}

func newGRPCServer(conf config.Env, svc app.ModuleAccessService) (*grpc.Server, error) {
	return grpc.NewServer(grpc.Config{
		Address:      conf.GRPC.Address,
		GraceTimeout: conf.App.GraceTimeout,
	}, func(server *grpcserver.Server) error {
		accessinternalapi.RegisterAccessServiceServer(server, internalhandler.New(svc))
		interceptorgrpc.RegisterInterceptorServer(server, interceptorhandler.New(svc))
		reflection.Register(server)
		return nil
	})
}
