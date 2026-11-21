package main

import (
	"log/slog"

	grpcserver "google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/distributed-programming-2026/lib/runtime/grpc"
	"github.com/distributed-programming-2026/lib/runtime/http"

	internalapi "editor/api/server/internalapi"
	"editor/internal/editor/app"
	"editor/internal/editor/config"
	internalhandler "editor/internal/editor/infra/handlers/internalapi"
	publichandler "editor/internal/editor/infra/handlers/public"
)

func newHTTPServer(logger *slog.Logger, conf config.Env, svc app.EditorService) (*http.Server, error) {
	register, err := publichandler.Register(svc, logger)
	if err != nil {
		return nil, err
	}

	return http.NewServer(http.Config{
		Address:      conf.HTTP.Address,
		GraceTimeout: conf.App.GraceTimeout,
	}, register)
}

func newGRPCServer(_ *slog.Logger, conf config.Env, svc app.EditorService) (*grpc.Server, error) {
	return grpc.NewServer(grpc.Config{
		Address:      conf.GRPC.Address,
		GraceTimeout: conf.App.GraceTimeout,
	}, func(server *grpcserver.Server) error {
		internalapi.RegisterEditorServiceServer(server, internalhandler.New(svc))
		reflection.Register(server)
		return nil
	})
}
