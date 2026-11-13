package service

import (
	"editor/internal/editor/domain/model"
	"time"
	"uuid"

	"github.com/distributed-programming-2026/lib/event"

	editorevents "editor/api/events"
)

type DocumentService interface {
	Create(owner, workspace uuid.UUID, title, content string) (model.Document, error)
	Get(actor, workspace, id uuid.UUID) (model.Document, error)
	List(actor, workspace uuid.UUID) ([]model.Document, error)
	Update(actor, workspace, id uuid.UUID, title, content string, expectedVersion int64) (model.Document, error)
	Delete(actor, workspace, id uuid.UUID) error
	Grant(actor, workspace, documentID, userID uuid.UUID, role model.Role) error
	Revoke(actor, workspace, documentID, userID uuid.UUID) error
	Permissions(actor, workspace, documentID uuid.UUID) ([]model.Permission, error)
}

type service struct {
	repo        model.DocumentRepository
	dispatcher  event.Dispatcher
	resolveRole func(uuid.UUID, uuid.UUID) (model.Role, error)
}

func NewService(repo model.DocumentRepository, dispatcher event.Dispatcher, resolveRole func(uuid.UUID, uuid.UUID) (model.Role, error)) DocumentService {
	return &service{repo: repo, dispatcher: dispatcher, resolveRole: resolveRole}
}

func (s *service) Create(owner, workspace uuid.UUID, title, content string) (model.Document, error) {
	if owner == uuid.Nil() || workspace == uuid.Nil() || !model.ValidDocument(title, content) {
		return model.Document{}, model.ErrInvalidDocument
	}
	id, err := s.repo.NextID()
	if err != nil {
		return model.Document{}, err
	}

	now := time.Now().UTC()
	doc := model.Document{
		ID:          id,
		OwnerUserID: owner,
		WorkspaceID: workspace,
		Title:       title,
		Content:     content,
		Version:     1,
		CreatedAt:   now,
		UpdatedAt:   now,
		Role:        model.RoleOwner,
	}
	if err = s.repo.Create(doc); err != nil {
		return model.Document{}, err
	}
	if err = s.dispatchEvent("created", doc, owner); err != nil {
		return model.Document{}, err
	}
	return doc, nil
}

func (s *service) Get(actor, workspace, id uuid.UUID) (model.Document, error) {
	doc, role, err := s.authorized(actor, workspace, id)
	if err != nil {
		return model.Document{}, err
	}
	doc.Role = role
	return doc, nil
}

func (s *service) List(actor, workspace uuid.UUID) ([]model.Document, error) {
	if actor == uuid.Nil() || workspace == uuid.Nil() {
		return nil, model.ErrPermissionDenied
	}
	return s.repo.ListForUser(actor, workspace)
}

func (s *service) Update(actor, workspace, id uuid.UUID, title, content string, expectedVersion int64) (model.Document, error) {
	if expectedVersion < 1 || !model.ValidDocument(title, content) {
		return model.Document{}, model.ErrInvalidDocument
	}
	doc, role, err := s.authorized(actor, workspace, id)
	if err != nil {
		return model.Document{}, err
	}
	if role != model.RoleOwner && role != model.RoleEditor {
		return model.Document{}, model.ErrPermissionDenied
	}
	if doc.Version != expectedVersion {
		return model.Document{}, model.ErrVersionConflict
	}
	doc.Title, doc.Content, doc.Version, doc.UpdatedAt, doc.Role = title, content, expectedVersion+1, time.Now().UTC(), role
	if err = s.repo.Update(doc, expectedVersion); err != nil {
		return model.Document{}, err
	}
	if err = s.dispatchEvent("updated", doc, actor); err != nil {
		return model.Document{}, err
	}
	return doc, nil
}

func (s *service) Delete(actor, workspace, id uuid.UUID) error {
	doc, role, err := s.authorized(actor, workspace, id)
	if err != nil {
		return err
	}
	if role != model.RoleOwner {
		return model.ErrPermissionDenied
	}
	if err = s.repo.Delete(id); err != nil {
		return err
	}
	return s.dispatchEvent("deleted", doc, actor)
}

func (s *service) Grant(actor, workspace, documentID, userID uuid.UUID, role model.Role) error {
	doc, actorRole, err := s.authorized(actor, workspace, documentID)
	if err != nil {
		return err
	}
	if actorRole != model.RoleOwner {
		return model.ErrPermissionDenied
	}
	if userID == uuid.Nil() || userID == doc.OwnerUserID || (role != model.RoleViewer && role != model.RoleEditor) {
		return model.ErrInvalidPermission
	}
	now := time.Now().UTC()
	if err = s.repo.StorePermission(model.Permission{
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

func (s *service) Revoke(actor, workspace, documentID, userID uuid.UUID) error {
	_, role, err := s.authorized(actor, workspace, documentID)
	if err != nil {
		return err
	}
	if role != model.RoleOwner {
		return model.ErrPermissionDenied
	}
	if actor == userID {
		return model.ErrInvalidPermission
	}
	return s.repo.DeletePermission(documentID, userID)
}

func (s *service) Permissions(actor, workspace, documentID uuid.UUID) ([]model.Permission, error) {
	_, role, err := s.authorized(actor, workspace, documentID)
	if err != nil {
		return nil, err
	}
	if role != model.RoleOwner {
		return nil, model.ErrPermissionDenied
	}
	return s.repo.ListPermissions(documentID)
}

func (s *service) authorized(actor, workspace, id uuid.UUID) (model.Document, model.Role, error) {
	if actor == uuid.Nil() || workspace == uuid.Nil() || id == uuid.Nil() {
		return model.Document{}, "", model.ErrPermissionDenied
	}
	doc, err := s.repo.Find(id)
	if err != nil {
		return model.Document{}, "", err
	}
	if doc.WorkspaceID != workspace {
		return model.Document{}, "", model.ErrDocumentNotFound
	}
	if doc.OwnerUserID == actor {
		return doc, model.RoleOwner, nil
	}
	role, err := s.resolveRole(actor, id)
	if err != nil || (role != model.RoleViewer && role != model.RoleEditor) {
		if err != nil {
			return model.Document{}, "", err
		}
		return model.Document{}, "", model.ErrPermissionDenied
	}
	return doc, role, err
}

func (s *service) dispatchEvent(kind string, doc model.Document, actor uuid.UUID) error {
	return s.dispatcher.Dispatch(
		&editorevents.DocumentChanged{
			Type:           kind,
			DocumentId:     doc.ID.String(),
			ActorUserId:    actor.String(),
			Version:        doc.Version,
			OccurredAtUnix: time.Now().UTC().Unix(),
		})
}
