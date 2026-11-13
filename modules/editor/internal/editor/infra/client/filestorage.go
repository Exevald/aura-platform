package client

import (
	"context"
	"editor/internal/editor/app"
	"errors"
	api "filestorage/api/server/internalapi"
	"io"
	"time"
	"uuid"
)

type filestorage struct {
	client  api.FilestorageServiceClient
	timeout time.Duration
}

func NewFilestorage(c api.FilestorageServiceClient, t time.Duration) app.FileStorage {
	return &filestorage{c, t}
}

func (c *filestorage) Upload(ctx context.Context, w, u uuid.UUID, name, contentType string, data []byte) (app.StoredFile, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	stream, e := c.client.Upload(ctx)
	if e != nil {
		return app.StoredFile{}, e
	}
	if e = stream.Send(&api.UploadRequest{Value: &api.UploadRequest_Metadata{Metadata: &api.UploadMetadata{WorkspaceId: w.String(), OwnerUserId: u.String(), FileName: name, ContentType: contentType}}}); e != nil {
		return app.StoredFile{}, e
	}
	for len(data) > 0 {
		n := min(len(data), 64<<10)
		if e = stream.Send(&api.UploadRequest{Value: &api.UploadRequest_Chunk{Chunk: data[:n]}}); e != nil {
			return app.StoredFile{}, e
		}
		data = data[n:]
	}
	v, e := stream.CloseAndRecv()
	if e != nil {
		return app.StoredFile{}, e
	}
	id, e := uuid.Parse(v.Id)
	return app.StoredFile{ID: id, FileName: v.FileName, ContentType: v.ContentType, SizeBytes: v.SizeBytes}, e
}

func (c *filestorage) Download(ctx context.Context, w, u, id uuid.UUID) (app.StoredFile, []byte, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	stream, e := c.client.Download(ctx, &api.DownloadRequest{Id: id.String(), WorkspaceId: w.String(), ActorUserId: u.String()})
	if e != nil {
		return app.StoredFile{}, nil, e
	}
	var meta app.StoredFile
	var data []byte
	for {
		part, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			return meta, data, nil
		}
		if err != nil {
			return app.StoredFile{}, nil, err
		}
		switch v := part.Value.(type) {
		case *api.DownloadResponse_Metadata:
			fid, parseErr := uuid.Parse(v.Metadata.Id)
			if parseErr != nil {
				return app.StoredFile{}, nil, parseErr
			}
			meta = app.StoredFile{ID: fid, FileName: v.Metadata.FileName, ContentType: v.Metadata.ContentType, SizeBytes: v.Metadata.SizeBytes}
		case *api.DownloadResponse_Chunk:
			data = append(data, v.Chunk...)
		}
	}
}

func (c *filestorage) Delete(ctx context.Context, w, u, id uuid.UUID) error {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	_, e := c.client.Delete(ctx, &api.DeleteRequest{Id: id.String(), WorkspaceId: w.String(), ActorUserId: u.String()})
	return e
}
