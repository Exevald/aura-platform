package main

import (
	"context"
	"errors"
	"log/slog"

	"github.com/urfave/cli/v3"
	"golang.org/x/sync/errgroup"

	"auditor/internal/auditor/config"
)

func (a *application) serviceCmd() *cli.Command {
	return &cli.Command{
		Name: "service",
		Action: func(ctx context.Context, _ *cli.Command) error {
			return service(ctx, a.logger, a.conf)
		},
	}
}

func service(ctx context.Context, logger *slog.Logger, conf config.Env) (err error) {
	db, err := openDatabase(ctx, logger, conf)
	if err != nil {
		return err
	}
	defer func() {
		err = errors.Join(err, db.connector.Close())
	}()

	httpServer, err := newHTTPServer(conf)
	if err != nil {
		return err
	}

	grpcServer, err := newGRPCServer(conf)
	if err != nil {
		return err
	}

	errGroup, egCtx := errgroup.WithContext(ctx)
	errGroup.Go(func() error {
		return httpServer.Serve(egCtx)
	})
	errGroup.Go(func() error {
		return grpcServer.Serve(egCtx)
	})
	return errGroup.Wait()
}
