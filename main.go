package main

import (
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/buildwithgrove/path/envoy/auth_server/proto"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/buildwithgrove/path-auth-grpc-server/server"
	"github.com/buildwithgrove/path-auth-grpc-server/yaml"
)

const port = 50051
const endpointsYAMLPath = "gateway-endpoints.yaml"

func main() {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		panic(fmt.Sprintf("failed to listen: %v", err))
	}

	// Server loads GatewayEndpoints from a YAML file as the data source
	// To use a different data source, implement the DataSource interface
	// eg. postgres.NewPostgresDataSource()
	dataSource, err := yaml.NewYAMLDataSource(endpointsYAMLPath)
	if err != nil {
		log.Fatalf("Failed to create data source: %v", err)
	}

	server, err := server.NewServer(dataSource)
	if err != nil {
		panic(fmt.Sprintf("failed to create server: %v", err))
	}

	grpcServer := grpc.NewServer()
	proto.RegisterGatewayEndpointsServer(grpcServer, server)

	// Enable gRPC reflection
	reflection.Register(grpcServer)

	// Create a new HTTP server mux
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Create a new HTTP server that serves both gRPC and HTTP
	httpServer := &http.Server{
		Handler: h2c.NewHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.ProtoMajor == 2 && r.Header.Get("Content-Type") == "application/grpc" {
				grpcServer.ServeHTTP(w, r)
			} else {
				mux.ServeHTTP(w, r)
			}
		}), &http2.Server{}),
	}

	log.Printf("Server listening on port %d", port)
	if err := httpServer.Serve(lis); err != nil {
		panic(fmt.Sprintf("failed to serve: %v", err))
	}
}
