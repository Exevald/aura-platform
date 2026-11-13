package main

import (
	"log/slog"

	"github.com/distributed-programming-2026/go-sdk/pkg/amqp"
	"github.com/distributed-programming-2026/go-sdk/pkg/event"
	eventamqp "github.com/distributed-programming-2026/go-sdk/pkg/event/amqp"
	"github.com/distributed-programming-2026/go-sdk/pkg/event/inbox"

	"auditor/internal/auditor/app"
	"auditor/internal/auditor/config"
	eventhandler "auditor/internal/auditor/infra/events"
)

func newMessageHandler(logger *slog.Logger, conf config.Env, db *database, service *app.AuditService) (amqp.Handler, error) {
	handler, _, err := inbox.Decorate(eventhandler.NewHandler(service), db.unit, inbox.Config{Consumer: conf.AMQP.Queue})
	if err != nil {
		return nil, err
	}
	return eventamqp.NewHandler(event.ChainHandler(handler, event.HandlerLogging(logger)))
}
