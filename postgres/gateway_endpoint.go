package postgres

import (
	"github.com/buildwithgrove/path/envoy/auth_server/proto"

	grpc_server "github.com/buildwithgrove/path-auth-data-server/grpc"
	"github.com/buildwithgrove/path-auth-data-server/postgres/sqlc"
)

// GatewayEndpointRow is a struct that represents a row from the gateway_endpoints table
// in the new database schema. It is necessary to convert the existing `gateway_endpoints`
// table schema to the `GatewayEndpoint` struct expected by the PATH Go External Authorization Server.
type GatewayEndpointRow struct {
	EndpointID          string
	AuthType            grpc_server.AuthType
	ApiKey              string
	AuthorizedUsers     []string
	ThroughputLimit     int32
	CapacityLimit       int32
	CapacityLimitPeriod grpc_server.CapacityLimitPeriod
}

func convertSelectGatewayEndpointRow(r sqlc.SelectGatewayEndpointRow) *GatewayEndpointRow {
	return &GatewayEndpointRow{
		EndpointID:          r.EndpointID,
		AuthType:            r.AuthType,
		ApiKey:              r.ApiKey.String,
		AuthorizedUsers:     r.AuthorizedUsers,
		ThroughputLimit:     r.ThroughputLimit.Int32,
		CapacityLimit:       r.CapacityLimit.Int32,
		CapacityLimitPeriod: r.CapacityLimitPeriod,
	}
}

func convertSelectGatewayEndpointsRow(r sqlc.SelectGatewayEndpointsRow) *GatewayEndpointRow {
	return &GatewayEndpointRow{
		EndpointID:          r.EndpointID,
		AuthType:            r.AuthType,
		ApiKey:              r.ApiKey.String,
		AuthorizedUsers:     r.AuthorizedUsers,
		ThroughputLimit:     r.ThroughputLimit.Int32,
		CapacityLimit:       r.CapacityLimit.Int32,
		CapacityLimitPeriod: r.CapacityLimitPeriod,
	}
}

func (r *GatewayEndpointRow) convertToProto() *proto.GatewayEndpoint {
	return &proto.GatewayEndpoint{
		EndpointId: r.EndpointID,
		Auth:       r.getAuthDetails(),
		RateLimiting: &proto.RateLimiting{
			ThroughputLimit:     r.ThroughputLimit,
			CapacityLimit:       r.CapacityLimit,
			CapacityLimitPeriod: grpc_server.CapacityLimitPeriods[r.CapacityLimitPeriod],
		},
	}
}

func (r *GatewayEndpointRow) getAuthDetails() *proto.Auth {
	authProto := &proto.Auth{
		AuthType: grpc_server.AuthTypes[r.AuthType],
	}

	switch r.AuthType {

	case grpc_server.AuthTypeAPIKey:
		if r.ApiKey != "" {
			authProto.AuthTypeDetails = &proto.Auth_ApiKey{
				ApiKey: &proto.APIKey{
					ApiKey: r.ApiKey,
				},
			}
		}

	case grpc_server.AuthTypeJWT:
		if len(r.AuthorizedUsers) > 0 {
			jwtDetails := &proto.Auth_Jwt{
				Jwt: &proto.JWT{AuthorizedUsers: make(map[string]*proto.Empty)},
			}
			for _, user := range r.AuthorizedUsers {
				jwtDetails.Jwt.AuthorizedUsers[user] = &proto.Empty{}
			}
			authProto.AuthTypeDetails = jwtDetails
		}

	default:
		authProto.AuthTypeDetails = &proto.Auth_NoAuth{}

	}

	return authProto
}

func convertGatewayEndpointsRows(rows []sqlc.SelectGatewayEndpointsRow) *proto.AuthDataResponse {
	endpointsProto := make(map[string]*proto.GatewayEndpoint, len(rows))
	for _, row := range rows {
		gatewayEndpointRow := convertSelectGatewayEndpointsRow(row)
		endpointsProto[gatewayEndpointRow.EndpointID] = gatewayEndpointRow.convertToProto()
	}

	return &proto.AuthDataResponse{Endpoints: endpointsProto}
}
