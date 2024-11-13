package grpc

import (
	"context"
	"sync"

	"github.com/buildwithgrove/path/envoy/auth_server/proto"
	"github.com/pokt-network/poktroll/pkg/polylog"
)

// grpcServer handles fetching and streaming GatewayEndpoints from the configured AuthDataSource.
//
// It implements the gRPC server defined in PATH's Go External Authorization Server's `gateway_endpoint.proto` file.
//
// TODO_IMPROVE(@commoddity): Update this link to point to main once `envoy-grpc-auth-service` is merged.
// See: https://github.com/buildwithgrove/path/blob/envoy-grpc-auth-service/envoy/auth_server/proto/gateway_endpoint.proto
type grpcServer struct {
	proto.UnimplementedGatewayEndpointsServer

	authDataSource   AuthDataSource
	authDataUpdateCh chan *proto.AuthDataUpdate

	gatewayEndpoints   map[string]*proto.GatewayEndpoint
	gatewayEndpointsMu sync.RWMutex

	logger polylog.Logger
}

// NewGRPCServer creates a new grpcServer instance using the provided AuthDataSource.
func NewGRPCServer(authDataSource AuthDataSource, logger polylog.Logger) (*grpcServer, error) {

	server := &grpcServer{
		authDataSource:   authDataSource,
		gatewayEndpoints: make(map[string]*proto.GatewayEndpoint),
		authDataUpdateCh: make(chan *proto.AuthDataUpdate, 1_000),
		logger:           logger,
	}

	// Warm up the data store with the full set of GatewayEndpoints from the data source.
	authDataResponse, err := authDataSource.FetchAuthDataSync()
	if err != nil {
		return nil, err
	}
	server.gatewayEndpoints = authDataResponse.Endpoints

	// Start listening for updates from the data source.
	authDataUpdatesCh, err := authDataSource.AuthDataUpdatesChan()
	if err != nil {
		return nil, err
	}
	go server.handleDataSourceUpdates(authDataUpdatesCh)

	return server, nil
}

// FetchAuthDataSync handles the gRPC request to retrieve the full set of GatewayEndpoints data.
// This method is called from PADS to initialize the data store on startup.
func (s *grpcServer) FetchAuthDataSync(ctx context.Context, req *proto.AuthDataRequest) (*proto.AuthDataResponse, error) {
	s.gatewayEndpointsMu.RLock()
	defer s.gatewayEndpointsMu.RUnlock()
	return &proto.AuthDataResponse{Endpoints: s.gatewayEndpoints}, nil
}

// StreamAuthDataUpdates streams GatewayEndpoint updates to PATH's
// Go External Authorization Server whenever the data source changes.
// It uses gRPC streaming to send updates to PATH's External Authorization Server.
func (s *grpcServer) StreamAuthDataUpdates(req *proto.AuthDataUpdatesRequest, stream proto.GatewayEndpoints_StreamAuthDataUpdatesServer) error {
	for update := range s.authDataUpdateCh {

		if err := stream.Send(update); err != nil {
			s.logger.Error().Err(err).Msg("failed to stream auth data update to client")
			return err
		}

	}
	return nil
}

// handleDataSourceUpdates listens for updates from the DataSource's authDataUpdatesCh and
// updates the server's data store accordingly.
// Update may be one of: create, update, or delete.
func (s *grpcServer) handleDataSourceUpdates(authDataUpdatesCh <-chan *proto.AuthDataUpdate) {

	for authDataUpdate := range authDataUpdatesCh {
		logger := s.logger.With("endpoint_id", authDataUpdate.EndpointId)

		s.gatewayEndpointsMu.Lock()
		if authDataUpdate.Delete {

			logger.Info().Msg("deleted gateway endpoint")

			delete(s.gatewayEndpoints, authDataUpdate.EndpointId)
		} else {

			if _, ok := s.gatewayEndpoints[authDataUpdate.EndpointId]; !ok {
				logger.Info().Msg("created gateway endpoint")
			} else {
				logger.Info().Msg("updated gateway endpoint")
			}

			s.gatewayEndpoints[authDataUpdate.EndpointId] = authDataUpdate.GatewayEndpoint
		}
		s.gatewayEndpointsMu.Unlock()

		// Send the update to any clients streaming updates.
		s.authDataUpdateCh <- authDataUpdate
	}
}
