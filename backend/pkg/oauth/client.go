package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Client handles OAuth 2.1 authorization flows
type Client struct {
	config     ClientConfig
	httpClient *http.Client
}

// ClientConfig holds OAuth client configuration
type ClientConfig struct {
	ClientID         string
	ClientSecret     string // Optional for public clients
	AuthorizationURL string
	TokenURL         string
	RedirectURL      string
	Scopes           []string
	JWKSURL          string // For token validation
	Issuer           string
}

// TokenResponse represents the OAuth token response
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Scope        string `json:"scope,omitempty"`
	IDToken      string `json:"id_token,omitempty"` // OpenID Connect
}

// TokenError represents an OAuth token error response
type TokenError struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description,omitempty"`
	ErrorURI         string `json:"error_uri,omitempty"`
}

// AuthorizationRequest represents the OAuth authorization request parameters
type AuthorizationRequest struct {
	State         string
	PKCE          *PKCE
	Nonce         string // For OpenID Connect
	ResponseType  string // "code" for Authorization Code flow
	Scope         string
	RedirectURI   string
	CodeChallenge string
	CodeMethod    string
}

// NewClient creates a new OAuth 2.1 client
func NewClient(cfg ClientConfig) (*Client, error) {
	if cfg.ClientID == "" {
		return nil, fmt.Errorf("client_id is required")
	}
	if cfg.AuthorizationURL == "" {
		return nil, fmt.Errorf("authorization_url is required")
	}
	if cfg.TokenURL == "" {
		return nil, fmt.Errorf("token_url is required")
	}
	if cfg.RedirectURL == "" {
		return nil, fmt.Errorf("redirect_url is required")
	}

	return &Client{
		config: cfg,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

// BuildAuthorizationURL builds the authorization URL for the OAuth flow
// OAuth 2.1 requires PKCE and state parameters
func (c *Client) BuildAuthorizationURL(req *AuthorizationRequest) (string, error) {
	if req.State == "" {
		return "", fmt.Errorf("state parameter is required (OAuth 2.1)")
	}
	if req.PKCE == nil {
		return "", fmt.Errorf("PKCE is required (OAuth 2.1)")
	}

	u, err := url.Parse(c.config.AuthorizationURL)
	if err != nil {
		return "", fmt.Errorf("invalid authorization URL: %w", err)
	}

	q := u.Query()
	q.Set("response_type", "code") // OAuth 2.1: Only Authorization Code flow is allowed
	q.Set("client_id", c.config.ClientID)
	q.Set("redirect_uri", req.RedirectURI)
	if req.RedirectURI == "" {
		q.Set("redirect_uri", c.config.RedirectURL)
	}
	q.Set("state", req.State)

	// PKCE parameters (required in OAuth 2.1)
	q.Set("code_challenge", req.PKCE.CodeChallenge)
	q.Set("code_challenge_method", req.PKCE.Method)

	// Scope
	scope := req.Scope
	if scope == "" && len(c.config.Scopes) > 0 {
		scope = strings.Join(c.config.Scopes, " ")
	}
	if scope != "" {
		q.Set("scope", scope)
	}

	// Nonce for OpenID Connect
	if req.Nonce != "" {
		q.Set("nonce", req.Nonce)
	}

	u.RawQuery = q.Encode()
	return u.String(), nil
}

// ExchangeCode exchanges an authorization code for tokens
// OAuth 2.1 requires PKCE code_verifier
func (c *Client) ExchangeCode(ctx context.Context, code, codeVerifier, redirectURI string) (*TokenResponse, error) {
	if code == "" {
		return nil, fmt.Errorf("authorization code is required")
	}
	if codeVerifier == "" {
		return nil, fmt.Errorf("code_verifier is required (OAuth 2.1 PKCE)")
	}

	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("client_id", c.config.ClientID)
	data.Set("code_verifier", codeVerifier)

	if redirectURI != "" {
		data.Set("redirect_uri", redirectURI)
	} else {
		data.Set("redirect_uri", c.config.RedirectURL)
	}

	// Add client_secret if configured (for confidential clients)
	if c.config.ClientSecret != "" {
		data.Set("client_secret", c.config.ClientSecret)
	}

	return c.doTokenRequest(ctx, data)
}

// RefreshToken refreshes an access token using a refresh token
func (c *Client) RefreshToken(ctx context.Context, refreshToken string) (*TokenResponse, error) {
	if refreshToken == "" {
		return nil, fmt.Errorf("refresh_token is required")
	}

	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", refreshToken)
	data.Set("client_id", c.config.ClientID)

	if c.config.ClientSecret != "" {
		data.Set("client_secret", c.config.ClientSecret)
	}

	return c.doTokenRequest(ctx, data)
}

// doTokenRequest performs the token request
func (c *Client) doTokenRequest(ctx context.Context, data url.Values) (*TokenResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.config.TokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create token request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("token request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var tokenErr TokenError
		if err := json.Unmarshal(body, &tokenErr); err == nil && tokenErr.Error != "" {
			return nil, fmt.Errorf("token error: %s - %s", tokenErr.Error, tokenErr.ErrorDescription)
		}
		return nil, fmt.Errorf("token request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	return &tokenResp, nil
}

// RevokeToken revokes a token (access or refresh token)
func (c *Client) RevokeToken(ctx context.Context, token, tokenTypeHint, revokeURL string) error {
	if token == "" {
		return fmt.Errorf("token is required")
	}
	if revokeURL == "" {
		return fmt.Errorf("revoke_url is required")
	}

	data := url.Values{}
	data.Set("token", token)
	data.Set("client_id", c.config.ClientID)

	if tokenTypeHint != "" {
		data.Set("token_type_hint", tokenTypeHint)
	}

	if c.config.ClientSecret != "" {
		data.Set("client_secret", c.config.ClientSecret)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, revokeURL, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create revoke request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("revoke request failed: %w", err)
	}
	defer resp.Body.Close()

	// Per RFC 7009, the authorization server responds with HTTP 200 even if the token was invalid
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("revoke request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// BuildLogoutURL builds the logout URL for OpenID Connect RP-Initiated Logout
func (c *Client) BuildLogoutURL(logoutURL, idToken, postLogoutRedirectURI, state string) (string, error) {
	if logoutURL == "" {
		return "", fmt.Errorf("logout_url is required")
	}

	u, err := url.Parse(logoutURL)
	if err != nil {
		return "", fmt.Errorf("invalid logout URL: %w", err)
	}

	q := u.Query()
	q.Set("client_id", c.config.ClientID)

	if idToken != "" {
		q.Set("id_token_hint", idToken)
	}

	if postLogoutRedirectURI != "" {
		q.Set("post_logout_redirect_uri", postLogoutRedirectURI)
	}

	if state != "" {
		q.Set("state", state)
	}

	u.RawQuery = q.Encode()
	return u.String(), nil
}

// GetConfig returns a copy of the client configuration
func (c *Client) GetConfig() ClientConfig {
	return c.config
}
