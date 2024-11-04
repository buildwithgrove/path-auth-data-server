-- This file is used by SQLC to autogenerate the Go code needed by the database driver. 
-- It contains all queries used for fetching user data by the Gateway.
-- See: https://docs.sqlc.dev/en/latest/tutorials/getting-started-postgresql.html#schema-and-queries

-- name: SelectPortalApplications :many
SELECT 
    pa.id AS endpoint_id,
    pa.account_id,
    a.plan_type AS plan,
    p.throughput_limit AS capacity_limit,
    p.monthly_relay_limit AS throughput_limit,
    ARRAY_AGG(DISTINCT uap.provider_user_id)::VARCHAR[] AS authorized_users
FROM portal_applications pa
LEFT JOIN accounts a 
    ON pa.account_id = a.id
LEFT JOIN pay_plans p 
    ON a.plan_type = p.plan_type
LEFT JOIN account_users au
    ON a.id = au.account_id
LEFT JOIN user_auth_providers uap
    ON au.user_id = uap.user_id
GROUP BY 
    pa.id,
    a.plan_type,
    p.throughput_limit,
    p.monthly_relay_limit;

-- name: SelectPortalApplication :one
SELECT 
    pa.id AS endpoint_id,
    pa.account_id,
    a.plan_type AS plan,
    p.throughput_limit AS capacity_limit,
    p.monthly_relay_limit AS throughput_limit,
    ARRAY_AGG(DISTINCT uap.provider_user_id)::VARCHAR[] AS authorized_users
FROM portal_applications pa
LEFT JOIN accounts a 
    ON pa.account_id = a.id
LEFT JOIN pay_plans p 
    ON a.plan_type = p.plan_type
LEFT JOIN account_users au
    ON a.id = au.account_id
LEFT JOIN user_auth_providers uap
    ON au.user_id = uap.user_id
WHERE pa.id = $1
GROUP BY 
    pa.id,
    a.plan_type,
    p.throughput_limit,
    p.monthly_relay_limit;
