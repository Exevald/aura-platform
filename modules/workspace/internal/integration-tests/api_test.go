package integration_tests

import (
	"context"
	"net/http"
	"testing"
	"time"
	"uuid"

	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	internalapi "workspace/api/server/internalapi"
	publicapi "workspace/api/server/publicapi"
)

func TestWorkspaceLifecycle(t *testing.T) {
	userID, workspaceID := uuid.New(), uuid.New()
	connection, err := grpc.NewClient(testEnv.conf.GRPC.Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer connection.Close()
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	verifyPersonalWorkspace(ctx, t, internalapi.NewWorkspaceServiceClient(connection), userID, workspaceID)
	client, err := publicapi.NewClientWithResponses("http://" + testEnv.conf.HTTP.Address)
	require.NoError(t, err)

	verifyTeamInvitation(ctx, t, client, userID)
}

func verifyPersonalWorkspace(ctx context.Context, t *testing.T, client internalapi.WorkspaceServiceClient, userID, workspaceID uuid.UUID) {
	t.Helper()
	created, err := client.EnsurePersonalWorkspace(ctx, &internalapi.EnsurePersonalWorkspaceRequest{
		WorkspaceId: workspaceID.String(), UserId: userID.String(), Name: "Личное пространство",
	}, grpc.WaitForReady(true))
	require.NoError(t, err)
	require.Equal(t, workspaceID.String(), created.WorkspaceId)
	check, err := client.CheckMembership(ctx, &internalapi.CheckMembershipRequest{WorkspaceId: workspaceID.String(), UserId: userID.String()})
	require.NoError(t, err)
	require.True(t, check.Allowed)
	require.Equal(t, "owner", check.Role)
}

func verifyTeamInvitation(ctx context.Context, t *testing.T, client *publicapi.ClientWithResponses, userID uuid.UUID) {
	t.Helper()
	response, err := client.ListWorkspacesWithResponse(ctx, &publicapi.ListWorkspacesParams{XAuthUser: openapi_types.UUID(userID)})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, response.StatusCode())
	require.Len(t, *response.JSON200, 1)
	team, err := client.CreateWorkspaceWithResponse(ctx, &publicapi.CreateWorkspaceParams{XAuthUser: openapi_types.UUID(userID)}, publicapi.CreateWorkspaceJSONRequestBody{Name: "Команда"})
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, team.StatusCode())
	require.Equal(t, publicapi.Team, team.JSON201.Kind)
	invitee := uuid.New()
	invite, err := client.CreateWorkspaceInvitationWithResponse(ctx, team.JSON201.Id, &publicapi.CreateWorkspaceInvitationParams{XAuthUser: openapi_types.UUID(userID)}, publicapi.CreateWorkspaceInvitationJSONRequestBody{Email: "invitee@example.test", UserID: ptrUUID(invitee), Role: publicapi.CreateInvitationRoleEditor})
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, invite.StatusCode())
	invitations, err := client.ListWorkspaceInvitationsWithResponse(ctx, team.JSON201.Id, &publicapi.ListWorkspaceInvitationsParams{XAuthUser: openapi_types.UUID(userID)})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, invitations.StatusCode(), string(invitations.Body))
	require.Len(t, *invitations.JSON200, 1)
	require.Equal(t, invite.JSON201.Token, (*invitations.JSON200)[0].Token)
	require.NotNil(t, (*invitations.JSON200)[0].UserID)
	require.Equal(t, openapi_types.UUID(invitee), *(*invitations.JSON200)[0].UserID)
	accepted, err := client.AcceptWorkspaceInvitationWithResponse(ctx, invite.JSON201.Token, &publicapi.AcceptWorkspaceInvitationParams{XAuthUser: openapi_types.UUID(invitee)})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, accepted.StatusCode(), string(accepted.Body))
	require.Equal(t, publicapi.RoleEditor, accepted.JSON200.Role)
}

func ptrUUID(value uuid.UUID) *openapi_types.UUID {
	converted := openapi_types.UUID(value)
	return &converted
}
