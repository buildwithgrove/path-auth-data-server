package grpc

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/buildwithgrove/path-external-auth-server/proto"
	"github.com/pokt-network/poktroll/pkg/polylog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Client ID counter for generating unique client IDs
var clientIDCounter uint64

// generateUniqueClientID generates a unique ID for a client connection
func generateUniqueClientID() string {
	id := atomic.AddUint64(&clientIDCounter, 1)
	return fmt.Sprintf("%d-%d", id, time.Now().UnixNano())
}

// grpcServer handles fetching and streaming GatewayEndpoints from the configured AuthDataSource.
//
// It implements the gRPC server defined in PATH's Go External Authorization Server's `gateway_endpoint.proto` file.
//
// This implementation is optimized for a single client connection (PEAS) and handles reconnection gracefully:
// 1. It tracks a single active client stream
// 2. When updates occur with no active client, they're stored in a pending updates queue
// 3. When a client reconnects, all pending updates are sent immediately
// 4. If the client disconnects during sending, the server marks the stream as inactive and stores further updates
//
// TODO_IMPROVE(@commoddity): Update this link to point to main once `envoy-grpc-auth-service` is merged.
// See: https://github.com/buildwithgrove/path/blob/envoy-grpc-auth-service/envoy/auth_server/proto/gateway_endpoint.proto
type grpcServer struct {
	proto.UnimplementedGatewayEndpointsServer

	authDataSource   AuthDataSource
	authDataUpdateCh chan *proto.AuthDataUpdate

	gatewayEndpoints   map[string]*proto.GatewayEndpoint
	gatewayEndpointsMu sync.RWMutex

	// For a single client connection
	currentClientStream proto.GatewayEndpoints_StreamAuthDataUpdatesServer
	currentClientID     string
	currentStreamCtx    context.Context
	currentStreamMu     sync.Mutex
	pendingUpdates      []*proto.AuthDataUpdate
	pendingUpdatesMu    sync.Mutex
	isStreamActive      bool

	logger polylog.Logger
}

// NewGRPCServer creates a new grpcServer instance using the provided AuthDataSource.
func NewGRPCServer(authDataSource AuthDataSource, logger polylog.Logger) (*grpcServer, error) {

	server := &grpcServer{
		authDataSource:   authDataSource,
		authDataUpdateCh: make(chan *proto.AuthDataUpdate, 1_000),

		gatewayEndpoints: make(map[string]*proto.GatewayEndpoint),
		pendingUpdates:   make([]*proto.AuthDataUpdate, 0, 100),

		logger: logger,
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
// This method is called from PADS to warm up the data store on startup.
func (s *grpcServer) FetchAuthDataSync(ctx context.Context, req *proto.AuthDataRequest) (*proto.AuthDataResponse, error) {
	s.gatewayEndpointsMu.RLock()
	defer s.gatewayEndpointsMu.RUnlock()

	s.logger.Info().Int("num_gateway_endpoints", len(s.gatewayEndpoints)).Msg("fetching auth data sync")

	return &proto.AuthDataResponse{Endpoints: s.gatewayEndpoints}, nil
}

// StreamAuthDataUpdates streams GatewayEndpoint updates to PATH's
// Go External Authorization Server whenever the data source changes.
// It uses gRPC streaming to send updates to PATH's External Authorization Server.
func (s *grpcServer) StreamAuthDataUpdates(req *proto.AuthDataUpdatesRequest, stream proto.GatewayEndpoints_StreamAuthDataUpdatesServer) error {
	// Since we only have one client, we need to handle when a new client connects
	// while the previous one is still "active" (from our perspective)
	clientID := generateUniqueClientID()

	// Set the current stream and client ID
	s.currentStreamMu.Lock()
	s.currentClientStream = stream
	s.currentClientID = clientID
	s.currentStreamCtx = stream.Context()
	s.isStreamActive = true

	// Process any pending updates that accumulated while no client was connected
	pendingToProcess := make([]*proto.AuthDataUpdate, 0)
	s.pendingUpdatesMu.Lock()
	if len(s.pendingUpdates) > 0 {
		pendingToProcess = append(pendingToProcess, s.pendingUpdates...)
		s.pendingUpdates = s.pendingUpdates[:0] // Clear the slice but keep capacity
	}
	s.pendingUpdatesMu.Unlock()
	s.currentStreamMu.Unlock()

	s.logger.Info().
		Int("pending_updates", len(pendingToProcess)).
		Msg("client connected to stream auth data updates")

	// Send any pending updates first
	for _, update := range pendingToProcess {
		if err := stream.Send(update); err != nil {
			s.logger.Error().
				Err(err).
				Msg("failed to send pending update to client")

			s.currentStreamMu.Lock()
			s.isStreamActive = false
			s.currentStreamMu.Unlock()
			return err
		}
		s.logger.Info().
			Str("endpoint_id", update.EndpointId).
			Msg("sent pending update to client")
	}

	// Now wait for the context to be done
	<-stream.Context().Done()

	s.currentStreamMu.Lock()
	if s.currentClientID == clientID {
		s.isStreamActive = false
		s.logger.Info().Msg("client disconnected")
	} else {
		s.logger.Info().Msg("old client stream closed, but a newer one is already active")
	}
	s.currentStreamMu.Unlock()

	return status.Error(codes.Canceled, "client context canceled")
}

// sendUpdateToStream attempts to send an update to the current client stream if one exists
// If no active stream exists, it stores the update to be sent when a client connects
func (s *grpcServer) sendUpdateToStream(update *proto.AuthDataUpdate) {
	s.currentStreamMu.Lock()
	defer s.currentStreamMu.Unlock()

	if !s.isStreamActive || s.currentClientStream == nil {
		// No active stream, store update for later
		s.pendingUpdatesMu.Lock()
		s.pendingUpdates = append(s.pendingUpdates, update)
		count := len(s.pendingUpdates)
		s.pendingUpdatesMu.Unlock()

		s.logger.Info().
			Str("endpoint_id", update.EndpointId).
			Int("pending_count", count).
			Msg("no active client stream, stored update for later")
		return
	}

	// We have an active stream, try to send the update
	err := s.currentClientStream.Send(update)
	if err != nil {
		s.logger.Error().
			Err(err).
			Msg("failed to send update to client, marking stream as inactive")

		// Client likely disconnected, store update and mark stream as inactive
		s.isStreamActive = false

		s.pendingUpdatesMu.Lock()
		s.pendingUpdates = append(s.pendingUpdates, update)
		count := len(s.pendingUpdates)
		s.pendingUpdatesMu.Unlock()

		s.logger.Info().
			Int("pending_count", count).
			Msg("added update to pending queue after send failure")
	} else {
		s.logger.Info().
			Str("endpoint_id", update.EndpointId).
			Msg("sent update to client")
	}
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

		// Try to send the update directly to the client stream
		s.sendUpdateToStream(authDataUpdate)
	}
}

/* -------------------- Helpers -------------------- */
// IsRequestGRPC checks the true if the request is a gRPC request by checking the protocol and content type.
func IsRequestGRPC(req *http.Request) bool {
	return req.ProtoMajor == 2 && req.Header.Get("Content-Type") == "application/grpc"
}
