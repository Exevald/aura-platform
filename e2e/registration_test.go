package e2e

import (
	"net/http"
	"testing"
	"time"

	"github.com/oapi-codegen/runtime/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"

	"registerer/api/server/publicapi"
)

type counts struct {
	credentialsCount int
	usersCount       int
	accessesCount    int
}

func TestRegistrationCommitsEveryParticipant(t *testing.T) {
	ctx := env.Bind(t)
	before := rowCounts(t)
	response, err := deps.client.RegisterWithResponse(
		ctx,
		registration("e2e-happy", "e2e-happy@example.com"),
	)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, response.StatusCode())
	require.NotNil(t, response.JSON201)
	require.True(t, response.JSON201.Authenticated)
	requireCountsEventually(t, before, 1)
	requireEventTypesOnce(t, "UserCreated", "UserRegistered", "SessionCreated")

	duplicate, err := deps.client.RegisterWithResponse(ctx, registration("e2e-happy", "another-e2e-happy@example.com"))
	require.NoError(t, err)
	require.Equal(t, http.StatusConflict, duplicate.StatusCode())
}

func TestRegistrationFansOutPrepareAndCommit(t *testing.T) {
	ctx := env.Bind(t)
	before := rowCounts(t)
	prepareFanOutGate.arm()
	commitFanOutGate.arm()
	t.Cleanup(prepareFanOutGate.reset)
	t.Cleanup(commitFanOutGate.reset)

	response, err := deps.client.RegisterWithResponse(
		ctx,
		registration("e2e-fan-out", "e2e-fan-out@example.com"),
	)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, response.StatusCode())
	require.True(t, prepareFanOutGate.wasCompleted(), "all prepare calls must overlap")
	require.True(t, commitFanOutGate.wasCompleted(), "all commit calls must overlap")
	requireCountsEventually(t, before, 1)
	requireEventTypesOnce(t, "UserCreated", "UserRegistered", "SessionCreated")
}

func TestRegistrationFansOutRollbackAfterPrepareFailure(t *testing.T) {
	ctx := env.Bind(t)
	before := rowCounts(t)
	prepareFanOutGate.arm()
	rollbackFanOutGate.arm()
	t.Cleanup(prepareFanOutGate.reset)
	t.Cleanup(rollbackFanOutGate.reset)
	userProxy.failPrepareBefore.Store(true)

	response, err := deps.client.RegisterWithResponse(
		ctx,
		registration("e2e-fan-out-rollback", "e2e-fan-out-rollback@example.com"),
	)
	require.NoError(t, err)
	require.Equal(t, http.StatusServiceUnavailable, response.StatusCode())
	require.True(t, prepareFanOutGate.wasCompleted(), "all prepare calls must overlap")
	require.True(t, rollbackFanOutGate.wasCompleted(), "all rollback calls must overlap")
	requireCountsEventually(t, before, 0)
	requireNoEventTypes(t, "UserCreated", "UserRegistered")
}

func TestRegistrationAbortsWhenParticipantRejectsPrepare(t *testing.T) {
	userProxy.failPrepareBefore.Store(true)
	verifyAbortedRegistrationCanRetry(t, "e2e-prepare-failure")
}

func TestRegistrationAbortsAfterLostPrepareResponse(t *testing.T) {
	userProxy.failPrepareAfter.Store(true)
	verifyAbortedRegistrationCanRetry(t, "e2e-lost-prepare")
}

func TestRegistrationAbortsAfterLostAuthPrepareResponse(t *testing.T) {
	authProxy.failPrepareAfter.Store(true)
	verifyAbortedRegistrationCanRetry(t, "e2e-lost-auth-prepare")
}

func verifyAbortedRegistrationCanRetry(t *testing.T, login string) {
	t.Helper()
	ctx := env.Bind(t)
	before := rowCounts(t)
	input := registration(login, login+"@example.com")
	response, err := deps.client.RegisterWithResponse(ctx, input)
	require.NoError(t, err)
	require.Equal(t, http.StatusServiceUnavailable, response.StatusCode())
	requireCountsEventually(t, before, 0)
	requireNoEventTypes(t, "UserCreated", "UserRegistered")

	retry, err := deps.client.RegisterWithResponse(ctx, input)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, retry.StatusCode())
	requireCountsEventually(t, before, 1)
	requireEventTypesOnce(t, "UserCreated", "UserRegistered", "SessionCreated")
}

func TestRecoveryRedeliversCommitDecision(t *testing.T) {
	ctx := env.Bind(t)
	before := rowCounts(t)
	userProxy.failCommitRemaining.Store(1)
	response, err := deps.client.RegisterWithResponse(ctx, registration("e2e-commit-recovery", "e2e-commit-recovery@example.com"))
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, response.StatusCode())
	requireCountsEventually(t, before, 1)
	requireCommittedDecisionEventually(t)
	requireEventTypesOnce(t, "UserCreated", "UserRegistered", "SessionCreated")
}

func TestSessionFailureDoesNotAbortRegistration(t *testing.T) {
	ctx := env.Bind(t)
	before := rowCounts(t)
	sessions.failCreate.Store(true)
	response, err := deps.client.RegisterWithResponse(ctx, registration("e2e-session-failure", "e2e-session-failure@example.com"))
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, response.StatusCode())
	require.NotNil(t, response.JSON201)
	require.False(t, response.JSON201.Authenticated)
	requireCountsEventually(t, before, 1)
	requireEventTypesOnce(t, "UserCreated", "UserRegistered")
}

