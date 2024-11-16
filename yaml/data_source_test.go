package yaml

import (
	"os"
	"testing"
	"time"

	"github.com/buildwithgrove/path/envoy/auth_server/proto"
	"github.com/stretchr/testify/require"
)

func Test_LoadGatewayEndpointsFromYAML(t *testing.T) {
	tests := []struct {
		name         string
		filePath     string
		fileContents string
		want         *proto.AuthDataResponse
		wantErr      bool
	}{
		{
			name:     "should load valid gateway endpoints without error",
			filePath: "./testdata/gateway-endpoints.example.yaml",
			want: &proto.AuthDataResponse{
				Endpoints: map[string]*proto.GatewayEndpoint{
					"endpoint_1": {
						EndpointId: "endpoint_1",
						Auth: &proto.Auth{
							AuthType: proto.Auth_API_KEY_AUTH,
							AuthTypeDetails: &proto.Auth_ApiKey{
								ApiKey: &proto.APIKey{
									ApiKey: "api_key_1",
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
							AuthType: proto.Auth_JWT_AUTH,
							AuthTypeDetails: &proto.Auth_Jwt{
								Jwt: &proto.JWT{
									AuthorizedUsers: map[string]*proto.Empty{
										"auth0|user_2": {},
									},
								},
							},
						},
						RateLimiting: &proto.RateLimiting{},
						Metadata: map[string]string{
							"account_id": "account_2",
							"plan_type":  "PLAN_UNLIMITED",
						},
					},
					"endpoint_3": {
						EndpointId: "endpoint_3",
						Auth: &proto.Auth{
							AuthType:        proto.Auth_NO_AUTH,
							AuthTypeDetails: &proto.Auth_NoAuth{},
						},
						RateLimiting: &proto.RateLimiting{
							ThroughputLimit:     50,
							CapacityLimit:       200,
							CapacityLimitPeriod: proto.CapacityLimitPeriod_CAPACITY_LIMIT_PERIOD_MONTHLY,
						},
						Metadata: map[string]string{
							"account_id": "account_3",
							"plan_type":  "PLAN_UNLIMITED",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name:     "should return error for non-existent file",
			filePath: "./testdata/non_existent.yaml",
			wantErr:  true,
		},
		{
			name:         "should return error for invalid YAML",
			filePath:     "./testdata/invalid.yaml",
			fileContents: "invalid_yaml: [",
			wantErr:      true,
		},
		{
			name:     "should return error for missing endpoint_id",
			filePath: "./testdata/missing_endpoint_id.yaml",
			fileContents: `
endpoints:
  endpoint_1:
    auth:
      auth_type: "API_KEY_AUTH"
      api_key: "api_key_1"
    metadata:
      account_id: "account_1"
      plan_type: "PLAN_UNLIMITED"
`,
			wantErr: true,
		},
		{
			name:     "should return error for invalid capacity_limit_period",
			filePath: "./testdata/invalid_capacity_limit_period.yaml",
			fileContents: `
endpoints:
  endpoint_1:
    endpoint_id: "endpoint_1"
    auth:
      auth_type: "JWT_AUTH"
      jwt_authorized_users:
        "auth0|user_1": {}
    rate_limiting:
      capacity_limit: 100
      capacity_limit_period: "yearly"
    metadata:
      account_id: "account_1"
      plan_type: "PLAN_UNLIMITED"
`,
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)

			if test.fileContents != "" {
				err := os.WriteFile(test.filePath, []byte(test.fileContents), 0644)
				defer os.Remove(test.filePath)
				c.NoError(err)
			}

			yamlDataSource, err := NewYAMLDataSource(test.filePath)
			if test.wantErr {
				c.Error(err)
			} else {
				c.NoError(err)
				got, err := yamlDataSource.FetchAuthDataSync()
				c.NoError(err)
				c.EqualValues(test.want, got)
			}
		})
	}
}

func Test_watchFile(t *testing.T) {
	tests := []struct {
		name             string
		gatewayEndpoints string
		updatedData      string
		expectedUpdates  []*proto.AuthDataUpdate
	}{
		{
			name: "should detect and send updates on file change",
			gatewayEndpoints: `
endpoints:
  endpoint_1:
    endpoint_id: "endpoint_1"
    auth:
      auth_type: "API_KEY_AUTH"
      api_key: "api_key_1"
    metadata:
      account_id: "account_1"
      plan_type: "PLAN_UNLIMITED"
  endpoint_2:
    endpoint_id: "endpoint_2"
    auth:
      auth_type: "JWT_AUTH"
      jwt_authorized_users:
        "auth0|user_2": {}
    metadata:
      account_id: "account_2"
      plan_type: "PLAN_UNLIMITED"
`,
			updatedData: `
endpoints:
  endpoint_1:
    endpoint_id: "endpoint_1"
    auth:
      auth_type: "NO_AUTH"
    metadata:
      account_id: "account_1"
      plan_type: "PLAN_UNLIMITED"
  endpoint_2:
    endpoint_id: "endpoint_2"
    auth:
      auth_type: "NO_AUTH"
    metadata:
      account_id: "account_2"
      plan_type: "PLAN_FREE"
  endpoint_3:
    endpoint_id: "endpoint_3"
    rate_limiting:
      throughput_limit: 50
      capacity_limit: 200
      capacity_limit_period: "CAPACITY_LIMIT_PERIOD_MONTHLY"
    metadata:
      account_id: "account_3"
      plan_type: "PLAN_UNLIMITED"
`,
			expectedUpdates: []*proto.AuthDataUpdate{
				{
					EndpointId: "endpoint_1",
					GatewayEndpoint: &proto.GatewayEndpoint{
						EndpointId: "endpoint_1",
						Auth: &proto.Auth{
							AuthType:        proto.Auth_NO_AUTH,
							AuthTypeDetails: &proto.Auth_NoAuth{},
						},
						RateLimiting: &proto.RateLimiting{},
						Metadata: map[string]string{
							"account_id": "account_1",
							"plan_type":  "PLAN_UNLIMITED",
						},
					},
				},
				{
					EndpointId: "endpoint_2",
					GatewayEndpoint: &proto.GatewayEndpoint{
						EndpointId: "endpoint_2",
						Auth: &proto.Auth{
							AuthType:        proto.Auth_NO_AUTH,
							AuthTypeDetails: &proto.Auth_NoAuth{},
						},
						RateLimiting: &proto.RateLimiting{},
						Metadata: map[string]string{
							"account_id": "account_2",
							"plan_type":  "PLAN_FREE",
						},
					},
				},
				{
					EndpointId: "endpoint_3",
					GatewayEndpoint: &proto.GatewayEndpoint{
						EndpointId: "endpoint_3",
						Auth: &proto.Auth{
							AuthType:        proto.Auth_NO_AUTH,
							AuthTypeDetails: &proto.Auth_NoAuth{},
						},
						RateLimiting: &proto.RateLimiting{
							ThroughputLimit:     50,
							CapacityLimit:       200,
							CapacityLimitPeriod: proto.CapacityLimitPeriod_CAPACITY_LIMIT_PERIOD_MONTHLY,
						},
						Metadata: map[string]string{
							"account_id": "account_3",
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

			filePath := "./testdata/temp_gateway_endpoints.yaml"
			err := os.WriteFile(filePath, []byte(test.gatewayEndpoints), 0644)
			c.NoError(err)
			defer os.Remove(filePath)

			yamlDataSource, err := NewYAMLDataSource(filePath)
			c.NoError(err)

			go yamlDataSource.watchFile()

			// small delay to ensure the file system processes the write
			<-time.After(500 * time.Millisecond)

			err = os.WriteFile(filePath, []byte(test.updatedData), 0644)
			c.NoError(err)

			var receivedUpdates []*proto.AuthDataUpdate
			timeout := time.After(2 * time.Second)

			for range test.expectedUpdates {
				select {
				case update := <-yamlDataSource.authDataUpdatesCh:
					receivedUpdates = append(receivedUpdates, update)
				case <-timeout:
					t.Fatal("expected update not received")
				}
			}

			expectedUpdatesMap := make(map[string]*proto.AuthDataUpdate)
			for _, expectedUpdate := range test.expectedUpdates {
				expectedUpdatesMap[expectedUpdate.EndpointId] = expectedUpdate
			}
			receivedUpdatesMap := make(map[string]*proto.AuthDataUpdate)
			for _, receivedUpdate := range receivedUpdates {
				receivedUpdatesMap[receivedUpdate.EndpointId] = receivedUpdate
			}

			c.EqualValues(expectedUpdatesMap, receivedUpdatesMap)
		})
	}
}

func Test_handleUpdates(t *testing.T) {
	tests := []struct {
		name             string
		gatewayEndpoints map[string]*proto.GatewayEndpoint
		newEndpoints     map[string]*proto.GatewayEndpoint
		expectedUpdates  []*proto.AuthDataUpdate
	}{
		{
			name: "should send updates for new and modified endpoints",
			gatewayEndpoints: map[string]*proto.GatewayEndpoint{
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
					Metadata: map[string]string{
						"account_id": "account_1",
						"plan_type":  "PLAN_UNLIMITED",
					},
				},
			},
			newEndpoints: map[string]*proto.GatewayEndpoint{
				"endpoint_1": {
					EndpointId: "endpoint_1",
					Auth: &proto.Auth{
						AuthType:        proto.Auth_NO_AUTH,
						AuthTypeDetails: &proto.Auth_NoAuth{},
					},
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
					Metadata: map[string]string{
						"account_id": "account_2",
						"plan_type":  "PLAN_FREE",
					},
				},
			},
			expectedUpdates: []*proto.AuthDataUpdate{
				{
					EndpointId: "endpoint_1",
					GatewayEndpoint: &proto.GatewayEndpoint{
						EndpointId: "endpoint_1",
						Auth: &proto.Auth{
							AuthType:        proto.Auth_NO_AUTH,
							AuthTypeDetails: &proto.Auth_NoAuth{},
						},
						Metadata: map[string]string{
							"account_id": "account_1",
							"plan_type":  "PLAN_UNLIMITED",
						},
					},
				},
				{
					EndpointId: "endpoint_2",
					GatewayEndpoint: &proto.GatewayEndpoint{
						EndpointId: "endpoint_2",
						Auth: &proto.Auth{
							AuthType:        proto.Auth_NO_AUTH,
							AuthTypeDetails: &proto.Auth_NoAuth{},
						},
						Metadata: map[string]string{
							"account_id": "account_2",
							"plan_type":  "PLAN_FREE",
						},
					},
				},
			},
		},
		{
			name: "should send delete updates for removed endpoints",
			gatewayEndpoints: map[string]*proto.GatewayEndpoint{
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
					Metadata: map[string]string{
						"account_id": "account_1",
						"plan_type":  "PLAN_UNLIMITED",
					},
				},
			},
			newEndpoints: map[string]*proto.GatewayEndpoint{},
			expectedUpdates: []*proto.AuthDataUpdate{
				{
					EndpointId: "endpoint_1",
					Delete:     true,
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)

			yamlDataSource := &yamlDataSource{
				gatewayEndpoints:  test.gatewayEndpoints,
				authDataUpdatesCh: make(chan *proto.AuthDataUpdate, len(test.expectedUpdates)),
			}

			yamlDataSource.handleUpdates(test.newEndpoints)

			for _, expectedUpdate := range test.expectedUpdates {
				select {
				case update := <-yamlDataSource.authDataUpdatesCh:
					c.EqualValues(expectedUpdate, update)
				default:
					t.Fatal("expected update not received")
				}
			}
		})
	}
}
