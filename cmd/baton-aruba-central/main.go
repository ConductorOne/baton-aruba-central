package main

import (
	"context"
	"fmt"
	"os"

	"github.com/conductorone/baton-sdk/pkg/config"
	"github.com/conductorone/baton-sdk/pkg/connectorbuilder"
	"github.com/conductorone/baton-sdk/pkg/connectorrunner"
	"github.com/conductorone/baton-sdk/pkg/types"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	arubaconfig "github.com/conductorone/baton-aruba-central/pkg/config"
	"github.com/conductorone/baton-aruba-central/pkg/connector"
)

var version = "dev"

func main() {
	ctx := context.Background()

	_, cmd, err := config.DefineConfiguration(
		ctx,
		"baton-aruba-central",
		getConnector,
		arubaconfig.ConfigurationSchema,
		connectorrunner.WithDefaultCapabilitiesConnectorBuilder(&connector.ArubaCentral{}),
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

func getConnector(ctx context.Context, v *viper.Viper) (types.ConnectorServer, error) {
	l := ctxzap.Extract(ctx)

	baseHost := v.GetString(arubaconfig.BaseHostField.FieldName)
	clientID := v.GetString(arubaconfig.ClientIDField.FieldName)
	clientSecret := v.GetString(arubaconfig.ClientSecretField.FieldName)
	accessToken := v.GetString(arubaconfig.AccessTokenField.FieldName)
	refreshToken := v.GetString(arubaconfig.RefreshTokenField.FieldName)
	username := v.GetString(arubaconfig.UsernameField.FieldName)
	password := v.GetString(arubaconfig.PasswordField.FieldName)
	customerID := v.GetString(arubaconfig.CustomerIDField.FieldName)

	useCodeFlow := username != "" && password != "" && customerID != ""
	useRefreshTokenFlow := accessToken != "" && refreshToken != ""

	if !useCodeFlow && !useRefreshTokenFlow {
		return nil, fmt.Errorf("either username, password, and customer-id or access-token and refresh-token are required")
	}

	base := connector.BaseConfig{
		BaseHost:     baseHost,
		ClientID:     clientID,
		ClientSecret: clientSecret,
	}

	var oauthConfig connector.OAuthConfig

	switch {
	case useCodeFlow:
		oauthConfig = &connector.CodeFlowConfig{
			BaseConfig: base,
			Username:   username,
			Password:   password,
			CustomerID: customerID,
		}

	case useRefreshTokenFlow:
		oauthConfig = &connector.RefreshTokenFlowConfig{
			BaseConfig:   base,
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
		}
	}

	cb, err := connector.New(ctx, baseHost, oauthConfig)
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
