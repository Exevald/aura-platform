package e2e

import (
	"context"
	"net/http"
	"testing"
	"time"
	"uuid"

	"github.com/oapi-codegen/runtime/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	accessapi "access/api/server/internalapi"
	accesspublicapi "access/api/server/publicapi"
	usergroupapi "usergroup/api/server/publicapi"
)

const editorModuleID = "b8547c76-51fd-4f7c-9c68-8f30c84f61cb"

type groupModuleScenario struct {
	t         *testing.T
	ctx       context.Context
	owner     uuid.UUID
	workspace uuid.UUID
	module    uuid.UUID
}

func TestGroupModuleAccessFollowsMembershipAndRevocation(t *testing.T) {
	s := &groupModuleScenario{
		t: t, ctx: env.Bind(t), owner: uuid.New(), workspace: uuid.New(), module: uuid.MustParse(editorModuleID),
	}
	first := s.registerMember("e2e-group-first")
	second := s.registerMember("e2e-group-second")
	group := s.createGroup("E2E primary")
	otherGroup := s.createGroup("E2E secondary")

	s.verifyInitialGrant(group, first)
	s.verifyMembershipChanges(group, first, second)
	s.verifyIndependentGroupSource(group, otherGroup, second)
	s.verifyGroupDeletionAndAuthorization(group, otherGroup, first, second)
}

func (s *groupModuleScenario) verifyInitialGrant(group, first uuid.UUID) {
	s.t.Helper()
	s.addMember(group, "e2e-group-first@example.com")
	require.Equal(s.t, http.StatusOK, s.setModules(s.owner, group, s.module))
	s.checkAccess(s.owner, true)
	s.checkAccess(first, true)
	require.Equal(s.t, 1, s.sourceCount(group, first))
	require.Equal(s.t, http.StatusOK, s.setModules(s.owner, group, s.module))
	require.Equal(s.t, 1, s.sourceCount(group, first))
}

func (s *groupModuleScenario) verifyMembershipChanges(group, first, second uuid.UUID) {
	s.t.Helper()
	s.addMember(group, "e2e-group-second@example.com")
	s.checkAccess(second, true)
	s.removeMember(group, first)
	s.checkAccess(first, false)
}

func (s *groupModuleScenario) verifyIndependentGroupSource(group, otherGroup, second uuid.UUID) {
	s.t.Helper()
	s.addMember(otherGroup, "e2e-group-second@example.com")
	require.Equal(s.t, http.StatusOK, s.setModules(s.owner, otherGroup, s.module))
	require.Eventually(s.t, func() bool { return s.sourceCount(otherGroup, second) == 1 }, 5*time.Second, 50*time.Millisecond)
	require.Equal(s.t, http.StatusOK, s.setModules(s.owner, group))
	s.checkAccess(second, true)
	require.Equal(s.t, 0, s.sourceCount(group, second))
}

func (s *groupModuleScenario) verifyGroupDeletionAndAuthorization(group, otherGroup, first, second uuid.UUID) {
	s.t.Helper()
	s.deleteGroup(otherGroup)
	s.checkAccess(second, false)
	s.checkAccess(s.owner, false)
	require.Equal(s.t, http.StatusNotFound, s.setModules(s.owner, otherGroup, s.module))
	require.Equal(s.t, http.StatusNotFound, s.setModules(first, group, s.module))
}

func (s *groupModuleScenario) registerMember(name string) uuid.UUID {
	s.t.Helper()
	response, err := deps.client.RegisterWithResponse(s.ctx, registration(name, name+"@example.com"))
	require.NoError(s.t, err)
	require.Equal(s.t, http.StatusCreated, response.StatusCode())
	require.NotNil(s.t, response.JSON201)
	requireEventTypesOnce(s.t, "UserCreated", "UserRegistered", "SessionCreated")
	return uuid.UUID(response.JSON201.UserID)
}

func (s *groupModuleScenario) createGroup(name string) uuid.UUID {
	s.t.Helper()
	response, err := deps.groups.CreateGroupWithResponse(s.ctx,
		&usergroupapi.CreateGroupParams{XAuthUser: types.UUID(s.owner), XWorkspaceID: types.UUID(s.workspace)},
		usergroupapi.GroupInput{Name: name})
	require.NoError(s.t, err)
	require.Equal(s.t, http.StatusCreated, response.StatusCode())
	require.NotNil(s.t, response.JSON201)
	return uuid.UUID(response.JSON201.Id)
}

func (s *groupModuleScenario) addMember(group uuid.UUID, email string) {
	s.t.Helper()
	response, err := deps.groups.AddGroupMemberWithResponse(s.ctx, types.UUID(group),
		&usergroupapi.AddGroupMemberParams{XAuthUser: types.UUID(s.owner), XWorkspaceID: types.UUID(s.workspace)},
		usergroupapi.MemberInput{Email: types.Email(email)})
	require.NoError(s.t, err)
	require.Equal(s.t, http.StatusOK, response.StatusCode())
}

func (s *groupModuleScenario) removeMember(group, member uuid.UUID) {
	s.t.Helper()
	response, err := deps.groups.RemoveGroupMemberWithResponse(s.ctx, types.UUID(group), types.UUID(member),
		&usergroupapi.RemoveGroupMemberParams{XAuthUser: types.UUID(s.owner), XWorkspaceID: types.UUID(s.workspace)})
	require.NoError(s.t, err)
	require.Equal(s.t, http.StatusNoContent, response.StatusCode())
}

func (s *groupModuleScenario) deleteGroup(group uuid.UUID) {
	s.t.Helper()
	response, err := deps.groups.DeleteGroupWithResponse(s.ctx, types.UUID(group),
		&usergroupapi.DeleteGroupParams{XAuthUser: types.UUID(s.owner), XWorkspaceID: types.UUID(s.workspace)})
	require.NoError(s.t, err)
	require.Equal(s.t, http.StatusNoContent, response.StatusCode())
}

func (s *groupModuleScenario) setModules(actor, group uuid.UUID, modules ...uuid.UUID) int {
	s.t.Helper()
	ids := make([]types.UUID, 0, len(modules))
	for _, id := range modules {
		ids = append(ids, types.UUID(id))
	}
	response, err := deps.access.SetGroupModulesWithResponse(s.ctx, types.UUID(group),
		&accesspublicapi.SetGroupModulesParams{XAuthUser: types.UUID(actor)},
		accesspublicapi.SetGroupModulesInput{ModuleIDs: ids})
	require.NoError(s.t, err)
	return response.StatusCode()
}

func (s *groupModuleScenario) checkAccess(user uuid.UUID, expected bool) {
	s.t.Helper()
	require.EventuallyWithT(s.t, func(c *assert.CollectT) {
		response, err := deps.accessGRPC.CheckModuleAccess(s.ctx,
			&accessapi.CheckModuleAccessRequest{UserId: user.String(), ModuleId: s.module.String()})
		if assert.NoError(c, err) {
			assert.Equal(c, expected, response.GetAllowed())
		}
	}, 5*time.Second, 50*time.Millisecond)
}

func (s *groupModuleScenario) sourceCount(group, user uuid.UUID) int {
	s.t.Helper()
	var count int
	err := deps.dbs[accessModule].QueryRowContext(s.ctx, `SELECT COUNT(*) FROM module_access a JOIN module_share_operation s ON a.source_id = CONCAT('group:', BIN_TO_UUID(s.share_id)) WHERE s.group_id = ? AND a.user_id = ? AND a.module_id = ?`, group[:], user[:], s.module[:]).Scan(&count)
	require.NoError(s.t, err)
	return count
}