func TestRecoverySurvivesCoordinatorRestart(t *testing.T) {
	ctx := env.Bind(t)
	before := rowCounts(t)
	userProxy.failCommitRemaining.Store(100)
	response, err := deps.client.RegisterWithResponse(ctx, registration("e2e-coordinator-restart", "e2e-coordinator-restart@example.com"))
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, response.StatusCode())

	testcontainers.RequireContainerExec(ctx, t, registererServiceContainer, []string{"sh", "-c", "kill -9 $(cat /tmp/registerer.pid)"})
	userProxy.failCommitRemaining.Store(0)
	requireCountsEventually(t, before, 1)
	requireCommittedDecisionEventually(t)
	requireEventTypesOnce(t, "UserCreated", "UserRegistered", "SessionCreated")
}

func requireCommittedDecisionEventually(t *testing.T) {
	t.Helper()
	ctx := env.Bind(t)
	require.EventuallyWithT(t, func(collect *assert.CollectT) {
		var state string
		err := deps.dbs[registererModule].QueryRowContext(ctx, "SELECT state FROM registration_transaction ORDER BY created_at DESC LIMIT 1").Scan(&state)
		require.NoError(collect, err)
		require.Equal(collect, "committed", state)
	}, 5*time.Second, 50*time.Millisecond)
}

func registration(login, email string) publicapi.RegisterJSONRequestBody {
	return publicapi.RegisterJSONRequestBody{
		Login: login, Password: "correct horse battery staple", DisplayName: login,
		FirstName: "End", LastName: "ToEnd", Email: types.Email(email), RememberMe: false,
	}
}

func rowCounts(t *testing.T) counts {
	t.Helper()
	ctx := env.Bind(t)
	var result counts
	require.NoError(t, deps.dbs[authModule].QueryRowContext(ctx, "SELECT COUNT(*) FROM credential").Scan(&result.credentialsCount))
	require.NoError(t, deps.dbs[userModule].QueryRowContext(ctx, "SELECT COUNT(*) FROM user WHERE deleted_at IS NULL").Scan(&result.usersCount))
	require.NoError(t, deps.dbs[accessModule].QueryRowContext(ctx, "SELECT COUNT(*) FROM module_access").Scan(&result.accessesCount))
	return result
}

func requireCountsEventually(t *testing.T, before counts, increment int) {
	t.Helper()
	require.EventuallyWithT(t, func(collect *assert.CollectT) {
		current := rowCountsCollect(collect, t)
		require.Equal(collect, before.credentialsCount+increment, current.credentialsCount)
		require.Equal(collect, before.usersCount+increment, current.usersCount)
		require.Equal(collect, before.accessesCount, current.accessesCount)
	}, 5*time.Second, 50*time.Millisecond)
}

func rowCountsCollect(collect *assert.CollectT, t *testing.T) counts {
	ctx := env.Bind(t)
	var result counts
	require.NoError(
		collect,
		deps.dbs[authModule].QueryRowContext(ctx, "SELECT COUNT(*) FROM credential").Scan(&result.credentialsCount),
	)
	require.NoError(
		collect,
		deps.dbs[userModule].QueryRowContext(ctx, "SELECT COUNT(*) FROM user WHERE deleted_at IS NULL").Scan(&result.usersCount),
	)
	require.NoError(
		collect,
		deps.dbs[accessModule].QueryRowContext(ctx, "SELECT COUNT(*) FROM module_access").Scan(&result.accessesCount),
	)
	return result
}

func requireEventTypesOnce(t *testing.T, expected ...string) {
	t.Helper()
	want := make(map[string]int, len(expected))
	for _, eventType := range expected {
		want[eventType] = 0
	}
	deadline := time.NewTimer(5 * time.Second)
	defer deadline.Stop()
	for {
		complete := true
		for _, count := range want {
			complete = complete && count == 1
		}
		if complete {
			settle := time.NewTimer(300 * time.Millisecond)
			for {
				select {
				case delivery := <-deps.events:
					if _, tracked := want[delivery.Type]; tracked {
						want[delivery.Type]++
						require.LessOrEqual(t, want[delivery.Type], 1, "event %s was published more than once", delivery.Type)
					}
				case <-settle.C:
					return
				}
			}
		}
		select {
		case delivery := <-deps.events:
			if _, tracked := want[delivery.Type]; tracked {
				want[delivery.Type]++
				require.LessOrEqual(t, want[delivery.Type], 1, "event %s was published more than once", delivery.Type)
			}
		case <-deadline.C:
			require.FailNow(t, "expected domain events were not published", "%v", want)
		}
	}
}

func requireNoEventTypes(t *testing.T, unexpected ...string) {
	t.Helper()
	blocked := make(map[string]struct{}, len(unexpected))
	for _, eventType := range unexpected {
		blocked[eventType] = struct{}{}
	}
	timer := time.NewTimer(300 * time.Millisecond)
	defer timer.Stop()
	for {
		select {
		case delivery := <-deps.events:
			if _, forbidden := blocked[delivery.Type]; forbidden {
				require.FailNow(t, "aborted registration published an event", delivery.Type)
			}
		case <-timer.C:
			return
		}
	}
}
