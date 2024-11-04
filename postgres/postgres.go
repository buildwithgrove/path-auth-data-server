package postgres

import (
	"context"
	"fmt"
	"regexp"

	"github.com/buildwithgrove/path/envoy/auth_server/proto"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/buildwithgrove/path-auth-grpc-server/server"
)

var _ server.DataSource = &postgresDataSource{} // postgresDriver implements the server.DataSource interface.

type (
	// The postgresDataSource struct satisfies the server.DataSource interface.
	postgresDataSource struct {
		driver *postgresDriver
	}
	// The postgresDriver struct wraps the sqlc generated queries and the pgxpool.Pool.
	postgresDriver struct {
		*Queries
		DB *pgxpool.Pool
	}
)

/* ---------- Postgres Connection Funcs ---------- */

// Regular expression to match a valid PostgreSQL connection string
var postgresConnectionStringRegex = regexp.MustCompile(`^postgres://[^:]+:[^@]+@[^:]+:\d+/.+$`)

/*
NewPostgresDataSource
- Ensures the connection string is valid.
- Parses the connection string into a pgx pool configuration object.
- Ensures connections are read-only.
- Creates a pool of connections to a PostgreSQL database using the provided connection string.
- Creates an instance of postgresDriver using the provided pgx connection and sqlc queries.
- Returns the created postgresDataSource instance.
*/
func NewPostgresDataSource(connectionString string) (*postgresDataSource, func(), error) {

	if !isValidPostgresConnectionString(connectionString) {
		return nil, nil, fmt.Errorf("invalid postgres connection string")
	}

	config, err := pgxpool.ParseConfig(connectionString)
	if err != nil {
		return nil, nil, err
	}

	// Enforce that connections are read-only, as the PATH auth gRPC server does not make any modifications to the database.
	config.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		_, err := conn.Exec(ctx, "SET SESSION CHARACTERISTICS AS TRANSACTION READ ONLY")
		return err
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, nil, fmt.Errorf("pgxpool.NewWithConfig: %v", err)
	}

	cleanup := pool.Close

	driver := &postgresDriver{
		Queries: New(pool),
		DB:      pool,
	}

	return &postgresDataSource{driver: driver}, cleanup, nil
}

// isValidPostgresConnectionString checks if a string is a valid PostgreSQL connection string.
func isValidPostgresConnectionString(s string) bool {
	return postgresConnectionStringRegex.MatchString(s)
}

/* ---------- Data Source Funcs ---------- */

// GetInitialData retrieves all GatewayEndpoints from the database and returns them as a map.
func (d *postgresDataSource) GetInitialData() (*proto.InitialDataResponse, error) {

	rows, err := d.driver.Queries.SelectPortalApplications(context.Background())
	if err != nil {
		return nil, err
	}

	return convertPortalApplicationsRows(rows), nil
}

// TODO: Implement this
func (d *postgresDataSource) SubscribeUpdates() (<-chan *proto.Update, error) {
	return nil, nil
}

/* ---------- Struct Conversion Funcs ---------- */

type PortalApplicationRow struct {
	EndpointID      string   `json:"endpoint_id"`
	AccountID       string   `json:"account_id"`
	Plan            string   `json:"plan"`
	CapacityLimit   int32    `json:"capacity_limit"`
	ThroughputLimit int32    `json:"throughput_limit"`
	AuthorizedUsers []string `json:"authorized_users"`
}

func (r *SelectPortalApplicationsRow) toPortalApplicationRow() *PortalApplicationRow {
	return &PortalApplicationRow{
		EndpointID:      r.EndpointID,
		AccountID:       r.AccountID.String,
		Plan:            r.Plan.String,
		CapacityLimit:   r.CapacityLimit.Int32,
		ThroughputLimit: r.ThroughputLimit.Int32,
		AuthorizedUsers: r.AuthorizedUsers,
	}
}

func (r *SelectPortalApplicationRow) toPortalApplicationRow() *PortalApplicationRow {
	return &PortalApplicationRow{
		EndpointID:      r.EndpointID,
		AccountID:       r.AccountID.String,
		Plan:            r.Plan.String,
		CapacityLimit:   r.CapacityLimit.Int32,
		ThroughputLimit: r.ThroughputLimit.Int32,
		AuthorizedUsers: r.AuthorizedUsers,
	}
}

func (r *PortalApplicationRow) convertToProto() *proto.GatewayEndpoint {
	return &proto.GatewayEndpoint{
		EndpointId: r.EndpointID,
		UserAccount: &proto.UserAccount{
			AccountId: r.AccountID,
			PlanType:  r.Plan,
		},
		RateLimiting: &proto.RateLimiting{
			ThroughputLimit: int32(r.ThroughputLimit),
			CapacityLimit:   int32(r.CapacityLimit),
			// TODO_IMPROVE(@commoddity): Add support for different capacity limit periods
			CapacityLimitPeriod: proto.CapacityLimitPeriod_CAPACITY_LIMIT_PERIOD_MONTHLY,
		},
		Auth: &proto.Auth{
			AuthorizedUsers: convertToProtoAuthorizedUsers(r.AuthorizedUsers),
		},
	}
}

func convertToProtoAuthorizedUsers(users []string) map[string]*proto.Empty {
	authUsers := make(map[string]*proto.Empty, len(users))
	for _, user := range users {
		authUsers[user] = &proto.Empty{}
	}
	return authUsers
}

func convertPortalApplicationsRows(rows []SelectPortalApplicationsRow) *proto.InitialDataResponse {
	endpointsProto := make(map[string]*proto.GatewayEndpoint, len(rows))
	for _, row := range rows {
		portalAppRow := row.toPortalApplicationRow()
		endpointsProto[portalAppRow.EndpointID] = portalAppRow.convertToProto()
	}

	return &proto.InitialDataResponse{Endpoints: endpointsProto}
}
