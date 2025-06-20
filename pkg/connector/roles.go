package connector

import (
	"context"
	"fmt"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	"github.com/conductorone/baton-sdk/pkg/types/entitlement"
	"github.com/conductorone/baton-sdk/pkg/types/resource"
	"github.com/conductorone/baton-sonatype-nexus/pkg/client"
)

type roleBuilder struct {
	client *client.APIClient
}

func (o *roleBuilder) ResourceType(ctx context.Context) *v2.ResourceType {
	return roleResourceType
}

func (o *roleBuilder) List(ctx context.Context, parentResourceID *v2.ResourceId, pToken *pagination.Token) ([]*v2.Resource, string, annotations.Annotations, error) {
	roles, annos, err := o.client.ListRoles(ctx)
	if err != nil {
		return nil, "", nil, err
	}

	var resources []*v2.Resource
	for _, role := range roles {
		profile := map[string]interface{}{
			"role_id":     role.ID,
			"source":      role.Source,
			"description": role.Description,
			"name":        role.Name,
		}

		roleTraits := []resource.RoleTraitOption{
			resource.WithRoleProfile(profile),
		}

		roleResource, err := resource.NewRoleResource(
			role.ID,
			roleResourceType,
			role.Name,
			roleTraits,
		)
		if err != nil {
			return nil, "", nil, fmt.Errorf("error creating role resource: %w", err)
		}

		resources = append(resources, roleResource)
	}

	return resources, "", annos, nil
}

func (o *roleBuilder) Entitlements(ctx context.Context, resource *v2.Resource, pToken *pagination.Token) ([]*v2.Entitlement, string, annotations.Annotations, error) {
	var entitlements []*v2.Entitlement

	opts := []entitlement.EntitlementOption{
		entitlement.WithGrantableTo(userResourceType),
		entitlement.WithDescription(fmt.Sprintf("Membership to role: %s", resource.DisplayName)),
		entitlement.WithDisplayName(fmt.Sprintf("Role: %s", resource.DisplayName)),
	}

	entitlements = append(entitlements, entitlement.NewPermissionEntitlement(resource, "assigned", opts...))
	return entitlements, "", nil, nil
}

func (o *roleBuilder) Grants(ctx context.Context, resource *v2.Resource, pToken *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	return nil, "", nil, nil
}

func newRoleBuilder(client *client.APIClient) *roleBuilder {
	return &roleBuilder{
		client: client,
	}
}
