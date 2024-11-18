package yaml

import (
	"fmt"

	"github.com/buildwithgrove/path/envoy/auth_server/proto"
)

/* ----------------------------- GatewayEndpoints YAML Struct ----------------------------- */

// gatewayEndpointsYAML represents the structure of the YAML file, which contains a map of GatewayEndpoints.
type gatewayEndpointsYAML struct {
	Endpoints map[string]gatewayEndpointYAML `yaml:"endpoints"`
}

func (g *gatewayEndpointsYAML) convertToProto() *proto.AuthDataResponse {
	endpointsProto := make(map[string]*proto.GatewayEndpoint)
	for endpointID, endpointYAML := range g.Endpoints {
		endpointsProto[endpointID] = endpointYAML.convertToProto(endpointID)
	}
	return &proto.AuthDataResponse{Endpoints: endpointsProto}
}

func (g *gatewayEndpointsYAML) validate() error {
	for endpointID, endpoint := range g.Endpoints {
		if err := endpoint.validate(endpointID); err != nil {
			return fmt.Errorf("validation failed for endpoint %s: %w", endpointID, err)
		}
	}
	return nil
}
