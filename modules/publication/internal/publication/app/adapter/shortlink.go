package adapter

import (
	"context"
	"uuid"
)

type ShortlinkService interface {
	Ensure(context.Context, uuid.UUID, string) (uuid.UUID, string, error)
	Delete(context.Context, uuid.UUID) error
}
