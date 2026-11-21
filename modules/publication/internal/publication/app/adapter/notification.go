package adapter

import (
	"context"
	"uuid"
)

type NotificationService interface {
	Create(context.Context, string, uuid.UUID, uuid.UUID, string, string) error
}
