package connector

import (
	"context"
	"os"
	"testing"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sonatype-nexus/pkg/client"
	"github.com/stretchr/testify/assert"
)

var (
	ctx              = context.Background()
	parentResourceID = &v2.ResourceId{}
)

func initClient(t *testing.T) *client.APIClient {
	host := os.Getenv("NEXUS_HOST")
	username := os.Getenv("NEXUS_USERNAME")
	password := os.Getenv("NEXUS_PASSWORD")

	if host == "" || username == "" || password == "" {
		t.Skipf("Missing required environment variables: NEXUS_HOST, NEXUS_USERNAME, NEXUS_PASSWORD")
	}

	c, err := client.NewClient(ctx, host, username, password, nil)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	return c
}

func TestUserBuilderList(t *testing.T) {
	c := initClient(t)

	u := newUserBuilder(c)

	res, _, _, err := u.List(ctx, parentResourceID, nil)
	assert.Nil(t, err)
	assert.NotNil(t, res)

	t.Logf("Amount of users obtained: %d", len(res))
}

func TestRoleBuilderList(t *testing.T) {
	c := initClient(t)

	r := newRoleBuilder(c)

	res, _, _, err := r.List(ctx, parentResourceID, nil)
	assert.Nil(t, err)
	assert.NotNil(t, res)

	t.Logf("Amount of roles obtained: %d", len(res))
}

func TestUserBuilderGrants(t *testing.T) {
	c := initClient(t)

	u := newUserBuilder(c)

	// First get users
	users, _, _, err := u.List(ctx, parentResourceID, nil)
	assert.Nil(t, err)
	assert.NotNil(t, users)

	if len(users) == 0 {
		t.Skip("No users found to test grants")
	}

	// Test grants for the first user
	user := users[0]
	grants, _, _, err := u.Grants(ctx, user, nil)
	assert.Nil(t, err)
	assert.NotNil(t, grants)

	t.Logf("Amount of grants for user %s: %d", user.DisplayName, len(grants))
}

func TestRoleBuilderEntitlements(t *testing.T) {
	c := initClient(t)

	r := newRoleBuilder(c)

	// First get roles
	roles, _, _, err := r.List(ctx, parentResourceID, nil)
	assert.Nil(t, err)
	assert.NotNil(t, roles)

	if len(roles) == 0 {
		t.Skip("No roles found to test entitlements")
	}

	// Test entitlements for the first role
	role := roles[0]
	entitlements, _, _, err := r.Entitlements(ctx, role, nil)
	assert.Nil(t, err)
	assert.NotNil(t, entitlements)

	t.Logf("Amount of entitlements for role %s: %d", role.DisplayName, len(entitlements))
}
