package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// LINEProvider implements the Provider interface for LINE OAuth
type LINEProvider struct {
	config     ProviderConfig
	httpClient *http.Client
	logger     *slog.Logger
}

// LINEProviderOptions contains options for creating a LINE provider
type LINEProviderOptions struct {
	Config     ProviderConfig
	Logger     *slog.Logger
	HTTPClient *http.Client
}

// DefaultLINEConfig returns the default LINE OAuth configuration
func DefaultLINEConfig() ProviderConfig {
	return ProviderConfig{
		Type:             ProviderLINE,
		AuthorizationURL: "https://access.line.me/oauth2/v2.1/authorize",
		TokenURL:         "https://api.line.me/oauth2/v2.1/token",
		RevokeURL:        "https://api.line.me/oauth2/v2.1/revoke",
		Issuer:           "https://access.line.me",
		Scopes:           []string{"profile", "openid"},
	}
}

// LINEEndpoints holds LINE API endpoints
var LINEEndpoints = struct {
	VerifyURL  string
	ProfileURL string
}{
	VerifyURL:  "https://api.line.me/oauth2/v2.1/verify",
	ProfileURL: "https://api.line.me/v2/profile",
}

// NewLINEProvider creates a new LINE OAuth provider
func NewLINEProvider(opts LINEProviderOptions) (*LINEProvider, error) {
	if opts.Config.ClientID == "" {
		return nil, fmt.Errorf("client_id (channel_id) is required")
	}
	if opts.Config.ClientSecret == "" {
		return nil, fmt.Errorf("client_secret (channel_secret) is required")
	}
	if opts.Config.RedirectURL == "" {
		return nil, fmt.Errorf("redirect_url is required")
	}

	// Merge with defaults
	cfg := opts.Config
	defaults := DefaultLINEConfig()
	if cfg.AuthorizationURL == "" {
		cfg.AuthorizationURL = defaults.AuthorizationURL
	}
	if cfg.TokenURL == "" {
		cfg.TokenURL = defaults.TokenURL
	}
	if cfg.RevokeURL == "" {
		cfg.RevokeURL = defaults.RevokeURL
	}
	if cfg.Issuer == "" {
		cfg.Issuer = defaults.Issuer
	}
	if len(cfg.Scopes) == 0 {
		cfg.Scopes = defaults.Scopes
	}
	cfg.Type = ProviderLINE

	logger := opts.Logger
	if logger == nil {
		logger = slog.Default()
	}

	httpClient := opts.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 30 * time.Second}
	}

	return &LINEProvider{
		config:     cfg,
		httpClient: httpClient,
		logger:     logger,
	}, nil
}

// Type returns the provider type
func (p *LINEProvider) Type() ProviderType {
	return ProviderLINE
}

// BuildAuthorizationURL builds the authorization URL for the LINE OAuth flow
func (p *LINEProvider) BuildAuthorizationURL(req *AuthorizationRequest) (string, error) {
	if req.State == "" {
		return "", fmt.Errorf("state parameter is required (OAuth 2.1)")
	}
	if req.PKCE == nil {
		return "", fmt.Errorf("PKCE is required (OAuth 2.1)")
	}

	u, err := url.Parse(p.config.AuthorizationURL)
	if err != nil {
		return "", fmt.Errorf("invalid authorization URL: %w", err)
	}

	q := u.Query()
	q.Set("response_type", "code")
	q.Set("client_id", p.config.ClientID)

	redirectURI := req.RedirectURI
	if redirectURI == "" {
		redirectURI = p.config.RedirectURL
	}
	q.Set("redirect_uri", redirectURI)
	q.Set("state", req.State)

	// PKCE parameters
	q.Set("code_challenge", req.PKCE.CodeChallenge)
	q.Set("code_challenge_method", req.PKCE.Method)

	// Scope
	scope := req.Scope
	if scope == "" && len(p.config.Scopes) > 0 {
		scope = strings.Join(p.config.Scopes, " ")
	}
	if scope != "" {
		q.Set("scope", scope)
	}

	// Nonce for OpenID Connect
	if req.Nonce != "" {
		q.Set("nonce", req.Nonce)
	}

	// LINE-specific: bot_prompt option (optional)
	if botPrompt, ok := p.config.Options["bot_prompt"]; ok && botPrompt != "" {
		q.Set("bot_prompt", botPrompt)
	}

	u.RawQuery = q.Encode()
	return u.String(), nil
}

