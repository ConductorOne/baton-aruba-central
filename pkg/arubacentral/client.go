package arubacentral

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/uhttp"
)

const (
	UsersEndpoint  = "/platform/rbac/v1/users"
	RolesEndpoint  = "/platform/rbac/v1/roles"
	GroupsEndpoint = "/configuration/v2/groups"
)

type Client struct {
	httpClient *uhttp.BaseHttpClient
	baseHost   string
}

func NewClient(httpClient *http.Client, baseHost string) *Client {
	return &Client{
		httpClient: uhttp.NewBaseHttpClient(httpClient),
		baseHost:   baseHost,
	}
}

type PaginationVars struct {
	Limit  uint `json:"limit"`
	Offset uint `json:"offset"`
}

func NewPaginationVars(limit, offset uint) *PaginationVars {
	return &PaginationVars{Limit: limit, Offset: offset}
}

func (pgVars *PaginationVars) Apply(params *url.Values) {
	params.Set("limit", fmt.Sprint(pgVars.Limit))
	params.Set("offset", fmt.Sprint(pgVars.Offset))
}

type ListResponse[T any] struct {
	Items []T  `json:"items"`
	Total uint `json:"total"`
}

type ErrorResponse struct {
	Error string `json:"error"`
	Code  int    `json:"status_code"`
}

func (e *ErrorResponse) Message() string {
	return fmt.Sprintf("error: %s, status code: %d", e.Error, e.Code)
}

func WithRatelimitData(resource *v2.RateLimitDescription) uhttp.DoOption {
	return func(resp *uhttp.WrapperResponse) error {
		rl, err := extractRateLimitData(resp.StatusCode, &resp.Header)
		if err != nil {
			return err
		}

		resource.Limit = rl.Limit
		resource.Remaining = rl.Remaining
		resource.ResetAt = rl.ResetAt
		resource.Status = rl.Status

		return nil
	}
}

func (c *Client) ListUsers(ctx context.Context, pgVars *PaginationVars) ([]User, uint, *v2.RateLimitDescription, error) {
	u := &url.URL{
		Scheme: "https",
		Host:   c.baseHost,
		Path:   UsersEndpoint,
	}

	req, err := c.httpClient.NewRequest(ctx, http.MethodGet, u)
	if err != nil {
		return nil, 0, nil, err
	}

	params := &url.Values{}
	pgVars.Apply(params)
	params.Set("app_name", "nms")
	req.URL.RawQuery = params.Encode()

	var res ListResponse[User]
	var rl v2.RateLimitDescription
	resp, err := c.httpClient.Do(
		req,
		uhttp.WithJSONResponse(&res),
		uhttp.WithErrorResponse(&ErrorResponse{}),
		WithRatelimitData(&rl),
	)
	if err != nil {
		return nil, 0, &rl, err
	}

	defer resp.Body.Close()

	return res.Items, res.Total, &rl, nil
}

func (c *Client) ListRoles(ctx context.Context, pgVars *PaginationVars) ([]Role, uint, *v2.RateLimitDescription, error) {
	u := &url.URL{
		Scheme: "https",
		Host:   c.baseHost,
		Path:   RolesEndpoint,
	}

	req, err := c.httpClient.NewRequest(ctx, http.MethodGet, u)
	if err != nil {
		return nil, 0, nil, err
	}

	params := &url.Values{}
	pgVars.Apply(params)
	params.Set("app_name", "nms")
	req.URL.RawQuery = params.Encode()

	var res ListResponse[Role]
	var rl v2.RateLimitDescription
	resp, err := c.httpClient.Do(
		req,
		uhttp.WithJSONResponse(&res),
		uhttp.WithErrorResponse(&ErrorResponse{}),
		WithRatelimitData(&rl),
	)
	if err != nil {
		return nil, 0, &rl, err
	}

	defer resp.Body.Close()

	return res.Items, res.Total, &rl, nil
}

func (c *Client) ListGroups(ctx context.Context, pgVars *PaginationVars) ([]string, uint, *v2.RateLimitDescription, error) {
	u := &url.URL{
		Scheme: "https",
		Host:   c.baseHost,
		Path:   GroupsEndpoint,
	}

	req, err := c.httpClient.NewRequest(ctx, http.MethodGet, u)
	if err != nil {
		return nil, 0, nil, err
	}

	params := &url.Values{}
	pgVars.Apply(params)
	req.URL.RawQuery = params.Encode()

	var res struct {
		Items [][]string `json:"data"`
		Total uint       `json:"total"`
	}
	var rl v2.RateLimitDescription
	resp, err := c.httpClient.Do(
		req,
		uhttp.WithJSONResponse(&res),
		uhttp.WithErrorResponse(&ErrorResponse{}),
		WithRatelimitData(&rl),
	)
	if err != nil {
		return nil, 0, &rl, err
	}

	defer resp.Body.Close()

	var groups []string
	for _, group := range res.Items {
		groups = append(groups, group...)
	}

	return groups, res.Total, &rl, nil
}
