package yaml

import "github.com/buildwithgrove/path/envoy/auth_server/proto"

/* ----------------------------- GatewayEndpoints YAML Struct ----------------------------- */

// gatewayEndpointsYAML represents the structure of the YAML file, which contains a map of GatewayEndpoints.
type gatewayEndpointsYAML struct {
	Endpoints map[string]gatewayEndpointYAML `yaml:"endpoints"`
}

func (g *gatewayEndpointsYAML) convertToProto() *proto.AuthDataResponse {
	endpointsProto := make(map[string]*proto.GatewayEndpoint)
	for _, endpointYAML := range g.Endpoints {
		endpointsProto[endpointYAML.EndpointID] = endpointYAML.convertToProto()
	}
	return &proto.AuthDataResponse{Endpoints: endpointsProto}
}
