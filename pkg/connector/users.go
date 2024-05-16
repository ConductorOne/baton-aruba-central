package connector

import (
	"context"
	"fmt"

	"github.com/conductorone/baton-aruba-central/pkg/arubacentral"
	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	rs "github.com/conductorone/baton-sdk/pkg/types/resource"
)

type userBuilder struct {
	client       *arubacentral.Client
	resourceType *v2.ResourceType
}

func userResource(user *arubacentral.User) (*v2.Resource, error) {
	profile := map[string]interface{}{
		"login":      user.Username,
		"first_name": user.Name.First,
		"last_name":  user.Name.Last,
	}

	fullName := user.Name.First + " " + user.Name.Last
	resource, err := rs.NewUserResource(
		fullName,
		userResourceType,
		user.Username,
		[]rs.UserTraitOption{
			rs.WithUserProfile(profile),
		},
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
func (u *userBuilder) List(ctx context.Context, _ *v2.ResourceId, pToken *pagination.Token) ([]*v2.Resource, string, annotations.Annotations, error) {
	bag, offset, err := parsePageToken(pToken.Token, &v2.ResourceId{ResourceType: u.resourceType.Id})
	if err != nil {
		return nil, "", nil, fmt.Errorf("failed to parse page token: %w", err)
	}

	pgVars := arubacentral.NewPaginationVars(ResourcesPageSize, offset)
	users, total, rl, err := u.client.ListUsers(ctx, pgVars)
	if err != nil {
		return nil, "", annotations.New(rl), fmt.Errorf("failed to list users: %w", err)
	}

	var rv []*v2.Resource
	for _, user := range users {
		ur, err := userResource(&user) // #nosec G601
		if err != nil {
			return nil, "", nil, fmt.Errorf("failed to create user resource: %w", err)
		}

		rv = append(rv, ur)
	}

	nextPage := prepareNextToken(offset, total)
	next, err := bag.NextToken(nextPage)
	if err != nil {
		return nil, "", nil, fmt.Errorf("failed to prepare next page token: %w", err)
	}

	return rv, next, annotations.New(rl), nil
}

// Entitlements always returns an empty slice for users.
func (u *userBuilder) Entitlements(_ context.Context, resource *v2.Resource, _ *pagination.Token) ([]*v2.Entitlement, string, annotations.Annotations, error) {
	return nil, "", nil, nil
}

// Grants always returns an empty slice for users since they don't have any entitlements.
func (u *userBuilder) Grants(ctx context.Context, resource *v2.Resource, pToken *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	return nil, "", nil, nil
}

func newUserBuilder(client *arubacentral.Client) *userBuilder {
	return &userBuilder{
		client:       client,
		resourceType: userResourceType,
	}
}
