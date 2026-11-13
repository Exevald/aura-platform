package domain

import (
	"errors"
	"strings"
	"time"
	"unicode/utf8"
	"uuid"
)

type Kind string

const (
	KindPersonal Kind = "personal"
	KindTeam     Kind = "team"
)

type Role string

const (
	RoleOwner  Role = "owner"
	RoleEditor Role = "editor"
	RoleViewer Role = "viewer"
)

var (
	ErrNotFound  = errors.New("workspace not found")
	ErrForbidden = errors.New("workspace access denied")
	ErrInvalid   = errors.New("invalid workspace")
	ErrPersonal  = errors.New("personal workspace cannot be removed")
	ErrExpired   = errors.New("workspace invitation expired")
)

type Workspace struct {
	ID                   uuid.UUID
	Kind                 Kind
	Name                 string
	CreatedAt, UpdatedAt time.Time
}

type Member struct {
	WorkspaceID, UserID uuid.UUID
	Role                Role
	CreatedAt           time.Time
}

type Invitation struct {
	ID, WorkspaceID, UserID uuid.UUID
	Email                   string
	Role                    Role
	Token                   string
	ExpiresAt, CreatedAt    time.Time
}

type Repository interface {
	Store(Workspace) error
	Find(uuid.UUID) (Workspace, error)
	ListByUser(uuid.UUID) ([]Workspace, error)
	Delete(uuid.UUID) error
	StoreMember(Member) error
	DeleteMember(uuid.UUID, uuid.UUID) error
	FindMember(uuid.UUID, uuid.UUID) (Member, error)
	ListMembers(uuid.UUID) ([]Member, error)
	StoreInvitation(Invitation) error
	FindInvitationByToken(string) (Invitation, error)
	ListInvitations(uuid.UUID) ([]Invitation, error)
	DeleteInvitation(uuid.UUID, uuid.UUID) error
}

func validName(v string) bool {
	v = strings.TrimSpace(v)
	return v != "" && utf8.ValidString(v) && utf8.RuneCountInString(v) <= 120
}
