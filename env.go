package main

import (
	"fmt"
	"os"
)

// This file handles loading all environment variables for the PATH Auth Data Server.

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
