package integration_tests

import (
	"bytes"
	"context"
	"io"
	"net"
	"strconv"
	"sync"

	filestorageapi "filestorage/api/server/internalapi"
	"uuid"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type fakeFiles struct {
	filestorageapi.UnimplementedFilestorageServiceServer
	mu   sync.Mutex
	data map[string][]byte
}

func (f *fakeFiles) UploadCopy(_ context.Context, _, _, _, _ string, data []byte) (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	id := uuid.New().String()
	f.data[id] = append([]byte(nil), data...)
	return id, nil
}

func (f *fakeFiles) Upload(stream grpc.ClientStreamingServer[filestorageapi.UploadRequest, filestorageapi.FileMetadata]) error {
	var body bytes.Buffer
	var meta *filestorageapi.UploadMetadata
	for {
		p, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if m := p.GetMetadata(); m != nil {
			meta = m
		}
		if chunk := p.GetChunk(); len(chunk) > 0 {
			body.Write(chunk)
		}
	}
	if meta == nil {
		return status.Error(codes.InvalidArgument, "metadata required")
	}
	id, err := f.UploadCopy(stream.Context(), meta.WorkspaceId, meta.OwnerUserId, meta.FileName, meta.ContentType, body.Bytes())
	if err != nil {
		return err
	}
	return stream.SendAndClose(&filestorageapi.FileMetadata{Id: id, WorkspaceId: meta.WorkspaceId, OwnerUserId: meta.OwnerUserId, FileName: meta.FileName, ContentType: meta.ContentType, SizeBytes: int64(body.Len())})
}

func (f *fakeFiles) DownloadCopy(_ context.Context, _, id string) ([]byte, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	v, ok := f.data[id]
	if !ok {
		return nil, status.Error(codes.NotFound, "file not found")
	}
	return append([]byte(nil), v...), nil
}

func (f *fakeFiles) Download(req *filestorageapi.DownloadRequest, stream grpc.ServerStreamingServer[filestorageapi.DownloadResponse]) error {
	data, err := f.DownloadCopy(stream.Context(), req.WorkspaceId, req.Id)
	if err != nil {
		return err
	}
	return stream.Send(&filestorageapi.DownloadResponse{Value: &filestorageapi.DownloadResponse_Chunk{Chunk: data}})
}

func (f *fakeFiles) DeleteCopy(_ context.Context, _, _, id string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.data, id)
	return nil
}

func (f *fakeFiles) Delete(ctx context.Context, req *filestorageapi.DeleteRequest) (*filestorageapi.DeleteResponse, error) {
	return &filestorageapi.DeleteResponse{}, f.DeleteCopy(ctx, req.WorkspaceId, req.ActorUserId, req.Id)
}

type filePort struct{ f *fakeFiles }

func (p filePort) Upload(ctx context.Context, w, u uuid.UUID, n, t string, b []byte) (uuid.UUID, error) {
	id, err := p.f.UploadCopy(ctx, w.String(), u.String(), n, t, b)
	if err != nil {
		return uuid.Nil(), err
	}
	return uuid.Parse(id)
}

func (p filePort) Download(ctx context.Context, w, id uuid.UUID) ([]byte, error) {
	return p.f.DownloadCopy(ctx, w.String(), id.String())
}

func (p filePort) Delete(ctx context.Context, w, u, id uuid.UUID) error {
	return p.f.DeleteCopy(ctx, w.String(), u.String(), id.String())
}

func (e *environment) startFiles() error {
	e.files = &fakeFiles{data: make(map[string][]byte)}
	// #nosec G102 -- the test container reaches this short-lived fixture through host port forwarding.
	listener, err := net.Listen("tcp", "0.0.0.0:0")
	if err != nil {
		return err
	}
	e.filePort = listener.Addr().(*net.TCPAddr).Port
	server := grpc.NewServer()
	filestorageapi.RegisterFilestorageServiceServer(server, e.files)
	go func() { _ = server.Serve(listener) }()
	e.closers = append(e.closers, func() error { server.Stop(); return nil })
	return nil
}

func (e *environment) fileAddress() string {
	return "host.testcontainers.internal:" + strconv.Itoa(e.filePort)
}
