package adapter

import (
	"context"
	"errors"
	"fmt"
	"io"
	port "publication/internal/publication/app/adapter"
	"uuid"

	api "filestorage/api/server/internalapi"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func NewFilestorageService(client api.FilestorageServiceClient) port.FilestorageService {
	return &filestorageAdapter{client}
}

type filestorageAdapter struct {
	client api.FilestorageServiceClient
}

func (s *filestorageAdapter) Upload(ctx context.Context, workspace, owner uuid.UUID, name, contentType string, body []byte) (uuid.UUID, error) {
	stream, err := s.client.Upload(ctx)
	if err != nil {
		return uuid.Nil(), err
	}
	if err = stream.Send(&api.UploadRequest{Value: &api.UploadRequest_Metadata{Metadata: &api.UploadMetadata{WorkspaceId: workspace.String(), OwnerUserId: owner.String(), FileName: name, ContentType: contentType}}}); err != nil {
		return uuid.Nil(), err
	}
	for len(body) > 0 {
		n := min(len(body), 64<<10)
		if err = stream.Send(&api.UploadRequest{Value: &api.UploadRequest_Chunk{Chunk: body[:n]}}); err != nil {
			return uuid.Nil(), err
		}
		body = body[n:]
	}
	meta, err := stream.CloseAndRecv()
	if err != nil {
		return uuid.Nil(), err
	}
	id, err := uuid.Parse(meta.Id)
	if err != nil {
		return uuid.Nil(), fmt.Errorf("invalid filestorage ID %q: %w", meta.Id, err)
	}
	return id, nil
}

func (s *filestorageAdapter) Download(ctx context.Context, workspace, id uuid.UUID) ([]byte, error) {
	stream, err := s.client.Download(ctx, &api.DownloadRequest{WorkspaceId: workspace.String(), Id: id.String()})
	if err != nil {
		return nil, err
	}
	var data []byte
	for {
		part, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			return data, nil
		}
		if err != nil {
			return nil, err
		}
		data = append(data, part.GetChunk()...)
	}
}

func (s *filestorageAdapter) Delete(ctx context.Context, workspace, owner, id uuid.UUID) error {
	_, err := s.client.Delete(ctx, &api.DeleteRequest{WorkspaceId: workspace.String(), ActorUserId: owner.String(), Id: id.String()})
	if status.Code(err) == codes.NotFound {
		return nil
	}
	return err
}
