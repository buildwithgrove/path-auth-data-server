package server

import (
	"context"
	"sync"

	proto "github.com/buildwithgrove/path/envoy/auth_server/proto"
)

// Server implements the gRPC server for GatewayEndpoints.
// It uses a DataSource to retrieve initial data and updates.
type Server struct {
	proto.UnimplementedGatewayEndpointsServer
	GatewayEndpoints map[string]*proto.GatewayEndpoint
	updateCh         chan *proto.Update
	mu               sync.RWMutex
	dataSource       DataSource
}

// NewServer creates a new Server instance using the provided DataSource.
func NewServer(dataSource DataSource) (*Server, error) {
	server := &Server{
		GatewayEndpoints: make(map[string]*proto.GatewayEndpoint),
		updateCh:         make(chan *proto.Update, 100),
		dataSource:       dataSource,
	}

	initialData, err := dataSource.GetInitialData()
	if err != nil {
		return nil, err
	}

	server.mu.Lock()
	server.GatewayEndpoints = initialData.Endpoints
	server.mu.Unlock()

	updatesCh, err := dataSource.SubscribeUpdates()
	if err != nil {
		return nil, err
	}

	go server.handleDataSourceUpdates(updatesCh)

	return server, nil
}

// GetInitialData handles the gRPC request to retrieve initial GatewayEndpoints data.
func (s *Server) GetInitialData(ctx context.Context, req *proto.InitialDataRequest) (*proto.InitialDataResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return &proto.InitialDataResponse{Endpoints: s.GatewayEndpoints}, nil
}

// StreamUpdates streams updates to the client whenever the data source changes.
func (s *Server) StreamUpdates(req *proto.UpdatesRequest, stream proto.GatewayEndpoints_StreamUpdatesServer) error {
	for update := range s.updateCh {
		if err := stream.Send(update); err != nil {
			return err
		}
	}
	return nil
}

// handleDataSourceUpdates listens for updates from the DataSource and updates the server state accordingly.
func (s *Server) handleDataSourceUpdates(updatesCh <-chan *proto.Update) {
	for update := range updatesCh {
		s.mu.Lock()
		if update.Delete {
			delete(s.GatewayEndpoints, update.EndpointId)
		} else {
			s.GatewayEndpoints[update.EndpointId] = update.GatewayEndpoint
		}
		s.mu.Unlock()

		// Send the update to any clients streaming updates.
		s.updateCh <- update
	}
}
