package workflow

import (
	"errors"

	"go.temporal.io/sdk/workflow"

	workflowapi "publication/api/workflow"
)

type Workflows struct{}

type publish struct {
	*workflowapi.PublishWorkflowInput
}

type unpublish struct {
	*workflowapi.UnpublishWorkflowInput
}

func (*Workflows) Publish(
	_ workflow.Context,
	input *workflowapi.PublishWorkflowInput,
) (workflowapi.PublishWorkflow, error) {
	return &publish{input}, nil
}

func (*Workflows) Unpublish(
	_ workflow.Context,
	input *workflowapi.UnpublishWorkflowInput,
) (workflowapi.UnpublishWorkflow, error) {
	return &unpublish{input}, nil
}

func (w *publish) Execute(ctx workflow.Context) (*workflowapi.PublishResponse, error) {
	id := &workflowapi.ReleaseID{Id: w.Req.Id}
	if _, err := workflowapi.ExportSnapshot(ctx, id); err != nil {
		return nil, errors.Join(err, compensate(ctx, id, false))
	}
	if _, err := workflowapi.CreateShortLink(ctx, id); err != nil {
		return nil, errors.Join(err, compensate(ctx, id, true))
	}
	if _, err := workflowapi.Activate(ctx, id); err != nil {
		return nil, errors.Join(err, compensate(ctx, id, true))
	}
	if _, err := workflowapi.NotifyMembers(ctx, id); err != nil {
		if _, markErr := workflowapi.MarkNotificationFailed(ctx, id); markErr != nil {
			return nil, errors.Join(err, markErr)
		}
	}
	return &workflowapi.PublishResponse{Id: id.Id}, nil
}

func (w *unpublish) Execute(ctx workflow.Context) (*workflowapi.UnpublishResponse, error) {
	id := &workflowapi.ReleaseID{Id: w.Req.Id}
	if _, err := workflowapi.Close(ctx, id); err != nil {
		return nil, err
	}
	if _, err := workflowapi.DeleteShortLink(ctx, id); err != nil {
		return nil, err
	}
	if _, err := workflowapi.DeleteSnapshot(ctx, id); err != nil {
		return nil, err
	}
	return &workflowapi.UnpublishResponse{Id: id.Id}, nil
}

func compensate(ctx workflow.Context, id *workflowapi.ReleaseID, link bool) error {
	ctx, _ = workflow.NewDisconnectedContext(ctx)
	if link {
		if _, err := workflowapi.DeleteShortLink(ctx, id); err != nil {
			return err
		}
	}
	_, err := workflowapi.DeleteSnapshot(ctx, id)
	return err
}
