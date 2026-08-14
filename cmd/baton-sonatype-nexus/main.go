package main

import (
	"context"

	"github.com/conductorone/baton-sdk/pkg/config"
	"github.com/conductorone/baton-sdk/pkg/connectorrunner"
	cfg "github.com/conductorone/baton-sonatype-nexus/pkg/config"
	"github.com/conductorone/baton-sonatype-nexus/pkg/connector"
)

var version = "dev"

func main() {
	ctx := context.Background()
	config.RunConnector(
		ctx,
		"baton-sonatype-nexus",
		version,
		cfg.Configuration,
		connector.New,
		connectorrunner.WithProvisioningEnabled(),
		connectorrunner.WithDefaultCapabilitiesConnectorBuilderV2(&connector.Connector{}),
	)
}
