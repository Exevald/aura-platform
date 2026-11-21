package integration_tests

import (
	"context"
	"io"
	"publication/internal/publication/app/workflow"
	serviceeditor "publication/internal/publication/infra/adapter"
	"publication/internal/publication/infra/mysql/storage"
	"sync"
	"time"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/worker"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

	editorapi "editor/api/server/internalapi"
	notificationapi "notification/api/server/internalapi"
	workflowapi "publication/api/workflow"
	shortlinkapi "shortlink/api/server/internalapi"
	workspaceapi "workspace/api/server/internalapi"
)

type flowFakes struct {
	mu                sync.Mutex
	snapshot          *editorapi.PublicationSnapshot
	files             map[string][]byte
	members           []string
	links             map[string]string
	notifications     map[string]int
	snapshotCalls     int
	changeAfterFirst  bool
	failLink          bool
	failNotifications bool
}

func (f *flowFakes) reset(snapshot *editorapi.PublicationSnapshot, files map[string][]byte, members []string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.snapshot = snapshot
	f.files = files
	f.members = members
	f.links = make(map[string]string)
	f.notifications = make(map[string]int)
	f.snapshotCalls = 0
	f.changeAfterFirst = false
	f.failLink = false
	f.failNotifications = false
}

type fakeEditor struct {
	editorapi.EditorServiceClient
	f *flowFakes
}

func (c fakeEditor) GetPublicationSnapshot(_ context.Context, req *editorapi.GetPublicationSnapshotRequest, _ ...grpc.CallOption) (*editorapi.PublicationSnapshot, error) {
	c.f.mu.Lock()
	defer c.f.mu.Unlock()
	c.f.snapshotCalls++
	if c.f.snapshot == nil || req.DocumentId != c.f.snapshot.DocumentId {
		return nil, status.Error(codes.NotFound, "document not found")
	}
	if req.ExpectedVersion != c.f.snapshot.Version || (c.f.changeAfterFirst && c.f.snapshotCalls > 1) {
		return nil, status.Error(codes.Aborted, "document version changed")
	}
	return proto.Clone(c.f.snapshot).(*editorapi.PublicationSnapshot), nil
}

func (c fakeEditor) DownloadPublicationAttachment(_ context.Context, req *editorapi.DownloadPublicationAttachmentRequest, _ ...grpc.CallOption) (grpc.ServerStreamingClient[editorapi.PublicationAttachmentChunk], error) {
	c.f.mu.Lock()
	defer c.f.mu.Unlock()
	data, ok := c.f.files[req.FileId]
	if !ok {
		return nil, status.Error(codes.NotFound, "file not found")
	}
	return &fakeAttachmentStream{data: append([]byte(nil), data...)}, nil
}

type fakeAttachmentStream struct {
	grpc.ClientStream
	data []byte
	done bool
}

func (s *fakeAttachmentStream) Recv() (*editorapi.PublicationAttachmentChunk, error) {
	if s.done {
		return nil, io.EOF
	}
	s.done = true
	return &editorapi.PublicationAttachmentChunk{Data: s.data}, nil
}

type fakeShortlink struct {
	shortlinkapi.ShortlinkServiceClient
	f *flowFakes
}

func (c fakeShortlink) EnsureLink(_ context.Context, req *shortlinkapi.EnsureLinkRequest, _ ...grpc.CallOption) (*shortlinkapi.EnsureLinkResponse, error) {
	c.f.mu.Lock()
	defer c.f.mu.Unlock()
	if c.f.failLink {
		return nil, temporal.NewNonRetryableApplicationError("shortlink unavailable", "Injected", nil)
	}
	if _, ok := c.f.links[req.IdempotencyKey]; !ok {
		c.f.links[req.IdempotencyKey] = req.TargetUrl
	}
	return &shortlinkapi.EnsureLinkResponse{LinkId: req.IdempotencyKey, ShortPath: "/s/Publish123"}, nil
}

func (c fakeShortlink) DeleteOwnedLink(_ context.Context, req *shortlinkapi.DeleteOwnedLinkRequest, _ ...grpc.CallOption) (*shortlinkapi.DeleteOwnedLinkResponse, error) {
	c.f.mu.Lock()
	defer c.f.mu.Unlock()
	delete(c.f.links, req.IdempotencyKey)
	return &shortlinkapi.DeleteOwnedLinkResponse{}, nil
}

type fakeWorkspace struct {
	workspaceapi.WorkspaceServiceClient
	f *flowFakes
}

func (c fakeWorkspace) ListMembers(_ context.Context, _ *workspaceapi.ListMembersRequest, _ ...grpc.CallOption) (*workspaceapi.ListMembersResponse, error) {
	c.f.mu.Lock()
	defer c.f.mu.Unlock()
	return &workspaceapi.ListMembersResponse{UserIds: append([]string(nil), c.f.members...)}, nil
}

type fakeNotifications struct {
	notificationapi.NotificationServiceClient
	f *flowFakes
}

func (c fakeNotifications) CreateNotification(_ context.Context, req *notificationapi.CreateNotificationRequest, _ ...grpc.CallOption) (*notificationapi.CreateNotificationResponse, error) {
	c.f.mu.Lock()
	defer c.f.mu.Unlock()
	if c.f.failNotifications {
		return nil, temporal.NewNonRetryableApplicationError("notifications unavailable", "Injected", nil)
	}
	if c.f.notifications[req.IdempotencyKey] == 0 {
		c.f.notifications[req.IdempotencyKey] = 1
	}
	return &notificationapi.CreateNotificationResponse{NotificationId: req.IdempotencyKey, Created: true}, nil
}

func (e *environment) startWorker(ctx context.Context) error {
	var err error
	for start := time.Now(); time.Since(start) < time.Minute; {
		e.temporal, err = client.Dial(client.Options{HostPort: e.temporalAddress, Namespace: "default"})
		if err == nil {
			break
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(250 * time.Millisecond):
		}
	}
	if err != nil {
		return err
	}
	e.flowFakes = &flowFakes{}
	e.worker = worker.New(e.temporal, workflowapi.PublicationFlowTaskQueue, worker.Options{})
	workflowapi.RegisterPublicationFlowWorkflows(e.worker, &workflow.Workflows{})
	workflowapi.RegisterPublicationFlowActivities(e.worker, &workflow.Activities{Steps: &workflow.ReleaseSteps{
		ReleaseStorage:         &storage.ReleaseStore{Client: e.connector.TransactionalClient()},
		AttachmentStorage:      &storage.AttachmentStore{Client: e.connector.TransactionalClient()},
		AttachmentQueryService: &storage.AttachmentQueries{Client: e.connector.TransactionalClient()},
		FilestorageService:     filePort{e.files},
		EditorService:          serviceeditor.NewEditorService(fakeEditor{f: e.flowFakes}),
		ShortlinkService:       serviceeditor.NewShortlinkService(fakeShortlink{f: e.flowFakes}),
		WorkspaceService:       serviceeditor.NewWorkspaceService(fakeWorkspace{f: e.flowFakes}),
		NotificationService:    serviceeditor.NewNotificationService(fakeNotifications{f: e.flowFakes}),
		PublicBaseURL:          "http://aura-platform.home",
	}})
	if err = e.worker.Start(); err != nil {
		e.temporal.Close()
		return err
	}
	e.closers = append(e.closers, func() error { e.worker.Stop(); e.temporal.Close(); return nil })
	return nil
}
