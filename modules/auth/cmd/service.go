package main

import (
	"context"
	"errors"
	"log/slog"

	"github.com/urfave/cli/v3"
	"golang.org/x/sync/errgroup"

	"auth/internal/auth/app"
	"auth/internal/auth/config"
	eventinfra "auth/internal/auth/infra/event"
	authmysql "auth/internal/auth/infra/mysql"
	"auth/internal/auth/infra/password"
	registrationinfra "auth/internal/auth/infra/registration"
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

	clients, err := newServiceClients(conf)
	if err != nil {
		return err
	}
	defer func() {
		err = errors.Join(err, clients.Close())
	}()

	dispatcherFactory := eventinfra.NewDispatcherFactory(appID, dispatcher)
	if err = db.xa.Validate(ctx); err != nil {
		return err
	}
	hasher := password.NewArgon2ID()
	svc := app.NewAuthService(
		authmysql.NewUnitOfWork(db.unit),
		dispatcherFactory,
		hasher,
		clients.sessions,
	)
	registration := registrationinfra.New(db.xa, dispatcherFactory, hasher)
	sso := app.NewSSOService(clients.sessions, authmysql.NewSSOTicketStore(db.connector.TransactionalClient()))
	httpServer, err := newHTTPServer(logger, conf, svc, sso)
	if err != nil {
		return err
	}

	grpcServer, err := newGRPCServer(conf, registration)
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
