package connector

import (
	"context"
	"fmt"
	"io"

	cfg "github.com/conductorone/baton-aruba-central/pkg/config"

	"github.com/conductorone/baton-aruba-central/pkg/arubacentral"
	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/cli"
	"github.com/conductorone/baton-sdk/pkg/connectorbuilder"
)

type ArubaCentral struct {
	client *arubacentral.Client
}

var _ connectorbuilder.ConnectorBuilderV2 = &ArubaCentral{}

// ResourceSyncers returns a ResourceSyncer for each resource type that should be synced from the upstream service.
func (ac *ArubaCentral) ResourceSyncers(ctx context.Context) []connectorbuilder.ResourceSyncerV2 {
	return []connectorbuilder.ResourceSyncerV2{
		newUserBuilder(ac.client),
		newRoleBuilder(ac.client),
		newGroupBuilder(ac.client),
	}
}

// Asset takes an input AssetRef and attempts to fetch it using the connector's authenticated http client
// It streams a response, always starting with a metadata object, following by chunked payloads for the asset.
func (ac *ArubaCentral) Asset(ctx context.Context, asset *v2.AssetRef) (string, io.ReadCloser, error) {
	return "", nil, nil
}

// Metadata returns metadata about the connector.
func (ac *ArubaCentral) Metadata(ctx context.Context) (*v2.ConnectorMetadata, error) {
	return &v2.ConnectorMetadata{
		DisplayName: "ArubaCentral",
		Description: "Connector syncing ArubaCentral users, roles and groups to Baton",
	}, nil
}

// Validate is called to ensure that the connector is properly configured. It should exercise any API credentials
// to be sure that they are valid.
func (ac *ArubaCentral) Validate(ctx context.Context) (annotations.Annotations, error) {
	pgVars := arubacentral.NewPaginationVars(1, 0)
	_, _, rl, err := ac.client.ListUsers(ctx, pgVars)
	if err != nil {
		return annotations.New(rl), err
	}

	return annotations.New(rl), nil
}

// New returns a new instance of the connector.
func New(ctx context.Context, cfg *cfg.ArubaCentral, opts *cli.ConnectorOpts) (connectorbuilder.ConnectorBuilderV2, []connectorbuilder.Opt, error) {
	useCodeFlow := cfg.Username != "" && cfg.Password != "" && cfg.CustomerId != ""
	useRefreshTokenFlow := cfg.AccessToken != "" && cfg.RefreshToken != ""

	if !useCodeFlow && !useRefreshTokenFlow {
		return nil, nil, fmt.Errorf("either username, password, and customer-id or access-token and refresh-token are required")
	}

	base := BaseConfig{
		BaseHost:     cfg.ApiBaseHost,
		ClientID:     cfg.ArubaCentralClientId,
		ClientSecret: cfg.ArubaCentralClientSecret,
	}

	var oauthConfig OAuthConfig

	switch {
	case useCodeFlow:
		oauthConfig = &CodeFlowConfig{
			BaseConfig: base,
			Username:   cfg.Username,
			Password:   cfg.Password,
			CustomerID: cfg.CustomerId,
		}

	case useRefreshTokenFlow:
		oauthConfig = &RefreshTokenFlowConfig{
			BaseConfig:   base,
			AccessToken:  cfg.AccessToken,
			RefreshToken: cfg.RefreshToken,
		}
	}

	httpClient, err := oauthConfig.GetClient(ctx)
	if err != nil {
		return nil, nil, err
	}

	return &ArubaCentral{
		client: arubacentral.NewClient(httpClient, base.BaseHost),
	}, nil, nil
}
