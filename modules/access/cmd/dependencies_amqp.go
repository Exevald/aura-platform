package main

import (
	"log/slog"

	"github.com/distributed-programming-2026/go-sdk/pkg/amqp"

	"access/internal/access/config"
)

func newConnection(logger *slog.Logger, conf config.Env) amqp.Connection {
	return amqp.NewConnection(appID, amqp.ConnectionConfig{
		URL:            conf.AMQP.URL,
		ConnectTimeout: conf.AMQP.ConnectTimeout,
	}, logger)
}
