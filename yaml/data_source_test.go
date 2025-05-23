package yaml

import (
	"os"
	"sort"
	"testing"
	"time"

	"github.com/buildwithgrove/path-external-auth-server/proto"
	"github.com/pokt-network/poktroll/pkg/polylog/polyzero"
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
					"endpoint_1_static_key": {
						EndpointId: "endpoint_1_static_key",
						Auth: &proto.Auth{
							AuthType: &proto.Auth_StaticApiKey{
								StaticApiKey: &proto.StaticAPIKey{
									ApiKey: "api_key_1",
								},
							},
						},
						Metadata: &proto.Metadata{
							AccountId: "account_1",
							PlanType:  "PLAN_UNLIMITED",
							Email:     "amos.burton@opa.belt",
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
							Email:     "frodo.baggins@shire.io",
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
  "":
    auth:
      api_key: "api_key_1"
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

			yamlDataSource, err := NewYAMLDataSource(test.filePath, polyzero.NewLogger())
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
  endpoint_1_static_key:
    endpoint_id: "endpoint_1_static_key"
    auth:
      api_key: "api_key_1"
    metadata:
      account_id: "account_1"
      plan_type: "PLAN_UNLIMITED"
  endpoint_2_no_auth:
    endpoint_id: "endpoint_2_no_auth"
    metadata:
      account_id: "account_2"
      plan_type: "PLAN_UNLIMITED"
`,
			updatedData: `
endpoints:
  endpoint_1_static_key:
    endpoint_id: "endpoint_1_static_key"
    metadata:
      account_id: "account_1"
      plan_type: "PLAN_UNLIMITED"
  endpoint_2_no_auth:
    endpoint_id: "endpoint_2_no_auth"
    auth:
      api_key: "api_key_2"
    metadata:
      account_id: "account_2"
      plan_type: "PLAN_FREE"
`,
			expectedUpdates: []*proto.AuthDataUpdate{
				{
					EndpointId: "endpoint_1_static_key",
					GatewayEndpoint: &proto.GatewayEndpoint{
						EndpointId: "endpoint_1_static_key",
						Auth: &proto.Auth{
							AuthType: &proto.Auth_NoAuth{},
						},
						Metadata: &proto.Metadata{
							AccountId: "account_1",
							PlanType:  "PLAN_UNLIMITED",
						},
					},
				},
				{
					EndpointId: "endpoint_2_no_auth",
					GatewayEndpoint: &proto.GatewayEndpoint{
						EndpointId: "endpoint_2_no_auth",
						Auth: &proto.Auth{
							AuthType: &proto.Auth_StaticApiKey{
								StaticApiKey: &proto.StaticAPIKey{
									ApiKey: "api_key_2",
								},
							},
						},
						Metadata: &proto.Metadata{
							AccountId: "account_2",
							PlanType:  "PLAN_FREE",
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

			yamlDataSource, err := NewYAMLDataSource(filePath, polyzero.NewLogger())
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
			},
			newEndpoints: map[string]*proto.GatewayEndpoint{
				"endpoint_1_static_key": {
					EndpointId: "endpoint_1_static_key",
					Auth: &proto.Auth{
						AuthType: &proto.Auth_NoAuth{},
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
			expectedUpdates: []*proto.AuthDataUpdate{
				{
					EndpointId: "endpoint_1_static_key",
					GatewayEndpoint: &proto.GatewayEndpoint{
						EndpointId: "endpoint_1_static_key",
						Auth: &proto.Auth{
							AuthType: &proto.Auth_NoAuth{},
						},
						Metadata: &proto.Metadata{
							AccountId: "account_1",
							PlanType:  "PLAN_UNLIMITED",
						},
					},
				},
				{
					EndpointId: "endpoint_2_no_auth",
					GatewayEndpoint: &proto.GatewayEndpoint{
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
		},
		{
			name: "should send delete updates for removed endpoints",
			gatewayEndpoints: map[string]*proto.GatewayEndpoint{
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
			},
			newEndpoints: map[string]*proto.GatewayEndpoint{},
			expectedUpdates: []*proto.AuthDataUpdate{
				{
					EndpointId: "endpoint_1_static_key",
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

			// Sort the expected updates and received updates by EndpointId
			sort.Slice(test.expectedUpdates, func(i, j int) bool {
				return test.expectedUpdates[i].EndpointId < test.expectedUpdates[j].EndpointId
			})

			receivedUpdates := make([]*proto.AuthDataUpdate, 0, len(test.expectedUpdates))
			for range test.expectedUpdates {
				select {
				case update := <-yamlDataSource.authDataUpdatesCh:
					receivedUpdates = append(receivedUpdates, update)
				default:
					t.Fatal("expected update not received")
				}
			}

			sort.Slice(receivedUpdates, func(i, j int) bool {
				return receivedUpdates[i].EndpointId < receivedUpdates[j].EndpointId
			})

			// Compare the sorted updates
			c.EqualValues(test.expectedUpdates, receivedUpdates)
		})
	}
}
