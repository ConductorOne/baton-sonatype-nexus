package main

import (
	cfg "github.com/conductorone/baton-sonatype-nexus/pkg/config"
	"github.com/conductorone/baton-sdk/pkg/config"
)

func main() {
	config.Generate("sonatype-nexus", cfg.Config)
}
