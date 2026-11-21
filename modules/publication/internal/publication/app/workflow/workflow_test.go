package workflow

import (
	"context"
	"testing"

	workflowapi "publication/api/workflow"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/testsuite"
)

func TestPublishRunsAllSteps(t *testing.T) {
	env := new(testsuite.WorkflowTestSuite).NewTestWorkflowEnvironment()
	client := workflowapi.NewTestPublicationFlowClient(env, &Workflows{}, &Activities{})
	id := &workflowapi.ReleaseID{Id: "release-1"}
	env.OnActivity(workflowapi.ExportSnapshotActivityName, mock.Anything, id).Return(&workflowapi.Snapshot{Id: id.Id}, nil).Once()
	env.OnActivity(workflowapi.CreateShortLinkActivityName, mock.Anything, id).Return(&workflowapi.ShortLink{Id: "link-1"}, nil).Once()
	env.OnActivity(workflowapi.ActivateActivityName, mock.Anything, id).Return(id, nil).Once()
	env.OnActivity(workflowapi.NotifyMembersActivityName, mock.Anything, id).Return(id, nil).Once()
	result, err := client.Publish(context.Background(), &workflowapi.PublishRequest{Id: id.Id})
	require.NoError(t, err)
	require.Equal(t, id.Id, result.Id)
	env.AssertExpectations(t)
}

func TestPublishCompensatesFailedLink(t *testing.T) {
	env := new(testsuite.WorkflowTestSuite).NewTestWorkflowEnvironment()
	client := workflowapi.NewTestPublicationFlowClient(env, &Workflows{}, &Activities{})
	id := &workflowapi.ReleaseID{Id: "release-2"}
	env.OnActivity(workflowapi.ExportSnapshotActivityName, mock.Anything, id).Return(&workflowapi.Snapshot{Id: id.Id}, nil).Once()
	env.OnActivity(workflowapi.CreateShortLinkActivityName, mock.Anything, id).Return((*workflowapi.ShortLink)(nil), temporal.NewNonRetryableApplicationError("link failed", "Conflict", nil)).Once()
	env.OnActivity(workflowapi.DeleteShortLinkActivityName, mock.Anything, id).Return(id, nil).Once()
	env.OnActivity(workflowapi.DeleteSnapshotActivityName, mock.Anything, id).Return(id, nil).Once()
	_, err := client.Publish(context.Background(), &workflowapi.PublishRequest{Id: id.Id})
	require.ErrorContains(t, err, "link failed")
	env.AssertExpectations(t)
}

func TestNotificationFailureKeepsPublication(t *testing.T) {
	env := new(testsuite.WorkflowTestSuite).NewTestWorkflowEnvironment()
	client := workflowapi.NewTestPublicationFlowClient(env, &Workflows{}, &Activities{})
	id := &workflowapi.ReleaseID{Id: "release-3"}
	env.OnActivity(workflowapi.ExportSnapshotActivityName, mock.Anything, id).Return(&workflowapi.Snapshot{Id: id.Id}, nil).Once()
	env.OnActivity(workflowapi.CreateShortLinkActivityName, mock.Anything, id).Return(&workflowapi.ShortLink{Id: "link-3"}, nil).Once()
	env.OnActivity(workflowapi.ActivateActivityName, mock.Anything, id).Return(id, nil).Once()
	env.OnActivity(workflowapi.NotifyMembersActivityName, mock.Anything, id).Return((*workflowapi.ReleaseID)(nil), temporal.NewNonRetryableApplicationError("delivery failed", "Unavailable", nil)).Once()
	env.OnActivity(workflowapi.MarkNotificationFailedActivityName, mock.Anything, id).Return(id, nil).Once()
	result, err := client.Publish(context.Background(), &workflowapi.PublishRequest{Id: id.Id})
	require.NoError(t, err)
	require.Equal(t, id.Id, result.Id)
	env.AssertExpectations(t)
}
