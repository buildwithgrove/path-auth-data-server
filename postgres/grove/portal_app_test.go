package grove

import (
	"testing"

	"github.com/buildwithgrove/path-external-auth-server/proto"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"

	"github.com/buildwithgrove/path-auth-data-server/postgres/grove/sqlc"
)

func Test_sqlcPortalAppsToProto(t *testing.T) {
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
					ID:                "endpoint_1_static_key",
					AccountID:         pgtype.Text{String: "account_1", Valid: true},
					Plan:              pgtype.Text{String: "PLAN_UNLIMITED", Valid: true},
					SecretKeyRequired: pgtype.Bool{Bool: true, Valid: true},
					SecretKey:         pgtype.Text{String: "secret_key_1", Valid: true},
				},
				{
					ID:                "endpoint_2_no_auth",
					AccountID:         pgtype.Text{String: "account_2", Valid: true},
					Plan:              pgtype.Text{String: "PLAN_FREE", Valid: true},
					SecretKeyRequired: pgtype.Bool{Bool: false, Valid: true},
					SecretKey:         pgtype.Text{String: "secret_key_2", Valid: true},
				},
			},
			expected: &proto.AuthDataResponse{
				Endpoints: map[string]*proto.GatewayEndpoint{
					"endpoint_1_static_key": {
						EndpointId: "endpoint_1_static_key",
						Auth: &proto.Auth{
							AuthType: &proto.Auth_StaticApiKey{
								StaticApiKey: &proto.StaticAPIKey{
									ApiKey: "secret_key_1",
								},
							},
						},
						Metadata: &proto.Metadata{
							AccountId: "account_1",
							PlanType:  "PLAN_UNLIMITED",
						},
					},
					"endpoint_2_no_auth": {
						EndpointId: "endpoint_2_no_auth",
						Auth: &proto.Auth{
							AuthType: &proto.Auth_NoAuth{},
						},
						Metadata: &proto.Metadata{
							AccountId: "account_2",
							PlanType:  "PLAN_FREE",
						},
					},
				},
			},
			wantErr: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := sqlcPortalAppsToProto(test.rows)
			require.Equal(t, test.expected, result)
		})
	}
}
