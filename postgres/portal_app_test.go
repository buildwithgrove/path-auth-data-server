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
					EndpointID:        "endpoint_1",
					AccountID:         pgtype.Text{String: "account_1", Valid: true},
					Plan:              pgtype.Text{String: "PLAN_UNLIMITED", Valid: true},
					SecretKeyRequired: pgtype.Bool{Bool: true, Valid: true},
					SecretKey:         pgtype.Text{String: "secret_key_1", Valid: true},
				},
				{
					EndpointID:        "endpoint_2",
					AccountID:         pgtype.Text{String: "account_2", Valid: true},
					Plan:              pgtype.Text{String: "PLAN_FREE", Valid: true},
					SecretKeyRequired: pgtype.Bool{Bool: false, Valid: true},
					SecretKey:         pgtype.Text{String: "secret_key_2", Valid: true},
					CapacityLimit:     pgtype.Int4{Int32: 30, Valid: true},
					ThroughputLimit:   pgtype.Int4{Int32: 1000, Valid: true},
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
						RateLimiting: &proto.RateLimiting{},
						Metadata: map[string]string{
							"account_id": "account_1",
							"plan_type":  "PLAN_UNLIMITED",
						},
					},
					"endpoint_2": {
						EndpointId: "endpoint_2",
						Auth: &proto.Auth{
							AuthType:        proto.Auth_NO_AUTH,
							AuthTypeDetails: &proto.Auth_NoAuth{},
						},
						RateLimiting: &proto.RateLimiting{
							ThroughputLimit:     1_000,
							CapacityLimit:       30,
							CapacityLimitPeriod: proto.CapacityLimitPeriod_CAPACITY_LIMIT_PERIOD_MONTHLY,
						},
						Metadata: map[string]string{
							"account_id": "account_2",
							"plan_type":  "PLAN_FREE",
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
