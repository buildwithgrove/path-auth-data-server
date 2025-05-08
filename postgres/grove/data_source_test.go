package grove

import (
	"context"
	"flag"
	"os"
	"testing"

	"github.com/pokt-network/poktroll/pkg/polylog/polyzero"
	"github.com/stretchr/testify/require"

	"github.com/buildwithgrove/path-external-auth-server/proto"
)

var connectionString string

func TestMain(m *testing.M) {
	flag.Parse()
	if testing.Short() {
		return
	}

	// Initialize the ephemeral postgres docker container
	pool, resource, databaseURL := setupPostgresDocker()
	connectionString = databaseURL

	// Run DB integration test
	exitCode := m.Run()

	// Cleanup the ephemeral postgres docker container
	cleanupPostgresDocker(m, pool, resource)
	os.Exit(exitCode)
}

func Test_Integration_FetchAuthDataSync(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping driver integration test")
	}

	tests := []struct {
		name     string
		expected *proto.AuthDataResponse
	}{
		{
			name: "should retrieve all gateway endpoints data correctly",
			expected: &proto.AuthDataResponse{
				Endpoints: map[string]*proto.GatewayEndpoint{
					"endpoint_1_no_auth": {
						EndpointId: "endpoint_1_no_auth",
						Auth: &proto.Auth{
							AuthType: &proto.Auth_NoAuth{},
						},
						Metadata: &proto.Metadata{
							AccountId: "account_1",
							PlanType:  "PLAN_FREE",
						},
					},
					"endpoint_2_static_key": {
						EndpointId: "endpoint_2_static_key",
						Auth: &proto.Auth{
							AuthType: &proto.Auth_StaticApiKey{
								StaticApiKey: &proto.StaticAPIKey{
									ApiKey: "secret_key_2",
								},
							},
						},
						Metadata: &proto.Metadata{
							AccountId: "account_2",
							PlanType:  "PLAN_UNLIMITED",
						},
					},
					"endpoint_3_static_key": {
						EndpointId: "endpoint_3_static_key",
						Auth: &proto.Auth{
							AuthType: &proto.Auth_StaticApiKey{
								StaticApiKey: &proto.StaticAPIKey{
									ApiKey: "secret_key_3",
								},
							},
						},
						Metadata: &proto.Metadata{
							AccountId: "account_3",
							PlanType:  "PLAN_FREE",
						},
					},
					"endpoint_4_no_auth": {
						EndpointId: "endpoint_4_no_auth",
						Auth: &proto.Auth{
							AuthType: &proto.Auth_NoAuth{},
						},
						Metadata: &proto.Metadata{
							AccountId: "account_1",
							PlanType:  "PLAN_FREE",
						},
					},
					"endpoint_5_static_key": {
						EndpointId: "endpoint_5_static_key",
						Auth: &proto.Auth{
							AuthType: &proto.Auth_StaticApiKey{
								StaticApiKey: &proto.StaticAPIKey{
									ApiKey: "secret_key_5",
								},
							},
						},
						Metadata: &proto.Metadata{
							AccountId: "account_2",
							PlanType:  "PLAN_UNLIMITED",
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)

			dataSource, _, err := NewGrovePostgresDataSource(context.Background(), connectionString, polyzero.NewLogger())
			c.NoError(err)

			authData, err := dataSource.FetchAuthDataSync()
			c.NoError(err)
			c.Equal(test.expected, authData)
		})
	}
}
