//go:build !generate

package main

import (
	"context"
	"fmt"
	"os"

	"github.com/conductorone/baton-sdk/pkg/config"
	"github.com/conductorone/baton-sdk/pkg/connectorbuilder"
	"github.com/conductorone/baton-sdk/pkg/field"
	"github.com/conductorone/baton-sdk/pkg/types"
	cfg "github.com/conductorone/baton-sonatype-nexus/pkg/config"
	"github.com/conductorone/baton-sonatype-nexus/pkg/connector"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
)

var version = "dev"

func main() {
	ctx := context.Background()

	_, cmd, err := config.DefineConfiguration(
		ctx,
		"baton-sonatype-nexus",
		getConnector,
		cfg.Config,
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	cmd.Version = version

	err = cmd.Execute()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func getConnector(ctx context.Context, ghc *cfg.SonatypeNexus) (types.ConnectorServer, error) {
	l := ctxzap.Extract(ctx)

	if err := field.Validate(cfg.Config, ghc); err != nil {
		l.Error("error validating config", zap.Error(err))
		return nil, err
	}

	host := ghc.GetString(cfg.HostField.FieldName)
	if host == "" {
		l.Error("host is required")
		return nil, fmt.Errorf("host is required")
	}

	username := ghc.GetString(cfg.UsernameField.FieldName)
	if username == "" {
		l.Error("username is required")
		return nil, fmt.Errorf("username is required")
	}

	password := ghc.GetString(cfg.PasswordField.FieldName)
	if password == "" {
		l.Error("password is required")
		return nil, fmt.Errorf("password is required")
	}

	cb, err := connector.New(ctx, host, username, password)
	if err != nil {
		l.Error("error creating connector", zap.Error(err))
		return nil, err
	}

	connector, err := connectorbuilder.NewConnector(ctx, cb)
	if err != nil {
		l.Error("error creating connector", zap.Error(err))
		return nil, err
	}

	return connector, nil
}
