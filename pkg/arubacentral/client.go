package arubacentral

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

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

func (c *Client) ListUsers(ctx context.Context, pgVars *PaginationVars) ([]User, uint, error) {
	u := &url.URL{
		Scheme: "https",
		Host:   c.baseHost,
		Path:   UsersEndpoint,
	}

	req, err := c.httpClient.NewRequest(ctx, http.MethodGet, u)
	if err != nil {
		return nil, 0, err
	}

	params := &url.Values{}
	pgVars.Apply(params)
	params.Set("app_name", "nms")
	req.URL.RawQuery = params.Encode()

	var res ListResponse[User]
	resp, err := c.httpClient.Do(
		req,
		uhttp.WithJSONResponse(&res),
		uhttp.WithErrorResponse(&ErrorResponse{}),
	)
	if err != nil {
		return nil, 0, err
	}

	defer resp.Body.Close()

	return res.Items, res.Total, nil
}

func (c *Client) ListRoles(ctx context.Context, pgVars *PaginationVars) ([]Role, uint, error) {
	u := &url.URL{
		Scheme: "https",
		Host:   c.baseHost,
		Path:   RolesEndpoint,
	}

	req, err := c.httpClient.NewRequest(ctx, http.MethodGet, u)
	if err != nil {
		return nil, 0, err
	}

	params := &url.Values{}
	pgVars.Apply(params)
	params.Set("app_name", "nms")
	req.URL.RawQuery = params.Encode()

	var res ListResponse[Role]
	resp, err := c.httpClient.Do(
		req,
		uhttp.WithJSONResponse(&res),
		uhttp.WithErrorResponse(&ErrorResponse{}),
	)
	if err != nil {
		return nil, 0, err
	}

	defer resp.Body.Close()

	return res.Items, res.Total, nil
}

func (c *Client) ListGroups(ctx context.Context, pgVars *PaginationVars) ([]string, uint, error) {
	u := &url.URL{
		Scheme: "https",
		Host:   c.baseHost,
		Path:   GroupsEndpoint,
	}

	req, err := c.httpClient.NewRequest(ctx, http.MethodGet, u)
	if err != nil {
		return nil, 0, err
	}

	params := &url.Values{}
	pgVars.Apply(params)
	req.URL.RawQuery = params.Encode()

	var res struct {
		Items [][]string `json:"data"`
		Total uint       `json:"total"`
	}
	resp, err := c.httpClient.Do(
		req,
		uhttp.WithJSONResponse(&res),
		uhttp.WithErrorResponse(&ErrorResponse{}),
	)
	if err != nil {
		return nil, 0, err
	}

	defer resp.Body.Close()

	var groups []string
	for _, group := range res.Items {
		groups = append(groups, group...)
	}

	return groups, res.Total, nil
}
