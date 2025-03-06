package grpc

import (
	"github.com/buildwithgrove/path-external-auth-server/proto"
)

// AuthDataSource is an interface that abstracts the data source.
// It can be implemented by any data provider or source (e.g., YAML, Postgres).
type AuthDataSource interface {

	// FetchAuthDataSync fetches the full set of GatewayEndpoints from the data source.
	// It is called from PADS and is used to warm up the data store from the data source.
	//
	// eg. PADS -- requests Gateway Endpoints data --> Data Source -- responds with Gateway Endpoints --> PADS
	FetchAuthDataSync() (*proto.AuthDataResponse, error)

	// AuthDataUpdatesChan returns a channel that emits updates to the GatewayEndpoints.
	// These updates are streamed from the data source to the gRPC server.
	//
	// eg. Data Source -- data changes --> PADS -- streams updates --> Go External Authorization Server
	AuthDataUpdatesChan() (<-chan *proto.AuthDataUpdate, error)
}
