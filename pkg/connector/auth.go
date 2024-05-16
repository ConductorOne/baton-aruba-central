package connector

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"sync"
	"time"

	"github.com/conductorone/baton-sdk/pkg/uhttp"
	"golang.org/x/net/publicsuffix"
)

const (
	LoginEndpoint    = "/oauth2/authorize/central/api/login"
	AuthCodeEndpoint = "/oauth2/authorize/central/api"
	TokenEndpoint    = "/oauth2/token" // #nosec G101 (hardcoded credentials are not used here)
)

type Token struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    time.Time
}

type AuthMiddleware struct {
	Transport http.RoundTripper
	Token     *Token
	mu        sync.Mutex

	baseHost     string
	clientID     string
	clientSecret string
}

func (m *AuthMiddleware) RoundTrip(req *http.Request) (*http.Response, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// check if token is expired
	if m.Token == nil || m.Token.ExpiresIn.IsZero() || time.Now().After(m.Token.ExpiresIn) {
		if err := m.refreshToken(req.Context()); err != nil {
			return nil, fmt.Errorf("failed to refresh token: %w", err)
		}
	}

	// set auth header with new token
	req.Header.Set("Authorization", "Bearer "+m.Token.AccessToken)

	return m.Transport.RoundTrip(req)
}

func (m *AuthMiddleware) refreshToken(ctx context.Context) error {
	accessToken, refreshToken, expiresIn, err := refreshToken(
		ctx,
		&http.Client{
			Transport: m.Transport,
		},
		m.baseHost,
		m.clientID,
		m.clientSecret,
		m.Token.RefreshToken,
	)
	if err != nil {
		return err
	}

	m.Token = &Token{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    time.Now().Add(time.Duration(expiresIn) * time.Second),
	}

	return nil
}

type OAuthConfig interface {
	GetClient(ctx context.Context) (*http.Client, error)
}

type NoConfig struct{}

func (cfg *NoConfig) GetClient(ctx context.Context) (*http.Client, error) {
	return http.DefaultClient, nil
}

type BaseConfig struct {
	BaseHost     string
	ClientID     string
	ClientSecret string
}

type RefreshTokenFlowConfig struct {
	BaseConfig
	AccessToken  string
	RefreshToken string
}

func (cfg *RefreshTokenFlowConfig) GetClient(ctx context.Context) (*http.Client, error) {
	httpClient, err := uhttp.NewClient(ctx, uhttp.WithLogger(true, nil))
	if err != nil {
		return nil, err
	}

	return &http.Client{
		Transport: &AuthMiddleware{
			Transport: httpClient.Transport,
			Token: &Token{
				AccessToken:  cfg.AccessToken,
				RefreshToken: cfg.RefreshToken,
				ExpiresIn:    time.Time{},
			},
			baseHost:     cfg.BaseHost,
			clientID:     cfg.ClientID,
			clientSecret: cfg.ClientSecret,
		},
	}, nil
}

type CodeFlowConfig struct {
	BaseConfig
	Username   string
	Password   string
	CustomerID string
}

