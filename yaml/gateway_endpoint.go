package yaml

import (
	"fmt"

	"github.com/buildwithgrove/path-external-auth-server/proto"

	grpc_server "github.com/buildwithgrove/path-auth-data-server/grpc"
)

/* ----------------------------- GatewayEndpoint YAML Struct ----------------------------- */

type (
	// gatewayEndpointYAML represents the structure of a single GatewayEndpoint in the YAML file.
	gatewayEndpointYAML struct {
		// The authorization configuration for a gateway endpoint. If omitted, the endpoint will not require any authorization.
		Auth authYAML `yaml:"auth"`
		// The rate limiting configuration for a gateway endpoint. May be omitted for endpoints with no rate limiting.
		RateLimiting rateLimitingYAML `yaml:"rate_limiting"`
		// Metadata is an optional map of string keys to string values for additional information about the gateway endpoint.
		Metadata metadataYAML `yaml:"metadata"`
	}
	// authYAML represents the Auth section of a single GatewayEndpoint in the YAML file.
	authYAML struct {
		// APIKey is non-empty if the auth_type is AUTH_TYPE_API_KEY.
		APIKey *string `yaml:"api_key,omitempty"`
	}
	// rateLimitingYAML represents the RateLimiting section of a single GatewayEndpoint in the YAML file.
	rateLimitingYAML struct {
		// ThroughputLimit defines the endpoint's per-second (TPS) rate limit.
		ThroughputLimit int `yaml:"throughput_limit"`
		// CapacityLimit defines the endpoint's rate limit over longer periods.
		CapacityLimit int `yaml:"capacity_limit"`
		// CapacityLimitPeriod defines the period over which the capacity limit is enforced.
		CapacityLimitPeriod grpc_server.CapacityLimitPeriod `yaml:"capacity_limit_period"`
	}
	metadataYAML struct {
		Name        string `yaml:"name"`        // The name of the GatewayEndpoint
		AccountId   string `yaml:"account_id"`  // Unique identifier for the GatewayEndpoint's account
		UserId      string `yaml:"user_id"`     // Identifier for a specific user within the system
		PlanType    string `yaml:"plan_type"`   // Subscription or account plan type (e.g., "PLAN_FREE", "PLAN_UNLIMITED")
		Email       string `yaml:"email"`       // The email address associated with the GatewayEndpoint
		Environment string `yaml:"environment"` // The environment the GatewayEndpoint is in (e.g., "development", "staging", "production")
	}
)

func (e *gatewayEndpointYAML) convertToProto(endpointID string) *proto.GatewayEndpoint {
	return &proto.GatewayEndpoint{
		EndpointId: endpointID,
		Auth:       e.Auth.convertToProto(),
		RateLimiting: &proto.RateLimiting{
			ThroughputLimit:     int32(e.RateLimiting.ThroughputLimit),
			CapacityLimit:       int32(e.RateLimiting.CapacityLimit),
			CapacityLimitPeriod: grpc_server.CapacityLimitPeriods[e.RateLimiting.CapacityLimitPeriod],
		},
		Metadata: &proto.Metadata{
			Name:        e.Metadata.Name,
			AccountId:   e.Metadata.AccountId,
			UserId:      e.Metadata.UserId,
			PlanType:    e.Metadata.PlanType,
			Email:       e.Metadata.Email,
			Environment: e.Metadata.Environment,
		},
	}
}

func (a *authYAML) convertToProto() *proto.Auth {
	switch {

	case a.APIKey != nil:
		return &proto.Auth{
			AuthType: &proto.Auth_StaticApiKey{
				StaticApiKey: &proto.StaticAPIKey{
					ApiKey: *a.APIKey,
				},
			},
		}

	default:
		return &proto.Auth{
			AuthType: &proto.Auth_NoAuth{},
		}
	}
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
	switch {

	// API Key authorization requires an API key to be set for the endpoint.
	case a.APIKey != nil:
		if a.APIKey == nil || *a.APIKey == "" {
			return fmt.Errorf("api_key is required for auth_type: AUTH_TYPE_API_KEY")
		}

	// Default case means no auth is set for the endpoint, which
	// means no authorization fields may be set for the endpoint.
	default:
		if a.APIKey != nil && *a.APIKey != "" {
			return fmt.Errorf("api_key must not be set if no auth is set for the endpoint")
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
