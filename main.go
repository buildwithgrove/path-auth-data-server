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
	"google.golang.org/grpc/reflection"

	"github.com/buildwithgrove/path-auth-data-server/server"
	"github.com/buildwithgrove/path-auth-data-server/yaml"
)

const port = 50051

func main() {
	// Initialize new polylog logger
	logger := polyzero.NewLogger()

	yamlFilePath := os.Getenv(yamlFilePathEnv)
	if yamlFilePath == "" {
		panic(fmt.Errorf("YAML_FILEPATH environment variable is not set"))
	}

	// TODO_NEXT - Postgres data source added in subsequent PR
	// https://github.com/buildwithgrove/path-auth-data-server/pull/3
	dataSource, err := yaml.NewYAMLDataSource(yamlFilePath)
	if err != nil {
		panic(fmt.Errorf("failed to create YAML data source: %v", err))
	}

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

	logger.Info().Int("port", port).Msg("PATH Auth Data Server listening.")
	if err := httpServer.Serve(lis); err != nil {
		panic(fmt.Sprintf("failed to serve: %v", err))
	}
}
