package postgres

import (
	"testing"

	"github.com/buildwithgrove/path/envoy/auth_server/proto"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"

	"github.com/buildwithgrove/path-auth-data-server/postgres/sqlc"
)

func Test_convertPortalApplicationsRows(t *testing.T) {
	tests := []struct {
		name     string
		rows     []sqlc.SelectPortalApplicationsRow
		expected *proto.AuthDataResponse
		wantErr  bool
	}{
		{
			name: "should convert rows to auth data response successfully",
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
			expected: &proto.AuthDataResponse{
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
