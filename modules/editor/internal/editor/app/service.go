package app

import (
	"context"
	"editor/internal/editor/domain/model"
	domainservice "editor/internal/editor/domain/service"
	"errors"
	"time"
	"uuid"

	domainevent "github.com/distributed-programming-2026/lib/event"
)

type RepositoryProvider interface {
	Document(context.Context) model.DocumentRepository
	ModuleShare(context.Context) model.ModuleShareRepository
	Attachment(context.Context) model.AttachmentRepository
}

type UnitOfWork interface {
	Execute(context.Context, func(RepositoryProvider) error) error
}

type DispatcherFactory interface {
	NewDispatcher(context.Context) domainevent.Dispatcher
}

type User struct {
	ID                 uuid.UUID
	Email, DisplayName string
}

type UserClient interface {
	FindByEmail(context.Context, string) (User, error)
	Get(context.Context, uuid.UUID) (User, error)
}

type StoredFile struct {
	ID                    uuid.UUID
	FileName, ContentType string
	SizeBytes             int64
}

type FileStorage interface {
	Upload(context.Context, uuid.UUID, uuid.UUID, string, string, []byte) (StoredFile, error)
	Download(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) (StoredFile, []byte, error)
	Delete(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) error
}

type Permission struct {
	UserID             uuid.UUID
	Email, DisplayName string
	Role               model.Role
	UpdatedAt          time.Time
}

type EditorService interface {
	Create(ctx context.Context, actor, workspace uuid.UUID, title, content string) (out model.Document, err error)
	Get(ctx context.Context, actor, workspace, id uuid.UUID) (out model.Document, err error)
	List(ctx context.Context, actor, workspace uuid.UUID) (out []model.Document, err error)
	Update(ctx context.Context, actor, workspace, id uuid.UUID, title, content string, version int64) (out model.Document, err error)
	Delete(ctx context.Context, userID, workspace, id uuid.UUID) error
	Grant(ctx context.Context, userID, workspace, id uuid.UUID, email string, role model.Role) error
	Revoke(ctx context.Context, actor, workspace, id, userID uuid.UUID) error
	Permissions(ctx context.Context, actor, workspace, id uuid.UUID) ([]Permission, error)
	CreateModuleShare(ctx context.Context, actor, groupID uuid.UUID) (model.ModuleShare, error)
	GetModuleShare(ctx context.Context, actor, shareID uuid.UUID) (model.ModuleShare, error)
	ListModuleShares(ctx context.Context, actor uuid.UUID) ([]model.ModuleShare, error)
	DeleteModuleShare(ctx context.Context, actor, shareID uuid.UUID) (model.ModuleShare, error)
	ActivateModuleShare(ctx context.Context, shareID uuid.UUID) error
	RejectModuleShare(ctx context.Context, shareID uuid.UUID, reason string) error
	CompleteModuleShareRevocation(ctx context.Context, shareID uuid.UUID, reason string) error
	RecoverModuleShares(ctx context.Context, now time.Time, attemptTimeout time.Duration, maxAttempts int) error
	UploadAttachment(context.Context, uuid.UUID, uuid.UUID, uuid.UUID, string, string, []byte) (model.Attachment, error)
	ListAttachments(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) ([]model.Attachment, error)
	DownloadAttachment(context.Context, uuid.UUID, uuid.UUID, uuid.UUID, uuid.UUID) (model.Attachment, []byte, error)
	DeleteAttachment(context.Context, uuid.UUID, uuid.UUID, uuid.UUID, uuid.UUID) error
}

func NewService(
	unit UnitOfWork,
	dispatch DispatcherFactory,
	userAdapter UserClient,
	moduleID uuid.UUID,
	files ...FileStorage,
) EditorService {
	s := &service{
		unit:        unit,
		dispatch:    dispatch,
		userAdapter: userAdapter,
		moduleID:    moduleID,
	}
	if len(files) > 0 {
		s.files = files[0]
	}
	return s
}

type service struct {
	unit        UnitOfWork
	dispatch    DispatcherFactory
	userAdapter UserClient
	moduleID    uuid.UUID
	files       FileStorage
}

