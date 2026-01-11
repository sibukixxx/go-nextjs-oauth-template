// Package oauth provides OAuth 2.1 authentication support with multi-provider capabilities.
package oauth

import (
	"context"
	"fmt"
)

// ProviderType represents the type of OAuth provider
type ProviderType string

const (
	// ProviderGoogle represents Google OAuth provider
	ProviderGoogle ProviderType = "google"
	// ProviderLINE represents LINE OAuth provider
	ProviderLINE ProviderType = "line"
)

// String returns the string representation of the provider type
func (p ProviderType) String() string {
	return string(p)
}

// IsValid checks if the provider type is valid
func (p ProviderType) IsValid() bool {
	switch p {
	case ProviderGoogle, ProviderLINE:
		return true
	default:
		return false
	}
}

// ParseProviderType parses a string into a ProviderType
func ParseProviderType(s string) (ProviderType, error) {
	switch s {
	case "google":
		return ProviderGoogle, nil
	case "line":
		return ProviderLINE, nil
	default:
		return "", fmt.Errorf("unknown provider type: %s", s)
	}
}

// ProviderConfig holds common configuration for all OAuth providers
type ProviderConfig struct {
	// Provider identification
	Type ProviderType `json:"type"`

	// OAuth endpoints
	AuthorizationURL string `json:"authorization_url"`
	TokenURL         string `json:"token_url"`
	RevokeURL        string `json:"revoke_url,omitempty"`
	LogoutURL        string `json:"logout_url,omitempty"`

	// Client credentials
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	RedirectURL  string `json:"redirect_url"`

	// OAuth scopes
	Scopes []string `json:"scopes"`

	// Token validation
	Issuer  string `json:"issuer,omitempty"`
	JWKSURL string `json:"jwks_url,omitempty"`

	// Provider-specific options
	Options map[string]string `json:"options,omitempty"`
}

// Provider defines the interface for OAuth providers
type Provider interface {
	// Type returns the provider type
	Type() ProviderType

	// BuildAuthorizationURL builds the authorization URL for the OAuth flow
	BuildAuthorizationURL(req *AuthorizationRequest) (string, error)

	// ExchangeCode exchanges an authorization code for tokens
	ExchangeCode(ctx context.Context, code, codeVerifier, redirectURI string) (*TokenResponse, error)

	// RefreshToken refreshes an access token using a refresh token
	RefreshToken(ctx context.Context, refreshToken string) (*TokenResponse, error)

	// RevokeToken revokes a token (access or refresh token)
	RevokeToken(ctx context.Context, token, tokenTypeHint string) error

	// ValidateIDToken validates the ID token and returns the claims
	ValidateIDToken(ctx context.Context, idToken string, nonce string) (*Claims, error)

	// GetUserInfo fetches user information using the access token (optional)
	GetUserInfo(ctx context.Context, accessToken string) (*UserInfo, error)

	// GetConfig returns the provider configuration
	GetConfig() ProviderConfig

	// SupportsLogout returns whether the provider supports logout URL
	SupportsLogout() bool

	// BuildLogoutURL builds the logout URL (if supported)
	BuildLogoutURL(idToken, postLogoutRedirectURI, state string) (string, error)
}

// UserInfo represents user information from the OAuth provider
type UserInfo struct {
	ID          string            `json:"id"`          // Provider-specific user ID
	Email       string            `json:"email"`       // Email address (if available)
	Name        string            `json:"name"`        // Display name
	Picture     string            `json:"picture"`     // Profile picture URL
	DisplayName string            `json:"displayName"` // LINE specific display name
	Extra       map[string]string `json:"extra"`       // Additional provider-specific fields
}

// ProviderRegistry holds registered OAuth providers
type ProviderRegistry struct {
	providers map[ProviderType]Provider
}

// NewProviderRegistry creates a new provider registry
func NewProviderRegistry() *ProviderRegistry {
	return &ProviderRegistry{
		providers: make(map[ProviderType]Provider),
	}
}

// Register registers a provider
func (r *ProviderRegistry) Register(provider Provider) {
	r.providers[provider.Type()] = provider
}

// Get returns a provider by type
func (r *ProviderRegistry) Get(providerType ProviderType) (Provider, error) {
	provider, exists := r.providers[providerType]
	if !exists {
		return nil, fmt.Errorf("provider not found: %s", providerType)
	}
	return provider, nil
}

// GetByName returns a provider by name string
func (r *ProviderRegistry) GetByName(name string) (Provider, error) {
	providerType, err := ParseProviderType(name)
	if err != nil {
		return nil, err
	}
	return r.Get(providerType)
}

// Has checks if a provider is registered
func (r *ProviderRegistry) Has(providerType ProviderType) bool {
	_, exists := r.providers[providerType]
	return exists
}

// List returns all registered provider types
func (r *ProviderRegistry) List() []ProviderType {
	types := make([]ProviderType, 0, len(r.providers))
	for t := range r.providers {
		types = append(types, t)
	}
	return types
}
