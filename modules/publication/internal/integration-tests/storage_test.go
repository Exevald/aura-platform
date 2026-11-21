package integration_tests

import (
	"errors"
	"publication/internal/publication/app/service"
	storemysql "publication/internal/publication/infra/mysql/storage"
	"testing"

	"publication/internal/publication/app/model"
	"uuid"

	"github.com/stretchr/testify/require"
)

func TestReleaseStoreRejectsStaleRevision(t *testing.T) {
	store := &storemysql.ReleaseStore{Client: testEnv.connector.TransactionalClient()}
	created, err := store.Store(t.Context(), model.Release{
		ID: uuid.New(), OwnerID: uuid.New(), WorkspaceID: uuid.New(),
		DocumentID: uuid.New(), IdempotencyKey: uuid.New().String(),
		ExpectedVersion: 1, Status: model.ReleasePending,
	})
	require.NoError(t, err)
	require.Equal(t, int64(1), created.Revision)

	stale := created
	created.Title = "new title"
	updated, err := store.Store(t.Context(), created)
	require.NoError(t, err)
	require.Equal(t, int64(2), updated.Revision)

	stale.Title = "old title"
	_, err = store.Store(t.Context(), stale)
	require.True(t, errors.Is(err, service.ErrConflict), "expected revision conflict, got %v", err)
	got, err := store.Get(t.Context(), created.ID)
	require.NoError(t, err)
	require.Equal(t, "new title", got.Title)
}
