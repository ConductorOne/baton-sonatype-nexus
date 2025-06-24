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
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
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

func (o *roleBuilder) Grant(ctx context.Context, principal *v2.Resource, entitlement *v2.Entitlement) (annotations.Annotations, error) {
	l := ctxzap.Extract(ctx)

	if principal.Id.ResourceType != userResourceType.Id {
		l.Warn(
			"baton-sonatype-nexus: only users can be granted roles",
			zap.String("principal_type", principal.Id.ResourceType),
			zap.String("principal_id", principal.Id.Resource),
		)
		return nil, fmt.Errorf("baton-sonatype-nexus: only users can be granted roles")
	}

	userId := principal.Id.Resource
	roleId := entitlement.Resource.Id.Resource

	users, _, err := o.client.ListUsersByID(ctx, userId)
	if err != nil {
		return nil, fmt.Errorf("failed to list users by id: %w", err)
	}

	var targetUser *client.User
	for _, u := range users {
		if u.UserID == userId {
			targetUser = u
			break
		}
	}

	if targetUser == nil {
		return nil, fmt.Errorf("user %s not found", userId)
	}

	// Check if the user already has the role
	for _, existingRole := range targetUser.Roles {
		if existingRole == roleId {
			return annotations.New(&v2.GrantAlreadyExists{}), nil
		}
	}

	// Add the new role to the user's roles
	targetUser.Roles = append(targetUser.Roles, roleId)

	// Update the user in Nexus
	_, err = o.client.UpdateUser(ctx, userId, targetUser)
	if err != nil {
		return nil, fmt.Errorf("failed to update user roles: %w", err)
	}

	return nil, nil
}

func (o *roleBuilder) Revoke(ctx context.Context, grant *v2.Grant) (annotations.Annotations, error) {
	userId := grant.Principal.Id.Resource
	roleId := grant.Entitlement.Resource.Id.Resource

	users, _, err := o.client.ListUsersByID(ctx, userId)
	if err != nil {
		return nil, fmt.Errorf("failed to list users by id: %w", err)
	}

	var targetUser *client.User
	for _, u := range users {
		if u.UserID == userId {
			targetUser = u
			break
		}
	}

	if targetUser == nil {
		return nil, fmt.Errorf("user %s not found", userId)
	}

	found := false
	newRoles := make([]string, 0, len(targetUser.Roles))
	for _, r := range targetUser.Roles {
		if r == roleId {
			found = true
			continue // we don't add it, so it's removed.
		}
		newRoles = append(newRoles, r)
	}

	if !found {
		return annotations.New(&v2.GrantAlreadyRevoked{}), nil
	}

	targetUser.Roles = newRoles

	_, err = o.client.UpdateUser(ctx, userId, targetUser)
	if err != nil {
		return nil, fmt.Errorf("failed to update user roles: %w", err)
	}

	return nil, nil
}
