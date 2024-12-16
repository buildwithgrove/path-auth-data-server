// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: grove_queries.sql

package sqlc

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const deletePortalApplicationChanges = `-- name: DeletePortalApplicationChanges :exec
DELETE FROM portal_application_changes
WHERE id = ANY($1::int [])
`

func (q *Queries) DeletePortalApplicationChanges(ctx context.Context, changeIds []int32) error {
	_, err := q.db.Exec(ctx, deletePortalApplicationChanges, changeIds)
	return err
}

const getPortalApplicationChanges = `-- name: GetPortalApplicationChanges :many
SELECT id,
    portal_app_id,
    is_delete
FROM portal_application_changes
`

type GetPortalApplicationChangesRow struct {
	ID          int32  `json:"id"`
	PortalAppID string `json:"portal_app_id"`
	IsDelete    bool   `json:"is_delete"`
}

func (q *Queries) GetPortalApplicationChanges(ctx context.Context) ([]GetPortalApplicationChangesRow, error) {
	rows, err := q.db.Query(ctx, getPortalApplicationChanges)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetPortalApplicationChangesRow
	for rows.Next() {
		var i GetPortalApplicationChangesRow
		if err := rows.Scan(&i.ID, &i.PortalAppID, &i.IsDelete); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const selectPortalApplication = `-- name: SelectPortalApplication :one
SELECT 
    pa.id,
    pas.secret_key,
    pas.secret_key_required,
    pa.account_id,
    a.plan_type AS plan,
    p.throughput_limit AS throughput_limit,
    p.monthly_relay_limit AS capacity_limit
FROM portal_applications pa
LEFT JOIN portal_application_settings pas
    ON pa.id = pas.application_id
LEFT JOIN accounts a 
    ON pa.account_id = a.id
LEFT JOIN pay_plans p 
    ON a.plan_type = p.plan_type
WHERE pa.id = $1
GROUP BY 
    pa.id,
    pas.secret_key,
    pas.secret_key_required,
    a.plan_type,
    p.throughput_limit,
    p.monthly_relay_limit
`

type SelectPortalApplicationRow struct {
	ID                string      `json:"id"`
	SecretKey         pgtype.Text `json:"secret_key"`
	SecretKeyRequired pgtype.Bool `json:"secret_key_required"`
	AccountID         pgtype.Text `json:"account_id"`
	Plan              pgtype.Text `json:"plan"`
	ThroughputLimit   pgtype.Int4 `json:"throughput_limit"`
	CapacityLimit     pgtype.Int4 `json:"capacity_limit"`
}

func (q *Queries) SelectPortalApplication(ctx context.Context, id string) (SelectPortalApplicationRow, error) {
	row := q.db.QueryRow(ctx, selectPortalApplication, id)
	var i SelectPortalApplicationRow
	err := row.Scan(
		&i.ID,
		&i.SecretKey,
		&i.SecretKeyRequired,
		&i.AccountID,
		&i.Plan,
		&i.ThroughputLimit,
		&i.CapacityLimit,
	)
	return i, err
}

const selectPortalApplications = `-- name: SelectPortalApplications :many

SELECT 
    pa.id,
    pas.secret_key,
    pas.secret_key_required,
    pa.account_id,
    a.plan_type AS plan,
    p.throughput_limit AS throughput_limit,
    p.monthly_relay_limit AS capacity_limit
FROM portal_applications pa
LEFT JOIN portal_application_settings pas
    ON pa.id = pas.application_id
LEFT JOIN accounts a 
    ON pa.account_id = a.id
LEFT JOIN pay_plans p 
    ON a.plan_type = p.plan_type
GROUP BY 
    pa.id,
    pas.secret_key,
    pas.secret_key_required,
    a.plan_type,
    p.throughput_limit,
    p.monthly_relay_limit
`

type SelectPortalApplicationsRow struct {
	ID                string      `json:"id"`
	SecretKey         pgtype.Text `json:"secret_key"`
	SecretKeyRequired pgtype.Bool `json:"secret_key_required"`
	AccountID         pgtype.Text `json:"account_id"`
	Plan              pgtype.Text `json:"plan"`
	ThroughputLimit   pgtype.Int4 `json:"throughput_limit"`
	CapacityLimit     pgtype.Int4 `json:"capacity_limit"`
}

// This file is used by SQLC to autogenerate the Go code needed by the database driver.
// It contains all queries used for fetching user data by the Gateway.
// See: https://docs.sqlc.dev/en/latest/tutorials/getting-started-postgresql.html#schema-and-queries
func (q *Queries) SelectPortalApplications(ctx context.Context) ([]SelectPortalApplicationsRow, error) {
	rows, err := q.db.Query(ctx, selectPortalApplications)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []SelectPortalApplicationsRow
	for rows.Next() {
		var i SelectPortalApplicationsRow
		if err := rows.Scan(
			&i.ID,
			&i.SecretKey,
			&i.SecretKeyRequired,
			&i.AccountID,
			&i.Plan,
			&i.ThroughputLimit,
			&i.CapacityLimit,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
