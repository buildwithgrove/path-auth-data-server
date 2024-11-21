-- This file is used by SQLC to autogenerate the Go code needed by the database driver. 
-- It contains all queries used for fetching user data by the Gateway.
-- See: https://docs.sqlc.dev/en/latest/tutorials/getting-started-postgresql.html#schema-and-queries

-- name: SelectGatewayEndpoints :many
SELECT ge.id AS endpoint_id,
    ge.auth_type,
    ge.api_key,
    p.throughput_limit,
    p.capacity_limit,
    p.capacity_limit_period::capacity_limit_period AS capacity_limit_period,
    ARRAY_AGG(geu.auth_provider_user_id) FILTER (
        WHERE geu.auth_provider_user_id IS NOT NULL
    )::varchar [] AS authorized_users
FROM gateway_endpoints ge
    LEFT JOIN plans p ON ge.plan_name = p.name
    LEFT JOIN gateway_endpoint_users geu ON ge.id = geu.gateway_endpoint_id
GROUP BY ge.id,
    ge.auth_type,
    ge.api_key,
    p.throughput_limit,
    p.capacity_limit,
    p.capacity_limit_period;

-- name: SelectGatewayEndpoint :one
SELECT ge.id AS endpoint_id,
    ge.auth_type,
    ge.api_key,
    p.throughput_limit,
    p.capacity_limit,
    p.capacity_limit_period::capacity_limit_period AS capacity_limit_period,
    ARRAY_AGG(geu.auth_provider_user_id) FILTER (
        WHERE geu.auth_provider_user_id IS NOT NULL
    )::varchar [] AS authorized_users
FROM gateway_endpoints ge
    LEFT JOIN plans p ON ge.plan_name = p.name
    LEFT JOIN gateway_endpoint_users geu ON ge.id = geu.gateway_endpoint_id
WHERE ge.id = $1
GROUP BY ge.id,
    ge.auth_type,
    ge.api_key,
    p.throughput_limit,
    p.capacity_limit,
    p.capacity_limit_period;

-- name: GetGatewayEndpointChanges :many
SELECT id,
    gateway_endpoint_id,
    is_delete
FROM gateway_endpoint_changes;

-- name: DeleteGatewayEndpointChanges :exec
DELETE FROM gateway_endpoint_changes
WHERE id = ANY($1::int []);
