// Package crypto provides cryptographic utilities for password handling.
package crypto

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

const (
	// DefaultCost is the bcrypt cost factor (12 is recommended for 2024+)
	DefaultCost = 12
	// MinPasswordLength is the minimum password length
	MinPasswordLength = 8
	// MaxPasswordLength is the maximum password length (bcrypt limit is 72 bytes)
	MaxPasswordLength = 72
)

// PasswordHasher handles password hashing and verification using bcrypt.
type PasswordHasher struct {
	cost int
}

// NewPasswordHasher creates a new password hasher with the specified cost.
// If cost is outside valid bcrypt range, DefaultCost is used.
func NewPasswordHasher(cost int) *PasswordHasher {
	if cost < bcrypt.MinCost || cost > bcrypt.MaxCost {
		cost = DefaultCost
	}
	return &PasswordHasher{cost: cost}
}

// DefaultPasswordHasher returns a hasher with default settings.
func DefaultPasswordHasher() *PasswordHasher {
	return NewPasswordHasher(DefaultCost)
}

// HashPassword hashes a password using bcrypt.
// Returns an error if password length is invalid.
func (h *PasswordHasher) HashPassword(password string) (string, error) {
	if len(password) < MinPasswordLength {
		return "", fmt.Errorf("password must be at least %d characters", MinPasswordLength)
	}
	if len(password) > MaxPasswordLength {
		return "", fmt.Errorf("password must be at most %d characters", MaxPasswordLength)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), h.cost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hash), nil
}

// VerifyPassword checks if a password matches the hash.
// Returns nil on success, or an error if the password doesn't match.
func (h *PasswordHasher) VerifyPassword(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

// NeedsRehash checks if the password hash needs to be rehashed.
// This is useful when the cost factor has been changed.
func (h *PasswordHasher) NeedsRehash(hash string) bool {
	cost, err := bcrypt.Cost([]byte(hash))
	if err != nil {
		return true
	}
	return cost != h.cost
}
