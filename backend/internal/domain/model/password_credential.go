package model

import (
	"time"

	"github.com/google/uuid"
)

// PasswordCredential stores password hash for email/password authentication.
// It is linked to an Identity with provider="email".
type PasswordCredential struct {
	ID            uuid.UUID `json:"id"`
	IdentityID    uuid.UUID `json:"identity_id"`
	PasswordHash  string    `json:"-"` // Never expose in JSON
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	LastChangedAt time.Time `json:"last_changed_at"`
}

// NewPasswordCredential creates a new password credential.
func NewPasswordCredential(identityID uuid.UUID, passwordHash string) *PasswordCredential {
	now := time.Now()
	return &PasswordCredential{
		ID:            uuid.New(),
		IdentityID:    identityID,
		PasswordHash:  passwordHash,
		CreatedAt:     now,
		UpdatedAt:     now,
		LastChangedAt: now,
	}
}

// UpdatePassword updates the password hash and timestamps.
func (p *PasswordCredential) UpdatePassword(newHash string) {
	now := time.Now()
	p.PasswordHash = newHash
	p.UpdatedAt = now
	p.LastChangedAt = now
}
