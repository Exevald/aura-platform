package integration_tests

import (
	"net/http"
	"testing"
	"time"
	"uuid"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	accessapi "access/api/server/internalapi"
)

func TestServiceStartsAndAppliesMigrations(t *testing.T) {
	client := &http.Client{Timeout: time.Second}
	response, err := client.Get("http://" + testEnv.conf.HTTP.Address + "/healthz")
	require.NoError(t, err)
	require.NoError(t, response.Body.Close())
	require.Equal(t, http.StatusNoContent, response.StatusCode)

	var tableName string
	err = testEnv.connector.TransactionalClient().GetContext(
		t.Context(),
		"test.business-table",
		&tableName,
		"SELECT table_name FROM information_schema.tables WHERE table_schema = ? AND table_name = ?",
		testEnv.conf.MySQL.Database,
		"module_access",
	)
	require.NoError(t, err)
	require.Equal(t, "module_access", tableName)
}

func TestRegistrationDoesNotGrantModuleAccess(t *testing.T) {
	connection, err := grpc.NewClient(testEnv.conf.GRPC.Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, connection.Close()) })
	client := accessapi.NewModuleAccessServiceClient(connection)
	transactionID, userID, moduleID := uuid.New(), uuid.New(), uuid.New()

	_, err = client.PrepareModuleAccessRegistration(t.Context(), &accessapi.PrepareModuleAccessRegistrationRequest{
		TransactionId: transactionID.String(), UserId: userID.String(), ModuleId: moduleID.String(),
	})
	require.NoError(t, err)
	_, err = client.CommitModuleAccessRegistration(t.Context(), &accessapi.RegistrationTransactionRequest{TransactionId: transactionID.String()})
	require.NoError(t, err)
	check, err := client.CheckModuleAccess(t.Context(), &accessapi.CheckModuleAccessRequest{
		UserId: userID.String(), ModuleId: moduleID.String(),
	})
	require.NoError(t, err)
	require.False(t, check.GetAllowed())
}
