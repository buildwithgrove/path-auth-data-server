package yaml

import (
	"testing"

	"github.com/buildwithgrove/path/envoy/auth_server/proto"
	"github.com/stretchr/testify/require"
)

func Test_gatewayEndpointYAML_convertToProto(t *testing.T) {
	tests := []struct {
		name     string
		input    gatewayEndpointYAML
		expected *proto.GatewayEndpoint
	}{
		{
			name: "should convert gatewayEndpointYAML to proto format correctly",
			input: gatewayEndpointYAML{
				EndpointID: "endpoint_1",
				Auth: authYAML{
					AuthType: "JWT_AUTH",
					JWTAuthorizedUsers: &map[string]struct{}{
						"auth0|user_1": {},
					},
				},
				UserAccount: userAccountYAML{
					AccountID: "account_1",
					PlanType:  "PLAN_UNLIMITED",
				},
				RateLimiting: rateLimitingYAML{
					ThroughputLimit:     30,
					CapacityLimit:       100000,
					CapacityLimitPeriod: "CAPACITY_LIMIT_PERIOD_MONTHLY",
				},
			},
			expected: &proto.GatewayEndpoint{
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
				UserAccount: &proto.UserAccount{
					AccountId: "account_1",
					PlanType:  "PLAN_UNLIMITED",
				},
				RateLimiting: &proto.RateLimiting{
					ThroughputLimit:     30,
					CapacityLimit:       100000,
					CapacityLimitPeriod: proto.CapacityLimitPeriod_CAPACITY_LIMIT_PERIOD_MONTHLY,
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

func Test_authYAML_convertToProto(t *testing.T) {
	tests := []struct {
		name     string
		input    authYAML
		expected *proto.Auth
	}{
		{
			name: "should convert authYAML to proto format correctly",
			input: authYAML{
				AuthType: "JWT_AUTH",
				JWTAuthorizedUsers: &map[string]struct{}{
					"auth0|user_1": {},
					"auth0|user_2": {},
				},
			},
			expected: &proto.Auth{
				AuthType: proto.Auth_JWT_AUTH,
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
				AuthType: proto.Auth_NO_AUTH,
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
