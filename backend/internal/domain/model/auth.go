package model

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"time"

	"github.com/google/uuid"
)

// AuthSessionType represents the type of authentication session
type AuthSessionType string

const (
	AuthSessionTypeOAuthLogin    AuthSessionType = "oauth_login"
	AuthSessionTypeOAuthLink     AuthSessionType = "oauth_link"
	AuthSessionTypeEmailVerify   AuthSessionType = "email_verify"
	AuthSessionTypePasswordReset AuthSessionType = "password_reset"
)

// AuthSession represents a temporary authentication session
// Used for OAuth state, email verification, etc.
type AuthSession struct {
	ID           uuid.UUID       `json:"id"`
	SessionType  AuthSessionType `json:"session_type"`
	Provider     *string         `json:"provider,omitempty"`
	State        string          `json:"state"`
	Nonce        *string         `json:"nonce,omitempty"`
	CodeVerifier *string         `json:"code_verifier,omitempty"` // PKCE
	RedirectURI  *string         `json:"redirect_uri,omitempty"`
	AccountID    *uuid.UUID      `json:"account_id,omitempty"` // For link operations
	ExpiresAt    time.Time       `json:"expires_at"`
	UsedAt       *time.Time      `json:"used_at,omitempty"`
	CreatedAt    time.Time       `json:"created_at"`
}

// NewAuthSession creates a new auth session for OAuth login
func NewAuthSession(sessionType AuthSessionType, provider string, ttl time.Duration) *AuthSession {
	now := time.Now()
	state := generateSecureToken(32)
	nonce := generateSecureToken(32)

	return &AuthSession{
		ID:          uuid.New(),
		SessionType: sessionType,
		Provider:    &provider,
		State:       state,
		Nonce:       &nonce,
		ExpiresAt:   now.Add(ttl),
		CreatedAt:   now,
	}
}

// NewLinkSession creates a new auth session for linking a provider
func NewLinkSession(provider string, accountID uuid.UUID, ttl time.Duration) *AuthSession {
	session := NewAuthSession(AuthSessionTypeOAuthLink, provider, ttl)
	session.AccountID = &accountID
	return session
}

// IsExpired returns true if the session has expired
func (s *AuthSession) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

// IsUsed returns true if the session has been used
func (s *AuthSession) IsUsed() bool {
	return s.UsedAt != nil
}

// MarkUsed marks the session as used
func (s *AuthSession) MarkUsed() {
	now := time.Now()
	s.UsedAt = &now
}

// SetPKCE sets the PKCE code verifier
func (s *AuthSession) SetPKCE(codeVerifier string) {
	s.CodeVerifier = &codeVerifier
}

