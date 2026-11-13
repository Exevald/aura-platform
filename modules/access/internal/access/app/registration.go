package app

import (
	"context"
	"uuid"
)

type RegistrationParticipant interface {
	Prepare(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) error
	Commit(context.Context, uuid.UUID) error
	Rollback(context.Context, uuid.UUID) error
}
