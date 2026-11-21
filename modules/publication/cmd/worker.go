package main

import (
	"context"
	"errors"
	"log/slog"

	"github.com/urfave/cli/v3"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	editorapi "editor/api/server/internalapi"
	notificationapi "notification/api/server/internalapi"
	workflowapi "publication/api/workflow"
	"publication/internal/publication/app/workflow"
	"publication/internal/publication/config"
	serviceeditor "publication/internal/publication/infra/adapter"
	"publication/internal/publication/infra/mysql/storage"
	shortlinkapi "shortlink/api/server/internalapi"
	workspaceapi "workspace/api/server/internalapi"
)

func (a *application) workerCmd() *cli.Command {
	return &cli.Command{Name: "worker", Action: func(ctx context.Context, _ *cli.Command) error {
		return runWorker(ctx, a.logger, a.conf)
	}}
}

func runWorker(ctx context.Context, logger *slog.Logger, conf config.Env) (err error) {
	db, err := openDatabase(ctx, logger, conf)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, db.connector.Close()) }()
	files, fileConn, err := connectFiles(conf.Dependencies.FilestorageAddress)
	if err != nil {
		return err
	}
	defer fileConn.Close()
	dial := func(address string) (*grpc.ClientConn, error) {
		return grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}
	editorConn, err := dial(conf.Dependencies.EditorAddress)
	if err != nil {
		return err
	}
	defer editorConn.Close()
	linkConn, err := dial(conf.Dependencies.ShortlinkAddress)
	if err != nil {
		return err
	}
	defer linkConn.Close()
	workspaceConn, err := dial(conf.Dependencies.WorkspaceAddress)
	if err != nil {
		return err
	}
	defer workspaceConn.Close()
	notificationConn, err := dial(conf.Dependencies.NotificationAddress)
	if err != nil {
		return err
	}
	defer notificationConn.Close()
	c, err := client.Dial(client.Options{HostPort: conf.Temporal.Address, Namespace: conf.Temporal.Namespace})
	if err != nil {
		return err
	}
	defer c.Close()
	w := worker.New(c, workflowapi.PublicationFlowTaskQueue, worker.Options{})
	workflowapi.RegisterPublicationFlowWorkflows(w, &workflow.Workflows{})
	workflowapi.RegisterPublicationFlowActivities(w, &workflow.Activities{Steps: &workflow.ReleaseSteps{
		ReleaseStorage:         &storage.ReleaseStore{Client: db.client},
		AttachmentStorage:      &storage.AttachmentStore{Client: db.client},
		AttachmentQueryService: &storage.AttachmentQueries{Client: db.client},
		FilestorageService:     files,
		EditorService:          serviceeditor.NewEditorService(editorapi.NewEditorServiceClient(editorConn)),
		ShortlinkService:       serviceeditor.NewShortlinkService(shortlinkapi.NewShortlinkServiceClient(linkConn)),
		WorkspaceService:       serviceeditor.NewWorkspaceService(workspaceapi.NewWorkspaceServiceClient(workspaceConn)),
		NotificationService:    serviceeditor.NewNotificationService(notificationapi.NewNotificationServiceClient(notificationConn)),
		PublicBaseURL:          conf.Dependencies.PublicBaseURL,
	}})
	if err = w.Start(); err != nil {
		return err
	}
	logger.Info("publication worker started")
	<-ctx.Done()
	w.Stop()
	return nil
}
