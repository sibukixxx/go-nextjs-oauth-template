// Package service contains business logic for the application.
package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/your-org/go-nextjs-oauth-template/backend/internal/config"
	"github.com/your-org/go-nextjs-oauth-template/backend/pkg/oauth"
)

// MultiAuthService handles OAuth 2.1 authentication with multiple providers
type MultiAuthService struct {
	registry        *oauth.ProviderRegistry
	defaultProvider oauth.ProviderType
	logger          *slog.Logger

	// Session storage (in-memory for simplicity, use Redis in production)
	sessions  map[string]*MultiAuthSession
	sessionMu sync.RWMutex
}

// MultiAuthSession represents an authentication session with provider info
type MultiAuthSession struct {
	ID           string             `json:"id"`
	Provider     oauth.ProviderType `json:"provider"`
	State        string             `json:"state"`
	Nonce        string             `json:"nonce,omitempty"`
	CodeVerifier string             `json:"code_verifier"`
	RedirectURI  string             `json:"redirect_uri"`
	CreatedAt    time.Time          `json:"created_at"`
	ExpiresAt    time.Time          `json:"expires_at"`
}

// MultiLoginRequest represents a login initiation request with provider selection
type MultiLoginRequest struct {
	Provider    string `json:"provider,omitempty"` // "google" or "line"
	RedirectURI string `json:"redirect_uri,omitempty"`
	Scope       string `json:"scope,omitempty"`
}

// MultiLoginResponse represents the response for login initiation
type MultiLoginResponse struct {
	AuthorizationURL string `json:"authorization_url"`
	SessionID        string `json:"session_id"`
	State            string `json:"state"`
	Provider         string `json:"provider"`
}

// MultiCallbackRequest represents the OAuth callback request
type MultiCallbackRequest struct {
	Code      string `json:"code"`
	State     string `json:"state"`
	SessionID string `json:"session_id"`
}

// MultiAuthResult represents the result of a successful authentication
type MultiAuthResult struct {
	AccessToken  string          `json:"access_token"`
	TokenType    string          `json:"token_type"`
	ExpiresIn    int             `json:"expires_in"`
	RefreshToken string          `json:"refresh_token,omitempty"`
	IDToken      string          `json:"id_token,omitempty"`
	Claims       *oauth.Claims   `json:"claims,omitempty"`
	UserInfo     *oauth.UserInfo `json:"user_info,omitempty"`
	Provider     string          `json:"provider"`
}

// MultiAuthServiceConfig holds configuration for MultiAuthService
type MultiAuthServiceConfig struct {
	Config config.OAuthProvidersConfig
	Logger *slog.Logger
}

