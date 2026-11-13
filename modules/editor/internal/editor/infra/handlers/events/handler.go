package events

import (
	"context"
	"fmt"
	"uuid"

	"github.com/distributed-programming-2026/go-sdk/pkg/event"

	accessevents "access/api/events"
	"editor/internal/editor/app"
)

type handler struct{ service app.EditorService }

func NewHandler(service app.EditorService) event.Handler { return &handler{service: service} }

func (h *handler) Handle(ctx context.Context, envelope event.Envelope) error {
	switch envelope.Type() {
	case "GroupModuleAccessGranted":
		var value accessevents.GroupModuleAccessGranted
		if err := envelope.UnmarshalPayload(&value); err != nil {
			return err
		}
		id, err := uuid.Parse(value.ShareId)
		if err != nil {
			return err
		}
		return h.service.ActivateModuleShare(ctx, id)
	case "GroupModuleAccessRejected":
		var value accessevents.GroupModuleAccessRejected
		if err := envelope.UnmarshalPayload(&value); err != nil {
			return err
		}
		id, err := uuid.Parse(value.ShareId)
		if err != nil {
			return err
		}
		return h.service.RejectModuleShare(ctx, id, value.Reason)
	case "GroupModuleAccessRevoked":
		var value accessevents.GroupModuleAccessRevoked
		if err := envelope.UnmarshalPayload(&value); err != nil {
			return err
		}
		id, err := uuid.Parse(value.ShareId)
		if err != nil {
			return err
		}
		return h.service.CompleteModuleShareRevocation(ctx, id, value.Reason)
	default:
		return fmt.Errorf("unsupported event type %q", envelope.Type())
	}
}
