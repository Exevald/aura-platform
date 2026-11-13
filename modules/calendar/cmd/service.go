package main

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/distributed-programming-2026/go-sdk/pkg/logging"

	"github.com/urfave/cli/v3"
	"golang.org/x/sync/errgroup"

	"calendar/internal/calendar/app"
	"calendar/internal/calendar/config"
	eventinfra "calendar/internal/calendar/infra/event"
	calendarmysql "calendar/internal/calendar/infra/mysql"
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
	svc := app.NewService(calendarmysql.NewUnitOfWork(db.unit), dispatcherFactory, clients.notifications)
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
		ticker := time.NewTicker(conf.Reminders.Interval)
		defer ticker.Stop()
		for {
			select {
			case <-egCtx.Done():
				return egCtx.Err()
			case now := <-ticker.C:
				if err := svc.DeliverDueReminders(egCtx, now.UTC(), conf.Reminders.BatchSize); err != nil {
					logger.ErrorContext(egCtx, "deliver reminders", logging.Error(err))
				}
			}
		}
	})
	return errGroup.Wait()
}
