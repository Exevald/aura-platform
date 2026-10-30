package domain

import (
	"errors"
	"time"
	"uuid"

	"github.com/distributed-programming-2026/lib/event"

	editorevents "editor/api/events"
)

type DocumentService interface {
	Create(owner uuid.UUID, title, content string) (Document, error)
	Get(actor, id uuid.UUID) (Document, error)
	List(actor uuid.UUID) ([]Document, error)
	Update(actor, id uuid.UUID, title, content string, expectedVersion int64) (Document, error)
	Delete(actor, id uuid.UUID) error
	Grant(actor, documentID, userID uuid.UUID, role Role) error
	Revoke(actor, documentID, userID uuid.UUID) error
	Permissions(actor, documentID uuid.UUID) ([]Permission, error)
}

func NewService(repo DocumentRepository, dispatcher event.Dispatcher) DocumentService {
	return &service{repo: repo, dispatcher: dispatcher}
}

type service struct {
	repo       DocumentRepository
	dispatcher event.Dispatcher
}

func (s *service) Create(owner uuid.UUID, title, content string) (Document, error) {
	if owner == uuid.Nil() || !ValidDocument(title, content) {
		return Document{}, ErrInvalidDocument
	}
	id, err := s.repo.NextID()
	if err != nil {
		return Document{}, err
	}

	now := time.Now().UTC()
	doc := Document{
		ID:          id,
		OwnerUserID: owner,
		Title:       title,
		Content:     content,
		Version:     1,
		CreatedAt:   now,
		UpdatedAt:   now,
		Role:        RoleOwner,
	}
	if err = s.repo.Create(doc); err != nil {
		return Document{}, err
	}
	if err = s.dispatchEvent("created", doc, owner); err != nil {
		return Document{}, err
	}
	return doc, nil
}

func (s *service) Get(actor, id uuid.UUID) (Document, error) {
	doc, role, err := s.authorized(actor, id)
	if err != nil {
		return Document{}, err
	}
	doc.Role = role
	return doc, nil
}

func (s *service) List(actor uuid.UUID) ([]Document, error) {
	if actor == uuid.Nil() {
		return nil, ErrPermissionDenied
	}
	return s.repo.ListForUser(actor)
}

func (s *service) Update(actor, id uuid.UUID, title, content string, expectedVersion int64) (Document, error) {
	if expectedVersion < 1 || !ValidDocument(title, content) {
		return Document{}, ErrInvalidDocument
	}
	doc, role, err := s.authorized(actor, id)
	if err != nil {
		return Document{}, err
	}
	if role != RoleOwner && role != RoleEditor {
		return Document{}, ErrPermissionDenied
	}
	if doc.Version != expectedVersion {
		return Document{}, ErrVersionConflict
	}
	doc.Title, doc.Content, doc.Version, doc.UpdatedAt, doc.Role = title, content, expectedVersion+1, time.Now().UTC(), role
	if err = s.repo.Update(doc, expectedVersion); err != nil {
		return Document{}, err
	}
	if err = s.dispatchEvent("updated", doc, actor); err != nil {
		return Document{}, err
	}
	return doc, nil
}

func (s *service) Delete(actor, id uuid.UUID) error {
	doc, role, err := s.authorized(actor, id)
	if err != nil {
		return err
	}
	if role != RoleOwner {
		return ErrPermissionDenied
	}
	if err = s.repo.Delete(id); err != nil {
		return err
	}
	return s.dispatchEvent("deleted", doc, actor)
}

func (s *service) Grant(actor, documentID, userID uuid.UUID, role Role) error {
	doc, actorRole, err := s.authorized(actor, documentID)
	if err != nil {
		return err
	}
	if actorRole != RoleOwner {
		return ErrPermissionDenied
	}
	if userID == uuid.Nil() || userID == doc.OwnerUserID || (role != RoleViewer && role != RoleEditor) {
		return ErrInvalidPermission
	}
	now := time.Now().UTC()
	if err = s.repo.StorePermission(Permission{
		DocumentID: documentID,
		UserID:     userID,
		Role:       role,
		CreatedAt:  now,
		UpdatedAt:  now,
	}); err != nil {
		return err
	}
	return s.dispatchEvent("permission_changed", doc, actor)
}

func (s *service) Revoke(actor, documentID, userID uuid.UUID) error {
	_, role, err := s.authorized(actor, documentID)
	if err != nil {
		return err
	}
	if role != RoleOwner {
		return ErrPermissionDenied
	}
	if actor == userID {
		return ErrInvalidPermission
	}
	return s.repo.DeletePermission(documentID, userID)
}

func (s *service) Permissions(actor, documentID uuid.UUID) ([]Permission, error) {
	_, role, err := s.authorized(actor, documentID)
	if err != nil {
		return nil, err
	}
	if role != RoleOwner {
		return nil, ErrPermissionDenied
	}
	return s.repo.ListPermissions(documentID)
}

func (s *service) authorized(actor, id uuid.UUID) (Document, Role, error) {
	if actor == uuid.Nil() || id == uuid.Nil() {
		return Document{}, "", ErrPermissionDenied
	}
	doc, err := s.repo.Find(id)
	if err != nil {
		return Document{}, "", err
	}
	if doc.OwnerUserID == actor {
		return doc, RoleOwner, nil
	}
	role, err := s.repo.FindRole(id, actor)
	if errors.Is(err, ErrDocumentNotFound) {
		return Document{}, "", ErrPermissionDenied
	}
	return doc, role, err
}

func (s *service) dispatchEvent(kind string, doc Document, actor uuid.UUID) error {
	return s.dispatcher.Dispatch(
		&editorevents.DocumentChanged{
			Type:           kind,
			DocumentId:     doc.ID.String(),
			ActorUserId:    actor.String(),
			Version:        doc.Version,
			OccurredAtUnix: time.Now().UTC().Unix(),
		})
}
