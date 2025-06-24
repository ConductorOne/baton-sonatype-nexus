package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/uhttp"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
)

type APIClient struct {
	baseURL string
	wrapper *uhttp.BaseHttpClient
}

type NexusErrorResponse struct {
	ErrorMessage string `json:"message"`
}

func (e *NexusErrorResponse) Message() string {
	return e.ErrorMessage
}

// doRequest executes an HTTP request and processes the response.
func (c *APIClient) doRequest(ctx context.Context, method, endpointUrl string, reqBody, res any) (http.Header, annotations.Annotations, error) {
	logger := ctxzap.Extract(ctx)

	urlAddress, err := url.Parse(endpointUrl)
	if err != nil {
		return nil, nil, err
	}

	options := []uhttp.RequestOption{
		uhttp.WithContentTypeJSONHeader(),
		uhttp.WithAcceptJSONHeader(),
	}

	if reqBody != nil {
		options = append(options, uhttp.WithJSONBody(reqBody))
	}

	request, err := c.wrapper.NewRequest(ctx, method, urlAddress, options...)
	if err != nil {
		logger.Error("failed to create request", zap.Error(err))
		return nil, nil, err
	}

	var errorResp NexusErrorResponse
	var rateLimitDesc v2.RateLimitDescription
	doOptions := []uhttp.DoOption{
		uhttp.WithErrorResponse(&errorResp),
		uhttp.WithRatelimitData(&rateLimitDesc),
	}

	if res != nil {
		doOptions = append(doOptions, uhttp.WithJSONResponse(res))
	}

	resp, err := c.wrapper.Do(request, doOptions...)
	if err != nil {
		logger.Error("failed to execute request",
			zap.String("url", endpointUrl),
			zap.String("method", method),
			zap.Error(err),
		)
		return nil, nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		logger.Error("request failed",
			zap.String("url", endpointUrl),
			zap.String("method", method),
			zap.Int("status_code", resp.StatusCode),
			zap.String("error", errorResp.Message()),
		)
		return nil, nil, fmt.Errorf("request failed: %s", errorResp.Message())
	}

	annotation := annotations.Annotations{}
	annotation.Append(&rateLimitDesc)
	return resp.Header, annotation, nil
}

// NewClient creates a new Nexus API client.
func NewClient(ctx context.Context, baseURL, username, password string, httpClient *http.Client) (*APIClient, error) {
	if httpClient == nil {
		var err error
		httpClient, err = uhttp.NewBasicAuth(username, password).GetClient(ctx,
			uhttp.WithUserAgent("baton-sonatype-nexus"),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create http client: %w", err)
		}
	}

	wrapper := uhttp.NewBaseHttpClient(httpClient)

	return &APIClient{
		baseURL: baseURL,
		wrapper: wrapper,
	}, nil
}

// CreateUser creates a new user in Nexus.
func (c *APIClient) CreateUser(ctx context.Context, payload *UserCreatePayload) (*User, annotations.Annotations, error) {
	l := ctxzap.Extract(ctx)

	var createdUser User
	queryUrl := fmt.Sprintf("%s/service/rest/v1/security/users", c.baseURL)

	_, annotation, err := c.doRequest(ctx, http.MethodPost, queryUrl, payload, &createdUser)
	if err != nil {
		l.Error("Error creating user", zap.Error(err))
		return nil, nil, fmt.Errorf("error creating user: %w", err)
	}

	return &createdUser, annotation, nil
}

// ListUsers retrieves a list of users from the API.
// https://help.sonatype.com/en/api-reference.html#operations-tag-Security%20management:%20users .
func (c *APIClient) ListUsers(ctx context.Context) ([]*User, annotations.Annotations, error) {
	l := ctxzap.Extract(ctx)

	var users []*User
	queryUrl := fmt.Sprintf("%s/service/rest/v1/security/users", c.baseURL)

	_, annotation, err := c.doRequest(ctx, http.MethodGet, queryUrl, nil, &users)
	if err != nil {
		l.Error("Error getting users", zap.Error(err))
		return nil, nil, fmt.Errorf("error getting users: %w", err)
	}

	return users, annotation, nil
}

// ListRoles returns a list of all roles in Nexus.
func (c *APIClient) ListRoles(ctx context.Context) ([]Role, annotations.Annotations, error) {
	l := ctxzap.Extract(ctx)

	var roles []Role
	queryUrl := fmt.Sprintf("%s/service/rest/v1/security/roles", c.baseURL)

	_, annotation, err := c.doRequest(ctx, http.MethodGet, queryUrl, nil, &roles)
	if err != nil {
		l.Error("Error getting roles", zap.Error(err))
		return nil, nil, fmt.Errorf("error getting roles: %w", err)
	}

	return roles, annotation, nil
}

// DeleteUser deletes a user in Nexus.
func (c *APIClient) DeleteUser(ctx context.Context, userID string) (annotations.Annotations, error) {
	l := ctxzap.Extract(ctx)

	queryUrl := fmt.Sprintf("%s/service/rest/v1/security/users/%s", c.baseURL, url.PathEscape(userID))

	_, annotation, err := c.doRequest(ctx, http.MethodDelete, queryUrl, nil, nil)
	if err != nil {
		l.Error("Error deleting user", zap.String("user_id", userID), zap.Error(err))
		return nil, fmt.Errorf("error deleting user %s: %w", userID, err)
	}

	return annotation, nil
}

// UpdateUser updates a user in Nexus.
func (c *APIClient) UpdateUser(ctx context.Context, userID string, payload *User) (annotations.Annotations, error) {
	l := ctxzap.Extract(ctx)

	queryUrl := fmt.Sprintf("%s/service/rest/v1/security/users/%s", c.baseURL, url.PathEscape(userID))

	_, annotation, err := c.doRequest(ctx, http.MethodPut, queryUrl, payload, nil)
	if err != nil {
		l.Error("Error updating user", zap.String("user_id", userID), zap.Error(err))
		return nil, fmt.Errorf("error updating user %s: %w", userID, err)
	}

	return annotation, nil
}

func (c *APIClient) ListUsersByID(ctx context.Context, userId string) ([]*User, annotations.Annotations, error) {
	l := ctxzap.Extract(ctx)

	var users []*User
	queryUrl := fmt.Sprintf("%s/service/rest/v1/security/users?userId=%s", c.baseURL, userId)

	_, annotation, err := c.doRequest(ctx, http.MethodGet, queryUrl, nil, &users)
	if err != nil {
		l.Error("Error getting users by ID", zap.Error(err))
		return nil, nil, fmt.Errorf("error getting users by ID %s: %w", userId, err)
	}

	return users, annotation, nil
}
