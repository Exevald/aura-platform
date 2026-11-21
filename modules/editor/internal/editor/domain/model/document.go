package model

import (
	"errors"
	"time"
	"unicode/utf8"
	"uuid"
)

const MaxContentBytes = 1024 * 1024

const (
	RoleOwner  Role = "owner"
	RoleEditor Role = "editor"
	RoleViewer Role = "viewer"
)

var (
	ErrDocumentNotFound  = errors.New("document not found")
	ErrPermissionDenied  = errors.New("document permission denied")
	ErrInvalidDocument   = errors.New("invalid document")
	ErrInvalidPermission = errors.New("invalid document permission")
	ErrVersionConflict   = errors.New("document version conflict")
)

type Role string

type Document struct {
	ID          uuid.UUID
	OwnerUserID uuid.UUID
	WorkspaceID uuid.UUID
	Title       string
	Content     string
	Version     int64
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Role        Role
}

type Permission struct {
	DocumentID uuid.UUID
	UserID     uuid.UUID
	Role       Role
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type Attachment struct {
	DocumentID  uuid.UUID
	FileID      uuid.UUID
	WorkspaceID uuid.UUID
	FileName    string
	ContentType string
	SizeBytes   int64
	CreatedAt   time.Time
}

type AttachmentRepository interface {
	StoreAttachment(Attachment) error
	ListAttachments(uuid.UUID) ([]Attachment, error)
	FindAttachment(uuid.UUID, uuid.UUID) (Attachment, error)
	DeleteAttachment(uuid.UUID, uuid.UUID) error
}

type DocumentRepository interface {
	NextID() (uuid.UUID, error)
	Create(Document) error
	Find(uuid.UUID) (Document, error)
	ListForUser(uuid.UUID, uuid.UUID) ([]Document, error)
	Update(Document, int64) error
	Delete(uuid.UUID) error
	FindRole(uuid.UUID, uuid.UUID) (Role, error)
	ListPermissions(uuid.UUID) ([]Permission, error)
	StorePermission(Permission) error
	DeletePermission(uuid.UUID, uuid.UUID) error
}

func ValidDocument(title, content string) bool {
	return title != "" && utf8.ValidString(title) && utf8.RuneCountInString(title) <= 255 && len(content) <= MaxContentBytes && utf8.ValidString(content)
}
