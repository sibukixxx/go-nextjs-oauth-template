package service

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/your-org/go-nextjs-oauth-template/backend/internal/domain/model"
	"github.com/your-org/go-nextjs-oauth-template/backend/internal/domain/repository"
	"github.com/your-org/go-nextjs-oauth-template/backend/pkg/crypto"
)

// PasswordAuthService handles email/password authentication.
type PasswordAuthService struct {
	accountRepo     repository.AccountRepository
	identityRepo    repository.IdentityRepository
	credentialRepo  repository.PasswordCredentialRepository
	authSessionRepo repository.AuthSessionRepository
	auditLogRepo    repository.AuthAuditLogRepository

	passwordHasher *crypto.PasswordHasher
	emailSender    EmailSender
	emailTemplates *EmailTemplates
	logger         *slog.Logger

	emailVerificationTTL time.Duration
	passwordResetTTL     time.Duration
}

// PasswordAuthServiceConfig holds configuration for PasswordAuthService.
type PasswordAuthServiceConfig struct {
	AccountRepo     repository.AccountRepository
	IdentityRepo    repository.IdentityRepository
	CredentialRepo  repository.PasswordCredentialRepository
	AuthSessionRepo repository.AuthSessionRepository
	AuditLogRepo    repository.AuthAuditLogRepository

	PasswordHasher *crypto.PasswordHasher
	EmailSender    EmailSender
	EmailTemplates *EmailTemplates
	Logger         *slog.Logger

	EmailVerificationTTL time.Duration
	PasswordResetTTL     time.Duration
}

// NewPasswordAuthService creates a new password authentication service.
func NewPasswordAuthService(cfg PasswordAuthServiceConfig) *PasswordAuthService {
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}
	if cfg.PasswordHasher == nil {
		cfg.PasswordHasher = crypto.DefaultPasswordHasher()
	}
	if cfg.EmailVerificationTTL == 0 {
		cfg.EmailVerificationTTL = 24 * time.Hour
	}
	if cfg.PasswordResetTTL == 0 {
		cfg.PasswordResetTTL = 1 * time.Hour
	}

	return &PasswordAuthService{
		accountRepo:          cfg.AccountRepo,
		identityRepo:         cfg.IdentityRepo,
		credentialRepo:       cfg.CredentialRepo,
		authSessionRepo:      cfg.AuthSessionRepo,
		auditLogRepo:         cfg.AuditLogRepo,
		passwordHasher:       cfg.PasswordHasher,
		emailSender:          cfg.EmailSender,
		emailTemplates:       cfg.EmailTemplates,
		logger:               cfg.Logger,
		emailVerificationTTL: cfg.EmailVerificationTTL,
		passwordResetTTL:     cfg.PasswordResetTTL,
	}
}

// RegisterRequest represents a password registration request.
type RegisterRequest struct {
	Email    string
	Password string
}

// RegisterResult represents the result of registration.
type RegisterResult struct {
	AccountID uuid.UUID
	Message   string
}

// Register creates a new account with email/password.
func (s *PasswordAuthService) Register(ctx context.Context, req RegisterRequest) (*RegisterResult, error) {
	email := strings.ToLower(strings.TrimSpace(req.Email))

	// Validate email format
	if !isValidEmail(email) {
		return nil, fmt.Errorf("invalid email format")
	}

	// Check if email already exists
	existing, err := s.identityRepo.GetByProvider(ctx, model.IdentityProviderEmail, email)
	if err == nil && existing != nil {
		// Don't reveal that email exists - security best practice
		s.logger.Info("registration attempt for existing email", "email", email)
		return &RegisterResult{
			Message: "If this email is not registered, you will receive a verification email.",
		}, nil
	}

	// Hash password
	passwordHash, err := s.passwordHasher.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("password validation failed: %w", err)
	}

	// Create account
	account := model.NewAccount()
	account.SetPrimaryEmail(email)

	if err := s.accountRepo.Create(ctx, account); err != nil {
		return nil, fmt.Errorf("failed to create account: %w", err)
	}

	// Create identity (provider=email, subject=email)
	identity := model.NewIdentity(account.ID, model.IdentityProviderEmail, email)
	identity.SetEmail(email)

	if err := s.identityRepo.Create(ctx, identity); err != nil {
		return nil, fmt.Errorf("failed to create identity: %w", err)
	}

	// Create password credential
	credential := model.NewPasswordCredential(identity.ID, passwordHash)
	if err := s.credentialRepo.Create(ctx, credential); err != nil {
		return nil, fmt.Errorf("failed to create credential: %w", err)
	}

	// Create email verification session
	session := model.NewEmailVerificationSession(&account.ID, s.emailVerificationTTL)
	if err := s.authSessionRepo.Create(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to create verification session: %w", err)
	}

	// Send verification email
	verificationURL := s.emailTemplates.BaseURL + "/auth/verify-email?token=" + session.State
	emailMsg := s.emailTemplates.BuildVerificationEmail(EmailVerificationData{
		Email:           email,
		VerificationURL: verificationURL,
		ExpiresIn:       "24 hours",
	})

	if err := s.emailSender.Send(ctx, emailMsg); err != nil {
		s.logger.Error("failed to send verification email", "error", err, "email", email)
		// Don't fail registration if email fails
	}

	// Audit log
	s.createAuditLog(ctx, &account.ID, model.AuditActionPasswordRegister, true, nil)

	s.logger.Info("account registered", "account_id", account.ID, "email", email)

	return &RegisterResult{
		AccountID: account.ID,
		Message:   "Registration successful. Please check your email to verify your account.",
	}, nil
}

