package events

import (
	"context"
	"log/slog"
	"uuid"

	editorevents "editor/api/events"

	"github.com/distributed-programming-2026/go-sdk/pkg/event"
)

type handler struct {
	logger *slog.Logger
	closer interface {
		CloseByDocument(context.Context, uuid.UUID) error
	}
}

func NewHandler(logger *slog.Logger, closer interface {
	CloseByDocument(context.Context, uuid.UUID) error
}) event.Handler {
	return &handler{
		logger: logger,
		closer: closer,
	}
}

func (h *handler) Handle(ctx context.Context, envelope event.Envelope) error {
	if envelope.Type() == "DocumentChanged" {
		var changed editorevents.DocumentChanged
		if err := envelope.UnmarshalPayload(&changed); err != nil {
			return err
		}
		if changed.Type == "deleted" {
			id, err := uuid.Parse(changed.DocumentId)
			if err != nil {
				return err
			}
			return h.closer.CloseByDocument(ctx, id)
		}
		return nil
	}
	return nil
}
