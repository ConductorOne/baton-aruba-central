package arubacentral

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/conductorone/baton-sdk/pkg/uhttp"
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

func (pgVars *PaginationVars) Apply(req *http.Request) {
	query := url.Values{}
	query.Set("limit", fmt.Sprint(pgVars.Limit))
	query.Set("offset", fmt.Sprint(pgVars.Offset))
	req.URL.RawQuery = query.Encode()
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
		Path:   "/platform/rbac/v1/users",
	}

	req, err := c.httpClient.NewRequest(ctx, http.MethodGet, u)
	if err != nil {
		return nil, 0, err
	}

	pgVars.Apply(req)

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
