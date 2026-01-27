package main

import (
	"context"
	"fmt"
	"os"

	"github.com/conductorone/baton-sdk/pkg/config"
	"github.com/conductorone/baton-sdk/pkg/connectorbuilder"
	"github.com/conductorone/baton-sdk/pkg/types"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"

	cfg "github.com/conductorone/baton-aruba-central/pkg/config"
	"github.com/conductorone/baton-aruba-central/pkg/connector"
)

var version = "dev"

func main() {
	ctx := context.Background()

	_, cmd, err := config.DefineConfiguration(
		ctx,
		"baton-aruba-central",
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

func getConnector(ctx context.Context, c *cfg.ArubaCentral) (types.ConnectorServer, error) {
	l := ctxzap.Extract(ctx)

	useCodeFlow := c.Username != "" && c.Password != "" && c.CustomerId != ""
	useRefreshTokenFlow := c.AccessToken != "" && c.RefreshToken != ""

	if !useCodeFlow && !useRefreshTokenFlow {
		return nil, fmt.Errorf("either username, password, and customer-id or access-token and refresh-token are required")
	}

	base := connector.BaseConfig{
		BaseHost:     c.ApiBaseHost,
		ClientID:     c.ArubaCentralClientId,
		ClientSecret: c.ArubaCentralClientSecret,
	}

	var oauthConfig connector.OAuthConfig

	switch {
	case useCodeFlow:
		oauthConfig = &connector.CodeFlowConfig{
			BaseConfig: base,
			Username:   c.Username,
			Password:   c.Password,
			CustomerID: c.CustomerId,
		}

	case useRefreshTokenFlow:
		oauthConfig = &connector.RefreshTokenFlowConfig{
			BaseConfig:   base,
			AccessToken:  c.AccessToken,
			RefreshToken: c.RefreshToken,
		}
	}

	cb, err := connector.New(ctx, c.ApiBaseHost, oauthConfig)
	if err != nil {
		l.Error("error creating connector", zap.Error(err))
		return nil, err
	}

	conn, err := connectorbuilder.NewConnector(ctx, cb)
	if err != nil {
		l.Error("error creating connector", zap.Error(err))
		return nil, err
	}

	return conn, nil
}
