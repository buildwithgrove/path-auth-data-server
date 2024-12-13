package main

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/buildwithgrove/path/envoy/auth_server/proto"
	"github.com/pokt-network/poktroll/pkg/polylog"
	"github.com/pokt-network/poktroll/pkg/polylog/polyzero"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"

	grpc_server "github.com/buildwithgrove/path-auth-data-server/grpc"
	"github.com/buildwithgrove/path-auth-data-server/postgres"
	"github.com/buildwithgrove/path-auth-data-server/yaml"

	_ "github.com/joho/godotenv/autoload"
)

func main() {
	logger := polyzero.NewLogger()

	env, err := gatherEnvVars()
	if err != nil {
		panic(fmt.Errorf("failed to gather environment variables: %v", err))
	}

	// 1. Load the data source
	authDataSource, cleanup, err := getAuthDataSource(env, logger)
	if err != nil {
		panic(err)
	}
	defer cleanup()

	// 2. Initialize the gRPC server that will serve the Gateway Endpoints
	ln, err := net.Listen("tcp", fmt.Sprintf(":%s", env.port))
	if err != nil {
		panic(fmt.Sprintf("failed to listen: %v", err))
	}

	server, err := grpc_server.NewGRPCServer(authDataSource, logger)
	if err != nil {
		panic(fmt.Sprintf("failed to create server: %v", err))
	}

	grpcServer := grpc.NewServer()
	proto.RegisterGatewayEndpointsServer(grpcServer, server)

	// create a new HTTP server mux for health check on `/healthz`
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("OK")); err != nil {
			logger.Error().Err(err).Msg("failed to write health check response")
		}
	})

	// create a new HTTP handler that serves both gRPC (for Gateway Endpoints) and HTTP (for health check)
	grpcAndHTTPHandler := h2c.NewHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if grpc_server.IsRequestGRPC(r) {
			grpcServer.ServeHTTP(w, r)
		} else {
			mux.ServeHTTP(w, r)
		}
	}), &http2.Server{})

	logger.Info().Str(portEnv, env.port).Msg("PATH Auth Data Server listening.")

	httpServer := &http.Server{Handler: grpcAndHTTPHandler}
	if err := httpServer.Serve(ln); err != nil {
		panic(fmt.Sprintf("failed to serve: %v", err))
	}
}

/* ------------------------------- Get Auth Data Source ------------------------------- */

// getAuthDataSource returns an AuthDataSource and a cleanup function.
// The cleanup function must be invoked by the caller to ensure resources are released.
func getAuthDataSource(env envVars, logger polylog.Logger) (grpc_server.AuthDataSource, func(), error) {

	// Environment variables are validated in gatherEnvVars, so
	// only one variable is checked in the switch statement at a time.

	// The auth data source used depends on which environment variable is set.
	switch {

	// POSTGRES_CONNECTION_STRING - use a Postgres database as the data source
	case env.postgresConnectionString != "":
		return getPostgresAuthDataSource(env, logger)

	// YAML_FILEPATH - use a local YAML file as the data source
	case env.yamlFilepath != "":
		return getYAMLAuthDataSource(env, logger)

	// This should never happen.
	default:
		return nil, nil, fmt.Errorf("neither POSTGRES_CONNECTION_STRING nor YAML_FILEPATH is set")
	}
}

// getPostgresAuthDataSource initializes a Postgres data source and returns it along with a cleanup function.
// The cleanup function must be invoked by the caller to ensure resources are released.
func getPostgresAuthDataSource(env envVars, logger polylog.Logger) (grpc_server.AuthDataSource, func(), error) {
	logger.Info().Msg("Using Postgres data source")

	authDataSource, cleanup, err := postgres.NewPostgresDataSource(
		context.Background(),
		env.postgresConnectionString,
		logger,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create Postgres data source: %v", err)
	}

	return authDataSource, cleanup, nil
}

// getYAMLAuthDataSource initializes a YAML data source and returns it.
func getYAMLAuthDataSource(env envVars, logger polylog.Logger) (grpc_server.AuthDataSource, func(), error) {
	logger.Info().Msg("Using YAML data source")

	authDataSource, err := yaml.NewYAMLDataSource(env.yamlFilepath, logger)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create YAML data source: %v", err)
	}

	cleanup := func() {} // cleanup is a no-op for YAML data source

	return authDataSource, cleanup, nil
}
