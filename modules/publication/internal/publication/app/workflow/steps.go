package workflow

import (
	"context"
	"errors"
	"fmt"
	"publication/internal/publication/app/service"
	"slices"
	"strings"
	"time"
	"uuid"

	"publication/internal/publication/app/adapter"
	"publication/internal/publication/app/model"
)

type ReleaseSteps struct {
	ReleaseStorage         service.ReleaseStorage
	AttachmentStorage      service.AttachmentStorage
	AttachmentQueryService service.AttachmentQueryService
	FilestorageService     adapter.FilestorageService
	EditorService          adapter.EditorService
	ShortlinkService       adapter.ShortlinkService
	WorkspaceService       adapter.WorkspaceService
	NotificationService    adapter.NotificationService
	PublicBaseURL          string
}

func (s *ReleaseSteps) ExportSnapshot(ctx context.Context, id uuid.UUID) (int64, error) {
	r, err := s.ReleaseStorage.Get(ctx, id)
	if err != nil {
		return 0, err
	}
	if r.Status != model.ReleasePending {
		return 0, service.ErrConflict
	}
	snapshot, err := s.EditorService.Snapshot(ctx, r.DocumentID, r.WorkspaceID, r.OwnerID, r.ExpectedVersion)
	if err != nil {
		return 0, err
	}
	for _, file := range snapshot.Attachments {
		if _, existingErr := s.AttachmentStorage.Get(ctx, id, file.FileID); existingErr == nil {
			continue
		} else if !errors.Is(existingErr, service.ErrNotFound) {
			return 0, existingErr
		}
		data, err2 := s.EditorService.DownloadAttachment(ctx, r.DocumentID, r.WorkspaceID, r.OwnerID, file.FileID)
		if err2 != nil {
			return 0, err2
		}
		copyID, err2 := s.FilestorageService.Upload(ctx, r.WorkspaceID, r.OwnerID, file.FileName, file.ContentType, data)
		if err2 != nil {
			return 0, err2
		}
		if err2 = s.AttachmentStorage.Store(
			ctx,
			model.Attachment{
				ReleaseID:   id,
				FileID:      file.FileID,
				CopyID:      copyID,
				FileName:    file.FileName,
				ContentType: file.ContentType,
				SizeBytes:   file.SizeBytes,
			}); err2 != nil {
			_ = s.FilestorageService.Delete(context.WithoutCancel(ctx), r.WorkspaceID, r.OwnerID, copyID)
			return 0, err2
		}
	}
	latest, err := s.EditorService.Snapshot(ctx, r.DocumentID, r.WorkspaceID, r.OwnerID, r.ExpectedVersion)
	if err != nil {
		return 0, err
	}
	if !sameAttachments(snapshot.Attachments, latest.Attachments) {
		return 0, service.ErrConflict
	}
	content := snapshot.Content
	for _, file := range snapshot.Attachments {
		old := "/api/editor/v1/documents/" + r.DocumentID.String() + "/attachments/" + file.FileID.String() + "/content"
		content = strings.ReplaceAll(content, old, "/p/"+id.String()+"/assets/"+file.FileID.String())
	}
	r.Title, r.Content, r.SourceVersion = snapshot.Title, content, snapshot.Version
	_, err = s.ReleaseStorage.Store(ctx, r)

	return snapshot.Version, err
}

func (s *ReleaseSteps) CreateShortLink(ctx context.Context, id uuid.UUID) (linkID uuid.UUID, shortPath string, err error) {
	r, err := s.ReleaseStorage.Get(ctx, id)
	if err != nil {
		return uuid.Nil(), "", err
	}
	if r.Status != model.ReleasePending {
		return uuid.Nil(), "", service.ErrConflict
	}
	linkID, shortPath, err = s.ShortlinkService.Ensure(ctx, id, strings.TrimRight(s.PublicBaseURL, "/")+"/p/"+id.String())
	if err != nil {
		return uuid.Nil(), "", err
	}
	r.LinkID, r.ShortPath = linkID, shortPath
	_, err = s.ReleaseStorage.Store(ctx, r)
	return linkID, shortPath, err
}

