package connector

import (
	"context"
	"fmt"

	"github.com/conductorone/baton-aruba-central/pkg/arubacentral"
	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/connectorbuilder"
	ent "github.com/conductorone/baton-sdk/pkg/types/entitlement"
	"github.com/conductorone/baton-sdk/pkg/types/grant"
	rs "github.com/conductorone/baton-sdk/pkg/types/resource"
)

const GroupMembershipEntitlement = "member"

type groupBuilder struct {
	client       *arubacentral.Client
	resourceType *v2.ResourceType
}

var _ connectorbuilder.ResourceSyncerV2 = &groupBuilder{}

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

func (g *groupBuilder) List(ctx context.Context, _ *v2.ResourceId, opts rs.SyncOpAttrs) ([]*v2.Resource, *rs.SyncOpResults, error) {
	annos := annotations.New()
	bag, offset, err := parsePageToken(opts.PageToken.Token, &v2.ResourceId{ResourceType: g.resourceType.Id})
	if err != nil {
		return nil, &rs.SyncOpResults{Annotations: annos}, fmt.Errorf("failed to parse page token: %w", err)
	}

	pgVars := arubacentral.NewPaginationVars(ResourcesPageSize, offset)
	groups, total, rl, err := g.client.ListGroups(ctx, pgVars)
	annos.Update(rl)
	if err != nil {
		return nil, &rs.SyncOpResults{Annotations: annos}, fmt.Errorf("failed to list groups: %w", err)
	}

	var rv []*v2.Resource
	for _, group := range groups {
		ur, err := groupResource(group)
		if err != nil {
			return nil, &rs.SyncOpResults{Annotations: annos}, fmt.Errorf("failed to create group resource: %w", err)
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

func (g *groupBuilder) Entitlements(_ context.Context, resource *v2.Resource, opts rs.SyncOpAttrs) ([]*v2.Entitlement, *rs.SyncOpResults, error) {
	var rv []*v2.Entitlement

	assignmentOptions := []ent.EntitlementOption{
		ent.WithGrantableTo(userResourceType),
		ent.WithDisplayName(fmt.Sprintf("%s %s", resource.DisplayName, GroupMembershipEntitlement)),
		ent.WithDescription(fmt.Sprintf("%s group %s in Aruba Central", resource.DisplayName, GroupMembershipEntitlement)),
	}

	rv = append(rv, ent.NewAssignmentEntitlement(resource, GroupMembershipEntitlement, assignmentOptions...))

	return rv, &rs.SyncOpResults{}, nil
}

func (g *groupBuilder) Grants(ctx context.Context, resource *v2.Resource, opts rs.SyncOpAttrs) ([]*v2.Grant, *rs.SyncOpResults, error) {
	annos := annotations.New()
	bag, offset, err := parsePageToken(opts.PageToken.Token, &v2.ResourceId{ResourceType: userResourceType.Id})
	if err != nil {
		return nil, &rs.SyncOpResults{Annotations: annos}, fmt.Errorf("failed to parse page token: %w", err)
	}

	pgVars := arubacentral.NewPaginationVars(ResourcesPageSize, offset)
	users, total, rl, err := g.client.ListUsers(ctx, pgVars)
	annos.Update(rl)
	if err != nil {
		return nil, &rs.SyncOpResults{Annotations: annos}, fmt.Errorf("failed to list users: %w", err)
	}

	var rv []*v2.Grant
	for _, user := range users {
		if !user.ContainsGroup(resource.Id.Resource) {
			continue
		}

		uID, err := rs.NewResourceID(userResourceType, user.Username)
		if err != nil {
			return nil, &rs.SyncOpResults{Annotations: annos}, fmt.Errorf("failed to create user resource id: %w", err)
		}

		rv = append(rv, grant.NewGrant(resource, GroupMembershipEntitlement, uID))
	}

	nextPage := prepareNextToken(offset, total)
	next, err := bag.NextToken(nextPage)
	if err != nil {
		return nil, &rs.SyncOpResults{Annotations: annos}, fmt.Errorf("failed to prepare next page token: %w", err)
	}

	return rv, &rs.SyncOpResults{Annotations: annos, NextPageToken: next}, nil
}

func newGroupBuilder(client *arubacentral.Client) *groupBuilder {
	return &groupBuilder{
		client:       client,
		resourceType: groupResourceType,
	}
}
