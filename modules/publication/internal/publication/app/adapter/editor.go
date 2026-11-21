package adapter

import (
	"context"
	"errors"
	"uuid"
)

var ErrRejected = errors.New("editor rejected publication")

type Attachment struct {
	FileID      uuid.UUID
	FileName    string
	ContentType string
	SizeBytes   int64
}

type Snapshot struct {
	Title       string
	Content     string
	Version     int64
	Attachments []Attachment
}

type EditorService interface {
	Snapshot(context.Context, uuid.UUID, uuid.UUID, uuid.UUID, int64) (Snapshot, error)
	DownloadAttachment(context.Context, uuid.UUID, uuid.UUID, uuid.UUID, uuid.UUID) ([]byte, error)
}
