package postgres

import (
	"context"
	"fmt"
	"regexp"

	"github.com/buildwithgrove/path/envoy/auth_server/proto"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgxlisten"
	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/path-auth-data-server/postgres/sqlc"
	"github.com/buildwithgrove/path-auth-data-server/server"
)

var _ server.DataSource = &postgresDataSource{} // postgresDriver implements the server.DataSource interface.

type (
	// The postgresDataSource struct satisfies the server.DataSource interface.
	postgresDataSource struct {
		driver         *postgresDriver
		listener       *pgxlisten.Listener
		notificationCh chan *Notification
		updatesCh      chan *proto.Update
		logger         polylog.Logger
	}
	// The postgresDriver struct wraps the sqlc generated queries and the pgxpool.Pool.
	postgresDriver struct {
		*sqlc.Queries
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
func NewPostgresDataSource(ctx context.Context, connectionString string, logger polylog.Logger) (*postgresDataSource, func(), error) {

	if !isValidPostgresConnectionString(connectionString) {
		return nil, nil, fmt.Errorf("invalid postgres connection string")
	}

	config, err := pgxpool.ParseConfig(connectionString)
	if err != nil {
		return nil, nil, err
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, nil, fmt.Errorf("pgxpool.NewWithConfig: %v", err)
	}

	cleanup := func() {
		pool.Close()
	}

	driver := &postgresDriver{
		Queries: sqlc.New(pool),
		DB:      pool,
	}

	postgresDataSource := &postgresDataSource{
		driver:         driver,
		listener:       newPGXPoolListener(pool, logger),
		notificationCh: make(chan *Notification),
		updatesCh:      make(chan *proto.Update, 100_000),
		logger:         logger,
	}

	go postgresDataSource.listenForUpdates(ctx)

	return postgresDataSource, cleanup, nil
}

// isValidPostgresConnectionString checks if a string is a valid PostgreSQL connection string.
func isValidPostgresConnectionString(s string) bool {
	return postgresConnectionStringRegex.MatchString(s)
}

/* ---------- Data Source Funcs ---------- */

// FetchInitialData retrieves all GatewayEndpoints from the database and returns them as a map.
func (d *postgresDataSource) FetchInitialData() (*proto.InitialDataResponse, error) {

	rows, err := d.driver.Queries.SelectPortalApplications(context.Background())
	if err != nil {
		return nil, err
	}

	return convertPortalApplicationsRows(rows), nil
}

// TODO: Implement this
func (d *postgresDataSource) SubscribeUpdates() (<-chan *proto.Update, error) {
	return d.updatesCh, nil
}

/* ---------- Data Update Funcs ---------- */

const portalApplicationChangesChannel = "portal_application_changes"

type Notification struct {
	Payload string
}

type PGXNotificationHandler struct {
	outCh chan *Notification
}

func (h *PGXNotificationHandler) HandleNotification(ctx context.Context, n *pgconn.Notification, conn *pgx.Conn) error {
	h.outCh <- &Notification{Payload: n.Payload}
	return nil
}

// newPGXPoolListener creates a new pgxlisten.Listener with a connection from the provided pool and output channel.
func newPGXPoolListener(pool *pgxpool.Pool, logger polylog.Logger) *pgxlisten.Listener {
	connectFunc := func(ctx context.Context) (*pgx.Conn, error) {
		conn, err := pool.Acquire(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to acquire connection: %v", err)
		}
		return conn.Conn(), nil
	}

	listener := &pgxlisten.Listener{
		Connect: connectFunc,
		LogError: func(ctx context.Context, err error) {
			logger.Error().Err(err).Msg("listener error")
		},
	}

	return listener
}

func (d *postgresDataSource) listenForUpdates(ctx context.Context) {
	handler := &PGXNotificationHandler{outCh: d.notificationCh}
	d.listener.Handle(portalApplicationChangesChannel, handler)

	go func() {
		if err := d.listener.Listen(ctx); err != nil {
			d.logger.Error().Err(err).Msg("listener error")
		}
	}()

	go func() {
		for range d.notificationCh {
			// Process the notification
			if err := d.processPortalApplicationChanges(ctx); err != nil {
				d.logger.Error().Err(err).Msg("failed to process portal application changes")
			}
		}
	}()
}

func (d *postgresDataSource) processPortalApplicationChanges(ctx context.Context) error {

	changes, err := d.driver.GetPortalApplicationChanges(ctx)
	if err != nil {
		return err
	}

	if len(changes) == 0 {
		return nil
	}

	var changeIDs []int32

	for _, change := range changes {

		if change.IsDelete {
			update := &proto.Update{
				EndpointId: change.PortalAppID,
				Delete:     true,
			}
			d.updatesCh <- update
		} else {
			portalAppRow, err := d.driver.SelectPortalApplication(ctx, change.PortalAppID)
			if err != nil {
				d.logger.Error().Err(err).Msg("failed to get portal application")
				continue
			}

			gatewayEndpoint := convertSelectPortalApplicationRow(portalAppRow).convertToProto()

			// Send the update
			update := &proto.Update{
				EndpointId:      gatewayEndpoint.EndpointId,
				GatewayEndpoint: gatewayEndpoint,
			}
			d.updatesCh <- update
		}

		changeIDs = append(changeIDs, change.ID)
	}

	// Use the autogenerated method to delete the processed changes
	err = d.driver.DeletePortalApplicationChanges(ctx, changeIDs)
	if err != nil {
		return err
	}

	return nil
}

/* ---------- Struct Conversion Funcs ---------- */

type PortalApplicationRow struct {
	EndpointID        string   `json:"endpoint_id"`
	AccountID         string   `json:"account_id"`
	SecretKeyRequired bool     `json:"secret_key_required"`
	Plan              string   `json:"plan"`
	CapacityLimit     int32    `json:"capacity_limit"`
	ThroughputLimit   int32    `json:"throughput_limit"`
	AuthorizedUsers   []string `json:"authorized_users"`
}

func convertSelectPortalApplicationsRow(r sqlc.SelectPortalApplicationsRow) *PortalApplicationRow {
	return &PortalApplicationRow{
		EndpointID:        r.EndpointID,
		AccountID:         r.AccountID.String,
		SecretKeyRequired: r.SecretKeyRequired.Bool,
		Plan:              r.Plan.String,
		CapacityLimit:     r.CapacityLimit.Int32,
		ThroughputLimit:   r.ThroughputLimit.Int32,
		AuthorizedUsers:   r.AuthorizedUsers,
	}
}

func convertSelectPortalApplicationRow(r sqlc.SelectPortalApplicationRow) *PortalApplicationRow {
	return &PortalApplicationRow{
		EndpointID:        r.EndpointID,
		AccountID:         r.AccountID.String,
		SecretKeyRequired: r.SecretKeyRequired.Bool,
		Plan:              r.Plan.String,
		CapacityLimit:     r.CapacityLimit.Int32,
		ThroughputLimit:   r.ThroughputLimit.Int32,
		AuthorizedUsers:   r.AuthorizedUsers,
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
			// The current Portal DB only supports monthly capacity limit periods
			CapacityLimitPeriod: proto.CapacityLimitPeriod_CAPACITY_LIMIT_PERIOD_MONTHLY,
		},
		Auth: &proto.Auth{
			// TODO_IMPROVE(@commoddity): Add a dedicated field for requiring auth for backwards compatibility
			RequireAuth:     r.SecretKeyRequired,
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

func convertPortalApplicationsRows(rows []sqlc.SelectPortalApplicationsRow) *proto.InitialDataResponse {
	endpointsProto := make(map[string]*proto.GatewayEndpoint, len(rows))
	for _, row := range rows {
		portalAppRow := convertSelectPortalApplicationsRow(row)
		endpointsProto[portalAppRow.EndpointID] = portalAppRow.convertToProto()
	}

	return &proto.InitialDataResponse{Endpoints: endpointsProto}
}
