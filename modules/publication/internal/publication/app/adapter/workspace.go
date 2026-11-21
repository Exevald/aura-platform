package adapter

import (
	"context"
	"uuid"
)

type WorkspaceService interface {
	Members(context.Context, uuid.UUID, uuid.UUID) ([]uuid.UUID, error)
}
