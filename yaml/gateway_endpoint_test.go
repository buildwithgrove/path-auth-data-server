package yaml

import (
	"testing"

	"github.com/buildwithgrove/path/envoy/auth_server/proto"
	"github.com/stretchr/testify/require"

	grpc_server "github.com/buildwithgrove/path-auth-data-server/grpc"
)

func Test_gatewayEndpointYAML_convertToProto(t *testing.T) {
	tests := []struct {
		name       string
		endpointID string
		input      gatewayEndpointYAML
		expected   *proto.GatewayEndpoint
	}{
		{
			name:       "should convert gatewayEndpointYAML to proto format correctly",
			endpointID: "endpoint_1",
			input: gatewayEndpointYAML{
				Auth: authYAML{
					AuthType: "AUTH_TYPE_JWT",
					JWTAuthorizedUsers: []string{
						"auth0|user_1",
					},
				},
				RateLimiting: rateLimitingYAML{
					ThroughputLimit:     30,
					CapacityLimit:       100_000,
					CapacityLimitPeriod: grpc_server.CapacityLimitPeriodMonthly,
				},
				Metadata: metadataYAML{
					Name:      "grove_city_test_endpoint",
					AccountId: "account_1",
					PlanType:  "PLAN_UNLIMITED",
				},
			},
			expected: &proto.GatewayEndpoint{
				EndpointId: "endpoint_1",
				Auth: &proto.Auth{
					AuthType: proto.Auth_AUTH_TYPE_JWT,
					AuthTypeDetails: &proto.Auth_Jwt{
						Jwt: &proto.JWT{
							AuthorizedUsers: map[string]*proto.Empty{
								"auth0|user_1": {},
							},
						},
					},
				},
				RateLimiting: &proto.RateLimiting{
					ThroughputLimit:     30,
					CapacityLimit:       100_000,
					CapacityLimitPeriod: proto.CapacityLimitPeriod_CAPACITY_LIMIT_PERIOD_MONTHLY,
				},
				Metadata: &proto.Metadata{
					Name:      "grove_city_test_endpoint",
					AccountId: "account_1",
					PlanType:  "PLAN_UNLIMITED",
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)

			result := test.input.convertToProto(test.endpointID)
			c.Equal(test.expected, result)
		})
	}
}

func Test_authYAML_convertToProto(t *testing.T) {
	tests := []struct {
		name     string
		input    authYAML
		expected *proto.Auth
	}{
		{
			name: "should convert authYAML to proto format correctly",
			input: authYAML{
				AuthType: "AUTH_TYPE_JWT",
				JWTAuthorizedUsers: []string{
					"auth0|user_1",
					"auth0|user_2",
				},
			},
			expected: &proto.Auth{
				AuthType: proto.Auth_AUTH_TYPE_JWT,
				AuthTypeDetails: &proto.Auth_Jwt{
					Jwt: &proto.JWT{
						AuthorizedUsers: map[string]*proto.Empty{
							"auth0|user_1": {},
							"auth0|user_2": {},
						},
					},
				},
			},
		},
		{
			name: "should handle empty authorized users",
			input: authYAML{
				AuthType: "NO_AUTH",
			},
			expected: &proto.Auth{
				AuthType:        proto.Auth_AUTH_TYPE_UNSPECIFIED,
				AuthTypeDetails: &proto.Auth_NoAuth{},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)

			result := test.input.convertToProto()
			c.Equal(test.expected, result)
		})
	}
}

func Test_gatewayEndpointYAML_validate(t *testing.T) {
	tests := []struct {
		name       string
		endpointID string
		input      gatewayEndpointYAML
		wantErr    bool
	}{
		{
			name:       "valid endpoint with AUTH_TYPE_JWT",
			endpointID: "endpoint_1",
			input: gatewayEndpointYAML{
				Auth: authYAML{
					AuthType: "AUTH_TYPE_JWT",
					JWTAuthorizedUsers: []string{
						"auth0|user_1",
					},
				},
				RateLimiting: rateLimitingYAML{
					ThroughputLimit:     30,
					CapacityLimit:       100_000,
					CapacityLimitPeriod: grpc_server.CapacityLimitPeriodMonthly,
				},
			},
			wantErr: false,
		},
		{
			name:       "missing endpoint_id",
			endpointID: "",
			input: gatewayEndpointYAML{
				Auth: authYAML{
					AuthType: "AUTH_TYPE_API_KEY",
					APIKey:   stringPtr("some_api_key"),
				},
			},
			wantErr: true,
		},
		{
			name:       "invalid capacity_limit_period",
			endpointID: "endpoint_2",
			input: gatewayEndpointYAML{
				Auth: authYAML{
					AuthType: "AUTH_TYPE_API_KEY",
					APIKey:   stringPtr("some_api_key"),
				},
				RateLimiting: rateLimitingYAML{
					CapacityLimit:       100_000,
					CapacityLimitPeriod: "CAPACITY_LIMIT_PERIOD_YEARLY",
				},
			},
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)

			err := test.input.validate(test.endpointID)
			if test.wantErr {
				c.Error(err)
			} else {
				c.NoError(err)
			}
		})
	}
}

func Test_authYAML_validate(t *testing.T) {
	tests := []struct {
		name    string
		input   authYAML
		wantErr bool
	}{
		{
			name: "valid AUTH_TYPE_API_KEY",
			input: authYAML{
				AuthType: "AUTH_TYPE_API_KEY",
				APIKey:   stringPtr("some_api_key"),
			},
			wantErr: false,
		},
		{
			name: "missing api_key for AUTH_TYPE_API_KEY",
			input: authYAML{
				AuthType: "AUTH_TYPE_API_KEY",
			},
			wantErr: true,
		},
		{
			name: "valid AUTH_TYPE_JWT",
			input: authYAML{
				AuthType: "AUTH_TYPE_JWT",
				JWTAuthorizedUsers: []string{
					"user_1",
				},
			},
			wantErr: false,
		},
		{
			name: "missing jwt_authorized_users for AUTH_TYPE_JWT",
			input: authYAML{
				AuthType: "AUTH_TYPE_JWT",
			},
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)

			err := test.input.validate()
			if test.wantErr {
				c.Error(err)
			} else {
				c.NoError(err)
			}
		})
	}
}

func stringPtr(s string) *string {
	return &s
}
