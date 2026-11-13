package registration

import (
	"context"
	"uuid"

	sdkmysql "github.com/distributed-programming-2026/go-sdk/pkg/mysql"
	"github.com/distributed-programming-2026/lib/xa"

	"auth/internal/auth/app"
	"auth/internal/auth/domain"
	mysqlinfra "auth/internal/auth/infra/mysql"
)

type registrationParticipant struct {
	manager           *xa.Manager
	dispatcherFactory app.DispatcherFactory
	hasher            app.PasswordHasher
}

func New(
	manager *xa.Manager,
	dispatcherFactory app.DispatcherFactory,
	hasher app.PasswordHasher,
) app.RegistrationParticipant {
	return &registrationParticipant{
		manager:           manager,
		dispatcherFactory: dispatcherFactory,
		hasher:            hasher,
	}
}

func (p *registrationParticipant) Prepare(
	ctx context.Context,
	transactionID,
	userID uuid.UUID,
	login,
	password string,
) (uuid.UUID, error) {
	var credential domain.Credential
	err := p.manager.Prepare(
		ctx,
		p.registrationXID(transactionID),
		func(xaCtx context.Context, client sdkmysql.ClientContext) error {
			provider := mysqlinfra.NewRepositoryProvider(client)
			service := domain.NewAuthService(
				provider.Credential(xaCtx),
				p.hasher,
				p.dispatcherFactory.NewDispatcher(xaCtx),
				domain.SystemClock{},
			)
			var createErr error
			credential, createErr = service.CreateCredential(userID, login, password)
			return createErr
		})
	if err != nil {
		return uuid.Nil(), err
	}
	return credential.ID, nil
}

func (p *registrationParticipant) Commit(ctx context.Context, transactionID uuid.UUID) error {
	return p.manager.Commit(ctx, p.registrationXID(transactionID))
}

func (p *registrationParticipant) Rollback(ctx context.Context, transactionID uuid.UUID) error {
	return p.manager.Rollback(ctx, p.registrationXID(transactionID))
}

func (p *registrationParticipant) registrationXID(transactionID uuid.UUID) xa.XID {
	return xa.XID{
		GlobalID: transactionID.String(),
		Branch:   "auth",
	}
}
