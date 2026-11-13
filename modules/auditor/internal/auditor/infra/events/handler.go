package events

import (
	"context"

	"github.com/distributed-programming-2026/go-sdk/pkg/event"

	"auditor/internal/auditor/app"
)

type handler struct{ service *app.AuditService }

func NewHandler(service *app.AuditService) event.Handler { return &handler{service: service} }

func (h *handler) Handle(ctx context.Context, envelope event.Envelope) error {
	payload, err := envelope.Marshal()
	if err != nil {
		return err
	}
	_, err = h.service.Record(ctx, app.AuditEvent{EventID: envelope.ID(), EventType: envelope.Type(), Producer: envelope.Producer(), CorrelationID: envelope.CorrelationID(), ResourceType: envelope.Producer(), Payload: payload, OccurredAt: envelope.OccurredAt()})
	return err
}
