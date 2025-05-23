-- This file is used by SQLC to autogenerate the Go code needed by the database driver. 
-- It contains all queries used for fetching user data by the Gateway.
-- See: https://docs.sqlc.dev/en/latest/tutorials/getting-started-postgresql.html#schema-and-queries

-- name: SelectPortalApplications :many
SELECT 
    pa.id,
    pas.secret_key,
    pas.secret_key_required,
    pa.account_id,
    a.plan_type AS plan
FROM portal_applications pa
LEFT JOIN portal_application_settings pas
    ON pa.id = pas.application_id
LEFT JOIN accounts a 
    ON pa.account_id = a.id
WHERE pa.deleted = false
GROUP BY 
    pa.id,
    pas.secret_key,
    pas.secret_key_required,
    a.plan_type;

-- name: SelectPortalApplication :one
SELECT 
    pa.id,
    pas.secret_key,
    pas.secret_key_required,
    pa.account_id,
    a.plan_type AS plan
FROM portal_applications pa
LEFT JOIN portal_application_settings pas
    ON pa.id = pas.application_id
LEFT JOIN accounts a 
    ON pa.account_id = a.id
WHERE pa.id = $1 AND pa.deleted = false
GROUP BY 
    pa.id,
    pas.secret_key,
    pas.secret_key_required,
    a.plan_type;

-- name: GetPortalApplicationChanges :many
SELECT id,
    portal_app_id,
    is_delete
FROM portal_application_changes;

-- name: DeletePortalApplicationChanges :exec
DELETE FROM portal_application_changes
WHERE id = ANY(@change_ids::int []);
