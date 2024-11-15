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

	grpc_server "github.com/buildwithgrove/path-auth-data-server/grpc"
	"github.com/buildwithgrove/path-auth-data-server/postgres/sqlc"
)

var _ grpc_server.AuthDataSource = &postgresDataSource{} // postgresDataSource implements the grpc_server.AuthDataSource interface.

type (
	// postgresDataSource implements the AuthDataSource interface for a Postgres database.
	// The database driver is generated using SQLC: https://docs.sqlc.dev/en/latest/index.html
	//
	// The schema defined in ./postgres/sqlc/schema.sql is compatible with the existing
	// Grove Portal database schema to allow PATH to stream updates from Grove Portal DB.
	//
	// For the current Grove Portal DB schema as defined in the Portal HTTP DB (PHD) repo:
	// https://github.com/pokt-foundation/portal-http-db/blob/master/postgres-driver/sqlc/schema.sql
	postgresDataSource struct {
		driver   *postgresDriver
		listener *pgxlisten.Listener

		notificationCh chan *Notification
		updatesCh      chan *proto.AuthDataUpdate

		logger polylog.Logger
	}
	// The postgresDriver struct wraps the SQLC generated queries and the pgxpool.Pool.
	// See: https://docs.sqlc.dev/en/latest/tutorials/getting-started-postgresql.html
	postgresDriver struct {
		*sqlc.Queries
		DB *pgxpool.Pool
	}
)

/* ---------- Postgres Connection Funcs ---------- */

// Regular expression to match a valid PostgreSQL connection string
var postgresConnectionStringRegex = regexp.MustCompile(`^postgres(?:ql)?:\/\/[^:]+:[^@]+@[^:]+:\d+\/[^?]+(?:\?.+)?$`)

/*
NewPostgresDataSource
- Ensures the connection string is valid.
- Parses the connection string into a pgx pool configuration object.
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
		updatesCh:      make(chan *proto.AuthDataUpdate, 100_000),
		logger:         logger,
	}

	// Start listening for updates from the Postgres database
	go postgresDataSource.listenForUpdates(ctx)

	return postgresDataSource, cleanup, nil
}

// isValidPostgresConnectionString checks if a string is a valid PostgreSQL connection string.
func isValidPostgresConnectionString(s string) bool {
	return postgresConnectionStringRegex.MatchString(s)
}

/* ---------- Data Source Funcs ---------- */

// FetchAuthDataSync loads the full set of GatewayEndpoints from the Postgres database.
func (d *postgresDataSource) FetchAuthDataSync() (*proto.AuthDataResponse, error) {

	rows, err := d.driver.Queries.SelectPortalApplications(context.Background())
	if err != nil {
		return nil, err
	}

	return convertPortalApplicationsRows(rows), nil
}

// AuthDataUpdatesChan returns a channel that streams updates when the Postgres database changes.
func (d *postgresDataSource) AuthDataUpdatesChan() (<-chan *proto.AuthDataUpdate, error) {
	return d.updatesCh, nil
}

/* ---------- Data Update Listener Funcs ---------- */

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
// It listens for updates from the Postgres database, using the function and triggers defined in “./postgres/sqlc/schema.sql#L56-162“
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
			update := &proto.AuthDataUpdate{
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
			update := &proto.AuthDataUpdate{
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
