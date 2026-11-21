package adapter

import (
	"context"
	"errors"
	"fmt"
	"io"
	"uuid"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	api "editor/api/server/internalapi"
	port "publication/internal/publication/app/adapter"
)

func NewEditorService(client api.EditorServiceClient) port.EditorService {
	return editorAdapter{client}
}

type editorAdapter struct {
	client api.EditorServiceClient
}

func (s editorAdapter) Snapshot(ctx context.Context, document, workspace, owner uuid.UUID, version int64) (port.Snapshot, error) {
	r, err := s.client.GetPublicationSnapshot(
		ctx,
		&api.GetPublicationSnapshotRequest{
			DocumentId:      document.String(),
			WorkspaceId:     workspace.String(),
			ActorUserId:     owner.String(),
			ExpectedVersion: version,
		})
	if err != nil {
		return port.Snapshot{}, convertError(err)
	}
	out := port.Snapshot{Title: r.Title, Content: r.Content, Version: r.Version}
	for _, a := range r.Attachments {
		fileID, err2 := uuid.Parse(a.FileId)
		if err2 != nil {
			return port.Snapshot{}, fmt.Errorf("%w: invalid attachment ID %q: %v", port.ErrRejected, a.FileId, err2)
		}
		out.Attachments = append(
			out.Attachments,
			port.Attachment{
				FileID:      fileID,
				FileName:    a.FileName,
				ContentType: a.ContentType,
				SizeBytes:   a.SizeBytes,
			})
	}
	return out, nil
}

func (s editorAdapter) DownloadAttachment(ctx context.Context, document, workspace, owner, file uuid.UUID) ([]byte, error) {
	stream, err := s.client.DownloadPublicationAttachment(
		ctx,
		&api.DownloadPublicationAttachmentRequest{
			DocumentId:  document.String(),
			WorkspaceId: workspace.String(),
			ActorUserId: owner.String(),
			FileId:      file.String(),
		})
	if err != nil {
		return nil, convertError(err)
	}
	var data []byte
	for {
		chunk, err2 := stream.Recv()
		if errors.Is(err2, io.EOF) {
			return data, nil
		}
		if err2 != nil {
			return nil, convertError(err2)
		}
		if len(data)+len(chunk.Data) > 32<<20 {
			return nil, fmt.Errorf("%w: attachment too large", port.ErrRejected)
		}
		data = append(data, chunk.Data...)
	}
}

func convertError(err error) error {
	switch status.Code(err) {
	case codes.NotFound, codes.PermissionDenied, codes.Aborted, codes.InvalidArgument:
		return fmt.Errorf("%w: %v", port.ErrRejected, err)
	default:
		return err
	}
}
