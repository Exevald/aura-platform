package events

import (
	"context"
	"log/slog"

	"github.com/distributed-programming-2026/go-sdk/pkg/event"
)

type handler struct {
	logger *slog.Logger
}

func NewHandler(logger *slog.Logger) event.Handler {
	return &handler{
		logger: logger,
	}
}

func (h *handler) Handle(ctx context.Context, envelope event.Envelope) error {
	h.logger.InfoContext(ctx, "realtime event received", slog.String("event_id", envelope.ID()), slog.String("type", envelope.Type()))
	return nil
}
