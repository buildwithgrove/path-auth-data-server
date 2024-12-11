package main

import (
	"fmt"
	"os"
)

// This file handles loading all environment variables for the PATH Auth Data Server.

const (
	postgresConnectionStringEnv = "POSTGRES_CONNECTION_STRING"
	yamlFilePathEnv             = "YAML_FILEPATH"

	portEnv     = "PORT"
	defaultPort = "50051"
)

type envVars struct {
	postgresConnectionString string
	yamlFilepath             string
	port                     string
}

func gatherEnvVars() (envVars, error) {
	env := envVars{
		postgresConnectionString: os.Getenv(postgresConnectionStringEnv),
		yamlFilepath:             os.Getenv(yamlFilePathEnv),
		port:                     os.Getenv(portEnv),
	}
	return env, env.validateAndHydrate()
}

// validateAndHydrate validates the required environment variables are set,
// confirms that only one data source will be used,
// and hydrates defaults for any optional values that are not set.
func (env *envVars) validateAndHydrate() error {
	if env.postgresConnectionString == "" && env.yamlFilepath == "" {
		return fmt.Errorf("neither %s nor %s is set", postgresConnectionStringEnv, yamlFilePathEnv)
	}
	if env.postgresConnectionString != "" && env.yamlFilepath != "" {
		return fmt.Errorf("only one of %s and %s can be set", postgresConnectionStringEnv, yamlFilePathEnv)
	}
	if env.port == "" {
		env.port = defaultPort
	}
	return nil
}
