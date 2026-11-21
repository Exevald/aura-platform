package workflow

import (
	"context"
	"errors"
	"publication/internal/publication/app/adapter"
	"publication/internal/publication/app/service"
	"uuid"

	workflowapi "publication/api/workflow"

	"go.temporal.io/sdk/temporal"
)

type Activities struct{ Steps *ReleaseSteps }

func (a *Activities) ExportSnapshot(ctx context.Context, req *workflowapi.ReleaseID) (*workflowapi.Snapshot, error) {
	id, err := parseReleaseID(req.Id)
	if err != nil {
		return nil, err
	}
	version, err := a.Steps.ExportSnapshot(ctx, id)
	if err != nil {
		return nil, activityError(err)
	}
	return &workflowapi.Snapshot{Id: req.Id, SourceVersion: version}, nil
}

func (a *Activities) CreateShortLink(ctx context.Context, req *workflowapi.ReleaseID) (*workflowapi.ShortLink, error) {
	id, err := parseReleaseID(req.Id)
	if err != nil {
		return nil, err
	}
	linkID, path, err := a.Steps.CreateShortLink(ctx, id)
	if err != nil {
		return nil, activityError(err)
	}
	return &workflowapi.ShortLink{Id: linkID.String(), Path: path}, nil
}

func (a *Activities) Activate(ctx context.Context, req *workflowapi.ReleaseID) (*workflowapi.ReleaseID, error) {
	id, err := parseReleaseID(req.Id)
	if err != nil {
		return nil, err
	}
	return req, activityError(a.Steps.Activate(ctx, id))
}

func (a *Activities) NotifyMembers(ctx context.Context, req *workflowapi.ReleaseID) (*workflowapi.ReleaseID, error) {
	id, err := parseReleaseID(req.Id)
	if err != nil {
		return nil, err
	}
	return req, activityError(a.Steps.NotifyMembers(ctx, id))
}

func (a *Activities) MarkNotificationFailed(ctx context.Context, req *workflowapi.ReleaseID) (*workflowapi.ReleaseID, error) {
	id, err := parseReleaseID(req.Id)
	if err != nil {
		return nil, err
	}
	return req, activityError(a.Steps.MarkNotificationFailed(ctx, id))
}

func (a *Activities) Close(ctx context.Context, req *workflowapi.ReleaseID) (*workflowapi.ReleaseID, error) {
	id, err := parseReleaseID(req.Id)
	if err != nil {
		return nil, err
	}
	return req, activityError(a.Steps.Close(ctx, id))
}

func (a *Activities) DeleteShortLink(ctx context.Context, req *workflowapi.ReleaseID) (*workflowapi.ReleaseID, error) {
	id, err := parseReleaseID(req.Id)
	if err != nil {
		return nil, err
	}
	return req, activityError(a.Steps.DeleteShortLink(ctx, id))
}

func (a *Activities) DeleteSnapshot(ctx context.Context, req *workflowapi.ReleaseID) (*workflowapi.ReleaseID, error) {
	id, err := parseReleaseID(req.Id)
	if err != nil {
		return nil, err
	}
	return req, activityError(a.Steps.DeleteSnapshot(ctx, id))
}

func parseReleaseID(value string) (uuid.UUID, error) {
	id, err := uuid.Parse(value)
	if err != nil {
		return uuid.Nil(), temporal.NewNonRetryableApplicationError("invalid release ID", "InvalidArgument", err)
	}
	return id, nil
}

func activityError(err error) error {
	switch {
	case errors.Is(err, service.ErrConflict):
		return temporal.NewNonRetryableApplicationError(err.Error(), "Conflict", err)
	case errors.Is(err, adapter.ErrRejected):
		return temporal.NewNonRetryableApplicationError(err.Error(), "EditorRejected", err)
	default:
		return err
	}
}
