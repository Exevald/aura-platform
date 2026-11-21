package integration_tests

import (
	"testing"
	"time"
	"uuid"

	"editor/internal/editor/domain/model"
	"editor/internal/editor/infra/mysql/repository"

	sdkmysql "github.com/distributed-programming-2026/go-sdk/pkg/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

func TestAttachmentRepositoryRoundTrip(t *testing.T) {
	env := newEditorEnvironment(t)
	documentID, workspaceID, ownerID, fileID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	now := time.Now().UTC()
	_, err := env.db.ExecContext(t.Context(), `INSERT INTO document(id, owner_user_id, workspace_id, title, content, version, created_at, updated_at) VALUES (?, ?, ?, 'Test', '', 1, ?, ?)`, documentID[:], ownerID[:], workspaceID[:], now, now)
	require.NoError(t, err)

	client := sdkmysql.NewTransactionalClientFromSQLx(sqlx.NewDb(env.db, "mysql"))
	attachments := repository.NewAttachment(t.Context(), client)
	want := model.Attachment{DocumentID: documentID, FileID: fileID, WorkspaceID: workspaceID, FileName: "photo.png", ContentType: "image/png", SizeBytes: 4, CreatedAt: now}
	require.NoError(t, attachments.StoreAttachment(want))

	listed, err := attachments.ListAttachments(documentID)
	require.NoError(t, err)
	require.Len(t, listed, 1)
	require.Equal(t, fileID, listed[0].FileID)

	found, err := attachments.FindAttachment(documentID, fileID)
	require.NoError(t, err)
	require.Equal(t, want.FileID, found.FileID)
	require.Equal(t, want.FileName, found.FileName)

	require.NoError(t, attachments.DeleteAttachment(documentID, fileID))
	listed, err = attachments.ListAttachments(documentID)
	require.NoError(t, err)
	require.Empty(t, listed)
}
