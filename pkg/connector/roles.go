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
	"google.golang.org/protobuf/types/known/structpb"
)

const RoleMembershipEntitlement = "member"

type roleBuilder struct {
	client       *arubacentral.Client
	resourceType *v2.ResourceType
}

func roleResource(role *arubacentral.Role) (*v2.Resource, error) {
	var users structpb.ListValue
	for _, user := range role.Users {
		u, err := user.Marshall()
		if err != nil {
			return nil, fmt.Errorf("failed to marshal user: %w", err)
		}

		users.Values = append(users.Values, u)
	}

	var apps structpb.ListValue
	for _, app := range role.Applications {
		a, err := app.Marshall()
		if err != nil {
			return nil, fmt.Errorf("failed to marshal application: %w", err)
		}

		apps.Values = append(apps.Values, a)
	}

	profile := map[string]interface{}{
		"role_name":    role.RoleName,
		"no_of_users":  role.NoOfUsers,
		"users":        &users,
		"applications": &apps,
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

	// permission entitlements (permission type) - what permissions are granted by this - dynamic based on the role
	roleTrait, err := rs.GetRoleTrait(resource)
	if err != nil {
		return nil, "", nil, fmt.Errorf("failed to get role trait: %w", err)
	}

	apps, ok := getValuesFromProfile(roleTrait.Profile, "applications", &arubacentral.Application{})
	if !ok {
		return nil, "", nil, fmt.Errorf("failed to get applications from role profile")
	}

	for _, app := range apps {
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

	return rv, "", nil, nil
}

func (r *roleBuilder) Grants(ctx context.Context, resource *v2.Resource, _ *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	roleTrait, err := rs.GetRoleTrait(resource)
	if err != nil {
		return nil, "", nil, fmt.Errorf("failed to get role trait: %w", err)
	}

	userIDs, ok := getValuesFromProfile(roleTrait.Profile, "users", new(arubacentral.UserString))
	if !ok {
		return nil, "", nil, fmt.Errorf("failed to get users from role profile")
	}

	if len(userIDs) == 0 {
		return nil, "", nil, nil
	}

	apps, ok := getValuesFromProfile(roleTrait.Profile, "applications", &arubacentral.Application{})
	if !ok {
		return nil, "", nil, fmt.Errorf("failed to get applications from role profile")
	}

	var rv []*v2.Grant
	for _, userID := range userIDs {
		uID, err := rs.NewResourceID(userResourceType, string(*userID))
		if err != nil {
			return nil, "", nil, fmt.Errorf("failed to create user resource id: %w", err)
		}

		// membership grants
		rv = append(rv, grant.NewGrant(resource, RoleMembershipEntitlement, uID))

		// permission grants
		for _, app := range apps {
			appEntitlementName := fmt.Sprintf("%s-%s", app.Name, app.Permission)
			rv = append(rv, grant.NewGrant(resource, appEntitlementName, uID))

			for _, module := range app.Modules {
				moduleEntitlementName := fmt.Sprintf("%s-%s", appEntitlementName, module.Name)
				rv = append(rv, grant.NewGrant(resource, moduleEntitlementName, uID))
			}
		}
	}

	return rv, "", nil, nil
}

func newRoleBuilder(client *arubacentral.Client) *roleBuilder {
	return &roleBuilder{
		client:       client,
		resourceType: roleResourceType,
	}
}