// LoginRequest represents a password login request.
type LoginRequest struct {
	Email    string
	Password string
}

// LoginResult represents the result of a successful login.
type LoginResult struct {
	AccountID         uuid.UUID
	IdentityID        uuid.UUID
	NeedsVerification bool
}

// Login authenticates a user with email/password.
func (s *PasswordAuthService) Login(ctx context.Context, req LoginRequest) (*LoginResult, error) {
	email := strings.ToLower(strings.TrimSpace(req.Email))

	// Find identity by email
	identity, err := s.identityRepo.GetByProvider(ctx, model.IdentityProviderEmail, email)
	if err != nil {
		s.createAuditLog(ctx, nil, model.AuditActionLoginFailed, false, strPtr("email not found"))
		return nil, fmt.Errorf("invalid email or password")
	}

	// Get account
	account, err := s.accountRepo.GetByID(ctx, identity.AccountID)
	if err != nil || !account.CanLogin() {
		s.createAuditLog(ctx, &identity.AccountID, model.AuditActionLoginFailed, false, strPtr("account not active"))
		return nil, fmt.Errorf("invalid email or password")
	}

	// Get password credential
	credential, err := s.credentialRepo.GetByIdentityID(ctx, identity.ID)
	if err != nil {
		s.createAuditLog(ctx, &account.ID, model.AuditActionLoginFailed, false, strPtr("no password credential"))
		return nil, fmt.Errorf("invalid email or password")
	}

	// Verify password
	if err := s.passwordHasher.VerifyPassword(req.Password, credential.PasswordHash); err != nil {
		s.createAuditLog(ctx, &account.ID, model.AuditActionLoginFailed, false, strPtr("invalid password"))
		return nil, fmt.Errorf("invalid email or password")
	}

	// Check if password needs rehashing
	if s.passwordHasher.NeedsRehash(credential.PasswordHash) {
		newHash, err := s.passwordHasher.HashPassword(req.Password)
		if err == nil {
			credential.UpdatePassword(newHash)
			_ = s.credentialRepo.Update(ctx, credential)
		}
	}

	// Update login stats
	identity.RecordLogin()
	_ = s.identityRepo.Update(ctx, identity)

	// Audit log
	s.createAuditLog(ctx, &account.ID, model.AuditActionPasswordLogin, true, nil)

	s.logger.Info("login successful", "account_id", account.ID)

	return &LoginResult{
		AccountID:         account.ID,
		IdentityID:        identity.ID,
		NeedsVerification: !account.HasVerifiedEmail(),
	}, nil
}

// ForgotPasswordRequest represents a forgot password request.
type ForgotPasswordRequest struct {
	Email string
}

// ForgotPassword initiates the password reset flow.
func (s *PasswordAuthService) ForgotPassword(ctx context.Context, req ForgotPasswordRequest) error {
	email := strings.ToLower(strings.TrimSpace(req.Email))

	// Find identity by email
	identity, err := s.identityRepo.GetByProvider(ctx, model.IdentityProviderEmail, email)
	if err != nil {
		// Don't reveal if email exists
		s.logger.Info("forgot password for non-existent email", "email", email)
		return nil
	}

	account, err := s.accountRepo.GetByID(ctx, identity.AccountID)
	if err != nil {
		return nil
	}

	// Create password reset session
	session := model.NewPasswordResetSession(account.ID, s.passwordResetTTL)
	if err := s.authSessionRepo.Create(ctx, session); err != nil {
		return fmt.Errorf("failed to create reset session: %w", err)
	}

	// Send reset email
	resetURL := s.emailTemplates.BaseURL + "/auth/reset-password?token=" + session.State
	emailMsg := s.emailTemplates.BuildPasswordResetEmail(PasswordResetData{
		Email:     email,
		ResetURL:  resetURL,
		ExpiresIn: "1 hour",
	})

	if err := s.emailSender.Send(ctx, emailMsg); err != nil {
		s.logger.Error("failed to send reset email", "error", err, "email", email)
		return fmt.Errorf("failed to send email")
	}

	s.logger.Info("password reset email sent", "account_id", account.ID)

	return nil
}

// ResetPasswordRequest represents a password reset request.
type ResetPasswordRequest struct {
	Token       string
	NewPassword string
}

