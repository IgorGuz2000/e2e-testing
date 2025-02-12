// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
// or more contributor license agreements. Licensed under the Elastic License;
// you may not use this file except in compliance with the Elastic License.

package main

import (
	"fmt"

	log "github.com/sirupsen/logrus"
)

// FleetConfig represents the configuration for Fleet Server when building the enrollment command
type FleetConfig struct {
	EnrollmentToken          string
	ElasticsearchPort        int
	ElasticsearchURI         string
	ElasticsearchCredentials string
	KibanaPort               int
	KibanaURI                string
	// server
	BootstrapFleetServer bool
	ServerPolicyID       string
}

// NewFleetConfig builds a new configuration for the fleet agent, defaulting ES credentials, URI and port.
// If the 'bootstrappFleetServer' flag is true, the it will create the config for the initial fleet server
// used to bootstrap Fleet Server
// If the 'fleetServerMode' flag is true, the it will create the config for an agent using an existing Fleet
// Server to connect to Fleet. It will also retrieve the default policy ID for fleet server
func NewFleetConfig(token string, bootstrapFleetServer bool, fleetServerMode bool) (*FleetConfig, error) {
	cfg := &FleetConfig{
		BootstrapFleetServer:     bootstrapFleetServer,
		EnrollmentToken:          token,
		ElasticsearchCredentials: "elastic:changeme",
		ElasticsearchPort:        9200,
		ElasticsearchURI:         "elasticsearch",
		KibanaPort:               5601,
		KibanaURI:                "kibana",
	}

	if fleetServerMode {
		defaultFleetServerPolicy, err := getAgentDefaultPolicy("is_default_fleet_server")
		if err != nil {
			return nil, err
		}

		cfg.ServerPolicyID = defaultFleetServerPolicy.Path("id").Data().(string)

		log.WithFields(log.Fields{
			"elasticsearch":     cfg.ElasticsearchURI,
			"elasticsearchPort": cfg.ElasticsearchPort,
			"policyID":          cfg.ServerPolicyID,
			"token":             cfg.EnrollmentToken,
		}).Debug("Fleet Server config created")
	}

	return cfg, nil
}

func (cfg FleetConfig) flags() []string {
	if cfg.BootstrapFleetServer {
		// TO-DO: remove all code to calculate the fleet-server policy, because it's inferred by the fleet-server
		return []string{
			"--force",
			"--fleet-server-es", fmt.Sprintf("http://%s@%s:%d", cfg.ElasticsearchCredentials, cfg.ElasticsearchURI, cfg.ElasticsearchPort),
		}
	}

	/*
		// agent using an already bootstrapped fleet-server
		fleetServerHost := "https://hostname_of_the_bootstrapped_fleet_server:8220"
		return []string{
			"-e", "-v", "--force", "--insecure",
			// ensure the enrollment belongs to the default policy
			"--enrollment-token=" + cfg.EnrollmentToken,
			"--url", fleetServerHost,
		}
	*/

	baseFlags := []string{"-e", "-v", "--force", "--insecure", "--enrollment-token=" + cfg.EnrollmentToken}

	if cfg.ServerPolicyID != "" {
		baseFlags = append(baseFlags, "--fleet-server-insecure-http", "--fleet-server", fmt.Sprintf("http://%s@%s:%d", cfg.ElasticsearchCredentials, cfg.ElasticsearchURI, cfg.ElasticsearchPort), "--fleet-server-host=http://0.0.0.0", "--fleet-server-policy", cfg.ServerPolicyID)
	}

	return append(baseFlags, "--kibana-url", fmt.Sprintf("http://%s@%s:%d", cfg.ElasticsearchCredentials, cfg.KibanaURI, cfg.KibanaPort))
}

func (fts *FleetTestSuite) anAgentIsDeployedToFleetWithInstallerInFleetMode(image string, installerType string) error {
	fts.ElasticAgentStopped = true
	return fts.anAgentIsDeployedToFleetWithInstallerAndFleetServer(image, installerType, true)
}

// bootstrapFleetServer runs a command for the elastic-agent
func bootstrapFleetServer(profile string, image string, service string, binary string, cfg *FleetConfig) error {
	log.Debug("Bootstrapping Fleet Server")

	args := []string{
		"-f", "--fleet-server-insecure-http",
		"--fleet-server", fmt.Sprintf("http://%s@%s:%d", cfg.ElasticsearchCredentials, cfg.ElasticsearchURI, cfg.ElasticsearchPort),
	}

	err := runElasticAgentCommand(profile, image, service, binary, "install", args)
	if err != nil {
		return fmt.Errorf("Failed to install the agent with subcommand: %v", err)
	}

	return nil
}
