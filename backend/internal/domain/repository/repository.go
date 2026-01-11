package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/your-org/go-nextjs-oauth-template/backend/internal/domain/model"
)

// AccountRepository defines the interface for account persistence
type AccountRepository interface {
	Create(ctx context.Context, account *model.Account) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Account, error)
	GetByEmail(ctx context.Context, email string) (*model.Account, error)
	Update(ctx context.Context, account *model.Account) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// IdentityRepository defines the interface for identity persistence
type IdentityRepository interface {
	Create(ctx context.Context, identity *model.Identity) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Identity, error)
	GetByProvider(ctx context.Context, provider model.IdentityProvider, subject string) (*model.Identity, error)
	GetByAccountID(ctx context.Context, accountID uuid.UUID) ([]model.Identity, error)
	Update(ctx context.Context, identity *model.Identity) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// AuthSessionRepository defines the interface for auth session persistence
type AuthSessionRepository interface {
	Create(ctx context.Context, session *model.AuthSession) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.AuthSession, error)
	GetByState(ctx context.Context, state string) (*model.AuthSession, error)
	Update(ctx context.Context, session *model.AuthSession) error
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteExpired(ctx context.Context) error
}

// RefreshTokenRepository defines the interface for refresh token persistence
type RefreshTokenRepository interface {
	Create(ctx context.Context, token *model.RefreshToken) error
	GetByTokenHash(ctx context.Context, tokenHash string) (*model.RefreshToken, error)
	Update(ctx context.Context, token *model.RefreshToken) error
	Revoke(ctx context.Context, id uuid.UUID, reason string) error
	RevokeAllByAccountID(ctx context.Context, accountID uuid.UUID, reason string) error
	DeleteExpired(ctx context.Context) error
}

// AuthAuditLogRepository defines the interface for audit log persistence
type AuthAuditLogRepository interface {
	Create(ctx context.Context, log *model.AuthAuditLog) error
	GetByAccountID(ctx context.Context, accountID uuid.UUID, limit, offset int) ([]model.AuthAuditLog, error)
}

// PasswordCredentialRepository defines the interface for password credential persistence
type PasswordCredentialRepository interface {
	Create(ctx context.Context, credential *model.PasswordCredential) error
	GetByIdentityID(ctx context.Context, identityID uuid.UUID) (*model.PasswordCredential, error)
	Update(ctx context.Context, credential *model.PasswordCredential) error
	Delete(ctx context.Context, id uuid.UUID) error
}
