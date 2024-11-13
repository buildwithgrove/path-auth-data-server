package main

import (
	"fmt"
	"net"
	"net/http"
	"os"

	"github.com/buildwithgrove/path/envoy/auth_server/proto"
	"github.com/pokt-network/poktroll/pkg/polylog/polyzero"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"

	grpc_server "github.com/buildwithgrove/path-auth-data-server/grpc"
	"github.com/buildwithgrove/path-auth-data-server/yaml"

	_ "github.com/joho/godotenv/autoload"
)

const (
	yamlFilePathEnv = "YAML_FILEPATH"
	portEnv         = "PORT"

	defaultPort = "50051"
)

type envVars struct {
	yamlFilepath string
	port         string
}

func gatherEnvVars() (envVars, error) {
	yamlFilepath := os.Getenv(yamlFilePathEnv)
	if yamlFilepath == "" {
		return envVars{}, fmt.Errorf("YAML_FILEPATH environment variable is not set")
	}

	port := os.Getenv(portEnv)
	if port == "" {
		port = defaultPort
	}

	return envVars{
		yamlFilepath: yamlFilepath,
		port:         port,
	}, nil
}

func main() {
	// Initialize new polylog logger
	logger := polyzero.NewLogger()

	env, err := gatherEnvVars()
	if err != nil {
		panic(fmt.Errorf("failed to gather environment variables: %v", err))
	}

	// TODO_UPNEXT(@commoddity): Add implementation for concrete data sources: Postgres(#3)
	authDataSource, err := yaml.NewYAMLDataSource(env.yamlFilepath)
	if err != nil {
		panic(fmt.Errorf("failed to create YAML data source: %v", err))
	}

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

	// Create a new HTTP server mux to allow health checks
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("OK")); err != nil {
			logger.Error().Err(err).Msg("failed to write health check response")
		}
	})

	// Create a new HTTP handler that serves both gRPC and HTTP
	grpcAndHTTPHandler := h2c.NewHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if grpc_server.IsRequestGRPC(r) {
			grpcServer.ServeHTTP(w, r)
		} else {
			mux.ServeHTTP(w, r)
		}
	}), &http2.Server{})

	// Create a new HTTP server that uses the gRPC and HTTP handler
	httpServer := &http.Server{Handler: grpcAndHTTPHandler}

	logger.Info().Str("port", env.port).Msg("PATH Auth Data Server listening.")

	if err := httpServer.Serve(ln); err != nil {
		panic(fmt.Sprintf("failed to serve: %v", err))
	}
}