func (cfg *CodeFlowConfig) GetClient(ctx context.Context) (*http.Client, error) {
	loginURL := &url.URL{
		Scheme: "https",
		Host:   cfg.BaseHost,
		Path:   LoginEndpoint,
	}
	authCodeURL := &url.URL{
		Scheme: "https",
		Host:   cfg.BaseHost,
		Path:   AuthCodeEndpoint,
	}
	tokenURL := &url.URL{
		Scheme: "https",
		Host:   cfg.BaseHost,
		Path:   TokenEndpoint,
	}

	// prepare a http client with logger and cookie jar to enable code flow (for parsing csrf token from cookies)
	httpClient, err := uhttp.NewClient(ctx, uhttp.WithLogger(true, nil))
	if err != nil {
		return nil, err
	}

	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		return nil, err
	}

	httpClient = &http.Client{
		Transport: httpClient.Transport,
		Jar:       jar,
	}

	// 1. login and get CSRF token
	err = loginAndGetCSRF(ctx, httpClient, loginURL.String(), cfg.ClientID, cfg.Username, cfg.Password)
	if err != nil {
		return nil, err
	}

	var csrfToken string
	cookies := jar.Cookies(loginURL)
	for _, cookie := range cookies {
		if cookie.Name == "X-CSRF-TOKEN" {
			csrfToken = cookie.Value
		}
	}

	// 2. get auth code
	authCode, err := getAuthCode(ctx, httpClient, authCodeURL.String(), cfg.ClientID, cfg.CustomerID, csrfToken)
	if err != nil {
		return nil, err
	}

	// 3. exchange auth code for access token
	accessToken, refreshToken, expiresIn, err := exchangeCodeForToken(ctx, httpClient, tokenURL.String(), cfg.ClientID, cfg.ClientSecret, authCode)
	if err != nil {
		return nil, err
	}

	return &http.Client{
		Transport: &AuthMiddleware{
			Transport: httpClient.Transport,
			Token: &Token{
				AccessToken:  accessToken,
				RefreshToken: refreshToken,
				ExpiresIn:    time.Now().Add(time.Duration(expiresIn) * time.Second),
			},
			baseHost:     cfg.BaseHost,
			clientID:     cfg.ClientID,
			clientSecret: cfg.ClientSecret,
		},
	}, nil
}

func loginAndGetCSRF(ctx context.Context, httpClient *http.Client, loginURL, clientID, username, password string) error {
	body := struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}{
		Username: username,
		Password: password,
	}

	b, err := json.Marshal(body)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, loginURL, bytes.NewReader(b))
	if err != nil {
		return err
	}

	// set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// add query params
	queryParams := url.Values{}
	queryParams.Set("client_id", clientID)
	req.URL.RawQuery = queryParams.Encode()

	// send request
	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	return nil
}

func getAuthCode(ctx context.Context, httpClient *http.Client, authCodeURL, clientID, customerID, csrfToken string) (string, error) {
	body := struct {
		CustomerID string `json:"customer_id"`
	}{
		CustomerID: customerID,
	}

	b, err := json.Marshal(body)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, authCodeURL, bytes.NewReader(b))
	if err != nil {
		return "", err
	}

	// set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-CSRF-TOKEN", csrfToken)

	// add query params
	queryParams := url.Values{}
	queryParams.Set("client_id", clientID)
	queryParams.Set("response_type", "code")
	queryParams.Set("scope", "all")
	req.URL.RawQuery = queryParams.Encode()

	// send request
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var respBody struct {
		AuthCode string `json:"auth_code"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		return "", err
	}

	return respBody.AuthCode, nil
}

func exchangeCodeForToken(ctx context.Context, httpClient *http.Client, tokenURL, clientID, clientSecret, authCode string) (string, string, int, error) {
	body := struct {
		ClientID     string `json:"client_id"`
		ClientSecret string `json:"client_secret"`
		GrantType    string `json:"grant_type"`
		AuthCode     string `json:"code"`
	}{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		GrantType:    "authorization_code",
		AuthCode:     authCode,
	}

	b, err := json.Marshal(body)
	if err != nil {
		return "", "", 0, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, bytes.NewReader(b))
	if err != nil {
		return "", "", 0, err
	}

	// set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// send request
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", "", 0, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", 0, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var respBody struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		return "", "", 0, err
	}

	return respBody.AccessToken, respBody.RefreshToken, respBody.ExpiresIn, nil
}

func refreshToken(ctx context.Context, httpClient *http.Client, baseHost string, clientID, clientSecret, refreshToken string) (string, string, int, error) {
	tokenURL := &url.URL{
		Scheme: "https",
		Host:   baseHost,
		Path:   TokenEndpoint,
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL.String(), nil)
	if err != nil {
		return "", "", 0, err
	}

	// set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// add query params
	queryParams := url.Values{}
	queryParams.Set("client_id", clientID)
	queryParams.Set("client_secret", clientSecret)
	queryParams.Set("grant_type", "refresh_token")
	queryParams.Set("refresh_token", refreshToken)
	req.URL.RawQuery = queryParams.Encode()

	// send request
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", "", 0, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", 0, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var respBody struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		return "", "", 0, err
	}

	return respBody.AccessToken, respBody.RefreshToken, respBody.ExpiresIn, nil
}
