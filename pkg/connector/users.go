package connector

import (
	"context"
	"fmt"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/connectorbuilder"
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

func (b *userBuilder) CreateAccountCapabilityDetails(
	_ context.Context,
) (*v2.CredentialDetailsAccountProvisioning, annotations.Annotations, error) {
	return &v2.CredentialDetailsAccountProvisioning{
		SupportedCredentialOptions: []v2.CapabilityDetailCredentialOption{
			v2.CapabilityDetailCredentialOption_CAPABILITY_DETAIL_CREDENTIAL_OPTION_RANDOM_PASSWORD,
		},
		PreferredCredentialOption: v2.CapabilityDetailCredentialOption_CAPABILITY_DETAIL_CREDENTIAL_OPTION_RANDOM_PASSWORD,
	}, nil, nil
}

func (o *userBuilder) CreateAccount(
	ctx context.Context,
	accountInfo *v2.AccountInfo,
	credentialOptions *v2.CredentialOptions,
) (
	connectorbuilder.CreateAccountResponse,
	[]*v2.PlaintextData,
	annotations.Annotations,
	error,
) {
	profile := accountInfo.GetProfile().AsMap()

	userId, ok := profile["userId"].(string)
	if !ok {
		return nil, nil, nil, fmt.Errorf("missing or invalid 'userId' in profile")
	}

	firstName, ok := profile["firstName"].(string)
	if !ok {
		return nil, nil, nil, fmt.Errorf("missing or invalid 'firstName' in profile")
	}

	lastName, ok := profile["lastName"].(string)
	if !ok {
		return nil, nil, nil, fmt.Errorf("missing or invalid 'lastName' in profile")
	}

	emailAddress, ok := profile["emailAddress"].(string)
	if !ok {
		return nil, nil, nil, fmt.Errorf("missing or invalid 'emailAddress' in profile")
	}

	status, ok := profile["status"].(string)
	if !ok || status == "" {
		status = "active"
	}

	role := "nx-anonymous"

	generatedPassword, err := generateCredentials(credentialOptions)
	if err != nil {
		return nil, nil, nil, err
	}

	payload := &client.UserCreatePayload{
		UserID:       userId,
		FirstName:    firstName,
		LastName:     lastName,
		EmailAddress: emailAddress,
		Password:     generatedPassword,
		Status:       status,
		Roles:        []string{role},
	}

	createdUser, annotations, err := o.client.CreateUser(ctx, payload)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to create user: %w", err)
	}

	resource, err := o.userToResource(createdUser)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to create resource from user: %w", err)
	}

	response := &v2.CreateAccountResponse_SuccessResult{
		Resource:              resource,
		IsCreateAccountResult: true,
	}

	plaintextData := []*v2.PlaintextData{
		{
			Name:        "password",
			Description: "Generated password for the new user",
			Bytes:       []byte(generatedPassword),
		},
	}

	return response, plaintextData, annotations, nil
}

func (o *userBuilder) userToResource(user *client.User) (*v2.Resource, error) {
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
		return nil, fmt.Errorf("error creating user resource: %w", err)
	}

	return userResource, nil
}

func (o *userBuilder) Delete(ctx context.Context, resourceId *v2.ResourceId) (annotations.Annotations, error) {
	if resourceId.ResourceType != userResourceType.Id {
		return nil, fmt.Errorf("baton-sonatype-nexus: non-user resource passed to user delete")
	}
	_, err := o.client.DeleteUser(ctx, resourceId.Resource)
	if err != nil {
		return nil, fmt.Errorf("failed to delete user: %w", err)
	}

	return nil, nil
}
