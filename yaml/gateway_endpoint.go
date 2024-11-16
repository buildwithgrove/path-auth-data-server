package yaml

import (
	"fmt"

	"github.com/buildwithgrove/path/envoy/auth_server/proto"
)

type yamlAuthType string

const (
	yamlAuthTypeAPIKey yamlAuthType = "API_KEY_AUTH"
	yamlAuthTypeJWT    yamlAuthType = "JWT_AUTH"
)

/* ----------------------------- GatewayEndpoint YAML Struct ----------------------------- */

type (
	// gatewayEndpointYAML represents the structure of a single GatewayEndpoint in the YAML file.
	gatewayEndpointYAML struct {
		EndpointID   string            `yaml:"endpoint_id"`
		Auth         authYAML          `yaml:"auth"`
		RateLimiting rateLimitingYAML  `yaml:"rate_limiting"`
		Metadata     map[string]string `yaml:"metadata"`
	}
	// authYAML represents the Auth section of a single GatewayEndpoint in the YAML file.
	authYAML struct {
		AuthType           yamlAuthType         `yaml:"auth_type"`
		APIKey             *string              `yaml:"api_key,omitempty"`
		JWTAuthorizedUsers *map[string]struct{} `yaml:"jwt_authorized_users,omitempty"`
	}
	// rateLimitingYAML represents the RateLimiting section of a single GatewayEndpoint in the YAML file.
	rateLimitingYAML struct {
		ThroughputLimit     int    `yaml:"throughput_limit"`
		CapacityLimit       int    `yaml:"capacity_limit"`
		CapacityLimitPeriod string `yaml:"capacity_limit_period"`
	}
)

func (e *gatewayEndpointYAML) convertToProto() *proto.GatewayEndpoint {

	metadata := map[string]string{}
	for key, value := range e.Metadata {
		metadata[key] = value
	}

	return &proto.GatewayEndpoint{
		EndpointId: e.EndpointID,
		Auth:       e.Auth.convertToProto(),
		RateLimiting: &proto.RateLimiting{
			ThroughputLimit:     int32(e.RateLimiting.ThroughputLimit),
			CapacityLimit:       int32(e.RateLimiting.CapacityLimit),
			CapacityLimitPeriod: proto.CapacityLimitPeriod(proto.CapacityLimitPeriod_value[e.RateLimiting.CapacityLimitPeriod]),
		},
		Metadata: metadata,
	}
}

func (a *authYAML) convertToProto() *proto.Auth {
	authProto := &proto.Auth{
		AuthType: proto.Auth_AuthType(proto.Auth_AuthType_value[string(a.AuthType)]),
	}

	switch a.AuthType {

	case yamlAuthTypeAPIKey:
		if a.APIKey != nil {
			authProto.AuthTypeDetails = &proto.Auth_ApiKey{
				ApiKey: &proto.APIKey{
					ApiKey: *a.APIKey,
				},
			}
		}

	case yamlAuthTypeJWT:
		if a.JWTAuthorizedUsers != nil {
			authProto.AuthTypeDetails = &proto.Auth_Jwt{
				Jwt: &proto.JWT{
					AuthorizedUsers: make(map[string]*proto.Empty),
				},
			}
			for user := range *a.JWTAuthorizedUsers {
				jwtDetails := authProto.AuthTypeDetails.(*proto.Auth_Jwt)
				jwtDetails.Jwt.AuthorizedUsers[user] = &proto.Empty{}
			}
		}

	default:
		authProto.AuthTypeDetails = &proto.Auth_NoAuth{}

	}

	return authProto
}

func (e *gatewayEndpointYAML) validate() error {
	if e.EndpointID == "" {
		return fmt.Errorf("endpoint_id is required")
	}
	if err := e.Auth.validate(); err != nil {
		return err
	}
	if e.RateLimiting.CapacityLimit != 0 {
		if e.RateLimiting.CapacityLimitPeriod != proto.CapacityLimitPeriod_name[1] &&
			e.RateLimiting.CapacityLimitPeriod != proto.CapacityLimitPeriod_name[2] &&
			e.RateLimiting.CapacityLimitPeriod != proto.CapacityLimitPeriod_name[3] {
			return fmt.Errorf("capacity_limit_period must be one of %s, %s, or %s",
				proto.CapacityLimitPeriod_name[1],
				proto.CapacityLimitPeriod_name[2],
				proto.CapacityLimitPeriod_name[3],
			)
		}
	}

	return nil
}

func (a *authYAML) validate() error {
	switch a.AuthType {
	case yamlAuthTypeAPIKey:
		if a.APIKey == nil || *a.APIKey == "" {
			return fmt.Errorf("api_key is required for API_KEY_AUTH")
		}
	case yamlAuthTypeJWT:
		if a.JWTAuthorizedUsers == nil || len(*a.JWTAuthorizedUsers) == 0 {
			return fmt.Errorf("jwt_authorized_users is required for JWT_AUTH")
		}
	default:
		return nil
	}
	return nil
}
