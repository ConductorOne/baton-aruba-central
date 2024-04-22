package main

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"

	"github.com/conductorone/baton-sdk/pkg/cli"
	"github.com/conductorone/baton-sdk/pkg/connectorbuilder"
	"github.com/conductorone/baton-sdk/pkg/types"
	"github.com/conductorone/baton-sdk/pkg/uhttp"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
	"golang.org/x/oauth2"

	"github.com/conductorone/baton-aruba-central/pkg/connector"
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

// TODO: use SDK helper when available
type OAuth2RefreshToken struct {
	cfg          *oauth2.Config
	accessToken  string
	refreshToken string
}

func NewOAuth2RefreshToken(clientID, clientSecret, redirectURI, tokenURL, accessToken, refreshToken string) *OAuth2RefreshToken {
	return &OAuth2RefreshToken{
		cfg: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			Scopes:       []string{"all"},
			RedirectURL:  redirectURI,
			Endpoint: oauth2.Endpoint{
				TokenURL: tokenURL,
			},
		},
		accessToken:  accessToken,
		refreshToken: refreshToken,
	}
}

func (o *OAuth2RefreshToken) GetClient(ctx context.Context, options ...uhttp.Option) (*http.Client, error) {
	token := &oauth2.Token{
		AccessToken:  o.accessToken,
		RefreshToken: o.refreshToken,
		TokenType:    "Bearer",
	}
	httpClient := o.cfg.Client(ctx, token)

	return httpClient, nil
}

const (
	RedirectURI = "https://arubanetworks.com"
)

func getConnector(ctx context.Context, cfg *config) (types.ConnectorServer, error) {
	l := ctxzap.Extract(ctx)

	tokenURL, err := url.JoinPath(cfg.BaseHost, "/oauth2/token")
	if err != nil {
		return nil, fmt.Errorf("error creating token URL: %w", err)
	}

	credentials := NewOAuth2RefreshToken(
		cfg.ArubaClientID,
		cfg.ArubaClientSecret,
		RedirectURI,
		tokenURL,
		cfg.AccessToken,
		cfg.RefreshToken,
	)
	cb, err := connector.New(ctx, cfg.BaseHost, credentials)
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
