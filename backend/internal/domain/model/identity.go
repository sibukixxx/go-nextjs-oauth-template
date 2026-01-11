package model

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// IdentityProvider represents the authentication provider
type IdentityProvider string

const (
	IdentityProviderLine   IdentityProvider = "line"
	IdentityProviderGoogle IdentityProvider = "google"
	IdentityProviderApple  IdentityProvider = "apple"
	IdentityProviderEmail  IdentityProvider = "email"
)

// Identity represents a login method
// Multiple identities can be linked to a single account
type Identity struct {
	ID              uuid.UUID        `json:"id"`
	AccountID       uuid.UUID        `json:"account_id"`
	Provider        IdentityProvider `json:"provider"`
	ProviderSubject string           `json:"provider_subject"` // LINE userId / OpenID sub
	Email           *string          `json:"email,omitempty"`  // Email from provider (not authoritative)
	ProviderData    json.RawMessage  `json:"provider_data,omitempty"`
	LastLoginAt     *time.Time       `json:"last_login_at,omitempty"`
	LoginCount      int              `json:"login_count"`
	CreatedAt       time.Time        `json:"created_at"`
	UpdatedAt       time.Time        `json:"updated_at"`
}

// NewIdentity creates a new identity
func NewIdentity(accountID uuid.UUID, provider IdentityProvider, providerSubject string) *Identity {
	now := time.Now()
	return &Identity{
		ID:              uuid.New(),
		AccountID:       accountID,
		Provider:        provider,
		ProviderSubject: providerSubject,
		ProviderData:    json.RawMessage("{}"),
		LoginCount:      0,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
}

// RecordLogin updates the identity after a successful login
func (i *Identity) RecordLogin() {
	now := time.Now()
	i.LastLoginAt = &now
	i.LoginCount++
	i.UpdatedAt = now
}

// SetEmail sets the email from the provider
func (i *Identity) SetEmail(email string) {
	i.Email = &email
	i.UpdatedAt = time.Now()
}

// SetProviderData sets the provider-specific data
func (i *Identity) SetProviderData(data interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	i.ProviderData = jsonData
	i.UpdatedAt = time.Now()
	return nil
}

// GetProviderData unmarshals the provider data into the given target
func (i *Identity) GetProviderData(target interface{}) error {
	if len(i.ProviderData) == 0 {
		return nil
	}
	return json.Unmarshal(i.ProviderData, target)
}

// ProviderDataMap represents provider-specific data stored in Identity
type ProviderDataMap struct {
	// LINE specific
	DisplayName *string `json:"display_name,omitempty"`
	PictureURL  *string `json:"picture_url,omitempty"`
	StatusMsg   *string `json:"status_message,omitempty"`

	// Google specific
	Name          *string `json:"name,omitempty"`
	GivenName     *string `json:"given_name,omitempty"`
	FamilyName    *string `json:"family_name,omitempty"`
	Picture       *string `json:"picture,omitempty"`
	Locale        *string `json:"locale,omitempty"`
	VerifiedEmail *bool   `json:"verified_email,omitempty"`

	// Apple specific (minimal)
	AppleEmail *string `json:"apple_email,omitempty"`
}

// IsValidProvider checks if the provider is valid
func IsValidProvider(provider string) bool {
	switch IdentityProvider(provider) {
	case IdentityProviderLine, IdentityProviderGoogle, IdentityProviderApple, IdentityProviderEmail:
		return true
	default:
		return false
	}
}

// ParseIdentityProvider parses a string into an IdentityProvider
func ParseIdentityProvider(provider string) (IdentityProvider, error) {
	p := IdentityProvider(provider)
	switch p {
	case IdentityProviderLine, IdentityProviderGoogle, IdentityProviderApple, IdentityProviderEmail:
		return p, nil
	default:
		return "", fmt.Errorf("unknown identity provider: %s", provider)
	}
}
