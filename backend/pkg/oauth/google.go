package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"time"
)

// GoogleProvider implements the Provider interface for Google OAuth
type GoogleProvider struct {
	config       ProviderConfig
	client       *Client
	jwksClient   *JWKSClient
	jwtValidator *JWTValidator
	httpClient   *http.Client
	logger       *slog.Logger
}

// GoogleProviderOptions contains options for creating a Google provider
type GoogleProviderOptions struct {
	Config     ProviderConfig
	Logger     *slog.Logger
	HTTPClient *http.Client

	// JWT validation options
	Audiences         []string
	AuthorizedParties []string
	RequiredScopes    []string
	ClockSkew         time.Duration
}

// DefaultGoogleConfig returns the default Google OAuth configuration
func DefaultGoogleConfig() ProviderConfig {
	return ProviderConfig{
		Type:             ProviderGoogle,
		AuthorizationURL: "https://accounts.google.com/o/oauth2/v2/auth",
		TokenURL:         "https://oauth2.googleapis.com/token",
		RevokeURL:        "https://oauth2.googleapis.com/revoke",
		LogoutURL:        "https://accounts.google.com/logout",
		Issuer:           "https://accounts.google.com",
		JWKSURL:          "https://www.googleapis.com/oauth2/v3/certs",
		Scopes:           []string{"openid", "profile", "email"},
	}
}

