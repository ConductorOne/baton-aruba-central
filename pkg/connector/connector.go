package connector

import (
	"context"
	"io"

	"github.com/conductorone/baton-aruba-central/pkg/arubacentral"
	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/connectorbuilder"
)

type ArubaCentral struct {
	client *arubacentral.Client
}

// ResourceSyncers returns a ResourceSyncer for each resource type that should be synced from the upstream service.
func (ac *ArubaCentral) ResourceSyncers(ctx context.Context) []connectorbuilder.ResourceSyncer {
	return []connectorbuilder.ResourceSyncer{
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
func New(ctx context.Context, baseHost string, cfg OAuthConfig) (*ArubaCentral, error) {
	httpClient, err := cfg.GetClient(ctx)
	if err != nil {
		return nil, err
	}

	return &ArubaCentral{
		client: arubacentral.NewClient(httpClient, baseHost),
	}, nil
}
