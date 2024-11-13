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
		name     string
		filePath string
		want     *proto.AuthDataResponse
		wantErr  bool
	}{
		{
			name:     "should load valid gateway endpoints without error",
			filePath: "./testdata/gateway-endpoints.example.yaml",
			want: &proto.AuthDataResponse{
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
						RateLimiting: &proto.RateLimiting{},
					},
					"endpoint_2": {
						EndpointId: "endpoint_2",
						Auth: &proto.Auth{
							RequireAuth:     false,
							AuthorizedUsers: map[string]*proto.Empty{},
						},
						UserAccount: &proto.UserAccount{
							AccountId: "account_2",
							PlanType:  "PLAN_FREE",
						},
						RateLimiting: &proto.RateLimiting{
							ThroughputLimit:     30,
							CapacityLimit:       100000,
							CapacityLimitPeriod: proto.CapacityLimitPeriod_CAPACITY_LIMIT_PERIOD_MONTHLY,
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
			name:     "should return error for invalid YAML",
			filePath: "./testdata/invalid.yaml",
			wantErr:  true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)

			if test.name == "should return error for invalid YAML" {
				err := os.WriteFile(test.filePath, []byte("invalid_yaml: ["), 0644)
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
				c.Equal(test.want, got)
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
      require_auth: true
      authorized_users:
        "auth0|user_1": {}
    user_account:
      account_id: "account_1"
      plan_type: "PLAN_UNLIMITED"
`,
			updatedData: `
endpoints:
  endpoint_1:
    endpoint_id: "endpoint_1"
    auth:
      require_auth: false
    user_account:
      account_id: "account_1"
      plan_type: "PLAN_UNLIMITED"
  endpoint_2:
    endpoint_id: "endpoint_2"
    auth:
      require_auth: false
    user_account:
      account_id: "account_2"
      plan_type: "PLAN_FREE"
`,
			expectedUpdates: []*proto.AuthDataUpdate{
				{
					EndpointId: "endpoint_1",
					GatewayEndpoint: &proto.GatewayEndpoint{
						EndpointId: "endpoint_1",
						Auth: &proto.Auth{
							RequireAuth:     false,
							AuthorizedUsers: map[string]*proto.Empty{},
						},
						UserAccount: &proto.UserAccount{
							AccountId: "account_1",
							PlanType:  "PLAN_UNLIMITED",
						},
						RateLimiting: &proto.RateLimiting{},
					},
				},
				{
					EndpointId: "endpoint_2",
					GatewayEndpoint: &proto.GatewayEndpoint{
						EndpointId: "endpoint_2",
						Auth: &proto.Auth{
							RequireAuth:     false,
							AuthorizedUsers: map[string]*proto.Empty{},
						},
						UserAccount: &proto.UserAccount{
							AccountId: "account_2",
							PlanType:  "PLAN_FREE",
						},
						RateLimiting: &proto.RateLimiting{},
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
			<-time.After(100 * time.Millisecond)

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

			c.ElementsMatch(test.expectedUpdates, receivedUpdates)
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
						RequireAuth: true,
						AuthorizedUsers: map[string]*proto.Empty{
							"auth0|user_1": {},
						},
					},
					UserAccount: &proto.UserAccount{
						AccountId: "account_1",
						PlanType:  "PLAN_UNLIMITED",
					},
				},
			},
			newEndpoints: map[string]*proto.GatewayEndpoint{
				"endpoint_1": {
					EndpointId: "endpoint_1",
					Auth: &proto.Auth{
						RequireAuth: false,
					},
					UserAccount: &proto.UserAccount{
						AccountId: "account_1",
						PlanType:  "PLAN_UNLIMITED",
					},
				},
				"endpoint_2": {
					EndpointId: "endpoint_2",
					Auth: &proto.Auth{
						RequireAuth: false,
					},
					UserAccount: &proto.UserAccount{
						AccountId: "account_2",
						PlanType:  "PLAN_FREE",
					},
				},
			},
			expectedUpdates: []*proto.AuthDataUpdate{
				{
					EndpointId: "endpoint_1",
					GatewayEndpoint: &proto.GatewayEndpoint{
						EndpointId: "endpoint_1",
						Auth: &proto.Auth{
							RequireAuth: false,
						},
						UserAccount: &proto.UserAccount{
							AccountId: "account_1",
							PlanType:  "PLAN_UNLIMITED",
						},
					},
				},
				{
					EndpointId: "endpoint_2",
					GatewayEndpoint: &proto.GatewayEndpoint{
						EndpointId: "endpoint_2",
						Auth: &proto.Auth{
							RequireAuth: false,
						},
						UserAccount: &proto.UserAccount{
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
				"endpoint_1": {
					EndpointId: "endpoint_1",
					Auth: &proto.Auth{
						RequireAuth: true,
					},
					UserAccount: &proto.UserAccount{
						AccountId: "account_1",
						PlanType:  "PLAN_UNLIMITED",
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
					c.Equal(expectedUpdate, update)
				default:
					t.Fatal("expected update not received")
				}
			}
		})
	}
}
