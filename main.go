package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"

	"github.com/buildwithgrove/path/envoy/auth_server/proto"
	"github.com/pokt-network/poktroll/pkg/polylog"
	"github.com/pokt-network/poktroll/pkg/polylog/polyzero"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/buildwithgrove/path-auth-data-server/postgres"
	"github.com/buildwithgrove/path-auth-data-server/server"
	"github.com/buildwithgrove/path-auth-data-server/yaml"

	_ "github.com/joho/godotenv/autoload" // Autoload env vars
)

const (
	port                        = 50051
	postgresConnectionStringEnv = "POSTGRES_CONNECTION_STRING"
	yamlFilePathEnv             = "YAML_FILEPATH"
)

func main() {
	// Initialize new polylog logger
	logger := polyzero.NewLogger()

	// Load the data source from either:
	// 1. a Postgres database
	// 2. a YAML file
	dataSource, cleanup, err := getDataSource(logger)
	if err != nil {
		panic(err)
	}
	defer cleanup()

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		panic(fmt.Sprintf("failed to listen: %v", err))
	}

	server, err := server.NewServer(dataSource)
	if err != nil {
		panic(fmt.Sprintf("failed to create server: %v", err))
	}

	grpcServer := grpc.NewServer()
	proto.RegisterGatewayEndpointsServer(grpcServer, server)

	// Enable gRPC reflection
	reflection.Register(grpcServer)

	// Create a new HTTP server mux to allow health checks
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Create a new HTTP handler that serves both gRPC and HTTP
	grpcAndHTTPHandler := h2c.NewHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.ProtoMajor == 2 && r.Header.Get("Content-Type") == "application/grpc" {
			grpcServer.ServeHTTP(w, r)
		} else {
			mux.ServeHTTP(w, r)
		}
	}), &http2.Server{})

	// Create a new HTTP server that uses the gRPC and HTTP handler
	httpServer := &http.Server{Handler: grpcAndHTTPHandler}

	logger.Info().Int("port", port).Msg("PATH Auth gRPC server listening.")
	if err := httpServer.Serve(lis); err != nil {
		panic(fmt.Sprintf("failed to serve: %v", err))
	}
}

// getDataSource returns a DataSource and a cleanup function.
// The cleanup function should be deferred to ensure resources are released.
//
// The specific data source loaded depends on the environment variables:
// - POSTGRES_CONNECTION_STRING - use a Postgres database as the data source
// - YAML_FILEPATH - use a local YAML file as the data source
func getDataSource(logger polylog.Logger) (server.DataSource, func(), error) {
	postgresConnectionString := os.Getenv(postgresConnectionStringEnv)
	yamlFilePath := os.Getenv(yamlFilePathEnv)

	if postgresConnectionString != "" && yamlFilePath != "" {
		return nil, nil, fmt.Errorf("only one of POSTGRES_CONNECTION_STRING and YAML_FILEPATH can be set")
	}

	if postgresConnectionString != "" {
		logger.Info().Str(postgresConnectionStringEnv, postgresConnectionString).Msg("Using Postgres data source")

		dataSource, cleanup, err := postgres.NewPostgresDataSource(
			context.Background(),
			postgresConnectionString,
			logger,
		)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create Postgres data source: %v", err)
		}
		return dataSource, cleanup, nil
	}

	if yamlFilePath != "" {
		logger.Info().Str(yamlFilePathEnv, yamlFilePath).Msg("Using YAML data source")

		dataSource, err := yaml.NewYAMLDataSource(yamlFilePath)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create YAML data source: %v", err)
		}
		return dataSource, func() {}, nil
	}

	return nil, nil, fmt.Errorf("neither POSTGRES_CONNECTION_STRING nor YAML_FILEPATH is set")
}
