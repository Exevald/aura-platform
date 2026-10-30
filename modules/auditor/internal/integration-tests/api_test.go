package integration_tests

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
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
		"audit_record",
	)
	require.NoError(t, err)
	require.Equal(t, "audit_record", tableName)
}
