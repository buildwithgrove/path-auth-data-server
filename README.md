<div align="center">
<h1>PADS<br/>PATH Auth Data Server</h1>
<img src="https://storage.googleapis.com/grove-brand-assets/Presskit/Logo%20Joined-2.png" alt="Grove logo" width="500"/>

</div>
<br/>

# Table of Contents <!-- omit in toc -->

- [1. Introduction](#1-introduction)
- [2. Docker Image](#2-docker-image)
- [3. Gateway Endpoints](#3-gateway-endpoints)
- [4. Data Sources](#4-data-sources)
  - [4.1. YAML](#41-yaml)
  - [4.2. Postgres](#42-postgres)
- [5. gRPC Proto File](#5-grpc-proto-file)

## 1. Introduction

**PADS** (PATH Auth Data Server) is a gRPC server that provides `Gateway Endpoint` data to the `Go External Authorization Server` in order to enable authorization for the PATH Gateway.

## 2. Docker Image

The Docker image is built and pushed to the `GitHub Container Registry` automatically whenever a new release is created.

The image is available at `ghcr.io/buildwithgrove/path-auth-data-server:latest`.

## 3. Gateway Endpoints

The PATH repo contains the `auth_server` package, which contains the `Go External Authorization Server`.

This package defines the `gateway_endpoint.proto` file, which contains the definitions for the `GatewayEndpoints` that PADS must provides to the `Go External Authorization Server`.

A single `GatewayEndpoint` represents a single authorized endpoint of the PATH Gateway service, which may be authorized for use by any number of users.

```go
// Simplified representation of the GatewayEndpoint proto message that 
// PADS must provide to the `Go External Authorization Server`.
type GatewayEndpoint struct {
    EndpointId: string
    Auth: {
        RequireAuth: bool
        AuthorizedUsers: map[string]struct{}
    }
    UserAccount: {
        AccountId: string
        PlanType: string
    }
    RateLimiting: {
        ThroughputLimit: int32
        CapacityLimit: int32
        CapacityLimitPeriod: CapacityLimitPeriod [enum]
    }
}
```

## 4. Data Sources

The `server` package contains the `DataSource` interface, which abstracts the data source that provides GatewayEndpoints to the `Go External Authorization Server`.

```go
// DataSource is an interface that abstracts the data source.
// It can be implemented by any data provider (e.g., YAML, Postgres).
type DataSource interface {
	FetchInitialData() (*proto.InitialDataResponse, error)
	SubscribeUpdates() (<-chan *proto.Update, error)
}
```

- `FetchInitialData()` returns the initial data for the Gateway Endpoints.
  - This is called when the `Go External Authorization Server` starts to populate its Gateway Endpoint Data Store.
- `SubscribeUpdates()` returns a channel that receives updates to the Gateway Endpoints.
  - Updates are streamed as changes are made to the data source.

### 4.1. YAML

If the `YAML_FILEPATH` environment variable is set, PADS will load the data from a YAML file at the specified path.

Hot reloading is supported, so changes to the YAML file will be reflected in the `Go External Authorization Server` without the need to restart PADS.

The YAML file must be formatted as follows:

```yaml
endpoints:
  endpoint_1:
    endpoint_id: "endpoint_1"
    auth:
      require_auth: true
      authorized_users:
        "auth0|user_1": {}
    user_account:
      account_id: "account_1"
      plan_type: "PLAN_UNLIMITED"

  endpoint_2:
    endpoint_id: "endpoint_2"
    auth:
      require_auth: false
    user_account:
      account_id: "account_2"
      plan_type: "PLAN_FREE"
    rate_limiting:
      throughput_limit: 30
      capacity_limit: 100000
      capacity_limit_period: "CAPACITY_LIMIT_PERIOD_MONTHLY"
```

### 4.2. Postgres

If the `POSTGRES_CONNECTION_STRING` environment variable is set, PADS will connect to the specified Postgres database.

The connected Postgres database must contain the tables and schema defined in the [`postgres/sqlc/schema.sql`](postgres/sqlc/schema.sql) file.

## 5. gRPC Proto File

The `github.com/buildwithgrove/path/auth_server/proto` package contains the file `gateway_endpoint.proto`, which contains:
- The gRPC auto-generated Go struct definitions for the GatewayEndpoints.
- The `GetInitialData` and `StreamUpdates` methods that the `Go External Authorization Server` uses to populate and update its Gateway Endpoint Data Store.