// RefreshToken represents a refresh token for session management
type RefreshToken struct {
	ID              uuid.UUID  `json:"id"`
	AccountID       uuid.UUID  `json:"account_id"`
	TokenHash       string     `json:"-"` // Never expose
	DeviceInfo      DeviceInfo `json:"device_info"`
	ExpiresAt       time.Time  `json:"expires_at"`
	RevokedAt       *time.Time `json:"revoked_at,omitempty"`
	RevokedReason   *string    `json:"revoked_reason,omitempty"`
	PreviousTokenID *uuid.UUID `json:"previous_token_id,omitempty"`
	RotatedAt       *time.Time `json:"rotated_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
}

// DeviceInfo represents device information for a refresh token
type DeviceInfo struct {
	UserAgent string `json:"user_agent,omitempty"`
	IPAddress string `json:"ip_address,omitempty"`
	Device    string `json:"device,omitempty"`
	OS        string `json:"os,omitempty"`
	Browser   string `json:"browser,omitempty"`
}

// NewRefreshToken creates a new refresh token
func NewRefreshToken(accountID uuid.UUID, ttl time.Duration, deviceInfo DeviceInfo) (*RefreshToken, string) {
	now := time.Now()
	rawToken := generateSecureToken(48)
	tokenHash := hashToken(rawToken)

	return &RefreshToken{
		ID:         uuid.New(),
		AccountID:  accountID,
		TokenHash:  tokenHash,
		DeviceInfo: deviceInfo,
		ExpiresAt:  now.Add(ttl),
		CreatedAt:  now,
	}, rawToken
}

// IsValid returns true if the token is valid (not expired or revoked)
func (t *RefreshToken) IsValid() bool {
	return t.RevokedAt == nil && time.Now().Before(t.ExpiresAt)
}

// Revoke revokes the token with a reason
func (t *RefreshToken) Revoke(reason string) {
	now := time.Now()
	t.RevokedAt = &now
	t.RevokedReason = &reason
}

// Rotate creates a new token and marks this one as rotated
func (t *RefreshToken) Rotate(ttl time.Duration, deviceInfo DeviceInfo) (*RefreshToken, string) {
	newToken, rawToken := NewRefreshToken(t.AccountID, ttl, deviceInfo)
	newToken.PreviousTokenID = &t.ID

	now := time.Now()
	t.RotatedAt = &now

	return newToken, rawToken
}

// VerifyToken verifies if the raw token matches this token's hash
func (t *RefreshToken) VerifyToken(rawToken string) bool {
	return hashToken(rawToken) == t.TokenHash
}

// AuthAuditLog represents an authentication audit log entry
type AuthAuditLog struct {
	ID            uuid.UUID   `json:"id"`
	AccountID     *uuid.UUID  `json:"account_id,omitempty"`
	Action        string      `json:"action"`
	Provider      *string     `json:"provider,omitempty"`
	IPAddress     *string     `json:"ip_address,omitempty"`
	UserAgent     *string     `json:"user_agent,omitempty"`
	Success       bool        `json:"success"`
	FailureReason *string     `json:"failure_reason,omitempty"`
	Metadata      interface{} `json:"metadata,omitempty"`
	CreatedAt     time.Time   `json:"created_at"`
}

// Audit action constants
const (
	AuditActionLogin          = "login"
	AuditActionLogout         = "logout"
	AuditActionLoginFailed    = "login_failed"
	AuditActionTokenRefresh   = "token_refresh"
	AuditActionTokenRevoke    = "token_revoke"
	AuditActionProviderLink   = "provider_link"
	AuditActionProviderUnlink = "provider_unlink"
	AuditActionEmailChange    = "email_change"
	AuditActionEmailVerify    = "email_verify"
	AuditActionProfileUpdate  = "profile_update"
	AuditActionAccountCreate  = "account_create"
	AuditActionAccountDelete  = "account_delete"

	// Password authentication actions
	AuditActionPasswordRegister = "password_register"
	AuditActionPasswordLogin    = "password_login"
	AuditActionPasswordReset    = "password_reset"
	AuditActionPasswordChange   = "password_change"
)

// NewAuditLog creates a new audit log entry
func NewAuditLog(accountID *uuid.UUID, action string, success bool) *AuthAuditLog {
	return &AuthAuditLog{
		ID:        uuid.New(),
		AccountID: accountID,
		Action:    action,
		Success:   success,
		CreatedAt: time.Now(),
	}
}

// generateSecureToken generates a cryptographically secure random token
func generateSecureToken(length int) string {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		panic(err) // Should never happen
	}
	return base64.URLEncoding.EncodeToString(bytes)
}

// hashToken creates a SHA-256 hash of the token
func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// HashToken exports the hashToken function for external use
func HashToken(token string) string {
	return hashToken(token)
}

// NewEmailVerificationSession creates a new session for email verification.
func NewEmailVerificationSession(accountID *uuid.UUID, ttl time.Duration) *AuthSession {
	now := time.Now()
	token := generateSecureToken(32)

	return &AuthSession{
		ID:          uuid.New(),
		SessionType: AuthSessionTypeEmailVerify,
		State:       token,
		AccountID:   accountID,
		ExpiresAt:   now.Add(ttl),
		CreatedAt:   now,
	}
}

// NewPasswordResetSession creates a new session for password reset.
func NewPasswordResetSession(accountID uuid.UUID, ttl time.Duration) *AuthSession {
	now := time.Now()
	token := generateSecureToken(32)

	return &AuthSession{
		ID:          uuid.New(),
		SessionType: AuthSessionTypePasswordReset,
		State:       token,
		AccountID:   &accountID,
		ExpiresAt:   now.Add(ttl),
		CreatedAt:   now,
	}
}
