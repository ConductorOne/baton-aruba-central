package connector

import (
	"context"
	"fmt"

	"github.com/conductorone/baton-aruba-central/pkg/arubacentral"
	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	ent "github.com/conductorone/baton-sdk/pkg/types/entitlement"
	"github.com/conductorone/baton-sdk/pkg/types/grant"
	rs "github.com/conductorone/baton-sdk/pkg/types/resource"
)

const GroupMembershipEntitlement = "member"

type groupBuilder struct {
	client       *arubacentral.Client
	resourceType *v2.ResourceType
}

func groupResource(group string) (*v2.Resource, error) {
	resource, err := rs.NewGroupResource(
		group,
		groupResourceType,
		group,
		nil,
	)
	if err != nil {
		return nil, err
	}

	return resource, nil
}

func (g *groupBuilder) ResourceType(ctx context.Context) *v2.ResourceType {
	return groupResourceType
}

func (g *groupBuilder) List(ctx context.Context, _ *v2.ResourceId, pToken *pagination.Token) ([]*v2.Resource, string, annotations.Annotations, error) {
	bag, offset, err := parsePageToken(pToken.Token, &v2.ResourceId{ResourceType: g.resourceType.Id})
	if err != nil {
		return nil, "", nil, fmt.Errorf("failed to parse page token: %w", err)
	}

	pgVars := arubacentral.NewPaginationVars(ResourcesPageSize, offset)
	groups, total, rl, err := g.client.ListGroups(ctx, pgVars)
	if err != nil {
		return nil, "", annotations.New(rl), fmt.Errorf("failed to list groups: %w", err)
	}

	var rv []*v2.Resource
	for _, group := range groups {
		ur, err := groupResource(group)
		if err != nil {
			return nil, "", nil, fmt.Errorf("failed to create group resource: %w", err)
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

func (g *groupBuilder) Entitlements(_ context.Context, resource *v2.Resource, _ *pagination.Token) ([]*v2.Entitlement, string, annotations.Annotations, error) {
	var rv []*v2.Entitlement

	assignmentOptions := []ent.EntitlementOption{
		ent.WithGrantableTo(userResourceType),
		ent.WithDisplayName(fmt.Sprintf("%s %s", resource.DisplayName, GroupMembershipEntitlement)),
		ent.WithDescription(fmt.Sprintf("%s group %s in Aruba Central", resource.DisplayName, GroupMembershipEntitlement)),
	}

	rv = append(rv, ent.NewAssignmentEntitlement(resource, GroupMembershipEntitlement, assignmentOptions...))

	return rv, "", nil, nil
}

func (g *groupBuilder) Grants(ctx context.Context, resource *v2.Resource, pToken *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	bag, offset, err := parsePageToken(pToken.Token, &v2.ResourceId{ResourceType: userResourceType.Id})
	if err != nil {
		return nil, "", nil, fmt.Errorf("failed to parse page token: %w", err)
	}

	pgVars := arubacentral.NewPaginationVars(ResourcesPageSize, offset)
	users, total, rl, err := g.client.ListUsers(ctx, pgVars)
	if err != nil {
		return nil, "", annotations.New(rl), fmt.Errorf("failed to list users: %w", err)
	}

	var rv []*v2.Grant
	for _, user := range users {
		if !user.ContainsGroup(resource.Id.Resource) {
			continue
		}

		uID, err := rs.NewResourceID(userResourceType, user.Username)
		if err != nil {
			return nil, "", nil, fmt.Errorf("failed to create user resource id: %w", err)
		}

		rv = append(rv, grant.NewGrant(resource, GroupMembershipEntitlement, uID))
	}

	nextPage := prepareNextToken(offset, total)
	next, err := bag.NextToken(nextPage)
	if err != nil {
		return nil, "", nil, fmt.Errorf("failed to prepare next page token: %w", err)
	}

	return rv, next, annotations.New(rl), nil
}

func newGroupBuilder(client *arubacentral.Client) *groupBuilder {
	return &groupBuilder{
		client:       client,
		resourceType: groupResourceType,
	}
}
