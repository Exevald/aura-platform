package main

import (
	"log/slog"

	"github.com/distributed-programming-2026/go-sdk/pkg/amqp"
	"github.com/distributed-programming-2026/go-sdk/pkg/event"
	eventamqp "github.com/distributed-programming-2026/go-sdk/pkg/event/amqp"
	"github.com/distributed-programming-2026/go-sdk/pkg/event/inbox"

	"bff/internal/bff/config"
	eventhandler "bff/internal/bff/infra/handlers/events"
)

func newMessageHandler(logger *slog.Logger, conf config.Env, db *database) (amqp.Handler, error) {
	handler, _, err := inbox.Decorate(eventhandler.NewHandler(logger), db.unit, inbox.Config{
		Consumer: conf.AMQP.Queue,
	})
	if err != nil {
		return nil, err
	}

	return eventamqp.NewHandler(event.ChainHandler(handler, event.HandlerLogging(logger)))
}
