package yaml

import (
	"testing"

	"github.com/buildwithgrove/path-external-auth-server/proto"
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
					"endpoint_1_static_key": {
						Auth: authYAML{
							APIKey: stringPtr("some_api_key"),
						},
						Metadata: metadataYAML{
							Name:      "grove_city_test_endpoint",
							AccountId: "account_1",
							PlanType:  "PLAN_UNLIMITED",
						},
					},
				},
			},
			expected: &proto.AuthDataResponse{
				Endpoints: map[string]*proto.GatewayEndpoint{
					"endpoint_1_static_key": {
						EndpointId: "endpoint_1_static_key",
						Auth: &proto.Auth{
							AuthType: &proto.Auth_StaticApiKey{
								StaticApiKey: &proto.StaticAPIKey{
									ApiKey: "some_api_key",
								},
							},
						},
						Metadata: &proto.Metadata{
							Name:      "grove_city_test_endpoint",
							AccountId: "account_1",
							PlanType:  "PLAN_UNLIMITED",
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
					"endpoint_1_no_auth": {
						Auth: authYAML{},
					},
					"endpoint_1_static_key": {
						Auth: authYAML{
							APIKey: stringPtr("some_api_key"),
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
							APIKey: stringPtr("some_api_key"),
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
