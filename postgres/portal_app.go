package postgres

import (
	"github.com/buildwithgrove/path/envoy/auth_server/proto"

	"github.com/buildwithgrove/path-auth-data-server/postgres/sqlc"
)

// PortalApplicationRow is a struct that represents a row from the portal_applications table
// in the existing Grove Portal Database. It is necessary to convert the existing `portal_applications`
// table schema to the new `GatewayEndpoint` struct expected by the PATH Go External Authorization Server.
type PortalApplicationRow struct {
	EndpointID        string `json:"endpoint_id"`
	AccountID         string `json:"account_id"`
	SecretKey         string `json:"secret_key"`
	SecretKeyRequired bool   `json:"secret_key_required"`
	Plan              string `json:"plan"`
	CapacityLimit     int32  `json:"capacity_limit"`
	ThroughputLimit   int32  `json:"throughput_limit"`
}

func convertSelectPortalApplicationsRow(r sqlc.SelectPortalApplicationsRow) *PortalApplicationRow {
	return &PortalApplicationRow{
		EndpointID:        r.EndpointID,
		AccountID:         r.AccountID.String,
		SecretKey:         r.SecretKey.String,
		SecretKeyRequired: r.SecretKeyRequired.Bool,
		Plan:              r.Plan.String,
		CapacityLimit:     r.CapacityLimit.Int32,
		ThroughputLimit:   r.ThroughputLimit.Int32,
	}
}

func convertSelectPortalApplicationRow(r sqlc.SelectPortalApplicationRow) *PortalApplicationRow {
	return &PortalApplicationRow{
		EndpointID:        r.EndpointID,
		AccountID:         r.AccountID.String,
		SecretKey:         r.SecretKey.String,
		SecretKeyRequired: r.SecretKeyRequired.Bool,
		Plan:              r.Plan.String,
		CapacityLimit:     r.CapacityLimit.Int32,
		ThroughputLimit:   r.ThroughputLimit.Int32,
	}
}

func (r *PortalApplicationRow) convertToProto() *proto.GatewayEndpoint {
	rateLimiting := &proto.RateLimiting{
		ThroughputLimit: int32(r.ThroughputLimit),
		CapacityLimit:   int32(r.CapacityLimit),
	}
	// The current Portal DB only supports monthly capacity limit periods
	if r.CapacityLimit > 0 {
		rateLimiting.CapacityLimitPeriod = proto.CapacityLimitPeriod_CAPACITY_LIMIT_PERIOD_MONTHLY
	}

	return &proto.GatewayEndpoint{
		EndpointId:   r.EndpointID,
		Auth:         r.getAuthDetails(),
		RateLimiting: rateLimiting,
		Metadata: &proto.Metadata{
			AccountId: r.AccountID,
			PlanType:  r.Plan,
		},
	}
}

func (r *PortalApplicationRow) getAuthDetails() *proto.Auth {
	if r.SecretKeyRequired {
		return &proto.Auth{
			AuthType: &proto.Auth_StaticApiKey{
				StaticApiKey: &proto.StaticAPIKey{
					ApiKey: r.SecretKey,
				},
			},
		}
	} else {
		return &proto.Auth{
			AuthType: &proto.Auth_NoAuth{},
		}
	}
}

// TODO - use this method when we have JWT auth
// func convertToProtoAuthorizedUsers(users []string) map[string]*proto.Empty {
// 	authUsers := make(map[string]*proto.Empty, len(users))
// 	for _, user := range users {
// 		authUsers[user] = &proto.Empty{}
// 	}
// 	return authUsers
// }

func convertPortalApplicationsRows(rows []sqlc.SelectPortalApplicationsRow) *proto.AuthDataResponse {
	endpointsProto := make(map[string]*proto.GatewayEndpoint, len(rows))
	for _, row := range rows {
		portalAppRow := convertSelectPortalApplicationsRow(row)
		endpointsProto[portalAppRow.EndpointID] = portalAppRow.convertToProto()
	}

	return &proto.AuthDataResponse{Endpoints: endpointsProto}
}
