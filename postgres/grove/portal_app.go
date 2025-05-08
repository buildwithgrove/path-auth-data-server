package grove

import (
	"github.com/buildwithgrove/path-external-auth-server/proto"

	"github.com/buildwithgrove/path-auth-data-server/postgres/grove/sqlc"
)

// portalApplicationRow is a struct that represents a row from the portal_applications table
// in the existing Grove Portal Database. It is necessary to convert the existing `portal_applications`
// table schema to the new `GatewayEndpoint` struct expected by the PATH Go External Authorization Server.
type portalApplicationRow struct {
	ID                string `json:"id"`                  // The PortalApp ID maps to the GatewayEndpoint.EndpointId
	SecretKey         string `json:"secret_key"`          // The PortalApp SecretKey maps to the GatewayEndpoint.Auth.AuthType.StaticApiKey.ApiKey
	SecretKeyRequired bool   `json:"secret_key_required"` // The PortalApp SecretKeyRequired determines whether the auth type is StaticApiKey or NoAuth
	AccountID         string `json:"account_id"`          // The PortalApp AccountID maps to the GatewayEndpoint.Metadata.AccountId
	Plan              string `json:"plan"`                // The PortalApp Plan maps to the GatewayEndpoint.Metadata.PlanType
}

// sqlcPortalAppsToPortalAppRow (not the plurality of Apps) converts a row from the
// `SelectPortalApplicationsRow` query to the intermediate portalApplicationRow struct.
// This is necessary because SQLC generates a specific struct for each query, which needs
// to be converted to a common struct before converting to the proto.GatewayEndpoint struct.
func sqlcPortalAppsToPortalAppRow(r sqlc.SelectPortalApplicationsRow) *portalApplicationRow {
	return &portalApplicationRow{
		ID:                r.ID,
		SecretKey:         r.SecretKey.String,
		SecretKeyRequired: r.SecretKeyRequired.Bool,
		AccountID:         r.AccountID.String,
		Plan:              r.Plan.String,
	}
}

// sqlcPortalAppToPortalAppRow (not the singularity of App) converts a row from the
// `SelectPortalApplicationRow` query to the intermediate portalApplicationRow struct.
// This is necessary because SQLC generates a specific struct for each query, which needs
// to be converted to a common struct before converting to the proto.GatewayEndpoint struct.
func sqlcPortalAppToPortalAppRow(r sqlc.SelectPortalApplicationRow) *portalApplicationRow {
	return &portalApplicationRow{
		ID:                r.ID,
		SecretKey:         r.SecretKey.String,
		SecretKeyRequired: r.SecretKeyRequired.Bool,
		AccountID:         r.AccountID.String,
		Plan:              r.Plan.String,
	}
}

func (r *portalApplicationRow) convertToProto() *proto.GatewayEndpoint {
	return &proto.GatewayEndpoint{
		EndpointId: r.ID,
		Auth:       r.getAuthDetails(),
		Metadata: &proto.Metadata{
			AccountId: r.AccountID,
			PlanType:  r.Plan,
		},
	}
}

func (r *portalApplicationRow) getAuthDetails() *proto.Auth {
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

func sqlcPortalAppsToProto(rows []sqlc.SelectPortalApplicationsRow) *proto.AuthDataResponse {
	endpointsProto := make(map[string]*proto.GatewayEndpoint, len(rows))
	for _, row := range rows {
		portalAppRow := sqlcPortalAppsToPortalAppRow(row)
		endpointsProto[portalAppRow.ID] = portalAppRow.convertToProto()
	}

	return &proto.AuthDataResponse{Endpoints: endpointsProto}
}
