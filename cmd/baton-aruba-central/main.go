package main

import (
	"context"
	"fmt"
	"os"

	"github.com/conductorone/baton-aruba-central/pkg/connector"
	"github.com/conductorone/baton-sdk/pkg/cli"
	"github.com/conductorone/baton-sdk/pkg/connectorbuilder"
	"github.com/conductorone/baton-sdk/pkg/types"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
)

var version = "dev"

func main() {
	ctx := context.Background()

	cfg := &config{}
	cmd, err := cli.NewCmd(ctx, "baton-aruba-central", cfg, validateConfig, getConnector)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	cmd.Version = version
	cmdFlags(cmd)

	err = cmd.Execute()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func getConnector(ctx context.Context, cfg *config) (types.ConnectorServer, error) {
	var oauthConfig connector.OAuthConfig
	var err error

	base := connector.BaseConfig{
		BaseHost:     cfg.BaseHost,
		ClientID:     cfg.ArubaClientID,
		ClientSecret: cfg.ArubaClientSecret,
	}

	switch {
	case cfg.ShouldUseOAuth2CodeFlow():
		oauthConfig = &connector.CodeFlowConfig{
			BaseConfig: base,
			Username:   cfg.Username,
			Password:   cfg.Password,
			CustomerID: cfg.CustomerID,
		}

	case cfg.ShouldUseOAuth2RefreshTokenFlow():
		oauthConfig = &connector.RefreshTokenFlowConfig{
			BaseConfig:   base,
			AccessToken:  cfg.AccessToken,
			RefreshToken: cfg.RefreshToken,
		}

	default:
		oauthConfig = &connector.NoConfig{}
	}

	l := ctxzap.Extract(ctx)
	cb, err := connector.New(ctx, cfg.BaseHost, oauthConfig)
	if err != nil {
		l.Error("error creating connector", zap.Error(err))
		return nil, err
	}

	c, err := connectorbuilder.NewConnector(ctx, cb)
	if err != nil {
		l.Error("error creating connector", zap.Error(err))
		return nil, err
	}

	return c, nil
}
