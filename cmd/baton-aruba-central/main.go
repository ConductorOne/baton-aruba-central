package main

import (
	"context"
	"fmt"
	"net/http"
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
	var httpClient *http.Client
	var err error
	switch {
	case cfg.ShouldUseOAuth2CodeFlow():
		httpClient, err = CodeFlow(ctx, cfg)
		if err != nil {
			return nil, err
		}

	case cfg.ShouldUseOAuth2RefreshTokenFlow():
		httpClient, err = RefreshTokenFlow(ctx, cfg)
		if err != nil {
			return nil, err
		}

	default:
		httpClient = http.DefaultClient
	}

	l := ctxzap.Extract(ctx)
	cb, err := connector.New(ctx, httpClient, cfg.BaseHost)
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
