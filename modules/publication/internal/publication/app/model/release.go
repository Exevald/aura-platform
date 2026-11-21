package model

import (
	"time"
	"uuid"
)

const (
	ReleasePending = "pending"
	ReleaseActive  = "active"
	ReleaseClosing = "closing"
	ReleaseClosed  = "closed"
	ReleaseFailed  = "failed"
)

type Release struct {
	ID              uuid.UUID
	OwnerID         uuid.UUID
	WorkspaceID     uuid.UUID
	DocumentID      uuid.UUID
	IdempotencyKey  string
	ExpectedVersion int64
	SourceVersion   int64
	Title           string
	Content         string
	Status          string
	LinkID          uuid.UUID
	ShortPath       string
	Failure         string
	DeliveryStatus  string
	CreatedAt       time.Time
	PublishedAt     time.Time
	Revision        int64
}

type Attachment struct {
	ReleaseID   uuid.UUID
	FileID      uuid.UUID
	CopyID      uuid.UUID
	FileName    string
	ContentType string
	SizeBytes   int64
}
