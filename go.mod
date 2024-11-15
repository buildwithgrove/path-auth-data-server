module github.com/buildwithgrove/path-auth-data-server

go 1.23.1

require (
	// TODO_NEXT(@commoddity): Update to a release version of the PATH auth_server package
	// once the `envoy-grpc-auth-service` branch is merged into `main`
	github.com/buildwithgrove/path/envoy/auth_server v0.0.0-20241113085325-36c02a256a51
	github.com/fsnotify/fsnotify v1.7.0
	github.com/joho/godotenv v1.5.1
	github.com/pokt-network/poktroll v0.0.9
	github.com/stretchr/testify v1.9.0
	go.uber.org/mock v0.4.0
	golang.org/x/net v0.28.0
	google.golang.org/grpc v1.67.1
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/rs/zerolog v1.32.0 // indirect
	golang.org/x/sys v0.25.0 // indirect
	golang.org/x/text v0.18.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20241007155032-5fefd90f89a9 // indirect
	google.golang.org/protobuf v1.35.1 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
)
