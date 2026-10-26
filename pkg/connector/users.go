package connector

import (
	"context"
	"fmt"

	"github.com/conductorone/baton-aruba-central/pkg/arubacentral"
	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/connectorbuilder"
	rs "github.com/conductorone/baton-sdk/pkg/types/resource"
)

type userBuilder struct {
	client       *arubacentral.Client
	resourceType *v2.ResourceType
}

var _ connectorbuilder.ResourceSyncerV2 = &userBuilder{}

func userResource(user *arubacentral.User) (*v2.Resource, error) {
	profile := map[string]any{
		"login":      user.Username,
		"first_name": user.Name.First,
		"last_name":  user.Name.Last,
	}

	traitOpts := []rs.UserTraitOption{}
	if user.Username != "" {
		traitOpts = append(traitOpts, rs.WithUserLogin(user.Username))
	}

	fullName := user.Name.First + " " + user.Name.Last
	resource, err := rs.NewUserResource(
		fullName,
		userResourceType,
		user.Username,
		traitOpts,
		rs.WithResourceProfile(profile),
	)
	if err != nil {
		return nil, err
	}

	return resource, nil
}

func (u *userBuilder) ResourceType(ctx context.Context) *v2.ResourceType {
	return userResourceType
}

// List returns all the users from the database as resource objects.
// Users include a UserTrait because they are the 'shape' of a standard user.
func (u *userBuilder) List(ctx context.Context, _ *v2.ResourceId, opts rs.SyncOpAttrs) ([]*v2.Resource, *rs.SyncOpResults, error) {
	annos := annotations.New()
	bag, offset, err := parsePageToken(opts.PageToken.Token, &v2.ResourceId{ResourceType: u.resourceType.Id})
	if err != nil {
		return nil, &rs.SyncOpResults{Annotations: annos}, fmt.Errorf("failed to parse page token: %w", err)
	}

	pgVars := arubacentral.NewPaginationVars(ResourcesPageSize, offset)
	users, total, rl, err := u.client.ListUsers(ctx, pgVars)
	annos.Update(rl)
	if err != nil {
		return nil, &rs.SyncOpResults{Annotations: annos}, fmt.Errorf("failed to list users: %w", err)
	}

	var rv []*v2.Resource
	for _, user := range users {
		ur, err := userResource(&user) // #nosec G601
		if err != nil {
			return nil, &rs.SyncOpResults{Annotations: annos}, fmt.Errorf("failed to create user resource: %w", err)
		}

		rv = append(rv, ur)
	}

	nextPage := prepareNextToken(offset, total)
	next, err := bag.NextToken(nextPage)
	if err != nil {
		return nil, &rs.SyncOpResults{Annotations: annos}, fmt.Errorf("failed to prepare next page token: %w", err)
	}

	return rv, &rs.SyncOpResults{Annotations: annos, NextPageToken: next}, nil
}

// Entitlements always returns an empty slice for users.
func (u *userBuilder) Entitlements(_ context.Context, resource *v2.Resource, opts rs.SyncOpAttrs) ([]*v2.Entitlement, *rs.SyncOpResults, error) {
	return nil, &rs.SyncOpResults{}, nil
}

// Grants always returns an empty slice for users since they don't have any entitlements.
func (u *userBuilder) Grants(ctx context.Context, resource *v2.Resource, opts rs.SyncOpAttrs) ([]*v2.Grant, *rs.SyncOpResults, error) {
	return nil, &rs.SyncOpResults{}, nil
}

func newUserBuilder(client *arubacentral.Client) *userBuilder {
	return &userBuilder{
		client:       client,
		resourceType: userResourceType,
	}
}