// ResetPassword resets the password using a valid token.
func (s *PasswordAuthService) ResetPassword(ctx context.Context, req ResetPasswordRequest) error {
	// Get session by token
	session, err := s.authSessionRepo.GetByState(ctx, req.Token)
	if err != nil || session == nil {
		return fmt.Errorf("invalid or expired reset token")
	}

	// Validate session
	if session.SessionType != model.AuthSessionTypePasswordReset {
		return fmt.Errorf("invalid token type")
	}
	if session.IsExpired() {
		return fmt.Errorf("reset token has expired")
	}
	if session.IsUsed() {
		return fmt.Errorf("reset token has already been used")
	}
	if session.AccountID == nil {
		return fmt.Errorf("invalid reset token")
	}

	// Get identity
	identities, err := s.identityRepo.GetByAccountID(ctx, *session.AccountID)
	if err != nil {
		return fmt.Errorf("account not found")
	}

	var emailIdentity *model.Identity
	for i := range identities {
		if identities[i].Provider == model.IdentityProviderEmail {
			emailIdentity = &identities[i]
			break
		}
	}
	if emailIdentity == nil {
		return fmt.Errorf("email identity not found")
	}

	// Hash new password
	passwordHash, err := s.passwordHasher.HashPassword(req.NewPassword)
	if err != nil {
		return fmt.Errorf("password validation failed: %w", err)
	}

	// Get and update credential
	credential, err := s.credentialRepo.GetByIdentityID(ctx, emailIdentity.ID)
	if err != nil {
		return fmt.Errorf("credential not found")
	}

	credential.UpdatePassword(passwordHash)
	if err := s.credentialRepo.Update(ctx, credential); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// Mark session as used
	session.MarkUsed()
	_ = s.authSessionRepo.Update(ctx, session)

	// Audit log
	s.createAuditLog(ctx, session.AccountID, model.AuditActionPasswordReset, true, nil)

	s.logger.Info("password reset successful", "account_id", session.AccountID)

	return nil
}

// VerifyEmail verifies an email using the token.
func (s *PasswordAuthService) VerifyEmail(ctx context.Context, token string) error {
	session, err := s.authSessionRepo.GetByState(ctx, token)
	if err != nil || session == nil {
		return fmt.Errorf("invalid or expired verification token")
	}

	if session.SessionType != model.AuthSessionTypeEmailVerify {
		return fmt.Errorf("invalid token type")
	}
	if session.IsExpired() {
		return fmt.Errorf("verification token has expired")
	}
	if session.IsUsed() {
		return fmt.Errorf("verification token has already been used")
	}
	if session.AccountID == nil {
		return fmt.Errorf("invalid verification token")
	}

	// Get account and verify email
	account, err := s.accountRepo.GetByID(ctx, *session.AccountID)
	if err != nil {
		return fmt.Errorf("account not found")
	}

	account.VerifyEmail()
	if err := s.accountRepo.Update(ctx, account); err != nil {
		return fmt.Errorf("failed to verify email: %w", err)
	}

	// Mark session as used
	session.MarkUsed()
	_ = s.authSessionRepo.Update(ctx, session)

	// Audit log
	s.createAuditLog(ctx, &account.ID, model.AuditActionEmailVerify, true, nil)

	s.logger.Info("email verified", "account_id", account.ID)

	return nil
}

// ResendVerificationEmail resends the verification email.
func (s *PasswordAuthService) ResendVerificationEmail(ctx context.Context, email string) error {
	email = strings.ToLower(strings.TrimSpace(email))

	// Find identity by email
	identity, err := s.identityRepo.GetByProvider(ctx, model.IdentityProviderEmail, email)
	if err != nil {
		// Don't reveal if email exists
		return nil
	}

	account, err := s.accountRepo.GetByID(ctx, identity.AccountID)
	if err != nil {
		return nil
	}

	// Check if already verified
	if account.HasVerifiedEmail() {
		return nil
	}

	// Create new email verification session
	session := model.NewEmailVerificationSession(&account.ID, s.emailVerificationTTL)
	if err := s.authSessionRepo.Create(ctx, session); err != nil {
		return fmt.Errorf("failed to create verification session: %w", err)
	}

	// Send verification email
	verificationURL := s.emailTemplates.BaseURL + "/auth/verify-email?token=" + session.State
	emailMsg := s.emailTemplates.BuildVerificationEmail(EmailVerificationData{
		Email:           email,
		VerificationURL: verificationURL,
		ExpiresIn:       "24 hours",
	})

	if err := s.emailSender.Send(ctx, emailMsg); err != nil {
		s.logger.Error("failed to send verification email", "error", err, "email", email)
		return fmt.Errorf("failed to send email")
	}

	s.logger.Info("verification email resent", "account_id", account.ID)

	return nil
}

// helper functions

func (s *PasswordAuthService) createAuditLog(ctx context.Context, accountID *uuid.UUID, action string, success bool, failureReason *string) {
	auditLog := model.NewAuditLog(accountID, action, success)
	auditLog.FailureReason = failureReason
	_ = s.auditLogRepo.Create(ctx, auditLog)
}

func strPtr(s string) *string {
	return &s
}

func isValidEmail(email string) bool {
	if len(email) < 5 || len(email) > 254 {
		return false
	}
	atIndex := strings.Index(email, "@")
	if atIndex < 1 {
		return false
	}
	dotIndex := strings.LastIndex(email, ".")
	if dotIndex < atIndex+2 || dotIndex >= len(email)-1 {
		return false
	}
	return true
}
