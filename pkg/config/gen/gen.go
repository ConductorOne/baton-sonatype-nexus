package main

import (
	"github.com/conductorone/baton-sdk/pkg/config"
	cfg "github.com/conductorone/baton-sonatype-nexus/pkg/config"
)

func main() {
	config.Generate("sonatype-nexus", cfg.Config)
}
