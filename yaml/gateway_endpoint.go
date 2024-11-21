package yaml

import (
	"fmt"

	"github.com/buildwithgrove/path/envoy/auth_server/proto"

	grpc_server "github.com/buildwithgrove/path-auth-data-server/grpc"
)

/* ----------------------------- GatewayEndpoint YAML Struct ----------------------------- */

type (
	// gatewayEndpointYAML represents the structure of a single GatewayEndpoint in the YAML file.
	gatewayEndpointYAML struct {
		Auth         authYAML          `yaml:"auth"`
		RateLimiting rateLimitingYAML  `yaml:"rate_limiting"`
		Metadata     map[string]string `yaml:"metadata"`
	}
	// authYAML represents the Auth section of a single GatewayEndpoint in the YAML file.
	authYAML struct {
		AuthType           grpc_server.AuthType `yaml:"auth_type"`
		APIKey             *string              `yaml:"api_key,omitempty"`
		JWTAuthorizedUsers []string             `yaml:"jwt_authorized_users,omitempty"`
	}
	// rateLimitingYAML represents the RateLimiting section of a single GatewayEndpoint in the YAML file.
	rateLimitingYAML struct {
		ThroughputLimit     int                             `yaml:"throughput_limit"`
		CapacityLimit       int                             `yaml:"capacity_limit"`
		CapacityLimitPeriod grpc_server.CapacityLimitPeriod `yaml:"capacity_limit_period"`
	}
)

func (e *gatewayEndpointYAML) convertToProto(endpointID string) *proto.GatewayEndpoint {

	metadata := map[string]string{}
	for key, value := range e.Metadata {
		metadata[key] = value
	}

	return &proto.GatewayEndpoint{
		EndpointId: endpointID,
		Auth:       e.Auth.convertToProto(),
		RateLimiting: &proto.RateLimiting{
			ThroughputLimit:     int32(e.RateLimiting.ThroughputLimit),
			CapacityLimit:       int32(e.RateLimiting.CapacityLimit),
			CapacityLimitPeriod: grpc_server.CapacityLimitPeriods[e.RateLimiting.CapacityLimitPeriod],
		},
		Metadata: metadata,
	}
}

func (a *authYAML) convertToProto() *proto.Auth {
	authProto := &proto.Auth{
		AuthType: grpc_server.AuthTypes[a.AuthType],
	}

	switch a.AuthType {

	case grpc_server.AuthTypeAPIKey:
		if a.APIKey != nil {
			authProto.AuthTypeDetails = &proto.Auth_ApiKey{
				ApiKey: *a.APIKey,
			}
		}

	case grpc_server.AuthTypeJWT:
		if a.JWTAuthorizedUsers != nil {
			jwtDetails := &proto.Auth_Jwt{
				Jwt: &proto.JWT{AuthorizedUsers: make(map[string]*proto.Empty)},
			}
			for _, user := range a.JWTAuthorizedUsers {
				jwtDetails.Jwt.AuthorizedUsers[user] = &proto.Empty{}
			}
			authProto.AuthTypeDetails = jwtDetails
		}

	default:
		authProto.AuthTypeDetails = &proto.Auth_NoAuth{}

	}

	return authProto
}

// gatewayEndpoint.validate ensures all fields set for the gateway endpoint are valid.
func (e *gatewayEndpointYAML) validate(endpointID string) error {
	if endpointID == "" {
		return fmt.Errorf("endpoint_id is required")
	}
	if err := e.Auth.validate(); err != nil {
		return err
	}
	if err := e.RateLimiting.validate(); err != nil {
		return err
	}
	return nil
}

// authYAML.validate ensures that the auth section of a GatewayEndpoint is valid by
// checking that the correct fields are set for the given auth type and are not set
// for any other auth type.
func (a *authYAML) validate() error {

	if a != nil && a.AuthType != "" && !a.AuthType.IsValid() {
		return fmt.Errorf("auth_type must be one of %s, %s, or %s",
			grpc_server.AuthTypeNoAuth,
			grpc_server.AuthTypeAPIKey,
			grpc_server.AuthTypeJWT,
		)
	}

	switch a.AuthType {

	// API Key authorization requires an API key to be set for the endpoint.
	case grpc_server.AuthTypeAPIKey:
		if len(a.JWTAuthorizedUsers) > 0 {
			return fmt.Errorf("jwt_authorized_users must not be set for auth_type: API_KEY_AUTH")
		}
		if a.APIKey == nil || *a.APIKey == "" {
			return fmt.Errorf("api_key is required for auth_type: API_KEY_AUTH")
		}

	// JWT authorization requires a list of authorized user IDs to be set for the endpoint.
	case grpc_server.AuthTypeJWT:
		if a.APIKey != nil && *a.APIKey != "" {
			return fmt.Errorf("api_key must not be set for auth_type: JWT_AUTH")
		}
		if len(a.JWTAuthorizedUsers) == 0 {
			return fmt.Errorf("jwt_authorized_users is required for auth_type: JWT_AUTH")
		}
		for _, user := range a.JWTAuthorizedUsers {
			if user == "" {
				return fmt.Errorf("jwt_authorized_users must not contain empty strings")
			}
		}

	// Default case means no auth is set for the endpoint, which
	// means no authorization fields may be set for the endpoint.
	default:
		if a.APIKey != nil && *a.APIKey != "" {
			return fmt.Errorf("api_key must not be set if no auth is set for the endpoint")
		}
		if len(a.JWTAuthorizedUsers) > 0 {
			return fmt.Errorf("jwt_authorized_users must not be set if no auth is set for the endpoint")
		}
	}

	return nil
}

// rateLimitingYAML.validate ensures that the rate limiting section of a GatewayEndpoint
// is valid by checking that the capacity limit period is one of the allowed values.
func (r *rateLimitingYAML) validate() error {
	if r.ThroughputLimit < 0 {
		return fmt.Errorf("throughput_limit must not be negative")
	}
	if r.CapacityLimit < 0 {
		return fmt.Errorf("capacity_limit must not be negative")
	}
	if r.CapacityLimit > 0 {
		if !r.CapacityLimitPeriod.IsValid() {
			return fmt.Errorf("capacity_limit_period must be one of %s, %s, or %s",
				grpc_server.CapacityLimitPeriodDaily,
				grpc_server.CapacityLimitPeriodWeekly,
				grpc_server.CapacityLimitPeriodMonthly,
			)
		}
	}
	return nil
}