func (s *service) UploadAttachment(ctx context.Context, actor, workspace, document uuid.UUID, name, contentType string, data []byte) (out model.Attachment, err error) {
	if s.files == nil {
		return out, errors.New("file storage unavailable")
	}
	var doc model.Document
	err = s.unit.Execute(ctx, func(p RepositoryProvider) error {
		doc, err = s.domainService(ctx, p).Get(actor, workspace, document)
		return err
	})
	if err != nil {
		return
	}
	if doc.Role != model.RoleOwner && doc.Role != model.RoleEditor {
		return out, model.ErrPermissionDenied
	}
	file, err := s.files.Upload(ctx, workspace, actor, name, contentType, data)
	if err != nil {
		return out, err
	}
	out = model.Attachment{DocumentID: document, FileID: file.ID, WorkspaceID: workspace, FileName: file.FileName, ContentType: file.ContentType, SizeBytes: file.SizeBytes, CreatedAt: time.Now().UTC()}
	err = s.unit.Execute(ctx, func(p RepositoryProvider) error { return p.Attachment(ctx).StoreAttachment(out) })
	if err != nil {
		_ = s.files.Delete(context.WithoutCancel(ctx), workspace, actor, file.ID)
	}
	return
}

func (s *service) ListAttachments(ctx context.Context, actor, workspace, document uuid.UUID) (out []model.Attachment, err error) {
	err = s.unit.Execute(ctx, func(p RepositoryProvider) error {
		if _, e := s.domainService(ctx, p).Get(actor, workspace, document); e != nil {
			return e
		}
		out, err = p.Attachment(ctx).ListAttachments(document)
		return err
	})
	return
}

func (s *service) DownloadAttachment(ctx context.Context, actor, workspace, document, file uuid.UUID) (out model.Attachment, data []byte, err error) {
	err = s.unit.Execute(ctx, func(p RepositoryProvider) error {
		if _, e := s.domainService(ctx, p).Get(actor, workspace, document); e != nil {
			return e
		}
		out, err = p.Attachment(ctx).FindAttachment(document, file)
		return err
	})
	if err != nil {
		return
	}
	_, data, err = s.files.Download(ctx, out.WorkspaceID, actor, file)
	return
}

func (s *service) DeleteAttachment(ctx context.Context, actor, workspace, document, file uuid.UUID) error {
	var attachment model.Attachment
	err := s.unit.Execute(ctx, func(p RepositoryProvider) error {
		doc, e := s.domainService(ctx, p).Get(actor, workspace, document)
		if e != nil {
			return e
		}
		if doc.Role != model.RoleOwner && doc.Role != model.RoleEditor {
			return model.ErrPermissionDenied
		}
		attachment, e = p.Attachment(ctx).FindAttachment(document, file)
		return e
	})
	if err != nil {
		return err
	}
	if err = s.files.Delete(ctx, attachment.WorkspaceID, actor, file); err != nil {
		return err
	}
	return s.unit.Execute(ctx, func(p RepositoryProvider) error { return p.Attachment(ctx).DeleteAttachment(document, file) })
}

func (s *service) CreateModuleShare(
	ctx context.Context,
	actor,
	groupID uuid.UUID,
) (share model.ModuleShare, err error) {
	err = s.unit.Execute(ctx, func(p RepositoryProvider) error {
		share, err = domainservice.NewModuleShareService(
			p.ModuleShare(ctx),
			s.dispatch.NewDispatcher(ctx),
		).Create(s.moduleID, groupID, actor)
		return err
	})
	return
}

func (s *service) GetModuleShare(
	ctx context.Context,
	actor,
	shareID uuid.UUID,
) (share model.ModuleShare, err error) {
	err = s.unit.Execute(ctx, func(p RepositoryProvider) error {
		share, err = domainservice.NewModuleShareService(p.ModuleShare(ctx), s.dispatch.NewDispatcher(ctx)).Get(shareID)
		if err == nil && share.OwnerUserID != actor {
			return model.ErrShareNotFound
		}
		return err
	})
	return
}

func (s *service) ListModuleShares(ctx context.Context, actor uuid.UUID) (out []model.ModuleShare, err error) {
	err = s.unit.Execute(ctx, func(p RepositoryProvider) error {
		out, err = domainservice.NewModuleShareService(
			p.ModuleShare(ctx),
			s.dispatch.NewDispatcher(ctx),
		).List(actor)
		return err
	})
	return
}

func (s *service) DeleteModuleShare(ctx context.Context, actor, shareID uuid.UUID) (out model.ModuleShare, err error) {
	err = s.unit.Execute(ctx, func(p RepositoryProvider) error {
		shareService := domainservice.NewModuleShareService(p.ModuleShare(ctx), s.dispatch.NewDispatcher(ctx))
		share, findErr := shareService.Get(shareID)
		if findErr != nil {
			return findErr
		}
		if share.OwnerUserID != actor {
			return model.ErrShareNotFound
		}
		out, err = shareService.Cancel(shareID, "revoked_by_owner")
		return err
	})
	return
}

func (s *service) ActivateModuleShare(ctx context.Context, shareID uuid.UUID) error {
	return s.unit.Execute(ctx, func(p RepositoryProvider) error {
		return domainservice.NewModuleShareService(p.ModuleShare(ctx), s.dispatch.NewDispatcher(ctx)).Activate(shareID)
	})
}

