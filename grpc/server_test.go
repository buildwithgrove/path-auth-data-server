package grpc

import (
	"context"
	"errors"
	"testing"

	"github.com/buildwithgrove/path/envoy/auth_server/proto"
	"github.com/pokt-network/poktroll/pkg/polylog/polyzero"
	"github.com/stretchr/testify/require"
	gomock "go.uber.org/mock/gomock"
	"google.golang.org/grpc"
)

func Test_FetchAuthDataSync(t *testing.T) {
	tests := []struct {
		name             string
		expectedResponse *proto.AuthDataResponse
		expectedError    error
	}{
		{
			name: "should return Gateway Endpoints successfully",
			expectedResponse: &proto.AuthDataResponse{
				Endpoints: map[string]*proto.GatewayEndpoint{
					"endpoint1": {
						EndpointId: "endpoint1",
						Auth: &proto.Auth{
							RequireAuth: true,
							AuthorizedUsers: map[string]*proto.Empty{
								"user1": {},
							},
						},
						UserAccount: &proto.UserAccount{
							AccountId: "account1",
							PlanType:  "basic",
						},
						RateLimiting: &proto.RateLimiting{
							ThroughputLimit:     100,
							CapacityLimit:       1000,
							CapacityLimitPeriod: proto.CapacityLimitPeriod_CAPACITY_LIMIT_PERIOD_DAILY,
						},
					},
				},
			},
			expectedError: nil,
		},
		{
			name:             "should return error when data source fails",
			expectedResponse: nil,
			expectedError:    errors.New("data source error"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockDataSource := NewMockAuthDataSource(ctrl)

			mockDataSource.EXPECT().FetchAuthDataSync().Return(test.expectedResponse, test.expectedError)
			if test.expectedError == nil {
				mockDataSource.EXPECT().AuthDataUpdatesChan().Return(nil, nil)
			}

			server, err := NewGRPCServer(mockDataSource, polyzero.NewLogger())
			c.Equal(test.expectedError, err)

			if test.expectedError == nil {
				ctx := context.Background()

				resp, err := server.FetchAuthDataSync(ctx, &proto.AuthDataRequest{})
				if test.expectedError == nil {
					c.Equal(test.expectedResponse, resp)
					c.NoError(err)
				} else {
					c.Nil(resp)
					c.EqualError(err, test.expectedError.Error())
				}
			}
		})
	}
}

func Test_StreamUpdates(t *testing.T) {
	tests := []struct {
		name          string
		updates       []*proto.AuthDataUpdate
		expectedError error
	}{
		{
			name: "should stream updates successfully",
			updates: []*proto.AuthDataUpdate{
				{
					EndpointId: "endpoint1",
					GatewayEndpoint: &proto.GatewayEndpoint{
						EndpointId: "endpoint1",
					},
					Delete: false,
				},
				{
					EndpointId: "endpoint2",
					Delete:     true,
				},
			},
			expectedError: nil,
		},
		{
			name:          "should handle no updates",
			updates:       []*proto.AuthDataUpdate{},
			expectedError: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockDataSource := NewMockAuthDataSource(ctrl)
			updateCh := make(chan *proto.AuthDataUpdate, len(test.updates))

			for _, update := range test.updates {
				updateCh <- update
			}
			close(updateCh)

			mockDataSource.EXPECT().FetchAuthDataSync().Return(&proto.AuthDataResponse{
				Endpoints: map[string]*proto.GatewayEndpoint{},
			}, nil)
			mockDataSource.EXPECT().AuthDataUpdatesChan().Return(updateCh, nil)

			server, err := NewGRPCServer(mockDataSource, polyzero.NewLogger())
			c.NoError(err)

			mockStream := &mockStreamServer{
				updates:         test.updates,
				updatesReceived: make(chan *proto.AuthDataUpdate, len(test.updates)),
			}

			go func() {
				err = server.StreamAuthDataUpdates(&proto.AuthDataUpdatesRequest{}, mockStream)
				c.Equal(test.expectedError, err)
			}()

			for _, expectedUpdate := range test.updates {
				receivedUpdate := <-mockStream.updatesReceived
				c.Equal(expectedUpdate, receivedUpdate)
			}
		})
	}
}

func Test_handleDataSourceUpdates(t *testing.T) {
	tests := []struct {
		name                     string
		gatewayEndpoints         map[string]*proto.GatewayEndpoint
		updates                  []*proto.AuthDataUpdate
		expectedDataAfterUpdates map[string]*proto.GatewayEndpoint
	}{
		{
			name: "should update server state with new updates",
			gatewayEndpoints: map[string]*proto.GatewayEndpoint{
				"endpoint1": {
					EndpointId: "endpoint1",
				},
			},
			updates: []*proto.AuthDataUpdate{
				{
					EndpointId: "endpoint1",
					GatewayEndpoint: &proto.GatewayEndpoint{
						EndpointId: "endpoint1",
					},
					Delete: false,
				},
				{
					EndpointId: "endpoint2",
					GatewayEndpoint: &proto.GatewayEndpoint{
						EndpointId: "endpoint2",
					},
					Delete: false,
				},
				{
					EndpointId: "endpoint1",
					Delete:     true,
				},
			},
			expectedDataAfterUpdates: map[string]*proto.GatewayEndpoint{
				"endpoint2": {
					EndpointId: "endpoint2",
				},
			},
		},
		{
			name: "should handle no updates",
			gatewayEndpoints: map[string]*proto.GatewayEndpoint{
				"endpoint1": {
					EndpointId: "endpoint1",
				},
			},
			updates: []*proto.AuthDataUpdate{},
			expectedDataAfterUpdates: map[string]*proto.GatewayEndpoint{
				"endpoint1": {
					EndpointId: "endpoint1",
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockDataSource := NewMockAuthDataSource(ctrl)
			updateCh := make(chan *proto.AuthDataUpdate, len(test.updates))

			for _, update := range test.updates {
				updateCh <- update
			}
			close(updateCh)

			mockDataSource.EXPECT().FetchAuthDataSync().Return(&proto.AuthDataResponse{Endpoints: test.gatewayEndpoints}, nil)
			mockDataSource.EXPECT().AuthDataUpdatesChan().Return(updateCh, nil)

			server, err := NewGRPCServer(mockDataSource, polyzero.NewLogger())
			c.NoError(err)

			server.handleDataSourceUpdates(updateCh)

			c.Equal(test.expectedDataAfterUpdates, server.gatewayEndpoints)
		})
	}
}

type mockStreamServer struct {
	grpc.ServerStream
	updates         []*proto.AuthDataUpdate
	updatesReceived chan *proto.AuthDataUpdate
}

func (m *mockStreamServer) Send(update *proto.AuthDataUpdate) error {
	for _, u := range m.updates {
		if u.EndpointId == update.EndpointId {
			m.updatesReceived <- update
			return nil
		}
	}
	return errors.New("unexpected update")
}

func (m *mockStreamServer) Context() context.Context {
	return context.Background()
}