// ExchangeCode exchanges an authorization code for tokens
func (p *LINEProvider) ExchangeCode(ctx context.Context, code, codeVerifier, redirectURI string) (*TokenResponse, error) {
	if code == "" {
		return nil, fmt.Errorf("authorization code is required")
	}
	if codeVerifier == "" {
		return nil, fmt.Errorf("code_verifier is required (OAuth 2.1 PKCE)")
	}

	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("client_id", p.config.ClientID)
	data.Set("client_secret", p.config.ClientSecret)
	data.Set("code_verifier", codeVerifier)

	if redirectURI != "" {
		data.Set("redirect_uri", redirectURI)
	} else {
		data.Set("redirect_uri", p.config.RedirectURL)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.config.TokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create token request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := p.httpClient.Do(req)
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

// RefreshToken refreshes an access token using a refresh token
func (p *LINEProvider) RefreshToken(ctx context.Context, refreshToken string) (*TokenResponse, error) {
	if refreshToken == "" {
		return nil, fmt.Errorf("refresh_token is required")
	}

	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", refreshToken)
	data.Set("client_id", p.config.ClientID)
	data.Set("client_secret", p.config.ClientSecret)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.config.TokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create refresh request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("refresh request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var tokenErr TokenError
		if err := json.Unmarshal(body, &tokenErr); err == nil && tokenErr.Error != "" {
			return nil, fmt.Errorf("refresh error: %s - %s", tokenErr.Error, tokenErr.ErrorDescription)
		}
		return nil, fmt.Errorf("refresh request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("failed to parse refresh response: %w", err)
	}

	return &tokenResp, nil
}

// RevokeToken revokes a token
func (p *LINEProvider) RevokeToken(ctx context.Context, token, tokenTypeHint string) error {
	if p.config.RevokeURL == "" {
		return fmt.Errorf("revoke URL is not configured")
	}
	if token == "" {
		return fmt.Errorf("token is required")
	}

	data := url.Values{}
	data.Set("access_token", token)
	data.Set("client_id", p.config.ClientID)
	data.Set("client_secret", p.config.ClientSecret)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.config.RevokeURL, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create revoke request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("revoke request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("revoke request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// LINEVerifyResponse represents the response from LINE's verify endpoint
type LINEVerifyResponse struct {
	Issuer    string `json:"iss"`
	Subject   string `json:"sub"`
	Audience  string `json:"aud"`
	ExpiresAt int64  `json:"exp"`
	IssuedAt  int64  `json:"iat"`
	Nonce     string `json:"nonce,omitempty"`
	Name      string `json:"name,omitempty"`
	Picture   string `json:"picture,omitempty"`
	Email     string `json:"email,omitempty"`
}

// ValidateIDToken validates the ID token using LINE's verify API
func (p *LINEProvider) ValidateIDToken(ctx context.Context, idToken string, nonce string) (*Claims, error) {
	if idToken == "" {
		return nil, fmt.Errorf("id_token is required")
	}

	// Use LINE's verify API
	data := url.Values{}
	data.Set("id_token", idToken)
	data.Set("client_id", p.config.ClientID)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, LINEEndpoints.VerifyURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create verify request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("verify request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read verify response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var tokenErr TokenError
		if err := json.Unmarshal(body, &tokenErr); err == nil && tokenErr.Error != "" {
			return nil, fmt.Errorf("verify error: %s - %s", tokenErr.Error, tokenErr.ErrorDescription)
		}
		return nil, fmt.Errorf("verify request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var verifyResp LINEVerifyResponse
	if err := json.Unmarshal(body, &verifyResp); err != nil {
		return nil, fmt.Errorf("failed to parse verify response: %w", err)
	}

	// Validate audience (client_id)
	if verifyResp.Audience != p.config.ClientID {
		return nil, fmt.Errorf("invalid audience: expected %q, got %q", p.config.ClientID, verifyResp.Audience)
	}

	// Validate nonce if provided
	if nonce != "" && verifyResp.Nonce != nonce {
		return nil, fmt.Errorf("nonce mismatch: expected %q, got %q", nonce, verifyResp.Nonce)
	}

	// Validate expiration
	if verifyResp.ExpiresAt < time.Now().Unix() {
		return nil, fmt.Errorf("token has expired")
	}

	// Convert to Claims
	claims := &Claims{
		Issuer:    verifyResp.Issuer,
		Subject:   verifyResp.Subject,
		Audience:  Audience{verifyResp.Audience},
		ExpiresAt: verifyResp.ExpiresAt,
		IssuedAt:  verifyResp.IssuedAt,
		Nonce:     verifyResp.Nonce,
		Extra: map[string]interface{}{
			"name":    verifyResp.Name,
			"picture": verifyResp.Picture,
			"email":   verifyResp.Email,
		},
	}

	return claims, nil
}

// LINEProfileResponse represents the response from LINE's profile endpoint
type LINEProfileResponse struct {
	UserID        string `json:"userId"`
	DisplayName   string `json:"displayName"`
	PictureURL    string `json:"pictureUrl,omitempty"`
	StatusMessage string `json:"statusMessage,omitempty"`
}

// GetUserInfo fetches user information using the access token
func (p *LINEProvider) GetUserInfo(ctx context.Context, accessToken string) (*UserInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, LINEEndpoints.ProfileURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create profile request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch profile: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read profile response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("profile request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var profile LINEProfileResponse
	if err := json.Unmarshal(body, &profile); err != nil {
		return nil, fmt.Errorf("failed to parse profile response: %w", err)
	}

	return &UserInfo{
		ID:          profile.UserID,
		Name:        profile.DisplayName,
		DisplayName: profile.DisplayName,
		Picture:     profile.PictureURL,
		Extra: map[string]string{
			"status_message": profile.StatusMessage,
		},
	}, nil
}

// GetConfig returns the provider configuration
func (p *LINEProvider) GetConfig() ProviderConfig {
	return p.config
}

// SupportsLogout returns whether the provider supports logout URL
// LINE does not support OpenID Connect RP-Initiated Logout
func (p *LINEProvider) SupportsLogout() bool {
	return false
}

// BuildLogoutURL is not supported by LINE
func (p *LINEProvider) BuildLogoutURL(idToken, postLogoutRedirectURI, state string) (string, error) {
	return "", fmt.Errorf("LINE does not support logout URL; use token revocation instead")
}

// VerifyAccessToken verifies an access token using LINE's verify API
func (p *LINEProvider) VerifyAccessToken(ctx context.Context, accessToken string) (*LINEAccessTokenVerifyResponse, error) {
	if accessToken == "" {
		return nil, fmt.Errorf("access_token is required")
	}

	verifyURL := "https://api.line.me/oauth2/v2.1/verify?access_token=" + url.QueryEscape(accessToken)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, verifyURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create verify request: %w", err)
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("verify request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read verify response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var tokenErr TokenError
		if err := json.Unmarshal(body, &tokenErr); err == nil && tokenErr.Error != "" {
			return nil, fmt.Errorf("verify error: %s - %s", tokenErr.Error, tokenErr.ErrorDescription)
		}
		return nil, fmt.Errorf("verify request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var verifyResp LINEAccessTokenVerifyResponse
	if err := json.Unmarshal(body, &verifyResp); err != nil {
		return nil, fmt.Errorf("failed to parse verify response: %w", err)
	}

	// Validate client_id
	if verifyResp.ClientID != p.config.ClientID {
		return nil, fmt.Errorf("invalid client_id: expected %q, got %q", p.config.ClientID, verifyResp.ClientID)
	}

	return &verifyResp, nil
}

// LINEAccessTokenVerifyResponse represents the response from LINE's access token verify endpoint
type LINEAccessTokenVerifyResponse struct {
	Scope     string `json:"scope"`
	ClientID  string `json:"client_id"`
	ExpiresIn int    `json:"expires_in"`
}
