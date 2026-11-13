package events

import (
	"context"
	"fmt"
	"uuid"

	"github.com/distributed-programming-2026/go-sdk/pkg/event"

	"access/internal/access/app"
	usergroupevents "usergroup/api/events"
)

type handler struct {
	moduleAccess app.ModuleAccessService
}

func NewHandler(moduleAccess app.ModuleAccessService) event.Handler {
	return &handler{moduleAccess: moduleAccess}
}

func (h *handler) Handle(ctx context.Context, envelope event.Envelope) error {
	switch envelope.Type() {
	case "GroupMembershipChanged":
		var value usergroupevents.GroupMembershipChanged
		if err := envelope.UnmarshalPayload(&value); err != nil {
			return err
		}
		groupID, err := uuid.Parse(value.GroupId)
		if err != nil {
			return err
		}
		userID, err := uuid.Parse(value.UserId)
		if err != nil {
			return err
		}
		ownerID, err := uuid.Parse(value.OwnerUserId)
		if err != nil {
			return err
		}
		if userID == ownerID {
			return nil
		}
		return h.moduleAccess.ChangeGroupMembership(ctx, groupID, userID, value.Type == "added")

	case "GroupDeleted":
		var value usergroupevents.GroupDeleted
		if err := envelope.UnmarshalPayload(&value); err != nil {
			return err
		}
		groupID, err := uuid.Parse(value.GroupId)
		if err != nil {
			return err
		}
		return h.moduleAccess.DeleteGroupModuleAccess(ctx, groupID)
	default:
		return fmt.Errorf("unsupported event type %q", envelope.Type())
	}
}
