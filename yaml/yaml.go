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

var _ grpc_server.DataSource = &yamlDataSource{} // yamlDataSource implements the DataSource interface

/* --------------------------- yamlDataSource Struct ---------------------------- */

// yamlDataSource implements the DataSource interface for YAML files.
type yamlDataSource struct {
	filename  string
	updatesCh chan *proto.Update
	endpoints map[string]*proto.GatewayEndpoint
	mu        sync.Mutex
}

// NewYAMLDataSource creates a new yamlDataSource for the specified filename.
func NewYAMLDataSource(filename string) (*yamlDataSource, error) {
	y := &yamlDataSource{
		filename:  filename,
		updatesCh: make(chan *proto.Update, 100_000),
	}

	initialData, err := y.loadGatewayEndpointsFromYAML()
	if err != nil {
		return nil, err
	}

	y.endpoints = initialData.Endpoints

	go y.watchFile()

	return y, nil
}

// FetchInitialData loads the initial data from the YAML file.
func (y *yamlDataSource) FetchInitialData() (*proto.InitialDataResponse, error) {
	return y.loadGatewayEndpointsFromYAML()
}

// SubscribeUpdates returns a channel that streams updates when the YAML file changes.
func (y *yamlDataSource) GetUpdatesChan() (<-chan *proto.Update, error) {
	return y.updatesCh, nil
}

// loadGatewayEndpointsFromYAML reads and parses the YAML file into proto format.
func (y *yamlDataSource) loadGatewayEndpointsFromYAML() (*proto.InitialDataResponse, error) {
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
	y.mu.Lock()
	defer y.mu.Unlock()

	endpoints := y.endpoints
	y.endpoints = newEndpoints

	// Send updates for new or modified endpoints
	for id, newEndpoint := range newEndpoints {
		update := &proto.Update{
			EndpointId:      id,
			GatewayEndpoint: newEndpoint,
		}
		y.updatesCh <- update
	}

	// Send delete updates for removed endpoints
	for id := range endpoints {
		if _, exists := newEndpoints[id]; !exists {
			update := &proto.Update{
				EndpointId: id,
				Delete:     true,
			}
			y.updatesCh <- update
		}
	}
}
