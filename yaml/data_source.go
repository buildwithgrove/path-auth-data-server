/*
Package yaml provides an implementation of the AuthDataSource interface for YAML files.

It loads YAML data from a file which must match the format defined in gateway-endpoints.schema.yaml.

An example `gateway-endpoints.example.yaml` is provided in the testdata directory.

This package also uses a file watcher to detect changes to the YAML file and sends updates
to the authDataUpdatesCh channel if any changes are detected.
*/
package yaml

import (
	"os"
	"sync"

	"github.com/buildwithgrove/path/envoy/auth_server/proto"
	"github.com/fsnotify/fsnotify"
	"github.com/pokt-network/poktroll/pkg/polylog"
	"gopkg.in/yaml.v3"

	grpc_server "github.com/buildwithgrove/path-auth-data-server/grpc"
)

// yamlDataSource implements the AuthDataSource interface
var _ grpc_server.AuthDataSource = &yamlDataSource{}

/* --------------------------- yamlDataSource Struct ---------------------------- */

// yamlDataSource implements the AuthDataSource interface for YAML files.
//
// It uses a file watcher to detect changes to the YAML file and sends updates to the
// authDataUpdatesCh channel if any changes are detected.
type yamlDataSource struct {
	filename string

	gatewayEndpoints   map[string]*proto.GatewayEndpoint
	gatewayEndpointsMu sync.Mutex

	authDataUpdatesCh chan *proto.AuthDataUpdate

	logger polylog.Logger
}

// NewYAMLDataSource creates a new yamlDataSource for the specified filename.
func NewYAMLDataSource(filename string, logger polylog.Logger) (*yamlDataSource, error) {

	dataSource := &yamlDataSource{
		filename:          filename,
		authDataUpdatesCh: make(chan *proto.AuthDataUpdate, 100_000),
		logger:            logger,
	}

	// Warm up the data store with the full set of GatewayEndpoints from the YAML file.
	gatewayEndpoints, err := dataSource.loadGatewayEndpointsFromYAML()
	if err != nil {
		return nil, err
	}
	dataSource.gatewayEndpoints = gatewayEndpoints.Endpoints

	// Watch the YAML file for changes.
	go dataSource.watchFile()

	return dataSource, nil
}

// FetchAuthDataSync loads the full set of GatewayEndpoints from the YAML file.
func (y *yamlDataSource) FetchAuthDataSync() (*proto.AuthDataResponse, error) {
	return y.loadGatewayEndpointsFromYAML()
}

// AuthDataUpdatesChan returns a channel that streams updates when the YAML file changes.
func (y *yamlDataSource) AuthDataUpdatesChan() (<-chan *proto.AuthDataUpdate, error) {
	return y.authDataUpdatesCh, nil
}

// loadGatewayEndpointsFromYAML reads and parses the YAML file into proto format.
func (y *yamlDataSource) loadGatewayEndpointsFromYAML() (*proto.AuthDataResponse, error) {
	data, err := os.ReadFile(y.filename)
	if err != nil {
		return nil, err
	}

	var endpointsYAML gatewayEndpointsYAML
	if err := yaml.Unmarshal(data, &endpointsYAML); err != nil {
		return nil, err
	}

	if err := endpointsYAML.validate(); err != nil {
		return nil, err
	}

	return endpointsYAML.convertToProto(), nil
}

// watchFile monitors the YAML file for changes and triggers updates.
func (y *yamlDataSource) watchFile() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		y.logger.Error().Err(err).Msg("failed to create file watcher")
		return
	}
	defer watcher.Close()

	err = watcher.Add(y.filename)
	if err != nil {
		y.logger.Error().Err(err).Msg("failed to add file to watcher")
		return
	}

	for {
		select {
		case event := <-watcher.Events:
			// Check if the write operation flag is set in the event
			isWriteEvent := event.Op & fsnotify.Write
			if isWriteEvent == fsnotify.Write {
				newData, err := y.loadGatewayEndpointsFromYAML()
				if err != nil {
					y.logger.Error().Err(err).Msg("error loading new data from updated YAML file")
					continue
				}
				y.handleUpdates(newData.Endpoints)
			}

		case err := <-watcher.Errors:
			y.logger.Error().Err(err).Msg("watcher error")
		}
	}
}

// handleUpdates compares old and new data and sends appropriate updates.
func (y *yamlDataSource) handleUpdates(newEndpoints map[string]*proto.GatewayEndpoint) {
	y.gatewayEndpointsMu.Lock()
	defer y.gatewayEndpointsMu.Unlock()

	// Save old set of gateway endpoints in order to
	// compare with the new set to handle deletions.
	oldGatewayEndpoints := y.gatewayEndpoints

	// Assign new set of gateway endpoints.
	y.gatewayEndpoints = newEndpoints

	// Send updates for new or modified endpoints.
	// The onus of determining if an endpoint is new is on the receiver.
	for id, newEndpoint := range newEndpoints {
		update := &proto.AuthDataUpdate{
			EndpointId:      id,
			GatewayEndpoint: newEndpoint,
		}
		y.authDataUpdatesCh <- update
	}

	// Send delete updates for removed endpoints
	for id := range oldGatewayEndpoints {
		if _, exists := newEndpoints[id]; !exists {
			update := &proto.AuthDataUpdate{
				EndpointId: id,
				Delete:     true,
			}
			y.authDataUpdatesCh <- update
		}
	}
}
