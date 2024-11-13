package yaml

import (
	"log"
	"os"
	"sync"

	"github.com/buildwithgrove/path/envoy/auth_server/proto"
	"github.com/fsnotify/fsnotify"
	"gopkg.in/yaml.v3"

	grpc_server "github.com/buildwithgrove/path-auth-data-server/grpc"
)

var _ grpc_server.AuthDataSource = &yamlDataSource{} // yamlDataSource implements the AuthDataSource interface

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
}

// NewYAMLDataSource creates a new yamlDataSource for the specified filename.
func NewYAMLDataSource(filename string) (*yamlDataSource, error) {

	y := &yamlDataSource{
		filename:          filename,
		authDataUpdatesCh: make(chan *proto.AuthDataUpdate, 100_000),
	}

	// Warm up the data store with the full set of GatewayEndpoints from the YAML file.
	gatewayEndpoints, err := y.loadGatewayEndpointsFromYAML()
	if err != nil {
		return nil, err
	}
	y.gatewayEndpoints = gatewayEndpoints.Endpoints

	// Watch the YAML file for changes.
	go y.watchFile()

	return y, nil
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

	return endpointsYAML.convertToProto(), nil
}

// watchFile monitors the YAML file for changes and triggers updates.
func (y *yamlDataSource) watchFile() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Printf("Failed to create file watcher: %v", err)
		return
	}
	defer watcher.Close()

	err = watcher.Add(y.filename)
	if err != nil {
		log.Printf("Failed to add file to watcher: %v", err)
		return
	}

	for {
		select {
		case event := <-watcher.Events:
			if event.Op&fsnotify.Write == fsnotify.Write {
				newData, err := y.loadGatewayEndpointsFromYAML()
				if err != nil {
					log.Printf("Error loading new data: %v", err)
					continue
				}
				y.handleUpdates(newData.Endpoints)
			}

		case err := <-watcher.Errors:
			log.Printf("Watcher error: %v", err)
		}
	}
}

// handleUpdates compares old and new data and sends appropriate updates.
func (y *yamlDataSource) handleUpdates(newEndpoints map[string]*proto.GatewayEndpoint) {
	y.gatewayEndpointsMu.Lock()
	defer y.gatewayEndpointsMu.Unlock()

	// Save old set of gateway endpoints in order to
	// compare with the new set to handle deletions.
	gatewayEndpoints := y.gatewayEndpoints

	// Assign new set of gateway endpoints.
	y.gatewayEndpoints = newEndpoints

	// Send updates for new or modified endpoints
	for id, newEndpoint := range newEndpoints {
		update := &proto.AuthDataUpdate{
			EndpointId:      id,
			GatewayEndpoint: newEndpoint,
		}
		y.authDataUpdatesCh <- update
	}

	// Send delete updates for removed endpoints
	for id := range gatewayEndpoints {
		if _, exists := newEndpoints[id]; !exists {
			update := &proto.AuthDataUpdate{
				EndpointId: id,
				Delete:     true,
			}
			y.authDataUpdatesCh <- update
		}
	}
}
