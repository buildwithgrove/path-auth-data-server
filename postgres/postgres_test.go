package postgres

import (
	"context"
	"flag"
	"os"
	"testing"

	"github.com/pokt-network/poktroll/pkg/polylog/polyzero"
	"github.com/stretchr/testify/require"

	"github.com/buildwithgrove/path/envoy/auth_server/proto"
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
					"endpoint_1": {
						EndpointId: "endpoint_1",
						Auth: &proto.Auth{
							AuthType: &proto.Auth_NoAuth{},
						},
						RateLimiting: &proto.RateLimiting{
							ThroughputLimit:     1000,
							CapacityLimit:       30,
							CapacityLimitPeriod: proto.CapacityLimitPeriod_CAPACITY_LIMIT_PERIOD_MONTHLY,
						},
						Metadata: &proto.Metadata{
							AccountId: "account_1",
							PlanType:  "PLAN_FREE",
						},
					},
					"endpoint_2": {
						EndpointId: "endpoint_2",
						Auth: &proto.Auth{
							AuthType: &proto.Auth_StaticApiKey{
								StaticApiKey: &proto.StaticAPIKey{
									ApiKey: "secr	et_key_2",
								},
							},
						},
						RateLimiting: &proto.RateLimiting{},
						Metadata: &proto.Metadata{
							AccountId: "account_2",
							PlanType:  "PLAN_UNLIMITED",
						},
					},
					"endpoint_3": {
						EndpointId: "endpoint_3",
						Auth: &proto.Auth{
							AuthType: &proto.Auth_StaticApiKey{
								StaticApiKey: &proto.StaticAPIKey{
									ApiKey: "secret_key_3",
								},
							},
						},
						RateLimiting: &proto.RateLimiting{
							ThroughputLimit:     1000,
							CapacityLimit:       30,
							CapacityLimitPeriod: proto.CapacityLimitPeriod_CAPACITY_LIMIT_PERIOD_MONTHLY,
						},
						Metadata: &proto.Metadata{
							AccountId: "account_3",
							PlanType:  "PLAN_FREE",
						},
					},
					"endpoint_4": {
						EndpointId: "endpoint_4",
						Auth: &proto.Auth{
							AuthType: &proto.Auth_NoAuth{},
						},
						RateLimiting: &proto.RateLimiting{
							ThroughputLimit:     1000,
							CapacityLimit:       30,
							CapacityLimitPeriod: proto.CapacityLimitPeriod_CAPACITY_LIMIT_PERIOD_MONTHLY,
						},
						Metadata: &proto.Metadata{
							AccountId: "account_1",
							PlanType:  "PLAN_FREE",
						},
					},
					"endpoint_5": {
						EndpointId: "endpoint_5",
						Auth: &proto.Auth{
							AuthType: &proto.Auth_StaticApiKey{
								StaticApiKey: &proto.StaticAPIKey{
									ApiKey: "secret_key_5",
								},
							},
						},
						RateLimiting: &proto.RateLimiting{},
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

			dataSource, _, err := NewPostgresDataSource(context.Background(), connectionString, polyzero.NewLogger())
			c.NoError(err)

			authData, err := dataSource.FetchAuthDataSync()
			c.NoError(err)
			c.Equal(test.expected, authData)
		})
	}
}
