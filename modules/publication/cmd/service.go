package main

import (
	"context"
	"errors"
	"log/slog"
	service2 "publication/internal/publication/app/service"
	"publication/internal/publication/infra/mysql/storage"
	"time"

	"github.com/urfave/cli/v3"
	"golang.org/x/sync/errgroup"

	workflowapi "publication/api/workflow"
	"publication/internal/publication/config"

	"go.temporal.io/sdk/client"
)

func (a *application) serviceCmd() *cli.Command {
	return &cli.Command{
		Name: "service",
		Action: func(ctx context.Context, _ *cli.Command) error {
			return service(ctx, a.logger, a.conf)
		},
	}
}

func service(ctx context.Context, logger *slog.Logger, conf config.Env) (err error) {
	db, err := openDatabase(ctx, logger, conf)
	if err != nil {
		return err
	}
	defer func() {
		err = errors.Join(err, db.connector.Close())
	}()

	temporalClient, err := client.NewLazyClient(client.Options{HostPort: conf.Temporal.Address, Namespace: conf.Temporal.Namespace})
	if err != nil {
		return err
	}
	defer temporalClient.Close()
	files, fileConn, err := connectFiles(conf.Dependencies.FilestorageAddress)
	if err != nil {
		return err
	}
	defer fileConn.Close()
	releases := service2.NewReleases(&storage.ReleaseStore{Client: db.client}, &storage.AttachmentStore{Client: db.client}, &storage.ReleaseQueries{Client: db.client}, &storage.AttachmentQueries{Client: db.client}, workflowapi.NewPublicationFlowClient(temporalClient), files)
	httpServer, err := newHTTPServer(conf, releases)
	if err != nil {
		return err
	}

	errGroup, egCtx := errgroup.WithContext(ctx)
	errGroup.Go(func() error {
		return httpServer.Serve(egCtx)
	})
	errGroup.Go(func() error {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for {
			if e := releases.Recover(egCtx); e != nil && egCtx.Err() == nil {
				logger.Warn("publication recovery retry", "error", e)
			}
			select {
			case <-egCtx.Done():
				return nil
			case <-ticker.C:
			}
		}
	})
	return errGroup.Wait()
}