func (s *service) RejectModuleShare(ctx context.Context, shareID uuid.UUID, reason string) error {
	return s.unit.Execute(ctx, func(p RepositoryProvider) error {
		return domainservice.NewModuleShareService(p.ModuleShare(ctx), s.dispatch.NewDispatcher(ctx)).Reject(shareID, reason)
	})
}

func (s *service) CompleteModuleShareRevocation(ctx context.Context, shareID uuid.UUID, reason string) error {
	return s.unit.Execute(ctx, func(p RepositoryProvider) error {
		return domainservice.NewModuleShareService(p.ModuleShare(ctx), s.dispatch.NewDispatcher(ctx)).Revoked(shareID, reason)
	})
}

func (s *service) RecoverModuleShares(ctx context.Context, now time.Time, attemptTimeout time.Duration, maxAttempts int) error {
	return s.unit.Execute(ctx, func(p RepositoryProvider) error {
		return domainservice.NewModuleShareService(p.ModuleShare(ctx), s.dispatch.NewDispatcher(ctx)).Recover(now, attemptTimeout, maxAttempts)
	})
}

func (s *service) Create(ctx context.Context, actor, workspace uuid.UUID, title, content string) (document model.Document, err error) {
	err = s.unit.Execute(ctx, func(p RepositoryProvider) error {
		document, err = s.domainService(ctx, p).Create(actor, workspace, title, content)
		return err
	})
	return
}

func (s *service) Get(ctx context.Context, actor, workspace, id uuid.UUID) (document model.Document, err error) {
	err = s.unit.Execute(ctx, func(p RepositoryProvider) error {
		document, err = s.domainService(ctx, p).Get(actor, workspace, id)
		return err
	})
	return
}

func (s *service) List(ctx context.Context, actor, workspace uuid.UUID) (documents []model.Document, err error) {
	err = s.unit.Execute(ctx, func(p RepositoryProvider) error {
		documents, err = s.domainService(ctx, p).List(actor, workspace)
		return err
	})
	return
}

func (s *service) Update(ctx context.Context, actor, workspace, id uuid.UUID, title, content string, version int64) (document model.Document, err error) {
	err = s.unit.Execute(ctx, func(p RepositoryProvider) error {
		document, err = s.domainService(ctx, p).Update(actor, workspace, id, title, content, version)
		return err
	})
	return
}

func (s *service) Delete(ctx context.Context, actor, workspace, id uuid.UUID) error {
	return s.unit.Execute(ctx, func(p RepositoryProvider) error {
		return s.domainService(ctx, p).Delete(actor, workspace, id)
	})
}

func (s *service) Grant(ctx context.Context, actor, workspace, id uuid.UUID, email string, role model.Role) error {
	user, err := s.userAdapter.FindByEmail(ctx, email)
	if err != nil {
		return err
	}
	return s.unit.Execute(ctx, func(p RepositoryProvider) error {
		return s.domainService(ctx, p).Grant(actor, workspace, id, user.ID, role)
	})
}

func (s *service) Revoke(ctx context.Context, actor, workspace, id, userID uuid.UUID) error {
	return s.unit.Execute(ctx, func(p RepositoryProvider) error {
		return s.domainService(ctx, p).Revoke(actor, workspace, id, userID)
	})
}

func (s *service) Permissions(ctx context.Context, actor, workspace, id uuid.UUID) ([]Permission, error) {
	var values []model.Permission
	err := s.unit.Execute(ctx, func(p RepositoryProvider) error {
		var listErr error
		values, listErr = s.domainService(ctx, p).Permissions(actor, workspace, id)
		return listErr
	})
	if err != nil {
		return nil, err
	}
	result := make([]Permission, 0, len(values))
	for _, value := range values {
		user, getErr := s.userAdapter.Get(ctx, value.UserID)
		if getErr != nil {
			return nil, getErr
		}
		result = append(result, Permission{
			UserID:      value.UserID,
			Email:       user.Email,
			DisplayName: user.DisplayName,
			Role:        value.Role,
			UpdatedAt:   value.UpdatedAt,
		})
	}
	return result, nil
}

func (s *service) domainService(ctx context.Context, p RepositoryProvider) domainservice.DocumentService {
	repository := p.Document(ctx)
	return domainservice.NewService(repository, s.dispatch.NewDispatcher(ctx), func(actor, document uuid.UUID) (model.Role, error) {
		role, err := repository.FindRole(document, actor)
		if errors.Is(err, model.ErrDocumentNotFound) {
			return "", nil
		}
		return role, err
	})
}
