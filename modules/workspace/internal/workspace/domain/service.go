package domain

import (
	"crypto/rand"
	"encoding/hex"
	"net/mail"
	"strings"
	"time"
	"uuid"
	workspaceevents "workspace/api/events"

	"github.com/distributed-programming-2026/lib/event"
)

type Service struct {
	repo   Repository
	events event.Dispatcher
}

func NewService(repo Repository, events event.Dispatcher) *Service {
	return &Service{repo: repo, events: events}
}

func (s *Service) EnsurePersonal(id, owner uuid.UUID, name string) (Workspace, error) {
	if id == uuid.Nil() || owner == uuid.Nil() || !validName(name) {
		return Workspace{}, ErrInvalid
	}
	if existing, err := s.repo.Find(id); err == nil {
		return existing, nil
	}
	return s.create(id, owner, KindPersonal, name)
}

func (s *Service) CreateTeam(owner uuid.UUID, name string) (Workspace, error) {
	if owner == uuid.Nil() || !validName(name) {
		return Workspace{}, ErrInvalid
	}
	return s.create(uuid.New(), owner, KindTeam, name)
}

func (s *Service) create(id, owner uuid.UUID, kind Kind, name string) (Workspace, error) {
	now := time.Now().UTC()
	value := Workspace{ID: id, Kind: kind, Name: strings.TrimSpace(name), CreatedAt: now, UpdatedAt: now}
	if err := s.repo.Store(value); err != nil {
		return Workspace{}, err
	}
	if err := s.repo.StoreMember(Member{WorkspaceID: id, UserID: owner, Role: RoleOwner, CreatedAt: now}); err != nil {
		return Workspace{}, err
	}
	if err := s.events.Dispatch(&workspaceevents.WorkspaceCreated{WorkspaceId: id.String(), OwnerUserId: owner.String(), Kind: string(kind), Name: value.Name, OccurredAtUnix: now.Unix()}); err != nil {
		return Workspace{}, err
	}
	return value, nil
}

func (s *Service) List(user uuid.UUID) ([]Workspace, error) {
	if user == uuid.Nil() {
		return nil, ErrInvalid
	}
	return s.repo.ListByUser(user)
}

func (s *Service) Get(user, id uuid.UUID) (Workspace, []Member, Role, error) {
	m, err := s.repo.FindMember(id, user)
	if err != nil {
		return Workspace{}, nil, "", ErrNotFound
	}
	w, err := s.repo.Find(id)
	if err != nil {
		return Workspace{}, nil, "", err
	}
	members, err := s.repo.ListMembers(id)
	return w, members, m.Role, err
}

func (s *Service) Check(user, id uuid.UUID) (Member, Workspace, error) {
	m, err := s.repo.FindMember(id, user)
	if err != nil {
		return Member{}, Workspace{}, ErrNotFound
	}
	w, err := s.repo.Find(id)
	return m, w, err
}

func (s *Service) PutMember(actor, id, user uuid.UUID, role Role) error {
	owner, _, err := s.Check(actor, id)
	if err != nil {
		return err
	}
	if owner.Role != RoleOwner {
		return ErrForbidden
	}
	if user == uuid.Nil() || (role != RoleEditor && role != RoleViewer) {
		return ErrInvalid
	}
	now := time.Now().UTC()
	if err = s.repo.StoreMember(Member{WorkspaceID: id, UserID: user, Role: role, CreatedAt: now}); err != nil {
		return err
	}
	return s.events.Dispatch(&workspaceevents.WorkspaceMembershipChanged{WorkspaceId: id.String(), UserId: user.String(), Role: string(role), Active: true, OccurredAtUnix: now.Unix()})
}

func (s *Service) RemoveMember(actor, id, user uuid.UUID) error {
	owner, _, err := s.Check(actor, id)
	if err != nil {
		return err
	}
	if owner.Role != RoleOwner {
		return ErrForbidden
	}
	if actor == user {
		return ErrInvalid
	}
	if err = s.repo.DeleteMember(id, user); err != nil {
		return err
	}
	now := time.Now().UTC()
	return s.events.Dispatch(&workspaceevents.WorkspaceMembershipChanged{WorkspaceId: id.String(), UserId: user.String(), Active: false, OccurredAtUnix: now.Unix()})
}

