package integration_tests

import (
	internalapi "appreg/api/server/internalapi"
	publicapi "appreg/api/server/publicapi"
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestRegistryAPIs(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	client, e := publicapi.NewClientWithResponses("http://" + testEnv.conf.HTTP.Address)
	require.NoError(t, e)
	res, e := client.ListAppsWithResponse(ctx)
	require.NoError(t, e)
	require.Equal(t, http.StatusOK, res.StatusCode())
	require.Len(t, *res.JSON200, 4)
	conn, e := grpc.NewClient(testEnv.conf.GRPC.Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, e)
	defer conn.Close()
	list, e := internalapi.NewAppregServiceClient(conn).ListApps(ctx, &internalapi.ListAppsRequest{DefaultEnabledOnly: true})
	require.NoError(t, e)
	require.Len(t, list.Apps, 4)
}
