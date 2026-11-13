package main

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/urfave/cli/v3"
	"golang.org/x/sync/errgroup"

	"registerer/internal/registerer/app"
	"registerer/internal/registerer/config"
	eventinfra "registerer/internal/registerer/infra/event"
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

	connection := newConnection(logger, conf)
	defer func() {
		err = errors.Join(err, connection.Stop())
	}()

	dispatcher, relay, err := newDispatcher(logger, conf, db, connection)
	if err != nil {
		return err
	}

	err = connection.Start()
	if err != nil {
		return err
	}

	dispatcherFactory := eventinfra.NewDispatcherFactory(appID, dispatcher)
	clients, err := newServiceClients(conf)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, clients.Close()) }()
	svc := app.NewService(
		appUnitOfWork{unit: db.unit},
		dispatcherFactory,
		clients.auth,
		clients.users,
		clients.sessions,
		clients.workspaces,
		clients.access,
	)
	httpServer, err := newHTTPServer(logger, conf, svc)
	if err != nil {
		return err
	}

	grpcServer, err := newGRPCServer(logger, conf, svc)
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
	errGroup.Go(func() error {
		return runRelay(egCtx, logger, relay, conf.Relay)
	})
	errGroup.Go(func() error {
		return runRegistrationRecovery(egCtx, logger, svc, conf.Recovery)
	})
	return errGroup.Wait()
}

func runRegistrationRecovery(ctx context.Context, logger *slog.Logger, svc *app.Service, conf config.Recovery) error {
	ticker := time.NewTicker(conf.Interval)
	defer ticker.Stop()
	for {
		if err := svc.RecoverRegistrations(ctx, conf.BatchSize); err != nil && ctx.Err() == nil {
			logger.ErrorContext(ctx, "registration recovery failed", "error", err)
		}
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
		}
	}
}
