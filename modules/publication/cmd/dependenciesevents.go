package main

import (
	"log/slog"
	service2 "publication/internal/publication/app/service"

	"github.com/distributed-programming-2026/go-sdk/pkg/amqp"
	"github.com/distributed-programming-2026/go-sdk/pkg/event"
	eventamqp "github.com/distributed-programming-2026/go-sdk/pkg/event/amqp"

	eventhandler "publication/internal/publication/infra/handlers/events"
)

func newMessageHandler(logger *slog.Logger, releases *service2.Releases) (amqp.Handler, error) {
	return eventamqp.NewHandler(event.ChainHandler(eventhandler.NewHandler(logger, releases), event.HandlerLogging(logger)))
}
