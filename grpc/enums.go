package grpc

import "github.com/buildwithgrove/path/envoy/auth_server/proto"

// These enums exist to simplify the usage of converting from data source types to
// the types expected by the proto package.

// AuthType is a string type representing the authorization type for a gateway
// endpoint, which matches the values in the AuthType enum in the proto package.
type AuthType string

const (
	AuthTypeNoAuth AuthType = "NO_AUTH"
	AuthTypeAPIKey AuthType = "API_KEY_AUTH"
	AuthTypeJWT    AuthType = "JWT_AUTH"
)

var AuthTypes = map[AuthType]proto.Auth_AuthType{
	AuthTypeNoAuth: proto.Auth_NO_AUTH,
	AuthTypeAPIKey: proto.Auth_API_KEY_AUTH,
	AuthTypeJWT:    proto.Auth_JWT_AUTH,
}

func (a AuthType) IsValid() bool {
	switch a {
	case AuthTypeNoAuth, AuthTypeAPIKey, AuthTypeJWT:
		return true
	default:
		return false
	}
}

// CapacityLimitPeriod is a string type representing the capacity limit period for
// a gateway endpoint, which maps to the CapacityLimitPeriod enum in the proto package.
type CapacityLimitPeriod string

const (
	CapacityLimitPeriodUnspecified CapacityLimitPeriod = "CAPACITY_LIMIT_PERIOD_UNSPECIFIED"
	CapacityLimitPeriodDaily       CapacityLimitPeriod = "CAPACITY_LIMIT_PERIOD_DAILY"
	CapacityLimitPeriodWeekly      CapacityLimitPeriod = "CAPACITY_LIMIT_PERIOD_WEEKLY"
	CapacityLimitPeriodMonthly     CapacityLimitPeriod = "CAPACITY_LIMIT_PERIOD_MONTHLY"
)

var CapacityLimitPeriods = map[CapacityLimitPeriod]proto.CapacityLimitPeriod{
	CapacityLimitPeriodUnspecified: proto.CapacityLimitPeriod_CAPACITY_LIMIT_PERIOD_UNSPECIFIED,
	CapacityLimitPeriodDaily:       proto.CapacityLimitPeriod_CAPACITY_LIMIT_PERIOD_DAILY,
	CapacityLimitPeriodWeekly:      proto.CapacityLimitPeriod_CAPACITY_LIMIT_PERIOD_WEEKLY,
	CapacityLimitPeriodMonthly:     proto.CapacityLimitPeriod_CAPACITY_LIMIT_PERIOD_MONTHLY,
}

func (p CapacityLimitPeriod) IsValid() bool {
	switch p {
	case CapacityLimitPeriodUnspecified, CapacityLimitPeriodDaily, CapacityLimitPeriodWeekly, CapacityLimitPeriodMonthly:
		return true
	default:
		return false
	}
}
