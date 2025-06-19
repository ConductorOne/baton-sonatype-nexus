package connector

import (
	"context"
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/conductorone/baton-sdk/pkg/uhttp"
	"github.com/conductorone/baton-sonatype-nexus/pkg/client"
	"github.com/conductorone/baton-sonatype-nexus/pkg/test"
)

// Tests that the client can fetch users.
func TestNexusClient_GetUsers(t *testing.T) {
	body, err := test.ReadFile("usersMock.json")
	if err != nil {
		t.Fatalf("Error reading body: %s", err)
	}
	mockResponse := &http.Response{
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(body)),
	}
	mockResponse.Header.Set("Content-Type", "application/json")

	testClient := test.NewTestClient(mockResponse, nil)

	ctx := context.Background()

	result, _, err := testClient.ListUsers(ctx)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	expectedCount := len(test.Users)
	if len(result) != expectedCount {
		t.Errorf("Expected count to be %d, got %d", expectedCount, len(result))
	}

	for index, user := range result {
		expectedUser := client.User{
			UserID:       test.Users[index]["userId"].(string),
			FirstName:    test.Users[index]["firstName"].(string),
			LastName:     test.Users[index]["lastName"].(string),
			EmailAddress: test.Users[index]["emailAddress"].(string),
			Status:       test.Users[index]["status"].(string),
			Source:       test.Users[index]["source"].(string),
		}

		// Check roles
		if rolesData, ok := test.Users[index]["roles"].([]interface{}); ok {
			expectedRoles := make([]string, len(rolesData))
			for i, role := range rolesData {
				if roleStr, ok := role.(string); ok {
					expectedRoles[i] = roleStr
				}
			}
			expectedUser.Roles = expectedRoles
		} else if rolesData, ok := test.Users[index]["roles"].([]string); ok {
			expectedUser.Roles = rolesData
		}

		if user.UserID != expectedUser.UserID ||
			user.FirstName != expectedUser.FirstName ||
			user.LastName != expectedUser.LastName ||
			user.EmailAddress != expectedUser.EmailAddress ||
			user.Status != expectedUser.Status ||
			user.Source != expectedUser.Source ||
			!reflect.DeepEqual(user.Roles, expectedUser.Roles) {
			t.Errorf("Unexpected user: got %+v, want %+v", user, expectedUser)
		}
	}
}

func TestNexusClient_GetUsers_RequestDetails(t *testing.T) {
	var capturedRequest *http.Request

	mockResponse := &http.Response{
		StatusCode: http.StatusOK,
		Body: io.NopCloser(strings.NewReader(`[
	{
		"userId": "test",
		"firstName": "Test",
		"lastName": "User",
		"emailAddress": "test@example.com",
		"source": "default",
		"status": "active",
		"readOnly": false,
		"roles": ["nx-admin"],
		"externalRoles": []
	}
]`)),
		Header: make(http.Header),
	}
	mockResponse.Header.Set("Content-Type", "application/json")

	mockTransport := &test.MockRoundTripper{
		Response: mockResponse,
		Err:      nil,
		RoundTripFunc: func(req *http.Request) (*http.Response, error) {
			capturedRequest = req
			return mockResponse, nil
		},
	}

	httpClient := &http.Client{Transport: mockTransport}
	baseHttpClient := uhttp.NewBaseHttpClient(httpClient)
	testClient := client.NewClient("http://localhost:8081", "admin", "admin123", baseHttpClient)

	ctx := context.Background()

	_, _, err := testClient.ListUsers(ctx)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if capturedRequest == nil {
		t.Fatal("capturedRequest is nil — the HTTP request was not captured")
	}

	expectedURL := "http://localhost:8081/service/rest/v1/security/users"
	actualURL := capturedRequest.URL.String()
	if actualURL != expectedURL {
		t.Errorf("Expected URL to be %s, got %s", expectedURL, actualURL)
	}

	expectedHeaders := map[string]string{
		"Accept":        "application/json",
		"Content-Type":  "application/json",
		"Authorization": "Basic YWRtaW46YWRtaW4xMjM=", // admin:admin123 base64 encoded.
	}

	for key, expectedValue := range expectedHeaders {
		if value := capturedRequest.Header.Get(key); value != expectedValue {
			t.Errorf("Expected header %s to be %s, got %s", key, expectedValue, value)
		}
	}
}

func TestNexusClient_GetRoles(t *testing.T) {
	body, err := test.ReadFile("rolesMock.json")
	if err != nil {
		t.Fatalf("Error reading body: %s", err)
	}
	mockResponse := &http.Response{
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(body)),
	}
	mockResponse.Header.Set("Content-Type", "application/json")

	testClient := test.NewTestClient(mockResponse, nil)

	ctx := context.Background()

	result, _, err := testClient.ListRoles(ctx)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	expectedCount := len(test.Roles)
	if len(result) != expectedCount {
		t.Errorf("Expected count to be %d, got %d", expectedCount, len(result))
	}

	for index, role := range result {
		expectedRole := client.Role{
			ID:          test.Roles[index]["id"].(string),
			Name:        test.Roles[index]["name"].(string),
			Description: test.Roles[index]["description"].(string),
			Source:      test.Roles[index]["source"].(string),
		}

		if role.ID != expectedRole.ID ||
			role.Name != expectedRole.Name ||
			role.Description != expectedRole.Description ||
			role.Source != expectedRole.Source {
			t.Errorf("Unexpected role: got %+v, want %+v", role, expectedRole)
		}
	}
}

func TestUserBuilder_List(t *testing.T) {
	// Test the user builder List method.
	body, err := test.ReadFile("usersMock.json")
	if err != nil {
		t.Fatalf("Error reading body: %s", err)
	}
	mockResponse := &http.Response{
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(body)),
	}
	mockResponse.Header.Set("Content-Type", "application/json")

	testClient := test.NewTestClient(mockResponse, nil)
	userBuilder := newUserBuilder(testClient)

	ctx := context.Background()
	resources, _, _, err := userBuilder.List(ctx, nil, nil)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if resources == nil {
		t.Fatal("Expected non-nil resources")
	}

	expectedCount := len(test.Users)
	if len(resources) != expectedCount {
		t.Errorf("Expected count to be %d, got %d", expectedCount, len(resources))
	}

	// Check that resources have the correct structure.
	for _, resource := range resources {
		if resource.Id == nil {
			t.Error("Expected resource ID to be set")
		}
		if resource.DisplayName == "" {
			t.Error("Expected resource display name to be set")
		}
		if resource.Id.ResourceType == "" {
			t.Error("Expected resource type to be set")
		}
	}
}