// NewGoogleProvider creates a new Google OAuth provider
func NewGoogleProvider(opts GoogleProviderOptions) (*GoogleProvider, error) {
	if opts.Config.ClientID == "" {
		return nil, fmt.Errorf("client_id is required")
	}
	if opts.Config.RedirectURL == "" {
		return nil, fmt.Errorf("redirect_url is required")
	}

	// Merge with defaults
	cfg := opts.Config
	defaults := DefaultGoogleConfig()
	if cfg.AuthorizationURL == "" {
		cfg.AuthorizationURL = defaults.AuthorizationURL
	}
	if cfg.TokenURL == "" {
		cfg.TokenURL = defaults.TokenURL
	}
	if cfg.RevokeURL == "" {
		cfg.RevokeURL = defaults.RevokeURL
	}
	if cfg.LogoutURL == "" {
		cfg.LogoutURL = defaults.LogoutURL
	}
	if cfg.Issuer == "" {
		cfg.Issuer = defaults.Issuer
	}
	if cfg.JWKSURL == "" {
		cfg.JWKSURL = defaults.JWKSURL
	}
	if len(cfg.Scopes) == 0 {
		cfg.Scopes = defaults.Scopes
	}
	cfg.Type = ProviderGoogle

	// Create OAuth client
	client, err := NewClient(ClientConfig{
		ClientID:         cfg.ClientID,
		ClientSecret:     cfg.ClientSecret,
		AuthorizationURL: cfg.AuthorizationURL,
		TokenURL:         cfg.TokenURL,
		RedirectURL:      cfg.RedirectURL,
		Scopes:           cfg.Scopes,
		JWKSURL:          cfg.JWKSURL,
		Issuer:           cfg.Issuer,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create OAuth client: %w", err)
	}

	logger := opts.Logger
	if logger == nil {
		logger = slog.Default()
	}

	httpClient := opts.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 30 * time.Second}
	}

	// Create JWKs client
	jwksClient, err := NewJWKSClient(JWKSConfig{
		JWKSURL:     cfg.JWKSURL,
		RefreshRate: 1 * time.Hour,
		CacheTTL:    24 * time.Hour,
		HTTPTimeout: 10 * time.Second,
		Logger:      logger,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create JWKs client: %w", err)
	}

	// Determine audiences
	audiences := opts.Audiences
	if len(audiences) == 0 {
		audiences = []string{cfg.ClientID}
	}

	// Create JWT validator
	clockSkew := opts.ClockSkew
	if clockSkew == 0 {
		clockSkew = 1 * time.Minute
	}

	jwtValidator, err := NewJWTValidator(JWTValidatorConfig{
		JWKSClient:        jwksClient,
		Issuer:            cfg.Issuer,
		Audiences:         audiences,
		AuthorizedParties: opts.AuthorizedParties,
		RequiredScopes:    opts.RequiredScopes,
		ClockSkew:         clockSkew,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create JWT validator: %w", err)
	}

	return &GoogleProvider{
		config:       cfg,
		client:       client,
		jwksClient:   jwksClient,
		jwtValidator: jwtValidator,
		httpClient:   httpClient,
		logger:       logger,
	}, nil
}

// Type returns the provider type
func (p *GoogleProvider) Type() ProviderType {
	return ProviderGoogle
}

// BuildAuthorizationURL builds the authorization URL for the OAuth flow
func (p *GoogleProvider) BuildAuthorizationURL(req *AuthorizationRequest) (string, error) {
	return p.client.BuildAuthorizationURL(req)
}

// ExchangeCode exchanges an authorization code for tokens
func (p *GoogleProvider) ExchangeCode(ctx context.Context, code, codeVerifier, redirectURI string) (*TokenResponse, error) {
	return p.client.ExchangeCode(ctx, code, codeVerifier, redirectURI)
}

// RefreshToken refreshes an access token using a refresh token
func (p *GoogleProvider) RefreshToken(ctx context.Context, refreshToken string) (*TokenResponse, error) {
	return p.client.RefreshToken(ctx, refreshToken)
}

// RevokeToken revokes a token
func (p *GoogleProvider) RevokeToken(ctx context.Context, token, tokenTypeHint string) error {
	if p.config.RevokeURL == "" {
		return fmt.Errorf("revoke URL is not configured")
	}
	return p.client.RevokeToken(ctx, token, tokenTypeHint, p.config.RevokeURL)
}

// ValidateIDToken validates the ID token and returns the claims
func (p *GoogleProvider) ValidateIDToken(ctx context.Context, idToken string, nonce string) (*Claims, error) {
	result, err := p.jwtValidator.Validate(idToken)
	if err != nil {
		return nil, fmt.Errorf("failed to validate ID token: %w", err)
	}

	// Verify nonce if provided
	if nonce != "" && result.Claims.Nonce != nonce {
		return nil, fmt.Errorf("nonce mismatch in ID token")
	}

	return result.Claims, nil
}

// GetUserInfo fetches user information using the access token
func (p *GoogleProvider) GetUserInfo(ctx context.Context, accessToken string) (*UserInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://www.googleapis.com/oauth2/v2/userinfo", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("user info request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var googleUser struct {
		ID            string `json:"id"`
		Email         string `json:"email"`
		VerifiedEmail bool   `json:"verified_email"`
		Name          string `json:"name"`
		GivenName     string `json:"given_name"`
		FamilyName    string `json:"family_name"`
		Picture       string `json:"picture"`
		Locale        string `json:"locale"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&googleUser); err != nil {
		return nil, fmt.Errorf("failed to decode user info: %w", err)
	}

	return &UserInfo{
		ID:      googleUser.ID,
		Email:   googleUser.Email,
		Name:    googleUser.Name,
		Picture: googleUser.Picture,
		Extra: map[string]string{
			"given_name":     googleUser.GivenName,
			"family_name":    googleUser.FamilyName,
			"locale":         googleUser.Locale,
			"verified_email": fmt.Sprintf("%t", googleUser.VerifiedEmail),
		},
	}, nil
}

// GetConfig returns the provider configuration
func (p *GoogleProvider) GetConfig() ProviderConfig {
	return p.config
}

// SupportsLogout returns whether the provider supports logout URL
func (p *GoogleProvider) SupportsLogout() bool {
	return p.config.LogoutURL != ""
}

// BuildLogoutURL builds the logout URL for OpenID Connect RP-Initiated Logout
func (p *GoogleProvider) BuildLogoutURL(idToken, postLogoutRedirectURI, state string) (string, error) {
	if p.config.LogoutURL == "" {
		return "", fmt.Errorf("logout URL is not configured")
	}

	u, err := url.Parse(p.config.LogoutURL)
	if err != nil {
		return "", fmt.Errorf("invalid logout URL: %w", err)
	}

	q := u.Query()
	q.Set("client_id", p.config.ClientID)

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

// GetJWTValidator returns the JWT validator for external use
func (p *GoogleProvider) GetJWTValidator() *JWTValidator {
	return p.jwtValidator
}
