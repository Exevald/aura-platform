package app

import (
	"context"
	"uuid"

	"github.com/distributed-programming-2026/go-sdk/pkg/mysql"
	"github.com/distributed-programming-2026/lib/xa"

	"user/internal/user/domain"
	usermysql "user/internal/user/infra/mysql"
)

type RegistrationParticipant interface {
	Prepare(context.Context, uuid.UUID, uuid.UUID, UserFieldsSpec) error
	Commit(context.Context, uuid.UUID) error
	Rollback(context.Context, uuid.UUID) error
}

type registrationParticipant struct {
	manager           *xa.Manager
	dispatcherFactory DispatcherFactory
}

func NewRegistrationParticipant(
	manager *xa.Manager,
	dispatcherFactory DispatcherFactory,
) RegistrationParticipant {
	return &registrationParticipant{
		manager:           manager,
		dispatcherFactory: dispatcherFactory,
	}
}

func (p *registrationParticipant) Prepare(
	ctx context.Context,
	transactionID,
	userID uuid.UUID,
	spec UserFieldsSpec,
) error {
	return p.manager.Prepare(ctx, userXID(transactionID), func(xaCtx context.Context, client mysql.ClientContext) error {
		provider := usermysql.NewRepositoryProvider(client)
		return domain.NewUserService(
			provider.UserRepository(xaCtx),
			p.dispatcherFactory.NewDispatcher(xaCtx),
		).CreateUserWithID(userID, spec.DisplayName, spec.FirstName, spec.LastName, spec.Email)
	})
}

func (p *registrationParticipant) Commit(ctx context.Context, transactionID uuid.UUID) error {
	return p.manager.Commit(ctx, userXID(transactionID))
}

func (p *registrationParticipant) Rollback(ctx context.Context, transactionID uuid.UUID) error {
	return p.manager.Rollback(ctx, userXID(transactionID))
}

func userXID(transactionID uuid.UUID) xa.XID {
	return xa.XID{GlobalID: transactionID.String(), Branch: "user"}
}
