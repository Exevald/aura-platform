package main

import (
	"context"
	"errors"
	"log/slog"

	"github.com/distributed-programming-2026/go-sdk/pkg/amqp"
	"github.com/urfave/cli/v3"

	domainevent "github.com/distributed-programming-2026/lib/event"
	runtimehttp "github.com/distributed-programming-2026/lib/runtime/http"

	"access/internal/access/app"
	"access/internal/access/config"
	eventinfra "access/internal/access/infra/event"
	accessmysql "access/internal/access/infra/mysql"
)

func (a *application) messageHandlerCmd() *cli.Command {
	return &cli.Command{Name: "message-handler", Action: func(ctx context.Context, _ *cli.Command) error { return messageHandler(ctx, a.logger, a.conf) }}
}

func messageHandler(ctx context.Context, logger *slog.Logger, conf config.Env) (err error) {
	db, err := openDatabase(ctx, logger, conf)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, db.connector.Close()) }()
	connection := newConnection(logger, conf)
	defer func() { err = errors.Join(err, connection.Stop()) }()
	dispatcher, _, err := newDispatcher(logger, conf, db, connection)
	if err != nil {
		return err
	}
	svc := app.NewModuleAccessService(accessmysql.NewUnitOfWork(db.unit), eventinfra.NewDispatcherFactory(appID, dispatcher))
	handler, err := newMessageHandler(logger, conf, db, svc)
	if err != nil {
		return err
	}
	consumerCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	consumer := connection.Consumer(consumerCtx, handler, domainevent.DomainExchange(), eventQueue(conf), eventBinding(conf), &amqp.QoSConfig{PrefetchCount: conf.AMQP.PrefetchCount})
	defer func() {
		cancel()
		closeCtx, closeCancel := context.WithTimeout(context.Background(), conf.App.GraceTimeout)
		defer closeCancel()
		err = errors.Join(err, consumer.Close(closeCtx))
	}()
	if err = connection.Start(); err != nil {
		return err
	}
	server, err := runtimehttp.NewServer(runtimehttp.Config{Address: conf.HTTP.Address, GraceTimeout: conf.App.GraceTimeout})
	if err != nil {
		return err
	}
	return server.Serve(ctx)
}
