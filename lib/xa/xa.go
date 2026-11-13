package xa

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"github.com/distributed-programming-2026/go-sdk/pkg/mysql"
	"github.com/distributed-programming-2026/go-sdk/pkg/uow"
	mysqldriver "github.com/go-sql-driver/mysql"
)

const formatID = 0x41555241

var (
	ErrXIDNotFound = errors.New("XA XID not found")
)

type XID struct {
	GlobalID string
	Branch   string
}

type Manager struct {
	client mysql.TransactionalClient
}

type ContextualUnitOfWork struct {
	delegate uow.LockableUnitOfWork
}

type clientContextKey struct{}

func NewManager(client mysql.TransactionalClient) *Manager {
	return &Manager{client: client}
}

func NewContextualUnitOfWork(delegate uow.LockableUnitOfWork) *ContextualUnitOfWork {
	return &ContextualUnitOfWork{delegate: delegate}
}

func (m *Manager) Validate(ctx context.Context) error {
	var enabled bool
	err := m.client.GetContext(
		ctx,
		"xa.detach-enabled",
		&enabled,
		"SELECT @@xa_detach_on_prepare",
	)
	if err != nil {
		return fmt.Errorf("check xa_detach_on_prepare: %w", err)
	}
	if !enabled {
		return errors.New("xa_detach_on_prepare must be enabled")
	}
	return nil
}

func (m *Manager) Prepare(
	ctx context.Context,
	xid XID,
	work func(context.Context, mysql.ClientContext) error,
) (err error) {
	if work == nil {
		return errors.New("XA prepare work must not be nil")
	}
	statement, err := xid.statement()
	if err != nil {
		return err
	}
	connection, err := m.client.Connection(ctx)
	if err != nil {
		return fmt.Errorf("open XA connection: %w", err)
	}
	defer func() { err = errors.Join(err, connection.Close()) }()

	_, err = connection.ExecContext(
		ctx,
		"xa.start",
		"XA START "+statement)
	if err != nil {
		return fmt.Errorf("XA START: %w", err)
	}

	active := true
	defer func() {
		if !active {
			return
		}
		_, endErr := connection.ExecContext(
			context.WithoutCancel(ctx),
			"xa.end-after-error",
			"XA END "+statement,
		)
		_, rollbackErr := connection.ExecContext(
			context.WithoutCancel(ctx),
			"xa.rollback-after-error",
			"XA ROLLBACK "+statement,
		)
		err = errors.Join(err, endErr, rollbackErr)
	}()

	if err = work(withClient(ctx, connection), connection); err != nil {
		return err
	}
	_, err = connection.ExecContext(ctx, "xa.end", "XA END "+statement)
	if err != nil {
		return fmt.Errorf("XA END: %w", err)
	}
	active = false

	_, err = connection.ExecContext(ctx, "xa.prepare", "XA PREPARE "+statement)
	if err != nil {
		_, rollbackErr := connection.ExecContext(
			context.WithoutCancel(ctx),
			"xa.rollback-after-prepare-error",
			"XA ROLLBACK "+statement,
		)
		return errors.Join(fmt.Errorf("XA PREPARE: %w", err), rollbackErr)
	}
	return nil
}

func (m *Manager) Commit(ctx context.Context, xid XID) error {
	return m.finish(ctx, xid, "COMMIT")
}

func (m *Manager) Rollback(ctx context.Context, xid XID) error {
	return m.finish(ctx, xid, "ROLLBACK")
}

func (u *ContextualUnitOfWork) ExecuteWithClientContext(
	ctx context.Context,
	locks []uow.LockOptions,
	callback func(mysql.ClientContext) error,
) error {
	if client, ok := ctx.Value(clientContextKey{}).(mysql.ClientContext); ok {
		if len(locks) != 0 {
			return errors.New("advisory locks are not supported inside an XA branch")
		}
		return callback(client)
	}
	return u.delegate.ExecuteWithClientContext(ctx, locks, callback)
}

func (m *Manager) finish(ctx context.Context, xid XID, operation string) error {
	statement, err := xid.statement()
	if err != nil {
		return err
	}
	_, err = m.client.ExecContext(ctx, "xa."+strings.ToLower(operation), "XA "+operation+" "+statement)
	if isUnknownXID(err) {
		return fmt.Errorf(
			"XA %s: %w: %v",
			operation,
			ErrXIDNotFound,
			err,
		)
	}
	if err != nil {
		return fmt.Errorf("XA %s: %w", operation, err)
	}

	return nil
}

func (xid XID) statement() (string, error) {
	global := strings.ReplaceAll(xid.GlobalID, "-", "")
	if len(global) != 32 {
		return "", errors.New("invalid XA global transaction id")
	}
	if _, err := hex.DecodeString(global); err != nil {
		return "", errors.New("invalid XA global transaction id")
	}
	if xid.Branch == "" || len(xid.Branch) > 64 {
		return "", errors.New("invalid XA branch qualifier")
	}

	return fmt.Sprintf("X'%s',X'%s',%d", global, hex.EncodeToString([]byte(xid.Branch)), formatID), nil
}

func isUnknownXID(err error) bool {
	var mysqlErr *mysqldriver.MySQLError
	return errors.As(err, &mysqlErr) && mysqlErr.Number == 1397
}

func withClient(ctx context.Context, client mysql.ClientContext) context.Context {
	return context.WithValue(ctx, clientContextKey{}, client)
}
