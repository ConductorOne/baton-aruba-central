package connector

import (
	"context"
	"fmt"
	"strings"

	"github.com/conductorone/baton-aruba-central/pkg/arubacentral"
	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	ent "github.com/conductorone/baton-sdk/pkg/types/entitlement"
	"github.com/conductorone/baton-sdk/pkg/types/grant"
	rs "github.com/conductorone/baton-sdk/pkg/types/resource"
)

const RoleMembershipEntitlement = "member"

type roleBuilder struct {
	client       *arubacentral.Client
	resourceType *v2.ResourceType
}

func roleResource(role *arubacentral.Role) (*v2.Resource, error) {
	profile := map[string]interface{}{
		"role_name":   role.RoleName,
		"no_of_users": role.NoOfUsers,
		"users":       strings.Join(role.Users, ","),
	}

	resource, err := rs.NewRoleResource(
		role.RoleName,
		roleResourceType,
		slugify(role.RoleName),
		[]rs.RoleTraitOption{
			rs.WithRoleProfile(profile),
		},
	)
	if err != nil {
		return nil, err
	}

	return resource, nil
}

func (r *roleBuilder) ResourceType(ctx context.Context) *v2.ResourceType {
	return roleResourceType
}

func (r *roleBuilder) List(ctx context.Context, _ *v2.ResourceId, pToken *pagination.Token) ([]*v2.Resource, string, annotations.Annotations, error) {
	bag, offset, err := parsePageToken(pToken.Token, &v2.ResourceId{ResourceType: r.resourceType.Id})
	if err != nil {
		return nil, "", nil, fmt.Errorf("failed to parse page token: %w", err)
	}

	pgVars := arubacentral.NewPaginationVars(ResourcesPageSize, offset)
	roles, total, rl, err := r.client.ListRoles(ctx, pgVars)
	if err != nil {
		return nil, "", annotations.New(rl), fmt.Errorf("failed to list roles: %w", err)
	}

	var rv []*v2.Resource
	for _, role := range roles {
		resource, err := roleResource(&role) // #nosec G601
		if err != nil {
			return nil, "", nil, fmt.Errorf("failed to create role resource: %w", err)
		}

		rv = append(rv, resource)
	}

	nextToken := prepareNextToken(offset, total)
	next, err := bag.NextToken(nextToken)
	if err != nil {
		return nil, "", nil, fmt.Errorf("failed to create next token: %w", err)
	}

	return rv, next, annotations.New(rl), nil
}

func (r *roleBuilder) Entitlements(ctx context.Context, resource *v2.Resource, _ *pagination.Token) ([]*v2.Entitlement, string, annotations.Annotations, error) {
	var rv []*v2.Entitlement

	// membership entitlements (assignment type) - what users are under this - manual
	assignmentOptions := []ent.EntitlementOption{
		ent.WithGrantableTo(userResourceType),
		ent.WithDisplayName(fmt.Sprintf("%s role membership", resource.DisplayName)),
		ent.WithDescription(fmt.Sprintf("%s role membership in Aruba Central", resource.DisplayName)),
	}

	rv = append(rv, ent.NewAssignmentEntitlement(resource, RoleMembershipEntitlement, assignmentOptions...))

	roleDetail, rl, err := r.client.GetRole(ctx, resource.DisplayName)
	if err != nil {
		return nil, "", annotations.New(rl), fmt.Errorf("failed to get role details: %w", err)
	}

	for _, app := range roleDetail.Applications {
		// create permission entitlement for each app and each module
		permissionOptions := []ent.EntitlementOption{
			ent.WithGrantableTo(userResourceType),
			ent.WithDisplayName(fmt.Sprintf("%s %s app permissions", app.Name, app.Permission)),
			ent.WithDescription(fmt.Sprintf("%s %s app permissions in Aruba Central", app.Name, app.Permission)),
		}

		appEntitlementName := fmt.Sprintf("%s-%s", app.Name, app.Permission)

		rv = append(rv, ent.NewPermissionEntitlement(resource, appEntitlementName, permissionOptions...))

		// create permission entitlement for each module
		for _, module := range app.Modules {
			permissionOptions := []ent.EntitlementOption{
				ent.WithGrantableTo(userResourceType),
				ent.WithDisplayName(fmt.Sprintf("%s %s %s module permissions", app.Name, module.Name, module.Permission)),
				ent.WithDescription(fmt.Sprintf("%s %s %s module permissions in Aruba Central", app.Name, module.Name, module.Permission)),
			}

			moduleEntitlementName := fmt.Sprintf("%s-%s", appEntitlementName, module.Name)

			rv = append(rv, ent.NewPermissionEntitlement(resource, moduleEntitlementName, permissionOptions...))
		}
	}

	return rv, "", annotations.New(rl), nil
}

func (r *roleBuilder) Grants(ctx context.Context, resource *v2.Resource, _ *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	roleDetail, rl, err := r.client.GetRole(ctx, resource.DisplayName)
	if err != nil {
		return nil, "", annotations.New(rl), fmt.Errorf("failed to get role details: %w", err)
	}

	if roleDetail.NoOfUsers == 0 {
		return nil, "", annotations.New(rl), nil
	}

	var rv []*v2.Grant
	for _, userID := range roleDetail.Users {
		uID, err := rs.NewResourceID(userResourceType, userID)
		if err != nil {
			return nil, "", nil, fmt.Errorf("failed to create user resource id: %w", err)
		}

		// membership grants
		rv = append(rv, grant.NewGrant(resource, RoleMembershipEntitlement, uID))

		// permission grants
		for _, app := range roleDetail.Applications {
			appEntitlementName := fmt.Sprintf("%s-%s", app.Name, app.Permission)
			rv = append(rv, grant.NewGrant(resource, appEntitlementName, uID))

			for _, module := range app.Modules {
				moduleEntitlementName := fmt.Sprintf("%s-%s", appEntitlementName, module.Name)
				rv = append(rv, grant.NewGrant(resource, moduleEntitlementName, uID))
			}
		}
	}

	return rv, "", annotations.New(rl), nil
}

func newRoleBuilder(client *arubacentral.Client) *roleBuilder {
	return &roleBuilder{
		client:       client,
		resourceType: roleResourceType,
	}
}
