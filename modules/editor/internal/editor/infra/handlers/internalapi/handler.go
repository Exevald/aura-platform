package internalapi

import (
	"context"
	"errors"
	"uuid"

	api "editor/api/server/internalapi"
	"editor/internal/editor/app"
	"editor/internal/editor/domain/model"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type handler struct {
	api.UnimplementedEditorServiceServer
	svc app.EditorService
}

func New(svc app.EditorService) api.EditorServiceServer { return &handler{svc: svc} }

func (*handler) Echo(context.Context, *api.EchoRequest) (*api.EchoResponse, error) {
	return nil, status.Error(codes.Unimplemented, "echo is not supported")
}

func parseIDs(document, workspace, actor string) (documentID, workspaceID, actorID uuid.UUID, parseErr error) {
	d, err := uuid.Parse(document)
	if err != nil {
		return uuid.Nil(), uuid.Nil(), uuid.Nil(), status.Error(codes.InvalidArgument, "invalid document")
	}
	w, err := uuid.Parse(workspace)
	if err != nil {
		return uuid.Nil(), uuid.Nil(), uuid.Nil(), status.Error(codes.InvalidArgument, "invalid workspace")
	}
	a, err := uuid.Parse(actor)
	if err != nil {
		return uuid.Nil(), uuid.Nil(), uuid.Nil(), status.Error(codes.InvalidArgument, "invalid actor")
	}
	return d, w, a, nil
}

func mapError(err error) error {
	if errors.Is(err, model.ErrDocumentNotFound) {
		return status.Error(codes.NotFound, "document not found")
	}
	if errors.Is(err, model.ErrPermissionDenied) {
		return status.Error(codes.PermissionDenied, "document permission denied")
	}
	return status.Error(codes.Internal, "editor operation failed")
}

func (h *handler) GetPublicationSnapshot(ctx context.Context, r *api.GetPublicationSnapshotRequest) (*api.PublicationSnapshot, error) {
	d, w, a, err := parseIDs(r.DocumentId, r.WorkspaceId, r.ActorUserId)
	if err != nil {
		return nil, err
	}
	doc, err := h.svc.Get(ctx, a, w, d)
	if err != nil {
		return nil, mapError(err)
	}
	if doc.Role != model.RoleOwner {
		return nil, status.Error(codes.PermissionDenied, "owner required")
	}
	if doc.Version != r.ExpectedVersion {
		return nil, status.Error(codes.Aborted, "document version changed")
	}
	items, err := h.svc.ListAttachments(ctx, a, w, d)
	if err != nil {
		return nil, mapError(err)
	}
	out := &api.PublicationSnapshot{DocumentId: d.String(), Title: doc.Title, Content: doc.Content, Version: doc.Version}
	for _, item := range items {
		out.Attachments = append(out.Attachments, &api.PublicationAttachment{FileId: item.FileID.String(), FileName: item.FileName, ContentType: item.ContentType, SizeBytes: item.SizeBytes})
	}
	return out, nil
}

func (h *handler) DownloadPublicationAttachment(r *api.DownloadPublicationAttachmentRequest, stream api.EditorService_DownloadPublicationAttachmentServer) error {
	d, w, a, err := parseIDs(r.DocumentId, r.WorkspaceId, r.ActorUserId)
	if err != nil {
		return err
	}
	f, err := uuid.Parse(r.FileId)
	if err != nil {
		return status.Error(codes.InvalidArgument, "invalid file")
	}
	doc, err := h.svc.Get(stream.Context(), a, w, d)
	if err != nil {
		return mapError(err)
	}
	if doc.Role != model.RoleOwner {
		return status.Error(codes.PermissionDenied, "owner required")
	}
	_, data, err := h.svc.DownloadAttachment(stream.Context(), a, w, d, f)
	if err != nil {
		return mapError(err)
	}
	for len(data) > 0 {
		n := min(len(data), 64<<10)
		if err := stream.Send(&api.PublicationAttachmentChunk{Data: data[:n]}); err != nil {
			return err
		}
		data = data[n:]
	}
	return nil
}
