package events

import (
	"context"
	"fmt"
	"log/slog"

	editorevents "editor/api/events"

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
	if envelope.Type() != "DocumentChanged" {
		return fmt.Errorf("unsupported event type %q", envelope.Type())
	}
	var changed editorevents.DocumentChanged
	err := envelope.UnmarshalPayload(&changed)
	if err != nil {
		return err
	}
	h.logger.InfoContext(ctx, "document event received", slog.String("document_id", changed.DocumentId), slog.String("type", changed.Type))
	return nil
}