// NewMultiAuthService creates a new multi-provider authentication service
func NewMultiAuthService(cfg MultiAuthServiceConfig) (*MultiAuthService, error) {
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}

	registry := oauth.NewProviderRegistry()

	// Register Google provider if enabled
	if cfg.Config.Google.Enabled {
		redirectURL := cfg.Config.Google.RedirectURL
		if redirectURL == "" {
			redirectURL = cfg.Config.CommonRedirectURL
		}

		googleProvider, err := oauth.NewGoogleProvider(oauth.GoogleProviderOptions{
			Config: oauth.ProviderConfig{
				Type:         oauth.ProviderGoogle,
				ClientID:     cfg.Config.Google.ClientID,
				ClientSecret: cfg.Config.Google.ClientSecret,
				RedirectURL:  redirectURL,
				Scopes:       cfg.Config.Google.Scopes,
			},
			Logger: cfg.Logger,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create Google provider: %w", err)
		}
		registry.Register(googleProvider)
		cfg.Logger.Info("Google OAuth provider registered")
	}

	// Register LINE provider if enabled
	if cfg.Config.LINE.Enabled {
		redirectURL := cfg.Config.LINE.RedirectURL
		if redirectURL == "" {
			redirectURL = cfg.Config.CommonRedirectURL
		}

		lineProvider, err := oauth.NewLINEProvider(oauth.LINEProviderOptions{
			Config: oauth.ProviderConfig{
				Type:         oauth.ProviderLINE,
				ClientID:     cfg.Config.LINE.ClientID,
				ClientSecret: cfg.Config.LINE.ClientSecret,
				RedirectURL:  redirectURL,
				Scopes:       cfg.Config.LINE.Scopes,
				Options:      cfg.Config.LINE.Options,
			},
			Logger: cfg.Logger,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create LINE provider: %w", err)
		}
		registry.Register(lineProvider)
		cfg.Logger.Info("LINE OAuth provider registered")
	}

	// Validate at least one provider is registered
	if len(registry.List()) == 0 {
		return nil, fmt.Errorf("at least one OAuth provider must be enabled")
	}

	// Determine default provider
	defaultProvider := oauth.ProviderGoogle
	if cfg.Config.DefaultProvider != "" {
		pt, err := oauth.ParseProviderType(cfg.Config.DefaultProvider)
		if err != nil {
			cfg.Logger.Warn("invalid default provider, falling back to google", "provider", cfg.Config.DefaultProvider)
		} else {
			defaultProvider = pt
		}
	}

	// If default provider is not registered, use first available
	if !registry.Has(defaultProvider) {
		providers := registry.List()
		if len(providers) > 0 {
			defaultProvider = providers[0]
			cfg.Logger.Info("default provider not available, using", "provider", defaultProvider)
		}
	}

	return &MultiAuthService{
		registry:        registry,
		defaultProvider: defaultProvider,
		logger:          cfg.Logger,
		sessions:        make(map[string]*MultiAuthSession),
	}, nil
}

// InitiateLogin starts the OAuth login flow for a specific provider
func (s *MultiAuthService) InitiateLogin(ctx context.Context, req MultiLoginRequest) (*MultiLoginResponse, error) {
	// Determine provider
	providerType := s.defaultProvider
	if req.Provider != "" {
		pt, err := oauth.ParseProviderType(req.Provider)
		if err != nil {
			return nil, fmt.Errorf("invalid provider: %w", err)
		}
		providerType = pt
	}

	// Get provider
	provider, err := s.registry.Get(providerType)
	if err != nil {
		return nil, fmt.Errorf("provider not available: %w", err)
	}

	// Generate PKCE parameters (required in OAuth 2.1)
	pkce, err := oauth.GeneratePKCE()
	if err != nil {
		return nil, fmt.Errorf("failed to generate PKCE: %w", err)
	}

	// Generate state parameter (required in OAuth 2.1 for CSRF protection)
	state, err := oauth.GenerateState()
	if err != nil {
		return nil, fmt.Errorf("failed to generate state: %w", err)
	}

	// Generate nonce for OpenID Connect
	nonce, err := oauth.GenerateNonce()
	if err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Generate session ID
	sessionID, err := generateMultiSessionID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate session ID: %w", err)
	}

	// Determine redirect URI
	redirectURI := req.RedirectURI
	if redirectURI == "" {
		redirectURI = provider.GetConfig().RedirectURL
	}

	// Build authorization URL
	authURL, err := provider.BuildAuthorizationURL(&oauth.AuthorizationRequest{
		State:       state,
		PKCE:        pkce,
		Nonce:       nonce,
		Scope:       req.Scope,
		RedirectURI: redirectURI,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to build authorization URL: %w", err)
	}

	// Store session
	session := &MultiAuthSession{
		ID:           sessionID,
		Provider:     providerType,
		State:        state,
		Nonce:        nonce,
		CodeVerifier: pkce.CodeVerifier,
		RedirectURI:  redirectURI,
		CreatedAt:    time.Now(),
		ExpiresAt:    time.Now().Add(10 * time.Minute),
	}

	s.sessionMu.Lock()
	s.sessions[sessionID] = session
	s.sessionMu.Unlock()

	s.logger.Info("login initiated",
		"session_id", sessionID,
		"provider", providerType,
		"redirect_uri", redirectURI,
	)

	return &MultiLoginResponse{
		AuthorizationURL: authURL,
		SessionID:        sessionID,
		State:            state,
		Provider:         providerType.String(),
	}, nil
}

// HandleCallback handles the OAuth callback
func (s *MultiAuthService) HandleCallback(ctx context.Context, req MultiCallbackRequest) (*MultiAuthResult, error) {
	if req.State == "" {
		return nil, fmt.Errorf("state parameter is required")
	}
	if req.Code == "" {
		return nil, fmt.Errorf("authorization code is required")
	}

	// Get session
	s.sessionMu.RLock()
	session, exists := s.sessions[req.SessionID]
	s.sessionMu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("session not found: %s", req.SessionID)
	}

	// Verify session hasn't expired
	if time.Now().After(session.ExpiresAt) {
		s.deleteSession(req.SessionID)
		return nil, fmt.Errorf("session has expired")
	}

	// Verify state matches
	if session.State != req.State {
		s.deleteSession(req.SessionID)
		return nil, fmt.Errorf("state mismatch: possible CSRF attack")
	}

	// Get provider
	provider, err := s.registry.Get(session.Provider)
	if err != nil {
		s.deleteSession(req.SessionID)
		return nil, fmt.Errorf("provider not available: %w", err)
	}

	// Exchange code for tokens (with PKCE code_verifier)
	tokenResp, err := provider.ExchangeCode(ctx, req.Code, session.CodeVerifier, session.RedirectURI)
	if err != nil {
		s.deleteSession(req.SessionID)
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}

	// Validate ID token if present (OpenID Connect)
	var claims *oauth.Claims
	if tokenResp.IDToken != "" {
		claims, err = provider.ValidateIDToken(ctx, tokenResp.IDToken, session.Nonce)
		if err != nil {
			s.deleteSession(req.SessionID)
			return nil, fmt.Errorf("failed to validate ID token: %w", err)
		}
	}

	// Get user info if ID token doesn't have all info
	var userInfo *oauth.UserInfo
	if tokenResp.AccessToken != "" {
		userInfo, err = provider.GetUserInfo(ctx, tokenResp.AccessToken)
		if err != nil {
			// Log but don't fail - some providers may not support userinfo endpoint
			s.logger.Warn("failed to get user info", "error", err, "provider", session.Provider)
		}
	}

	// Clean up session
	s.deleteSession(req.SessionID)

	s.logger.Info("callback successful",
		"session_id", req.SessionID,
		"provider", session.Provider,
		"has_id_token", tokenResp.IDToken != "",
		"has_refresh_token", tokenResp.RefreshToken != "",
	)

	return &MultiAuthResult{
		AccessToken:  tokenResp.AccessToken,
		TokenType:    tokenResp.TokenType,
		ExpiresIn:    tokenResp.ExpiresIn,
		RefreshToken: tokenResp.RefreshToken,
		IDToken:      tokenResp.IDToken,
		Claims:       claims,
		UserInfo:     userInfo,
		Provider:     session.Provider.String(),
	}, nil
}

