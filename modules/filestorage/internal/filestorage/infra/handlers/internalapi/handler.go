package internalapi

import (
	"bytes"
	"context"
	"errors"
	api "filestorage/api/server/internalapi"
	"filestorage/internal/filestorage/app"
	"filestorage/internal/filestorage/domain"
	"io"
	"log/slog"
	"uuid"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const maxUpload = 32 << 20

type handler struct {
	svc    *app.Service
	logger *slog.Logger
	api.UnimplementedFilestorageServiceServer
}

func NewHandler(s *app.Service, l *slog.Logger) api.FilestorageServiceServer {
	return &handler{svc: s, logger: l}
}

func (h *handler) Upload(stream api.FilestorageService_UploadServer) error {
	var meta *api.UploadMetadata
	var body bytes.Buffer
	for {
		part, e := stream.Recv()
		if errors.Is(e, io.EOF) {
			break
		}
		if e != nil {
			return e
		}
		switch x := part.Value.(type) {
		case *api.UploadRequest_Metadata:
			if meta != nil {
				return status.Error(codes.InvalidArgument, "duplicate metadata")
			}
			meta = x.Metadata
		case *api.UploadRequest_Chunk:
			if body.Len()+len(x.Chunk) > maxUpload {
				return status.Error(codes.ResourceExhausted, "file too large")
			}
			_, _ = body.Write(x.Chunk)
		}
	}
	if meta == nil {
		return status.Error(codes.InvalidArgument, "metadata is required")
	}
	w, e := uuid.Parse(meta.WorkspaceId)
	if e != nil {
		return status.Error(codes.InvalidArgument, "invalid workspace")
	}
	u, e := uuid.Parse(meta.OwnerUserId)
	if e != nil {
		return status.Error(codes.InvalidArgument, "invalid owner")
	}
	v, e := h.svc.Upload(stream.Context(), w, u, meta.FileName, meta.ContentType, int64(body.Len()), &body)
	if e != nil {
		return status.Error(codes.Internal, "upload failed")
	}
	return stream.SendAndClose(proto(v))
}

func (h *handler) Download(r *api.DownloadRequest, stream api.FilestorageService_DownloadServer) error {
	w, e := uuid.Parse(r.WorkspaceId)
	if e != nil {
		return status.Error(codes.InvalidArgument, "invalid workspace")
	}
	id, e := uuid.Parse(r.Id)
	if e != nil {
		return status.Error(codes.InvalidArgument, "invalid file")
	}
	v, body, e := h.svc.Download(stream.Context(), w, id)
	if errors.Is(e, domain.ErrNotFound) {
		return status.Error(codes.NotFound, e.Error())
	}
	if e != nil {
		return status.Error(codes.Internal, "download failed")
	}
	defer body.Close()
	if e = stream.Send(&api.DownloadResponse{Value: &api.DownloadResponse_Metadata{Metadata: proto(v)}}); e != nil {
		return e
	}
	buffer := make([]byte, 64<<10)
	for {
		n, readErr := body.Read(buffer)
		if n > 0 {
			if e = stream.Send(&api.DownloadResponse{Value: &api.DownloadResponse_Chunk{Chunk: append([]byte(nil), buffer[:n]...)}}); e != nil {
				return e
			}
		}
		if errors.Is(readErr, io.EOF) {
			return nil
		}
		if readErr != nil {
			return readErr
		}
	}
}

func (h *handler) Delete(ctx context.Context, r *api.DeleteRequest) (*api.DeleteResponse, error) {
	w, e := uuid.Parse(r.WorkspaceId)
	if e != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid workspace")
	}
	u, e := uuid.Parse(r.ActorUserId)
	if e != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid actor")
	}
	id, e := uuid.Parse(r.Id)
	if e != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid file")
	}
	e = h.svc.Delete(ctx, w, u, id)
	if errors.Is(e, domain.ErrNotFound) || errors.Is(e, domain.ErrForbidden) {
		return nil, status.Error(codes.NotFound, "file not found")
	}
	if e != nil {
		return nil, status.Error(codes.Internal, "delete failed")
	}
	return &api.DeleteResponse{}, nil
}

func proto(v domain.File) *api.FileMetadata {
	return &api.FileMetadata{Id: v.ID.String(), WorkspaceId: v.WorkspaceID.String(), OwnerUserId: v.OwnerUserID.String(), FileName: v.FileName, ContentType: v.ContentType, SizeBytes: v.SizeBytes, ChecksumSha256: v.ChecksumSHA256, CreatedAtUnix: v.CreatedAt.Unix()}
}
