package registration

import (
	"context"
	"uuid"

	"github.com/distributed-programming-2026/go-sdk/pkg/mysql"
	"github.com/distributed-programming-2026/lib/xa"

	"access/internal/access/app"
)

type registrationParticipant struct {
	manager *xa.Manager
}

func New(manager *xa.Manager) app.RegistrationParticipant {
	return &registrationParticipant{manager: manager}
}

func (p *registrationParticipant) Prepare(
	ctx context.Context,
	transactionID,
	_,
	_ uuid.UUID,
) error {
	return p.manager.Prepare(
		ctx,
		accessXID(transactionID),
		func(context.Context, mysql.ClientContext) error { return nil },
	)
}

func (p *registrationParticipant) Commit(ctx context.Context, transactionID uuid.UUID) error {
	return p.manager.Commit(ctx, accessXID(transactionID))
}

func (p *registrationParticipant) Rollback(ctx context.Context, transactionID uuid.UUID) error {
	return p.manager.Rollback(ctx, accessXID(transactionID))
}

func accessXID(transactionID uuid.UUID) xa.XID {
	return xa.XID{GlobalID: transactionID.String(), Branch: "access"}
}