func (s *ReleaseSteps) Activate(ctx context.Context, id uuid.UUID) error {
	r, err := s.ReleaseStorage.Get(ctx, id)
	if err != nil {
		return err
	}
	if r.Status == model.ReleaseActive {
		return nil
	}
	if r.Status != model.ReleasePending {
		return service.ErrConflict
	}
	r.Status = model.ReleaseActive
	r.PublishedAt = time.Now().UTC()
	_, err = s.ReleaseStorage.Store(ctx, r)
	return err
}

func (s *ReleaseSteps) NotifyMembers(ctx context.Context, id uuid.UUID) error {
	r, err := s.ReleaseStorage.Get(ctx, id)
	if err != nil {
		return err
	}
	if r.Status != model.ReleaseActive {
		return nil
	}
	members, err := s.WorkspaceService.Members(ctx, r.WorkspaceID, r.OwnerID)
	if err != nil {
		return err
	}
	for _, user := range members {
		if err = s.NotificationService.Create(ctx, "publication:"+r.ID.String()+":"+user.String(), user, r.WorkspaceID, "Опубликован документ: "+r.Title, strings.TrimRight(s.PublicBaseURL, "/")+r.ShortPath); err != nil {
			return err
		}
	}
	r.DeliveryStatus = "sent"
	_, err = s.ReleaseStorage.Store(ctx, r)
	return err
}

func (s *ReleaseSteps) MarkNotificationFailed(ctx context.Context, id uuid.UUID) error {
	r, err := s.ReleaseStorage.Get(ctx, id)
	if err != nil {
		return err
	}
	if r.Status != model.ReleaseActive {
		return nil
	}
	r.DeliveryStatus = "failed"
	_, err = s.ReleaseStorage.Store(ctx, r)
	return err
}

func (s *ReleaseSteps) Close(ctx context.Context, id uuid.UUID) error {
	r, err := s.ReleaseStorage.Get(ctx, id)
	if err != nil {
		return err
	}
	if r.Status != model.ReleasePending && r.Status != model.ReleaseActive {
		return nil
	}
	r.Status = model.ReleaseClosing
	_, err = s.ReleaseStorage.Store(ctx, r)
	return err
}

func (s *ReleaseSteps) DeleteShortLink(ctx context.Context, id uuid.UUID) error {
	_, err := s.ReleaseStorage.Get(ctx, id)
	if errors.Is(err, service.ErrNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	return s.ShortlinkService.Delete(ctx, id)
}

func (s *ReleaseSteps) DeleteSnapshot(ctx context.Context, id uuid.UUID) error {
	r, err := s.ReleaseStorage.Get(ctx, id)
	if err != nil {
		return err
	}
	attachments, err := s.AttachmentQueryService.List(ctx, id)
	if err != nil {
		return err
	}
	for _, attachment := range attachments {
		if err = s.FilestorageService.Delete(ctx, r.WorkspaceID, r.OwnerID, attachment.CopyID); err != nil {
			return err
		}
	}
	if err = s.AttachmentStorage.Delete(ctx, id); err != nil {
		return err
	}
	for range 3 {
		r, err = s.ReleaseStorage.Get(ctx, id)
		if err != nil {
			return err
		}
		switch r.Status {
		case model.ReleaseClosing:
			r.Status, r.Content = model.ReleaseClosed, ""
		case model.ReleasePending:
			r.Status, r.Failure = model.ReleaseFailed, "publication failed before activation"
		default:
			return nil
		}
		_, err = s.ReleaseStorage.Store(ctx, r)
		if errors.Is(err, service.ErrConflict) {
			continue
		}
		if err != nil {
			return fmt.Errorf("mark release cleaned: %w", err)
		}
		return nil
	}
	return fmt.Errorf("mark release cleaned: %w", service.ErrConflict)
}

func sameAttachments(a, b []adapter.Attachment) bool {
	if len(a) != len(b) {
		return false
	}
	x, y := make([]uuid.UUID, len(a)), make([]uuid.UUID, len(b))
	for i, v := range a {
		x[i] = v.FileID
	}
	for i, v := range b {
		y[i] = v.FileID
	}
	slices.SortFunc(x, uuid.UUID.Compare)
	slices.SortFunc(y, uuid.UUID.Compare)
	return slices.Equal(x, y)
}
