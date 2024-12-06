<div align="center">
<h1>PADS<br/>PATH Auth Data Server</h1>
<img src="https://storage.googleapis.com/grove-brand-assets/Presskit/Logo%20Joined-2.png" alt="Grove logo" width="500"/>

</div>
<br/>

# Table of Contents <!-- omit in toc -->

- [\<\<\<\<\<\<\< HEAD](#-head)
      - [3.1.2. YAML Schema](#312-yaml-schema)
    - [3.2. Postgres](#32-postgres)
      - [3.2.1. Grove Portal DB Compatibility](#321-grove-portal-db-compatibility)
      - [3.2.2. SQLC Autogeneration](#322-sqlc-autogeneration)
  - [4. gRPC Proto File](#4-grpc-proto-file)

## 1. Introduction

<!-- TODO_DOCUMENT(@commoddity): Make sure these docs are accessible in https://path.grove.city/ -->

**PADS** (PATH Auth Data Server) is a gRPC server that provides `Gateway Endpoint` data from a data source to the `Go External Authorization Server` in order to enable authorization for [the PATH Gateway](https://github.com/buildwithgrove/path). The nature of the data source is configurable, for example it could be a YAML file or a Postgres database.

## 2. Gateway Endpoints

[The PATH repo](https://github.com/buildwithgrove/path) contains [the `auth_server` package](https://github.com/buildwithgrove/path/tree/main/envoy/auth_server) which contains the `Go External Authorization Server`.

[This package also defines the `gateway_endpoint.proto` file](https://github.com/buildwithgrove/path/blob/main/envoy/auth_server/proto/gateway_endpoint.proto), which contains the definitions for the `GatewayEndpoints` that PADS must provides to the `Go External Authorization Server`.

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
        // 3. JSON Web Token
        Jwt struct {
            AuthorizedUsers map[string]struct{}
        }
    }
    RateLimiting struct {
        ThroughputLimit int32
        CapacityLimit int32
        CapacityLimitPeriod CapacityLimitPeriod
    }
}
```

## 3. Data Sources

The `server` package contains the `DataSource` interface, which abstracts the data source that provides GatewayEndpoints to the `Go External Authorization Server`.

```go
// AuthDataSource is an interface that abstracts the data source.
// It can be implemented by any data provider (e.g., YAML, Postgres).
type AuthDataSource interface {
	FetchAuthDataSync() (*proto.AuthDataResponse, error)
	AuthDataUpdatesChan() (<-chan *proto.AuthDataUpdate, error)
}

```

- `FetchAuthDataSync()` returns the full set of Gateway Endpoints.
  - This is called when `PADS` starts to populate its Gateway Endpoint Data Store.
- `AuthDataUpdatesChan()` returns a channel that receives auth data updates to the Gateway Endpoints.
  - Updates are streamed as changes are made to the data source.

### 3.1. YAML

If the `YAML_FILEPATH` environment variable is set, PADS will load the data from a YAML file at the specified path.

Hot reloading is supported, so changes to the YAML file will be reflected in the `Go External Authorization Server` without the need to restart PADS.

#### 3.1.1. Example YAML File

```yaml
endpoints:
  # 1. Example of a gateway endpoint using API Key Authorization
  # This endpoint has no rate limits defined (the rate_limiting field is omitted entirely in this case).
  endpoint_1:                                                # The unique identifier for a gateway endpoint.
  
    auth:                                                    # The auth field is required for all endpoints that use authorization. 
      api_key: "api_key_1"                                   # For API Key Authorization, the API key string is required.

  # 2. Example of a gateway endpoint using JWT Authorization
  endpoint_2:
    auth:      
      jwt_authorized_users:                                  # For JWT Authorization, the jwt_authorized_users array is required.
        - "auth0|user_1"                                     # The user ID of an authorized user (in this case, a user ID provided by Auth0).
        - "auth0|user_2"

  # 3. Example of a gateway endpoint with rate limiting enabled and no authorization required 
  # (The auth field is omitted entirely in this case).
  endpoint_3:
    rate_limiting:                                           # This endpoint has a rate limit defined
      throughput_limit: 30                                   # Throughput limit defines the endpoint's per-second (TPS) rate limit.
      capacity_limit: 100000                                 # Capacity limit defines the endpoint's rate limit over longer periods.
      capacity_limit_period: "CAPACITY_LIMIT_PERIOD_MONTHLY" # Capacity limit period defines the period over which the capacity limit is enforced.
```
[Full Example Gateway Endpoints YAML File](./yaml/testdata/gateway-endpoints.example.yaml)

#### 3.1.2. YAML Schema

[The YAML Schema](./yaml/gateway-endpoints.schema.yaml) defines the expected structure of the YAML file.

### 3.2. Postgres

If the `POSTGRES_CONNECTION_STRING` environment variable is set, PADS will connect to the specified Postgres database.

#### 3.2.1. Grove Portal DB Compatibility

The [Postgres Driver schema file](postgres/sqlc/schema.sql) uses tables from the existing Grove Portal database schema, allowing PATH to source its authorization data from the existing Grove Portal DB. 

It converts the data stored in the `portal_applications` table and its associated tables into the `proto.GatewayEndpoint` format expected by PATH's Go External Authorization Server.

It also listens for updates to the Grove Portal DB and streams updates to the `Go External Authorization Server` in real time as changes are made to the connected Postgres database.

[For the full Grove Portal DB schema, refer to the database schema defined in the Portal HTTP DB (PHD) repository](https://github.com/pokt-foundation/portal-http-**db**/blob/master/postgres-driver/sqlc/schema.sql).

#### 3.2.2. SQLC Autogeneration

<div align="center">
<a href="https://docs.sqlc.dev/en/stable">
<img src="https://docs.sqlc.dev/en/stable/_static/logo.png" alt="SQLC logo" width="150"/>
<div>https://docs.sqlc.dev/en/stable</div>
</a>
</div>
<br/>

The Postgres Driver uses `SQLC` to autogenerate the Postgres Go code needed to interact with the Postgres database


The Make target `make gen_sqlc` will regenerate the Go code from the SQLC schema file. This will output autogenerated code to the [postgres/sqlc/schema.sql](postgres/sqlc/schema.sql) and [postgres/sqlc/query.sql](postgres/sqlc/query.sql) files to the [postgres/sqlc](postgres/sqlc) directory.

`SQLC` configuration is defined in the [postgres/sqlc/sqlc.yaml](postgres/sqlc/sqlc.yaml) file.

## 4. gRPC Proto File

[The PATH `auth_server` package](https://github.com/buildwithgrove/path/tree/main/envoy/auth_server) contains the file `gateway_endpoint.proto`, which contains:

- The gRPC auto-generated Go struct definitions for the GatewayEndpoints.
- The `FetchAuthDataSync` and `StreamAuthDataUpdates` methods that the `Go External Authorization Server` uses to populate and update its Gateway Endpoint Data Store.

The autogenerated Go code from [the `gateway_endpoint.proto` file](https://github.com/buildwithgrove/path/blob/main/envoy/auth_server/proto/gateway_endpoint.proto) is installed in PADS from the `github.com/buildwithgrove/path/envoy/auth_server` package.
