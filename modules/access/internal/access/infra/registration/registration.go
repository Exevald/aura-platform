package registration

import (
	"access/internal/access/domain/model"
	"access/internal/access/domain/service"
	"context"
	"uuid"

	"github.com/distributed-programming-2026/go-sdk/pkg/mysql"
	"github.com/distributed-programming-2026/lib/xa"

	"access/internal/access/app"
	accessmysql "access/internal/access/infra/mysql"
)

type registrationParticipant struct {
	manager           *xa.Manager
	dispatcherFactory app.DispatcherFactory
}

func New(manager *xa.Manager, dispatcherFactory app.DispatcherFactory) app.RegistrationParticipant {
	return &registrationParticipant{
		manager:           manager,
		dispatcherFactory: dispatcherFactory,
	}
}

func (p *registrationParticipant) Prepare(
	ctx context.Context,
	transactionID,
	userID,
	moduleID uuid.UUID,
) error {
	return p.manager.Prepare(
		ctx,
		accessXID(transactionID),
		func(xaCtx context.Context, client mysql.ClientContext) error {
			provider := accessmysql.NewRepositoryProvider(client)
			accessID := uuid.New()
			return service.NewModuleAccessService(
				provider.ModuleAccess(xaCtx),
				p.dispatcherFactory.NewDispatcher(xaCtx),
			).CreateWithID(accessID, userID, moduleID, model.EffectAllow, "registration:"+transactionID.String())
		})
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
