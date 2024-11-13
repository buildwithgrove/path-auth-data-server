package main

import (
	"fmt"
	"net"
	"net/http"

	"github.com/buildwithgrove/path/envoy/auth_server/proto"
	"github.com/pokt-network/poktroll/pkg/polylog/polyzero"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	grpc_server "github.com/buildwithgrove/path-auth-data-server/grpc"
)

const port = 50051

func main() {
	// Initialize new polylog logger
	logger := polyzero.NewLogger()

	// TODO_UPNEXT(@commoddity): Add implementations for concrete data sources: YAML(#2) && Postgres(#3)
	var authDataSource grpc_server.AuthDataSource

	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		panic(fmt.Sprintf("failed to listen: %v", err))
	}

	server, err := grpc_server.NewGRPCServer(authDataSource, logger)
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
		if _, err := w.Write([]byte("OK")); err != nil {
			logger.Error().Err(err).Msg("failed to write health check response")
		}
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

	logger.Info().Int("port", port).Msg("PATH Auth Dataserver listening.")
	if err := httpServer.Serve(ln); err != nil {
		panic(fmt.Sprintf("failed to serve: %v", err))
	}
}