func (s *Service) Delete(actor, id uuid.UUID) error {
	m, w, err := s.Check(actor, id)
	if err != nil {
		return err
	}
	if m.Role != RoleOwner {
		return ErrForbidden
	}
	if w.Kind == KindPersonal {
		return ErrPersonal
	}
	if err = s.repo.Delete(id); err != nil {
		return err
	}
	now := time.Now().UTC()
	return s.events.Dispatch(&workspaceevents.WorkspaceDeleted{WorkspaceId: id.String(), OwnerUserId: actor.String(), OccurredAtUnix: now.Unix()})
}

func (s *Service) CreateInvitation(actor, workspace, user uuid.UUID, email string, role Role) (Invitation, error) {
	member, _, err := s.Check(actor, workspace)
	if err != nil {
		return Invitation{}, err
	}
	if member.Role != RoleOwner {
		return Invitation{}, ErrForbidden
	}
	address, err := mail.ParseAddress(strings.TrimSpace(email))
	if err != nil || (role != RoleEditor && role != RoleViewer) {
		return Invitation{}, ErrInvalid
	}
	bytes := make([]byte, 32)
	if _, err = rand.Read(bytes); err != nil {
		return Invitation{}, err
	}
	now := time.Now().UTC()
	v := Invitation{ID: uuid.New(), WorkspaceID: workspace, UserID: user, Email: strings.ToLower(address.Address), Role: role, Token: hex.EncodeToString(bytes), ExpiresAt: now.Add(7 * 24 * time.Hour), CreatedAt: now}
	if err = s.repo.StoreInvitation(v); err != nil {
		return Invitation{}, err
	}
	err = s.events.Dispatch(&workspaceevents.WorkspaceInvitationChanged{InvitationId: v.ID.String(), WorkspaceId: workspace.String(), Email: v.Email, Role: string(role), Action: "created", OccurredAtUnix: now.Unix()})
	return v, err
}

func (s *Service) ListInvitations(actor, workspace uuid.UUID) ([]Invitation, error) {
	member, _, err := s.Check(actor, workspace)
	if err != nil || member.Role != RoleOwner {
		return nil, ErrNotFound
	}
	return s.repo.ListInvitations(workspace)
}

func (s *Service) AcceptInvitation(actor uuid.UUID, token string) (Workspace, error) {
	v, err := s.repo.FindInvitationByToken(token)
	if err != nil {
		return Workspace{}, ErrNotFound
	}
	if time.Now().UTC().After(v.ExpiresAt) {
		return Workspace{}, ErrExpired
	}
	if v.UserID != uuid.Nil() && v.UserID != actor {
		return Workspace{}, ErrNotFound
	}
	if err = s.repo.StoreMember(Member{WorkspaceID: v.WorkspaceID, UserID: actor, Role: v.Role, CreatedAt: time.Now().UTC()}); err != nil {
		return Workspace{}, err
	}
	if err = s.repo.DeleteInvitation(v.WorkspaceID, v.ID); err != nil {
		return Workspace{}, err
	}
	now := time.Now().UTC()
	if err = s.events.Dispatch(&workspaceevents.WorkspaceMembershipChanged{WorkspaceId: v.WorkspaceID.String(), UserId: actor.String(), Role: string(v.Role), Active: true, OccurredAtUnix: now.Unix()}); err != nil {
		return Workspace{}, err
	}
	return s.repo.Find(v.WorkspaceID)
}

func (s *Service) RevokeInvitation(actor, workspace, invitation uuid.UUID) error {
	member, _, err := s.Check(actor, workspace)
	if err != nil || member.Role != RoleOwner {
		return ErrNotFound
	}
	return s.repo.DeleteInvitation(workspace, invitation)
}
