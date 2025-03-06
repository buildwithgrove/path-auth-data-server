package yaml

import (
	"testing"

	"github.com/buildwithgrove/path-external-auth-server/proto"
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
			endpointID: "endpoint_1_static_key",
			input: gatewayEndpointYAML{
				Auth: authYAML{
					APIKey: stringPtr("some_api_key"),
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
				EndpointId: "endpoint_1_static_key",
				Auth: &proto.Auth{
					AuthType: &proto.Auth_StaticApiKey{
						StaticApiKey: &proto.StaticAPIKey{
							ApiKey: "some_api_key",
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
				APIKey: stringPtr("some_api_key"),
			},
			expected: &proto.Auth{
				AuthType: &proto.Auth_StaticApiKey{
					StaticApiKey: &proto.StaticAPIKey{
						ApiKey: "some_api_key",
					},
				},
			},
		},
		{
			name:  "should handle empty authorized users",
			input: authYAML{},
			expected: &proto.Auth{
				AuthType: &proto.Auth_NoAuth{},
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
			name:       "valid endpoint with API key auth",
			endpointID: "endpoint_1_static_key",
			input: gatewayEndpointYAML{
				Auth: authYAML{
					APIKey: stringPtr("some_api_key"),
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
					APIKey: stringPtr("some_api_key"),
				},
			},
			wantErr: true,
		},
		{
			name:       "invalid capacity_limit_period",
			endpointID: "endpoint_1_static_key",
			input: gatewayEndpointYAML{
				Auth: authYAML{
					APIKey: stringPtr("some_api_key"),
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
			name: "valid API key auth",
			input: authYAML{
				APIKey: stringPtr("some_api_key"),
			},
			wantErr: false,
		},
		{
			name: "missing api_key for API key auth",
			input: authYAML{
				APIKey: stringPtr(""),
			},
			wantErr: true,
		},
		{
			name: "valid API key auth",
			input: authYAML{
				APIKey: stringPtr("some_api_key"),
			},
			wantErr: false,
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
