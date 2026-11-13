package main

import (
	"context"
	"errors"
	"log/slog"
	"uuid"

	"github.com/urfave/cli/v3"
	"golang.org/x/sync/errgroup"

	"editor/internal/editor/app"
	"editor/internal/editor/config"
	eventinfra "editor/internal/editor/infra/event"
	editormysql "editor/internal/editor/infra/mysql"
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
	users, userConnection, err := newUserClient(conf)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, userConnection.Close()) }()
	files, fileConnection, err := newFilestorageClient(conf)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, fileConnection.Close()) }()
	moduleID, err := uuid.Parse(conf.Dependencies.ModuleID)
	if err != nil {
		return err
	}
	svc := app.NewService(editormysql.NewUnitOfWork(db.unit), dispatcherFactory, users, moduleID, files)
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
	return errGroup.Wait()
}
