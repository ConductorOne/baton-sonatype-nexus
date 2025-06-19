package client

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"

	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/uhttp"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
)

const (
	getUsersEndpoint = "/service/rest/v1/security/users"
	getRolesEndpoint = "/service/rest/v1/security/roles"
)

type APIClient struct {
	wrapper  *uhttp.BaseHttpClient
	baseURL  string
	username string
	password string
}

// NewClient creates a new API client.
func NewClient(baseURL, username, password string, httpClient *uhttp.BaseHttpClient) *APIClient {
	if httpClient == nil {
		httpClient = uhttp.NewBaseHttpClient(http.DefaultClient)
	}

	return &APIClient{
		wrapper:  httpClient,
		baseURL:  baseURL,
		username: username,
		password: password,
	}
}

// ListUsers retrieves a list of users from the API.
// https://help.sonatype.com/en/api-reference.html#operations-tag-Security%20management:%20users .
func (c *APIClient) ListUsers(ctx context.Context) ([]User, annotations.Annotations, error) {
	l := ctxzap.Extract(ctx)
	var users []User

	queryUrl, err := url.JoinPath(c.baseURL, getUsersEndpoint)
	if err != nil {
		l.Error("Error creating users URL", zap.Error(err))
		return nil, nil, err
	}

	_, annotation, err := c.doRequest(ctx, http.MethodGet, queryUrl, &users)
	if err != nil {
		l.Error("Error getting users", zap.Error(err))
		return nil, nil, err
	}

	return users, annotation, nil
}

// ListRoles retrieves a list of roles from the API.
// https://help.sonatype.com/en/api-reference.html#operations-tag-Security%20management:%20roles .
func (c *APIClient) ListRoles(ctx context.Context) ([]Role, annotations.Annotations, error) {
	l := ctxzap.Extract(ctx)
	var roles []Role

	queryUrl, err := url.JoinPath(c.baseURL, getRolesEndpoint)
	if err != nil {
		l.Error("Error creating roles URL", zap.Error(err))
		return nil, nil, err
	}

	_, annotation, err := c.doRequest(ctx, http.MethodGet, queryUrl, &roles)
	if err != nil {
		l.Error("Error getting roles", zap.Error(err))
		return nil, nil, err
	}

	return roles, annotation, nil
}

// doRequest executes an HTTP request and processes the response.
func (c *APIClient) doRequest(ctx context.Context, method, endpointUrl string, res any) (http.Header, annotations.Annotations, error) {
	logger := ctxzap.Extract(ctx)

	urlAddress, err := url.Parse(endpointUrl)
	if err != nil {
		return nil, nil, err
	}

	options := []uhttp.RequestOption{
		uhttp.WithContentTypeJSONHeader(),
		uhttp.WithAcceptJSONHeader(),
		uhttp.WithHeader("Authorization", fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", c.username, c.password))))),
	}

	request, err := c.wrapper.NewRequest(ctx, method, urlAddress, options...)
	if err != nil {
		logger.Error("failed to create request", zap.Error(err))
		return nil, nil, err
	}

	annotation := annotations.Annotations{}
	doOptions := []uhttp.DoOption{}

	if res != nil {
		doOptions = append(doOptions, uhttp.WithJSONResponse(res))
	}

	response, err := c.wrapper.Do(request, doOptions...)
	if response != nil && response.Body != nil {
		defer response.Body.Close()
	}

	if err != nil {
		return nil, annotation, fmt.Errorf("error in Do: %w", err)
	}

	return response.Header, annotation, nil
}
