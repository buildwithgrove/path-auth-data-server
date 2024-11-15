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

func main() {
	// Initialize new polylog logger
	logger := polyzero.NewLogger()

	env, err := gatherEnvVars()
	if err != nil {
		panic(fmt.Errorf("failed to gather environment variables: %v", err))
	}

	// 1. Load the YAML data source
	// TODO_UPNEXT(@commoddity): Add implementation for concrete data sources: Postgres(#3)
	authDataSource, err := yaml.NewYAMLDataSource(env.yamlFilepath)
	if err != nil {
		panic(fmt.Errorf("failed to create YAML data source: %v", err))
	}

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

/* ---------------------- Environment Variables ---------------------- */

const (
	yamlFilePathEnv = "YAML_FILEPATH"

	portEnv     = "PORT"
	defaultPort = "50051"
)

type envVars struct {
	yamlFilepath string
	port         string
}

func gatherEnvVars() (envVars, error) {
	env := envVars{
		yamlFilepath: os.Getenv(yamlFilePathEnv),
		port:         os.Getenv(portEnv),
	}
	return env, env.validateAndHydrate()
}

// validateAndHydrate validates the required environment variables are set
// and hydrates defaults for any optional values that are not set.
func (env *envVars) validateAndHydrate() error {
	if env.yamlFilepath == "" {
		return fmt.Errorf("%s is not set", yamlFilePathEnv)
	}
	if env.port == "" {
		env.port = defaultPort
	}
	return nil
}