// RefreshToken refreshes an access token for a specific provider
func (s *MultiAuthService) RefreshToken(ctx context.Context, providerName, refreshToken string) (*MultiAuthResult, error) {
	providerType, err := oauth.ParseProviderType(providerName)
	if err != nil {
		return nil, fmt.Errorf("invalid provider: %w", err)
	}

	provider, err := s.registry.Get(providerType)
	if err != nil {
		return nil, fmt.Errorf("provider not available: %w", err)
	}

	tokenResp, err := provider.RefreshToken(ctx, refreshToken)
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	return &MultiAuthResult{
		AccessToken:  tokenResp.AccessToken,
		TokenType:    tokenResp.TokenType,
		ExpiresIn:    tokenResp.ExpiresIn,
		RefreshToken: tokenResp.RefreshToken,
		IDToken:      tokenResp.IDToken,
		Provider:     providerName,
	}, nil
}

// RevokeToken revokes a token for a specific provider
func (s *MultiAuthService) RevokeToken(ctx context.Context, providerName, token, tokenTypeHint string) error {
	providerType, err := oauth.ParseProviderType(providerName)
	if err != nil {
		return fmt.Errorf("invalid provider: %w", err)
	}

	provider, err := s.registry.Get(providerType)
	if err != nil {
		return fmt.Errorf("provider not available: %w", err)
	}

	if err := provider.RevokeToken(ctx, token, tokenTypeHint); err != nil {
		return fmt.Errorf("failed to revoke token: %w", err)
	}

	s.logger.Info("token revoked", "provider", providerName, "token_type_hint", tokenTypeHint)
	return nil
}

// GetAvailableProviders returns the list of available providers
func (s *MultiAuthService) GetAvailableProviders() []string {
	providers := s.registry.List()
	result := make([]string, len(providers))
	for i, p := range providers {
		result[i] = p.String()
	}
	return result
}

// GetDefaultProvider returns the default provider
func (s *MultiAuthService) GetDefaultProvider() string {
	return s.defaultProvider.String()
}

// GetProvider returns a provider by name
func (s *MultiAuthService) GetProvider(name string) (oauth.Provider, error) {
	return s.registry.GetByName(name)
}

// ValidateToken validates a token using the specified provider
func (s *MultiAuthService) ValidateToken(ctx context.Context, providerName, idToken string) (*oauth.Claims, error) {
	providerType, err := oauth.ParseProviderType(providerName)
	if err != nil {
		return nil, fmt.Errorf("invalid provider: %w", err)
	}

	provider, err := s.registry.Get(providerType)
	if err != nil {
		return nil, fmt.Errorf("provider not available: %w", err)
	}

	return provider.ValidateIDToken(ctx, idToken, "")
}

// deleteSession removes a session from storage
func (s *MultiAuthService) deleteSession(sessionID string) {
	s.sessionMu.Lock()
	delete(s.sessions, sessionID)
	s.sessionMu.Unlock()
}

// CleanupExpiredSessions removes expired sessions
func (s *MultiAuthService) CleanupExpiredSessions() {
	s.sessionMu.Lock()
	defer s.sessionMu.Unlock()

	now := time.Now()
	for id, session := range s.sessions {
		if now.After(session.ExpiresAt) {
			delete(s.sessions, id)
		}
	}
}

// generateMultiSessionID generates a cryptographically secure session ID
func generateMultiSessionID() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
