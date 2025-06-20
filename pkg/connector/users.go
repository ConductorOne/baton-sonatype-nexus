package connector

import (
	"context"
	"fmt"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	"github.com/conductorone/baton-sdk/pkg/types/grant"
	"github.com/conductorone/baton-sdk/pkg/types/resource"
	"github.com/conductorone/baton-sonatype-nexus/pkg/client"
)

type userBuilder struct {
	client *client.APIClient
}

func (o *userBuilder) ResourceType(ctx context.Context) *v2.ResourceType {
	return userResourceType
}

// List returns all the users from the database as resource objects.
func (o *userBuilder) List(ctx context.Context, parentResourceID *v2.ResourceId, pToken *pagination.Token) ([]*v2.Resource, string, annotations.Annotations, error) {
	users, annos, err := o.client.ListUsers(ctx)
	if err != nil {
		return nil, "", nil, err
	}

	var resources []*v2.Resource
	for _, user := range users {
		displayName := fmt.Sprintf("%s %s", user.FirstName, user.LastName)

		profile := map[string]interface{}{
			"user_id":    user.UserID,
			"email":      user.EmailAddress,
			"first_name": user.FirstName,
			"last_name":  user.LastName,
			"status":     user.Status,
			"source":     user.Source,
		}

		userTraits := []resource.UserTraitOption{
			resource.WithUserProfile(profile),
			resource.WithEmail(user.EmailAddress, true),
		}

		userResource, err := resource.NewUserResource(
			displayName,
			userResourceType,
			user.UserID,
			userTraits,
		)
		if err != nil {
			return nil, "", nil, fmt.Errorf("error creating user resource: %w", err)
		}

		resources = append(resources, userResource)
	}

	return resources, "", annos, nil
}

// Entitlements always returns an empty slice for users.
func (o *userBuilder) Entitlements(ctx context.Context, resource *v2.Resource, _ *pagination.Token) ([]*v2.Entitlement, string, annotations.Annotations, error) {
	return nil, "", nil, nil
}

// Grants returns the roles assigned to a user as grants.
func (o *userBuilder) Grants(ctx context.Context, res *v2.Resource, _ *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	userID := res.Id.Resource

	users, _, err := o.client.ListUsers(ctx)
	if err != nil {
		return nil, "", nil, err
	}

	roles, _, err := o.client.ListRoles(ctx)
	if err != nil {
		return nil, "", nil, err
	}

	var user *client.User
	for _, u := range users {
		if u.UserID == userID {
			user = u
			break
		}
	}
	if user == nil {
		return nil, "", nil, nil
	}

	var grants []*v2.Grant
	for _, role := range roles {
		for _, userRoleID := range user.Roles {
			if userRoleID == role.ID {
				roleResource := &v2.Resource{
					Id: &v2.ResourceId{
						ResourceType: roleResourceType.Id,
						Resource:     role.ID,
					},
				}
				g := grant.NewGrant(roleResource, "assigned", res)
				grants = append(grants, g)
				break
			}
		}
	}

	return grants, "", nil, nil
}

func newUserBuilder(client *client.APIClient) *userBuilder {
	return &userBuilder{
		client: client,
	}
}
