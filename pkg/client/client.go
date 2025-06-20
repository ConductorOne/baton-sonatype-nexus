package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

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
func (c *APIClient) doRequest(ctx context.Context, method, endpointUrl string, res any) (http.Header, annotations.Annotations, error) {
	logger := ctxzap.Extract(ctx)

	urlAddress, err := url.Parse(endpointUrl)
	if err != nil {
		return nil, nil, err
	}

	options := []uhttp.RequestOption{
		uhttp.WithContentTypeJSONHeader(),
		uhttp.WithAcceptJSONHeader(),
	}

	request, err := c.wrapper.NewRequest(ctx, method, urlAddress, options...)
	if err != nil {
		logger.Error("failed to create request", zap.Error(err))
		return nil, nil, err
	}

	var errorResp NexusErrorResponse
	doOptions := []uhttp.DoOption{
		uhttp.WithJSONResponse(res),
		uhttp.WithErrorResponse(&errorResp),
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

// ListUsers retrieves a list of users from the API.
// https://help.sonatype.com/en/api-reference.html#operations-tag-Security%20management:%20users .
func (c *APIClient) ListUsers(ctx context.Context) ([]*User, annotations.Annotations, error) {
	l := ctxzap.Extract(ctx)

	var users []*User
	queryUrl := fmt.Sprintf("%s/service/rest/v1/security/users", c.baseURL)

	_, annotation, err := c.doRequest(ctx, http.MethodGet, queryUrl, &users)
	if err != nil {
		l.Error("Error getting users", zap.Error(err))
		return nil, nil, err
	}

	return users, annotation, nil
}

// ListRoles returns a list of all roles in Nexus.
func (c *APIClient) ListRoles(ctx context.Context) ([]*Role, annotations.Annotations, error) {
	l := ctxzap.Extract(ctx)

	var roles []*Role
	queryUrl := fmt.Sprintf("%s/service/rest/v1/security/roles", c.baseURL)

	_, annotation, err := c.doRequest(ctx, http.MethodGet, queryUrl, &roles)
	if err != nil {
		l.Error("Error getting roles", zap.Error(err))
		return nil, nil, err
	}

	return roles, annotation, nil
}
