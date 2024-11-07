package yaml

import "github.com/buildwithgrove/path/envoy/auth_server/proto"

/* --------------------------------- YAML Structs -------------------------------- */

// gatewayEndpointsYAML represents the structure of the YAML file.
type (
	gatewayEndpointsYAML struct {
		Endpoints map[string]gatewayEndpointYAML `yaml:"endpoints"`
	}
	gatewayEndpointYAML struct {
		EndpointID   string           `yaml:"endpoint_id"`
		Auth         authYAML         `yaml:"auth"`
		UserAccount  userAccountYAML  `yaml:"user_account"`
		RateLimiting rateLimitingYAML `yaml:"rate_limiting"`
	}
	authYAML struct {
		RequireAuth     bool                `yaml:"require_auth"`
		AuthorizedUsers map[string]struct{} `yaml:"authorized_users"`
	}
	userAccountYAML struct {
		AccountID string `yaml:"account_id"`
		PlanType  string `yaml:"plan_type"`
	}
	rateLimitingYAML struct {
		ThroughputLimit     int    `yaml:"throughput_limit"`
		CapacityLimit       int    `yaml:"capacity_limit"`
		CapacityLimitPeriod string `yaml:"capacity_limit_period"`
	}
)

func (g *gatewayEndpointsYAML) convertToProto() *proto.InitialDataResponse {
	endpointsProto := make(map[string]*proto.GatewayEndpoint)
	for _, endpointYAML := range g.Endpoints {
		endpointsProto[endpointYAML.EndpointID] = endpointYAML.convertToProto()
	}
	return &proto.InitialDataResponse{Endpoints: endpointsProto}
}

func (e *gatewayEndpointYAML) convertToProto() *proto.GatewayEndpoint {
	return &proto.GatewayEndpoint{
		EndpointId: e.EndpointID,
		Auth:       e.Auth.convertToProto(),
		UserAccount: &proto.UserAccount{
			AccountId: e.UserAccount.AccountID,
			PlanType:  e.UserAccount.PlanType,
		},
		RateLimiting: &proto.RateLimiting{
			ThroughputLimit:     int32(e.RateLimiting.ThroughputLimit),
			CapacityLimit:       int32(e.RateLimiting.CapacityLimit),
			CapacityLimitPeriod: proto.CapacityLimitPeriod(proto.CapacityLimitPeriod_value[e.RateLimiting.CapacityLimitPeriod]),
		},
	}
}

func (a *authYAML) convertToProto() *proto.Auth {
	authProto := &proto.Auth{
		RequireAuth:     a.RequireAuth,
		AuthorizedUsers: make(map[string]*proto.Empty),
	}
	for user := range a.AuthorizedUsers {
		authProto.AuthorizedUsers[user] = &proto.Empty{}
	}
	return authProto
}
