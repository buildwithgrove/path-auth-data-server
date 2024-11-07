package server

import (
	"github.com/buildwithgrove/path/envoy/auth_server/proto"
)

// DataSource is an interface that abstracts the data source.
// It can be implemented by any data provider (e.g., YAML, Postgres).
type DataSource interface {
	FetchInitialData() (*proto.InitialDataResponse, error)
	SubscribeUpdates() (<-chan *proto.Update, error)
}
