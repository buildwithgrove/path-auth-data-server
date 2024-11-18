package yaml

import (
	"testing"

	"github.com/buildwithgrove/path/envoy/auth_server/proto"
	"github.com/stretchr/testify/require"
)

func Test_gatewayEndpointsYAML_convertToProto(t *testing.T) {
	tests := []struct {
		name     string
		input    gatewayEndpointsYAML
		expected *proto.AuthDataResponse
	}{
		{
			name: "should convert YAML to proto format correctly",
			input: gatewayEndpointsYAML{
				Endpoints: map[string]gatewayEndpointYAML{
					"endpoint_1": {
						Auth: authYAML{
							AuthType: "JWT_AUTH",
							JWTAuthorizedUsers: []string{
								"auth0|user_1",
							},
						},
						RateLimiting: rateLimitingYAML{
							ThroughputLimit:     30,
							CapacityLimit:       100_000,
							CapacityLimitPeriod: yamlCapacityLimitPeriodMonthly,
						},
						Metadata: map[string]string{
							"account_id": "account_1",
							"plan_type":  "PLAN_UNLIMITED",
						},
					},
				},
			},
			expected: &proto.AuthDataResponse{
				Endpoints: map[string]*proto.GatewayEndpoint{
					"endpoint_1": {
						EndpointId: "endpoint_1",
						Auth: &proto.Auth{
							AuthType: proto.Auth_JWT_AUTH,
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
						Metadata: map[string]string{
							"account_id": "account_1",
							"plan_type":  "PLAN_UNLIMITED",
						},
					},
				},
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

func Test_gatewayEndpointsYAML_validate(t *testing.T) {
	tests := []struct {
		name    string
		input   gatewayEndpointsYAML
		wantErr bool
	}{
		{
			name: "valid endpoints",
			input: gatewayEndpointsYAML{
				Endpoints: map[string]gatewayEndpointYAML{
					"endpoint_1": {
						Auth: authYAML{
							AuthType: "JWT_AUTH",
							JWTAuthorizedUsers: []string{
								"auth0|user_1",
							},
						},
						RateLimiting: rateLimitingYAML{
							ThroughputLimit:     30,
							CapacityLimit:       100_000,
							CapacityLimitPeriod: yamlCapacityLimitPeriodMonthly,
						},
					},
					"endpoint_2": {
						Auth: authYAML{
							AuthType: "API_KEY_AUTH",
							APIKey:   stringPtr("some_api_key"),
						},
						RateLimiting: rateLimitingYAML{
							ThroughputLimit:     50,
							CapacityLimit:       200_000,
							CapacityLimitPeriod: yamlCapacityLimitPeriodDaily,
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid endpoint with missing endpoint_id",
			input: gatewayEndpointsYAML{
				Endpoints: map[string]gatewayEndpointYAML{
					"": {
						Auth: authYAML{
							AuthType: "API_KEY_AUTH",
							APIKey:   stringPtr("some_api_key"),
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid endpoint with incorrect capacity_limit_period",
			input: gatewayEndpointsYAML{
				Endpoints: map[string]gatewayEndpointYAML{
					"endpoint_1": {
						Auth: authYAML{
							AuthType: "JWT_AUTH",
							JWTAuthorizedUsers: []string{
								"auth0|user_1",
							},
						},
						RateLimiting: rateLimitingYAML{
							CapacityLimit:       100_000,
							CapacityLimitPeriod: "CAPACITY_LIMIT_PERIOD_YEARLY",
						},
					},
				},
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
