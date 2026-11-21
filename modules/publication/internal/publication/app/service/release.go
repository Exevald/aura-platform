package service

import (
	"context"
	"errors"
	"publication/internal/publication/app/adapter"

	workflowapi "publication/api/workflow"
	"publication/internal/publication/app/model"
	"uuid"

	"go.temporal.io/api/serviceerror"
)

var ErrInvalidRequest = errors.New("invalid publication request")
var ErrIdempotencyConflict = errors.New("idempotency key belongs to another request")
var ErrNotFound = errors.New("publication not found")
var ErrConflict = errors.New("publication was changed concurrently")

type ReleaseStorage interface {
	Get(context.Context, uuid.UUID) (model.Release, error)
	Store(context.Context, model.Release) (model.Release, error)
	Delete(context.Context, uuid.UUID) error
}

type AttachmentStorage interface {
	Get(context.Context, uuid.UUID, uuid.UUID) (model.Attachment, error)
	Store(context.Context, model.Attachment) error
	Delete(context.Context, uuid.UUID) error
}

type ReleaseQueryService interface {
	List(context.Context, uuid.UUID, uuid.UUID) ([]model.Release, error)
	ByDocument(context.Context, uuid.UUID) ([]model.Release, error)
	Pending(context.Context) ([]model.Release, error)
}

type AttachmentQueryService interface {
	List(context.Context, uuid.UUID) ([]model.Attachment, error)
}

type Releases struct {
	releaseStorage            ReleaseStorage
	attachmentStorage         AttachmentStorage
	releaseQueryService       ReleaseQueryService
	attachmentQueryService    AttachmentQueryService
	publicationWorkflowClient workflowapi.PublicationFlowClient
	filestorageService        adapter.FilestorageService
}

func NewReleases(
	releaseStorage ReleaseStorage,
	attachmentStorage AttachmentStorage,
	releaseQueryService ReleaseQueryService,
	attachmentQueryService AttachmentQueryService,
	publicationWorkflowClient workflowapi.PublicationFlowClient,
	filestorageService adapter.FilestorageService,
) *Releases {
	return &Releases{
		releaseStorage:            releaseStorage,
		attachmentStorage:         attachmentStorage,
		releaseQueryService:       releaseQueryService,
		attachmentQueryService:    attachmentQueryService,
		publicationWorkflowClient: publicationWorkflowClient,
		filestorageService:        filestorageService,
	}
}

func (s *Releases) Create(ctx context.Context, owner, workspace, document uuid.UUID, key string, version int64) (model.Release, error) {
	if owner == uuid.Nil() || workspace == uuid.Nil() || document == uuid.Nil() || key == "" || len(key) > 255 || version < 1 {
		return model.Release{}, ErrInvalidRequest
	}
	r, err := s.releaseStorage.Store(ctx, model.Release{ID: uuid.New(), OwnerID: owner, WorkspaceID: workspace, DocumentID: document, IdempotencyKey: key, ExpectedVersion: version, Status: model.ReleasePending})
	if err != nil {
		return r, err
	}
	if r.DocumentID != document || r.WorkspaceID != workspace || r.ExpectedVersion != version {
		return model.Release{}, ErrIdempotencyConflict
	}
	if r.Status == model.ReleasePending {
		if err = s.startPublish(ctx, r.ID); err != nil {
			return r, err
		}
	}
	return r, nil
}

func (s *Releases) Get(ctx context.Context, owner, workspace, id uuid.UUID) (model.Release, error) {
	r, err := s.releaseStorage.Get(ctx, id)
	if err != nil {
		return r, err
	}
	if r.OwnerID != owner || r.WorkspaceID != workspace {
		return model.Release{}, ErrNotFound
	}
	return r, nil
}

func (s *Releases) List(ctx context.Context, owner, workspace uuid.UUID) ([]model.Release, error) {
	return s.releaseQueryService.List(ctx, owner, workspace)
}

func (s *Releases) Public(ctx context.Context, id uuid.UUID) (model.Release, []model.Attachment, error) {
	r, err := s.releaseStorage.Get(ctx, id)
	if err != nil {
		return model.Release{}, nil, err
	}
	if r.Status != model.ReleaseActive {
		return model.Release{}, nil, ErrNotFound
	}
	attachments, err := s.attachmentQueryService.List(ctx, id)
	return r, attachments, err
}

func (s *Releases) DownloadAttachment(ctx context.Context, id, file uuid.UUID) (model.Attachment, []byte, error) {
	r, err := s.releaseStorage.Get(ctx, id)
	if err != nil {
		return model.Attachment{}, nil, err
	}
	if r.Status != model.ReleaseActive {
		return model.Attachment{}, nil, ErrNotFound
	}
	attachment, err := s.attachmentStorage.Get(ctx, id, file)
	if err != nil {
		return model.Attachment{}, nil, err
	}
	data, err := s.filestorageService.Download(ctx, r.WorkspaceID, attachment.CopyID)
	return attachment, data, err
}

func (s *Releases) Close(ctx context.Context, owner, workspace, id uuid.UUID) error {
	for range 3 {
		r, err := s.Get(ctx, owner, workspace, id)
		if err != nil {
			return err
		}
		if r.Status == model.ReleaseClosed {
			return nil
		}
		if r.Status == model.ReleasePending || r.Status == model.ReleaseActive {
			r.Status = model.ReleaseClosing
			_, err = s.releaseStorage.Store(ctx, r)
			if errors.Is(err, ErrConflict) {
				continue
			}
			if err != nil {
				return err
			}
		}
		return s.startUnpublish(ctx, id)
	}
	return ErrConflict
}

func (s *Releases) CloseByDocument(ctx context.Context, document uuid.UUID) error {
	items, err := s.releaseQueryService.ByDocument(ctx, document)
	if err != nil {
		return err
	}
	for _, r := range items {
		if r.Status == model.ReleasePending || r.Status == model.ReleaseActive {
			r.Status = model.ReleaseClosing
			if _, err = s.releaseStorage.Store(ctx, r); err != nil && !errors.Is(err, ErrConflict) {
				return err
			}
		}
		if err = s.startUnpublish(ctx, r.ID); err != nil {
			return err
		}
	}
	return nil
}

func (s *Releases) Recover(ctx context.Context) error {
	items, err := s.releaseQueryService.Pending(ctx)
	if err != nil {
		return err
	}
	for _, r := range items {
		if r.Status == model.ReleasePending {
			err = s.startPublish(ctx, r.ID)
		} else {
			err = s.startUnpublish(ctx, r.ID)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Releases) startPublish(ctx context.Context, id uuid.UUID) error {
	_, err := s.publicationWorkflowClient.PublishAsync(ctx, &workflowapi.PublishRequest{Id: id.String()})
	if _, ok := errors.AsType[*serviceerror.WorkflowExecutionAlreadyStarted](err); ok {
		return nil
	}
	return err
}

func (s *Releases) startUnpublish(ctx context.Context, id uuid.UUID) error {
	_, err := s.publicationWorkflowClient.UnpublishAsync(ctx, &workflowapi.UnpublishRequest{Id: id.String()})
	if _, ok := errors.AsType[*serviceerror.WorkflowExecutionAlreadyStarted](err); ok {
		return nil
	}
	return err
}
