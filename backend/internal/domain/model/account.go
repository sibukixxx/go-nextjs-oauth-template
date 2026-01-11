package model

import (
	"time"

	"github.com/google/uuid"
)

// AccountStatus represents the status of an account
type AccountStatus string

const (
	AccountStatusActive    AccountStatus = "active"
	AccountStatusSuspended AccountStatus = "suspended"
	AccountStatusDeleted   AccountStatus = "deleted"
)

// Account represents a user account (main entity)
// This is the primary entity for billing, notifications, and user data
type Account struct {
	ID              uuid.UUID     `json:"id"`
	Status          AccountStatus `json:"status"`
	PrimaryEmail    *string       `json:"primary_email,omitempty"`
	EmailVerifiedAt *time.Time    `json:"email_verified_at,omitempty"`
	DisplayName     *string       `json:"display_name,omitempty"`
	BillingName     *string       `json:"billing_name,omitempty"`
	AvatarURL       *string       `json:"avatar_url,omitempty"`
	CreatedAt       time.Time     `json:"created_at"`
	UpdatedAt       time.Time     `json:"updated_at"`
	DeletedAt       *time.Time    `json:"deleted_at,omitempty"`

	// Relationships (loaded separately)
	Identities []Identity `json:"identities,omitempty"`
}

// NewAccount creates a new account with default values
func NewAccount() *Account {
	now := time.Now()
	return &Account{
		ID:        uuid.New(),
		Status:    AccountStatusActive,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// IsActive returns true if the account is active
func (a *Account) IsActive() bool {
	return a.Status == AccountStatusActive
}

// HasVerifiedEmail returns true if the primary email is verified
func (a *Account) HasVerifiedEmail() bool {
	return a.PrimaryEmail != nil && a.EmailVerifiedAt != nil
}

// CanLogin returns true if the account can be used for login
func (a *Account) CanLogin() bool {
	return a.Status == AccountStatusActive && a.DeletedAt == nil
}

// SetPrimaryEmail sets the primary email (unverified)
func (a *Account) SetPrimaryEmail(email string) {
	a.PrimaryEmail = &email
	a.EmailVerifiedAt = nil
	a.UpdatedAt = time.Now()
}

// VerifyEmail marks the primary email as verified
func (a *Account) VerifyEmail() {
	if a.PrimaryEmail != nil {
		now := time.Now()
		a.EmailVerifiedAt = &now
		a.UpdatedAt = now
	}
}

// Suspend suspends the account
func (a *Account) Suspend() {
	a.Status = AccountStatusSuspended
	a.UpdatedAt = time.Now()
}

// Activate activates the account
func (a *Account) Activate() {
	a.Status = AccountStatusActive
	a.UpdatedAt = time.Now()
}

// SoftDelete marks the account as deleted
func (a *Account) SoftDelete() {
	now := time.Now()
	a.Status = AccountStatusDeleted
	a.DeletedAt = &now
	a.UpdatedAt = now
}
