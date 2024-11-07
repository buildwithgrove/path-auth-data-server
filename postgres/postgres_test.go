package postgres

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pokt-network/poktroll/pkg/polylog/polyzero"
	"github.com/stretchr/testify/require"

	"github.com/buildwithgrove/path-auth-data-server/postgres/sqlc"
	"github.com/buildwithgrove/path/envoy/auth_server/proto"
)

var connectionString string

func TestMain(m *testing.M) {
	// Initialize the ephemeral postgres docker container
	pool, resource, databaseURL := setupPostgresDocker()
	connectionString = databaseURL

	// Run DB integration test
	exitCode := m.Run()

	// Cleanup the ephemeral postgres docker container
	cleanupPostgresDocker(m, pool, resource)
	os.Exit(exitCode)
}

func Test_Integration_FetchInitialData(t *testing.T) {
	tests := []struct {
		name     string
		expected *proto.InitialDataResponse
	}{
		{
			name: "should retrieve all initial data correctly",
			expected: &proto.InitialDataResponse{
				Endpoints: map[string]*proto.GatewayEndpoint{
					"endpoint_1": {
						EndpointId: "endpoint_1",
						UserAccount: &proto.UserAccount{
							AccountId: "account_1",
							PlanType:  "PLAN_FREE",
						},
						RateLimiting: &proto.RateLimiting{
							ThroughputLimit:     1000,
							CapacityLimit:       30,
							CapacityLimitPeriod: proto.CapacityLimitPeriod_CAPACITY_LIMIT_PERIOD_MONTHLY,
						},
						Auth: &proto.Auth{
							RequireAuth: false,
							AuthorizedUsers: map[string]*proto.Empty{
								"provider_user_1": {},
							},
						},
					},
					"endpoint_2": {
						EndpointId: "endpoint_2",
						UserAccount: &proto.UserAccount{
							AccountId: "account_2",
							PlanType:  "PLAN_UNLIMITED",
						},
						RateLimiting: &proto.RateLimiting{
							ThroughputLimit:     0,
							CapacityLimit:       0,
							CapacityLimitPeriod: proto.CapacityLimitPeriod_CAPACITY_LIMIT_PERIOD_MONTHLY,
						},
						Auth: &proto.Auth{
							RequireAuth: true,
							AuthorizedUsers: map[string]*proto.Empty{
								"provider_user_2": {},
							},
						},
					},
					"endpoint_3": {
						EndpointId: "endpoint_3",
						UserAccount: &proto.UserAccount{
							AccountId: "account_3",
							PlanType:  "PLAN_FREE",
						},
						RateLimiting: &proto.RateLimiting{
							ThroughputLimit:     1000,
							CapacityLimit:       30,
							CapacityLimitPeriod: proto.CapacityLimitPeriod_CAPACITY_LIMIT_PERIOD_MONTHLY,
						},
						Auth: &proto.Auth{
							RequireAuth: true,
							AuthorizedUsers: map[string]*proto.Empty{
								"provider_user_3": {},
							},
						},
					},
					"endpoint_4": {
						EndpointId: "endpoint_4",
						UserAccount: &proto.UserAccount{
							AccountId: "account_1",
							PlanType:  "PLAN_FREE",
						},
						RateLimiting: &proto.RateLimiting{
							ThroughputLimit:     1000,
							CapacityLimit:       30,
							CapacityLimitPeriod: proto.CapacityLimitPeriod_CAPACITY_LIMIT_PERIOD_MONTHLY,
						},
						Auth: &proto.Auth{
							RequireAuth: false,
							AuthorizedUsers: map[string]*proto.Empty{
								"provider_user_1": {},
							},
						},
					},
					"endpoint_5": {
						EndpointId: "endpoint_5",
						UserAccount: &proto.UserAccount{
							AccountId: "account_2",
							PlanType:  "PLAN_UNLIMITED",
						},
						RateLimiting: &proto.RateLimiting{
							ThroughputLimit:     0,
							CapacityLimit:       0,
							CapacityLimitPeriod: proto.CapacityLimitPeriod_CAPACITY_LIMIT_PERIOD_MONTHLY,
						},
						Auth: &proto.Auth{
							RequireAuth: true,
							AuthorizedUsers: map[string]*proto.Empty{
								"provider_user_2": {},
							},
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if testing.Short() {
				t.Skip("skipping driver integration test")
			}

			c := require.New(t)

			dataSource, cleanup, err := NewPostgresDataSource(context.Background(), connectionString, polyzero.NewLogger())
			c.NoError(err)
			defer cleanup()

			initialData, err := dataSource.FetchInitialData()
			c.NoError(err)
			c.Equal(test.expected, initialData)
		})
	}
}

func Test_convertPortalApplicationsRows(t *testing.T) {
	tests := []struct {
		name     string
		rows     []sqlc.SelectPortalApplicationsRow
		expected *proto.InitialDataResponse
		wantErr  bool
	}{
		{
			name: "should convert rows to initial data response successfully",
			rows: []sqlc.SelectPortalApplicationsRow{
				{
					EndpointID: "endpoint_1",
					AccountID:  pgtype.Text{String: "account_1", Valid: true},
					Plan:       pgtype.Text{String: "PLAN_FREE", Valid: true},
					CapacityLimit: pgtype.Int4{
						Int32: 30,
						Valid: true,
					},
					ThroughputLimit: pgtype.Int4{
						Int32: 1000,
						Valid: true,
					},
					AuthorizedUsers: []string{"provider_user_1"},
				},
			},
			expected: &proto.InitialDataResponse{
				Endpoints: map[string]*proto.GatewayEndpoint{
					"endpoint_1": {
						EndpointId: "endpoint_1",
						UserAccount: &proto.UserAccount{
							AccountId: "account_1",
							PlanType:  "PLAN_FREE",
						},
						RateLimiting: &proto.RateLimiting{
							ThroughputLimit:     1000,
							CapacityLimit:       30,
							CapacityLimitPeriod: proto.CapacityLimitPeriod_CAPACITY_LIMIT_PERIOD_MONTHLY,
						},
						Auth: &proto.Auth{
							AuthorizedUsers: map[string]*proto.Empty{
								"provider_user_1": {},
							},
						},
					},
				},
			},
			wantErr: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := convertPortalApplicationsRows(test.rows)
			require.Equal(t, test.expected, result)
		})
	}
}
