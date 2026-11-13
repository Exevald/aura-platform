package main

import (
	"context"
	"errors"
	"log/slog"

	"github.com/urfave/cli/v3"
	"golang.org/x/sync/errgroup"

	"access/internal/access/app"
	"access/internal/access/config"
	eventinfra "access/internal/access/infra/event"
	accessmysql "access/internal/access/infra/mysql"
	registrationinfra "access/internal/access/infra/registration"
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
	svc := app.NewModuleAccessService(accessmysql.NewUnitOfWork(db.unit), eventinfra.NewDispatcherFactory(appID, dispatcher))
	groups, groupConnection, err := newUserGroupClient(conf)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, groupConnection.Close()) }()
	registry, err := newAppRegistryClient(conf)
	if err != nil {
		return err
	}
	if err = db.xa.Validate(ctx); err != nil {
		return err
	}
	registration := registrationinfra.New(db.xa, eventinfra.NewDispatcherFactory(appID, dispatcher))

	err = connection.Start()
	if err != nil {
		return err
	}

	httpServer, err := newHTTPServer(logger, conf, svc, groups, registry)
	if err != nil {
		return err
	}

	grpcServer, err := newGRPCServer(conf, svc, registration)
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
