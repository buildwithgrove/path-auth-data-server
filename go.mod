module github.com/buildwithgrove/path-auth-grpc-server

go 1.23.1

require (
	// TODO_IMPROVE - update to tag version once PATH Envoy gRPC branch is merged
	github.com/buildwithgrove/path/envoy/auth_server v0.0.0-20241103145230-011b9a4dbbd6
	github.com/fsnotify/fsnotify v1.7.0
	golang.org/x/net v0.28.0
	google.golang.org/grpc v1.67.1
	google.golang.org/protobuf v1.35.1
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/kr/pretty v0.3.1 // indirect
	github.com/rogpeppe/go-internal v1.12.0 // indirect
	golang.org/x/sys v0.25.0 // indirect
	golang.org/x/text v0.18.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20241007155032-5fefd90f89a9 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
)
