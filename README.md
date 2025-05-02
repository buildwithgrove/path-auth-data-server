<div align="center">
<h1>🐾 PADS<br/>PATH Auth Data Server</h1>
<img src="https://storage.googleapis.com/grove-brand-assets/Presskit/Logo%20Joined-2.png" alt="Grove logo" width="500"/>

</div>
<br/>

# Table of Contents <!-- omit in toc -->

- [1. Introduction](#1-introduction)
- [2. Gateway Endpoints](#2-gateway-endpoints)
- [3. Data Sources](#3-data-sources)
  - [3.1. YAML](#31-yaml)
    - [3.1.1. Example YAML File](#311-example-yaml-file)
    - [3.1.2. YAML Schema](#312-yaml-schema)
  - [3.2. Postgres](#32-postgres)
    - [3.2.1. Grove Portal DB Driver](#321-grove-portal-db-driver)

## 1. Introduction

<!-- TODO_MVP(@commoddity): Move these documents over to path.grove.city -->

**PADS** (PATH Auth Data Server) is a gRPC server that provides `Gateway Endpoint` data from a data source to [`PEAS (PATH External Auth Server)`](https://github.com/buildwithgrove/path-external-auth-server) in order to enable authorization for [PATH](https://github.com/buildwithgrove/path). 

The nature of the data source is configurable, for example it could be a YAML file or a Postgres database.

## 2. Gateway Endpoints

The [PEAS repo](https://github.com/buildwithgrove/path-external-auth-server) contains [the `proto` package](https://github.com/buildwithgrove/path-external-auth-server/tree/main/proto) which contains [the `gateway_endpoint.proto` file](https://github.com/buildwithgrove/path-external-auth-server/blob/main/proto/gateway_endpoint.proto), which contains the definitions for the `GatewayEndpoints` that PADS must provides to `PEAS`.

A single `GatewayEndpoint` represents a single authorized endpoint of the PATH Gateway service, which may be authorized for use by any number of users.

```go
// Simplified representation of the GatewayEndpoint proto message that
// PADS must provide to the `Go External Authorization Server`.
type GatewayEndpoint struct {
    EndpointId string
    // AuthType will be one of the following structs:
    AuthType {
        // 1. No Authorization Required
        NoAuth struct{}
        // 2. Static API Key
        StaticApiKey struct {
          ApiKey string
        }
    }
    RateLimiting struct {
        ThroughputLimit int32
        CapacityLimit int32
        CapacityLimitPeriod CapacityLimitPeriod
    }
    Metadata struct {
        Name string
        AccountId string
        UserId string
        PlanType string
        Email string
        Environment string
    }
}
```

## 3. Data Sources

The `grpc` package contains the [`AuthDataSource`](https://github.com/buildwithgrove/path-auth-data-server/blob/main/grpc/data_source.go) interface, which abstracts the data source that provides `GatewayEndpoint`s to `PEAS`.

```go
type GatewayEndpointsClient interface {
	// FetchAuthDataSync requests the initial set of GatewayEndpoints from the remote gRPC server.
	FetchAuthDataSync(ctx context.Context, in *AuthDataRequest, opts ...grpc.CallOption) (*AuthDataResponse, error)
	// StreamAuthDataUpdates listens for updates from the remote gRPC server and streams them to the client.
	StreamAuthDataUpdates(ctx context.Context, in *AuthDataUpdatesRequest, opts ...grpc.CallOption) (GatewayEndpoints_StreamAuthDataUpdatesClient, error)
}
```

- `FetchAuthDataSync()` returns the full set of Gateway Endpoints.
  - This is called when `PADS` starts to populate its Gateway Endpoint Data Store.
- `StreamAuthDataUpdates()` returns a channel that receives auth data updates to the Gateway Endpoints.
  - Updates are streamed as changes are made to the data source.

### 3.1. YAML

If the `YAML_FILEPATH` environment variable is set, PADS will load the data from a YAML file at the specified path.

Hot reloading is supported, so changes to the YAML file will be reflected in the `Go External Authorization Server` without the need to restart PADS.

#### 3.1.1. Example YAML File

```yaml
endpoints:
  # 1. Example of a gateway endpoint using API Key Authorization
  # This endpoint has no rate limits defined (the rate_limiting field is omitted entirely in this case).
  endpoint_1_static_key: # The unique identifier for a gateway endpoint.
    auth: # The auth field is required for all endpoints that use authorization.
      # The sub-field 'api_key' is required for API Key Authorization.
      # If auth is not set, the endpoint will be treated as using no authorization.
      api_key: "api_key_1" # For API Key Authorization, the API key string is required.
    
    metadata: # Metadata fields may be any key-value pairs and are optional.
      plan_type: "PLAN_UNLIMITED" # Example of a key-value pair (in this case, a pricing plan).
      account_id: "account_1" # Example of a key-value pair (in this case, an account ID).
      email: "user1@example.com" # Example of a key-value pair (in this case, an owner email).

  # 2. Example of a gateway endpoint with no authorization (the auth field is omitted entirely in this case).
  endpoint_2_no_auth:
    rate_limiting: # This endpoint has a rate limit defined
      throughput_limit: 30 # Throughput limit defines the endpoint's per-second (TPS) rate limit.
      capacity_limit: 100000 # Capacity limit defines the endpoint's rate limit over longer periods.
      capacity_limit_period: "CAPACITY_LIMIT_PERIOD_MONTHLY" # Capacity limit period defines the period over which the capacity limit is enforced.
    
    metadata:
      plan_type: "PLAN_FREE"
      account_id: "account_2"
      email: "user2@example.com"
```

[Full Example Gateway Endpoints YAML File](./yaml/testdata/gateway-endpoints.example.yaml)

#### 3.1.2. YAML Schema

[The YAML Schema](./yaml/gateway-endpoints.schema.yaml) defines the expected structure of the YAML file.

### 3.2. Postgres

If the `POSTGRES_CONNECTION_STRING` environment variable is set, PADS will connect to the specified Postgres database.

#### 3.2.1. Grove Portal DB Driver

A highly opinionated Postgres driver that is compatible with the Grove Portal DB is provided in this repository for use in the Grove Portal's authentication implementation.

For more details, see the [Grove Portal DB Driver README.md](https://github.com/buildwithgrove/path-auth-data-server/blob/main/postgres/grove/README.md) documentation.
