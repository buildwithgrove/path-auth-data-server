package postgres

import (
	"testing"

	"github.com/buildwithgrove/path/envoy/auth_server/proto"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"

	grpc_server "github.com/buildwithgrove/path-auth-data-server/grpc"
	"github.com/buildwithgrove/path-auth-data-server/postgres/sqlc"
)

func Test_convertGatewayEndpointsRows(t *testing.T) {
	tests := []struct {
		name     string
		rows     []sqlc.SelectGatewayEndpointsRow
		expected *proto.AuthDataResponse
		wantErr  bool
	}{
		{
			name: "should convert rows to auth data response successfully",
			rows: []sqlc.SelectGatewayEndpointsRow{
				{
					EndpointID:          "endpoint_1",
					AuthType:            grpc_server.AuthTypeAPIKey,
					ApiKey:              pgtype.Text{String: "secret_key_1", Valid: true},
					ThroughputLimit:     pgtype.Int4{Int32: 1000, Valid: true},
					CapacityLimit:       pgtype.Int4{Int32: 30, Valid: true},
					CapacityLimitPeriod: grpc_server.CapacityLimitPeriodMonthly,
					AuthorizedUsers:     []string{"user1", "user2"},
				},
				{
					EndpointID:          "endpoint_2",
					AuthType:            grpc_server.AuthTypeNoAuth,
					ApiKey:              pgtype.Text{String: "", Valid: false},
					ThroughputLimit:     pgtype.Int4{Int32: 500, Valid: true},
					CapacityLimit:       pgtype.Int4{Int32: 20, Valid: true},
					CapacityLimitPeriod: grpc_server.CapacityLimitPeriodDaily,
					AuthorizedUsers:     []string{},
				},
			},
			expected: &proto.AuthDataResponse{
				Endpoints: map[string]*proto.GatewayEndpoint{
					"endpoint_1": {
						EndpointId: "endpoint_1",
						Auth: &proto.Auth{
							AuthType: proto.Auth_API_KEY_AUTH,
							AuthTypeDetails: &proto.Auth_ApiKey{
								ApiKey: &proto.APIKey{
									ApiKey: "secret_key_1",
								},
							},
						},
						RateLimiting: &proto.RateLimiting{
							ThroughputLimit:     1000,
							CapacityLimit:       30,
							CapacityLimitPeriod: proto.CapacityLimitPeriod_CAPACITY_LIMIT_PERIOD_MONTHLY,
						},
					},
					"endpoint_2": {
						EndpointId: "endpoint_2",
						Auth: &proto.Auth{
							AuthType:        proto.Auth_NO_AUTH,
							AuthTypeDetails: &proto.Auth_NoAuth{},
						},
						RateLimiting: &proto.RateLimiting{
							ThroughputLimit:     500,
							CapacityLimit:       20,
							CapacityLimitPeriod: proto.CapacityLimitPeriod_CAPACITY_LIMIT_PERIOD_DAILY,
						},
					},
				},
			},
			wantErr: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := convertGatewayEndpointsRows(test.rows)
			require.Equal(t, test.expected, result)
		})
	}
}
