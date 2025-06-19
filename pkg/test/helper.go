package test

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/conductorone/baton-sdk/pkg/uhttp"
	"github.com/conductorone/baton-sonatype-nexus/pkg/client"
)

// Mock data for Nexus users.
var (
	Users = []map[string]interface{}{
		{
			"userId":        "anonymous",
			"firstName":     "Anonymous",
			"lastName":      "User",
			"emailAddress":  "anonymous@example.org",
			"source":        "default",
			"status":        "active",
			"readOnly":      false,
			"roles":         []string{"nx-anonymous"},
			"externalRoles": []string{},
		},
		{
			"userId":        "admin",
			"firstName":     "Administrator",
			"lastName":      "User",
			"emailAddress":  "admin@example.org",
			"source":        "default",
			"status":        "changepassword",
			"readOnly":      false,
			"roles":         []string{"nx-admin"},
			"externalRoles": []string{},
		},
	}

	// Mock data for Nexus roles.
	Roles = []map[string]interface{}{
		{
			"id":          "nx-admin",
			"name":        "nx-admin",
			"description": "Administrator Role",
			"source":      "default",
		},
		{
			"id":          "nx-anonymous",
			"name":        "nx-anonymous",
			"description": "Anonymous Role",
			"source":      "default",
		},
	}
)

// Custom RoundTripper for testing.
type TestRoundTripper struct {
	response *http.Response
	err      error
}

type MockRoundTripper struct {
	Response      *http.Response
	Err           error
	RoundTripFunc func(*http.Request) (*http.Response, error)
}

func (m *MockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if m.RoundTripFunc != nil {
		return m.RoundTripFunc(req)
	}
	return m.Response, m.Err
}

func (t *TestRoundTripper) RoundTrip(*http.Request) (*http.Response, error) {
	return t.response, t.err
}

// Helper function to create a test client with custom transport.
func NewTestClient(response *http.Response, err error) *client.APIClient {
	transport := &TestRoundTripper{response: response, err: err}
	httpClient := &http.Client{Transport: transport}
	baseHttpClient := uhttp.NewBaseHttpClient(httpClient)

	// Use test credentials for Nexus.
	baseURL := "http://localhost:8081"
	username := "admin"
	password := "admin123"

	return client.NewClient(baseURL, username, password, baseHttpClient)
}

// ReadFile reads a file from the test directory.
func ReadFile(fileName string) (string, error) {
	// Try multiple possible paths for the file
	possiblePaths := []string{
		"pkg/test/" + fileName,
		"../test/" + fileName,
		"../../pkg/test/" + fileName,
		"test/" + fileName,
	}

	for _, path := range possiblePaths {
		data, err := os.ReadFile(path)
		if err == nil {
			return string(data), nil
		}
	}

	// If none of the paths work, return the error from the first attempt.
	data, err := os.ReadFile("pkg/test/" + fileName)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// LoadMockUsers loads users from the usersMock.json file.
func LoadMockUsers() ([]map[string]interface{}, error) {
	data, err := ReadFile("usersMock.json")
	if err != nil {
		return nil, err
	}

	var users []map[string]interface{}
	err = json.Unmarshal([]byte(data), &users)
	if err != nil {
		return nil, err
	}

	return users, nil
}

// LoadMockRoles loads roles from the rolesMock.json file.
func LoadMockRoles() ([]map[string]interface{}, error) {
	data, err := ReadFile("rolesMock.json")
	if err != nil {
		return nil, err
	}

	var roles []map[string]interface{}
	err = json.Unmarshal([]byte(data), &roles)
	if err != nil {
		return nil, err
	}

	return roles, nil
}
