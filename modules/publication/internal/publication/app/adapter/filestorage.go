package adapter

import (
	"context"
	"uuid"
)

type FilestorageService interface {
	Upload(context.Context, uuid.UUID, uuid.UUID, string, string, []byte) (uuid.UUID, error)
	Download(context.Context, uuid.UUID, uuid.UUID) ([]byte, error)
	Delete(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) error
}
