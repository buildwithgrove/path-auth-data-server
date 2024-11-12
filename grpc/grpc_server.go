package grpc

import (
	"context"
	"sync"

	"github.com/buildwithgrove/path/envoy/auth_server/proto"
)

// Server implements the gRPC server for GatewayEndpoints.
// It uses a DataSource to retrieve initial data and updates.
type GRPCServer struct {
	proto.UnimplementedGatewayEndpointsServer
	dataSource       DataSource
	gatewayEndpoints map[string]*proto.GatewayEndpoint
	updateCh         chan *proto.Update
	mu               sync.RWMutex
}

// NewGRPCServer creates a new GRPCServer instance using the provided DataSource.
func NewGRPCServer(dataSource DataSource) (*GRPCServer, error) {
	server := &GRPCServer{
		dataSource:       dataSource,
		gatewayEndpoints: make(map[string]*proto.GatewayEndpoint),
		updateCh:         make(chan *proto.Update, 100),
	}

	initialData, err := dataSource.FetchInitialData()
	if err != nil {
		return nil, err
	}

	server.gatewayEndpoints = initialData.Endpoints

	updatesCh, err := dataSource.GetUpdatesChan()
	if err != nil {
		return nil, err
	}

	go server.handleDataSourceUpdates(updatesCh)

	return server, nil
}

// GetInitialData handles the gRPC request to retrieve initial GatewayEndpoints data.
func (s *GRPCServer) GetInitialData(ctx context.Context, req *proto.InitialDataRequest) (*proto.InitialDataResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return &proto.InitialDataResponse{Endpoints: s.gatewayEndpoints}, nil
}

// StreamUpdates streams updates to the client whenever the data source changes.
func (s *GRPCServer) StreamUpdates(req *proto.UpdatesRequest, stream proto.GatewayEndpoints_StreamUpdatesServer) error {
	for update := range s.updateCh {
		if err := stream.Send(update); err != nil {
			return err
		}
	}
	return nil
}

// handleDataSourceUpdates listens for updates from the DataSource and updates the server state accordingly.
func (s *GRPCServer) handleDataSourceUpdates(updatesCh <-chan *proto.Update) {
	for update := range updatesCh {
		s.mu.Lock()
		if update.Delete {
			delete(s.gatewayEndpoints, update.EndpointId)
		} else {
			s.gatewayEndpoints[update.EndpointId] = update.GatewayEndpoint
		}
		s.mu.Unlock()

		// Send the update to any clients streaming updates.
		s.updateCh <- update
	}
}
