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
		expected *proto.InitialDataResponse
	}{
		{
			name: "should convert YAML to proto format correctly",
			input: gatewayEndpointsYAML{
				Endpoints: map[string]gatewayEndpointYAML{
					"endpoint_1": {
						EndpointID: "endpoint_1",
						Auth: authYAML{
							RequireAuth: true,
							AuthorizedUsers: map[string]struct{}{
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
				},
			},
			expected: &proto.InitialDataResponse{
				Endpoints: map[string]*proto.GatewayEndpoint{
					"endpoint_1": {
						EndpointId: "endpoint_1",
						Auth: &proto.Auth{
							RequireAuth: true,
							AuthorizedUsers: map[string]*proto.Empty{
								"auth0|user_1": {},
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
					RequireAuth: true,
					AuthorizedUsers: map[string]struct{}{
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
					RequireAuth: true,
					AuthorizedUsers: map[string]*proto.Empty{
						"auth0|user_1": {},
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
				RequireAuth: true,
				AuthorizedUsers: map[string]struct{}{
					"auth0|user_1": {},
					"auth0|user_2": {},
				},
			},
			expected: &proto.Auth{
				RequireAuth: true,
				AuthorizedUsers: map[string]*proto.Empty{
					"auth0|user_1": {},
					"auth0|user_2": {},
				},
			},
		},
		{
			name: "should handle empty authorized users",
			input: authYAML{
				RequireAuth:     false,
				AuthorizedUsers: map[string]struct{}{},
			},
			expected: &proto.Auth{
				RequireAuth:     false,
				AuthorizedUsers: map[string]*proto.Empty{},
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
