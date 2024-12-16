package grove

import (
	"github.com/buildwithgrove/path/envoy/auth_server/proto"

	"github.com/buildwithgrove/path-auth-data-server/postgres/grove/sqlc"
)

// PortalApplicationRow is a struct that represents a row from the portal_applications table
// in the existing Grove Portal Database. It is necessary to convert the existing `portal_applications`
// table schema to the new `GatewayEndpoint` struct expected by the PATH Go External Authorization Server.
type PortalApplicationRow struct {
	ID                string `json:"id"`                  // The PortalApp ID maps to the GatewayEndpoint.EndpointId
	SecretKey         string `json:"secret_key"`          // The PortalApp SecretKey maps to the GatewayEndpoint.Auth.AuthType.StaticApiKey.ApiKey
	SecretKeyRequired bool   `json:"secret_key_required"` // The PortalApp SecretKeyRequired determines whether the auth type is StaticApiKey or NoAuth
	CapacityLimit     int32  `json:"capacity_limit"`      // The PortalApp CapacityLimit maps to the GatewayEndpoint.RateLimiting.CapacityLimit
	ThroughputLimit   int32  `json:"throughput_limit"`    // The PortalApp ThroughputLimit maps to the GatewayEndpoint.RateLimiting.ThroughputLimit
	AccountID         string `json:"account_id"`          // The PortalApp AccountID maps to the GatewayEndpoint.Metadata.AccountId
	Plan              string `json:"plan"`                // The PortalApp Plan maps to the GatewayEndpoint.Metadata.PlanType
}

func convertSelectPortalApplicationsRow(r sqlc.SelectPortalApplicationsRow) *PortalApplicationRow {
	return &PortalApplicationRow{
		ID:                r.ID,
		SecretKey:         r.SecretKey.String,
		SecretKeyRequired: r.SecretKeyRequired.Bool,
		CapacityLimit:     r.CapacityLimit.Int32,
		ThroughputLimit:   r.ThroughputLimit.Int32,
		AccountID:         r.AccountID.String,
		Plan:              r.Plan.String,
	}
}

func convertSelectPortalApplicationRow(r sqlc.SelectPortalApplicationRow) *PortalApplicationRow {
	return &PortalApplicationRow{
		ID:                r.ID,
		SecretKey:         r.SecretKey.String,
		SecretKeyRequired: r.SecretKeyRequired.Bool,
		CapacityLimit:     r.CapacityLimit.Int32,
		ThroughputLimit:   r.ThroughputLimit.Int32,
		AccountID:         r.AccountID.String,
		Plan:              r.Plan.String,
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
		EndpointId:   r.ID,
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
	}
	
	return &proto.Auth{
		AuthType: &proto.Auth_NoAuth{},
	}
}

func convertPortalApplicationsRows(rows []sqlc.SelectPortalApplicationsRow) *proto.AuthDataResponse {
	endpointsProto := make(map[string]*proto.GatewayEndpoint, len(rows))
	for _, row := range rows {
		portalAppRow := convertSelectPortalApplicationsRow(row)
		endpointsProto[portalAppRow.ID] = portalAppRow.convertToProto()
	}

	return &proto.AuthDataResponse{Endpoints: endpointsProto}
}
