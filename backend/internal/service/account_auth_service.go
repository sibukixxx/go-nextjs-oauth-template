// Package service contains business logic for the application.
package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/your-org/go-nextjs-oauth-template/backend/internal/domain/model"
	"github.com/your-org/go-nextjs-oauth-template/backend/internal/domain/repository"
)

// AccountAuthService handles authentication with Account/Identity model
type AccountAuthService struct {
	multiAuth *MultiAuthService

	accountRepo      repository.AccountRepository
	identityRepo     repository.IdentityRepository
	authSessionRepo  repository.AuthSessionRepository
	refreshTokenRepo repository.RefreshTokenRepository
	auditLogRepo     repository.AuthAuditLogRepository

	jwtService *JWTService

	logger *slog.Logger
}

// AccountAuthServiceConfig holds configuration for AccountAuthService
type AccountAuthServiceConfig struct {
	MultiAuthService *MultiAuthService

	AccountRepo      repository.AccountRepository
	IdentityRepo     repository.IdentityRepository
	AuthSessionRepo  repository.AuthSessionRepository
	RefreshTokenRepo repository.RefreshTokenRepository
	AuditLogRepo     repository.AuthAuditLogRepository

	JWTService *JWTService
	Logger     *slog.Logger
}

// NewAccountAuthService creates a new AccountAuthService
func NewAccountAuthService(cfg AccountAuthServiceConfig) (*AccountAuthService, error) {
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}

	if cfg.MultiAuthService == nil {
		return nil, fmt.Errorf("MultiAuthService is required")
	}

	return &AccountAuthService{
		multiAuth:        cfg.MultiAuthService,
		accountRepo:      cfg.AccountRepo,
		identityRepo:     cfg.IdentityRepo,
		authSessionRepo:  cfg.AuthSessionRepo,
		refreshTokenRepo: cfg.RefreshTokenRepo,
		auditLogRepo:     cfg.AuditLogRepo,
		jwtService:       cfg.JWTService,
		logger:           cfg.Logger,
	}, nil
}

// LoginInitRequest represents a login initiation request
type LoginInitRequest struct {
	Provider    string `json:"provider,omitempty"`
	RedirectURI string `json:"redirect_uri,omitempty"`
}

// LoginInitResponse represents the response for login initiation
type LoginInitResponse struct {
	AuthorizationURL string `json:"authorization_url"`
	State            string `json:"state"`
	Provider         string `json:"provider"`
	SessionID        string `json:"session_id"`
}

// InitiateLogin starts the OAuth login flow
func (s *AccountAuthService) InitiateLogin(ctx context.Context, req LoginInitRequest) (*LoginInitResponse, error) {
	// Use the existing MultiAuthService to initiate login
	multiReq := MultiLoginRequest{
		Provider:    req.Provider,
		RedirectURI: req.RedirectURI,
	}

	resp, err := s.multiAuth.InitiateLogin(ctx, multiReq)
	if err != nil {
		return nil, fmt.Errorf("failed to initiate login: %w", err)
	}

	// Create auth session in database
	providerStr := resp.Provider
	session := &model.AuthSession{
		ID:          uuid.New(),
		SessionType: model.AuthSessionTypeOAuthLogin,
		Provider:    &providerStr,
		State:       resp.State,
		ExpiresAt:   time.Now().Add(10 * time.Minute),
		CreatedAt:   time.Now(),
	}

	if s.authSessionRepo != nil {
		if err := s.authSessionRepo.Create(ctx, session); err != nil {
			s.logger.Warn("failed to store auth session in database", "error", err)
			// Continue anyway - we have in-memory session from MultiAuthService
		}
	}

	return &LoginInitResponse{
		AuthorizationURL: resp.AuthorizationURL,
		State:            resp.State,
		Provider:         resp.Provider,
		SessionID:        resp.SessionID,
	}, nil
}

// LoginCallbackRequest represents the OAuth callback request
type LoginCallbackRequest struct {
	Code      string `json:"code"`
	State     string `json:"state"`
	SessionID string `json:"session_id"`
	// Device info for refresh token
	UserAgent string `json:"user_agent,omitempty"`
	IPAddress string `json:"ip_address,omitempty"`
}

// LoginCallbackResponse represents the response after successful login
type LoginCallbackResponse struct {
	AccessToken  string         `json:"access_token"`
	TokenType    string         `json:"token_type"`
	ExpiresIn    int            `json:"expires_in"`
	RefreshToken string         `json:"refresh_token,omitempty"`
	Account      *model.Account `json:"account"`
	IsNewAccount bool           `json:"is_new_account"`
}

// HandleLoginCallback handles the OAuth callback and creates/updates account
func (s *AccountAuthService) HandleLoginCallback(ctx context.Context, req LoginCallbackRequest) (*LoginCallbackResponse, error) {
	// Use MultiAuthService to handle the OAuth callback
	multiReq := MultiCallbackRequest{
		Code:      req.Code,
		State:     req.State,
		SessionID: req.SessionID,
	}

	result, err := s.multiAuth.HandleCallback(ctx, multiReq)
	if err != nil {
		provider := ""
		if result != nil {
			provider = result.Provider
		}
		s.logAuditEvent(ctx, nil, model.AuditActionLoginFailed, provider, req.IPAddress, req.UserAgent, false, err.Error())
		return nil, fmt.Errorf("OAuth callback failed: %w", err)
	}

	// Extract user info from the result
	providerSubject := ""
	email := ""
	displayName := ""
	avatarURL := ""

	if result.Claims != nil {
		providerSubject = result.Claims.Subject
		// Email, Name, Picture are in Extra map for OpenID Connect claims
		if emailVal, ok := result.Claims.Extra["email"].(string); ok {
			email = emailVal
		}
		if nameVal, ok := result.Claims.Extra["name"].(string); ok {
			displayName = nameVal
		}
		if pictureVal, ok := result.Claims.Extra["picture"].(string); ok {
			avatarURL = pictureVal
		}
	}

	if result.UserInfo != nil {
		if providerSubject == "" {
			providerSubject = result.UserInfo.ID
		}
		if email == "" {
			email = result.UserInfo.Email
		}
		if displayName == "" {
			displayName = result.UserInfo.Name
		}
		if avatarURL == "" {
			avatarURL = result.UserInfo.Picture
		}
	}

	if providerSubject == "" {
		return nil, fmt.Errorf("provider did not return a subject")
	}

	// Parse provider type
	providerType, err := model.ParseIdentityProvider(result.Provider)
	if err != nil {
		return nil, fmt.Errorf("unknown provider: %s", result.Provider)
	}

	// Find existing identity
	identity, err := s.identityRepo.GetByProvider(ctx, providerType, providerSubject)
	if err != nil {
		return nil, fmt.Errorf("failed to lookup identity: %w", err)
	}

	var account *model.Account
	isNewAccount := false

	if identity != nil {
		// Existing identity found - update login info and get account
		identity.RecordLogin()
		if err := s.identityRepo.Update(ctx, identity); err != nil {
			s.logger.Warn("failed to update identity login info", "error", err)
		}

		account, err = s.accountRepo.GetByID(ctx, identity.AccountID)
		if err != nil {
			return nil, fmt.Errorf("failed to get account: %w", err)
		}
		if account == nil {
			return nil, fmt.Errorf("account not found for identity")
		}
	} else {
		// New user - create account and identity
		isNewAccount = true

		// Create account
		account = model.NewAccount()
		if email != "" {
			account.SetPrimaryEmail(email)
		}
		if displayName != "" {
			account.DisplayName = &displayName
		}
		if avatarURL != "" {
			account.AvatarURL = &avatarURL
		}

		if err := s.accountRepo.Create(ctx, account); err != nil {
			return nil, fmt.Errorf("failed to create account: %w", err)
		}

		// Create identity
		identity = model.NewIdentity(account.ID, providerType, providerSubject)
		if email != "" {
			identity.Email = &email
		}
		identity.RecordLogin()

		if err := s.identityRepo.Create(ctx, identity); err != nil {
			// Rollback account creation would be ideal here
			return nil, fmt.Errorf("failed to create identity: %w", err)
		}

		s.logAuditEvent(ctx, &account.ID, model.AuditActionAccountCreate, result.Provider, req.IPAddress, req.UserAgent, true, "")
	}

	// Generate our own tokens
	deviceInfo := model.DeviceInfo{
		UserAgent: req.UserAgent,
		IPAddress: req.IPAddress,
	}

	// Create access token
	accessToken := ""
	expiresIn := 600 // 10 minutes
	if s.jwtService != nil {
		accessToken, err = s.jwtService.GenerateAccessToken(account.ID, identity.ID, identity.Provider)
		if err != nil {
			return nil, fmt.Errorf("failed to generate access token: %w", err)
		}
		expiresIn = s.jwtService.GetAccessTokenTTL()
	}

	// Create refresh token (14 days TTL)
	refreshTokenRaw := ""
	if s.refreshTokenRepo != nil {
		refreshToken, raw := model.NewRefreshToken(account.ID, 14*24*time.Hour, deviceInfo)
		if err := s.refreshTokenRepo.Create(ctx, refreshToken); err != nil {
			s.logger.Warn("failed to store refresh token", "error", err)
		} else {
			refreshTokenRaw = raw
		}
	}

	s.logAuditEvent(ctx, &account.ID, model.AuditActionLogin, result.Provider, req.IPAddress, req.UserAgent, true, "")

	return &LoginCallbackResponse{
		AccessToken:  accessToken,
		TokenType:    "Bearer",
		ExpiresIn:    expiresIn,
		RefreshToken: refreshTokenRaw,
		Account:      account,
		IsNewAccount: isNewAccount,
	}, nil
}

// LinkProviderRequest represents a request to link a new provider
type LinkProviderRequest struct {
	AccountID   uuid.UUID `json:"-"` // From authenticated user
	Provider    string    `json:"provider"`
	RedirectURI string    `json:"redirect_uri,omitempty"`
}

// InitiateLinkProvider starts the OAuth flow to link a new provider
func (s *AccountAuthService) InitiateLinkProvider(ctx context.Context, req LinkProviderRequest) (*LoginInitResponse, error) {
	// Check if account already has this provider
	providerType, err := model.ParseIdentityProvider(req.Provider)
	if err != nil {
		return nil, fmt.Errorf("invalid provider: %w", err)
	}

	identities, err := s.identityRepo.GetByAccountID(ctx, req.AccountID)
	if err != nil {
		return nil, fmt.Errorf("failed to get identities: %w", err)
	}

	for _, id := range identities {
		if id.Provider == providerType {
			return nil, fmt.Errorf("account already linked to %s", req.Provider)
		}
	}

	// Initiate OAuth flow with link intent
	multiReq := MultiLoginRequest{
		Provider:    req.Provider,
		RedirectURI: req.RedirectURI,
	}

	resp, err := s.multiAuth.InitiateLogin(ctx, multiReq)
	if err != nil {
		return nil, fmt.Errorf("failed to initiate link: %w", err)
	}

	// Store auth session with link intent
	providerStr := resp.Provider
	session := &model.AuthSession{
		ID:          uuid.New(),
		SessionType: model.AuthSessionTypeOAuthLink,
		Provider:    &providerStr,
		State:       resp.State,
		AccountID:   &req.AccountID,
		ExpiresAt:   time.Now().Add(10 * time.Minute),
		CreatedAt:   time.Now(),
	}

	if s.authSessionRepo != nil {
		if err := s.authSessionRepo.Create(ctx, session); err != nil {
			s.logger.Warn("failed to store link session", "error", err)
		}
	}

	return &LoginInitResponse{
		AuthorizationURL: resp.AuthorizationURL,
		State:            resp.State,
		Provider:         resp.Provider,
	}, nil
}

// HandleLinkCallback handles the OAuth callback for linking
func (s *AccountAuthService) HandleLinkCallback(ctx context.Context, accountID uuid.UUID, req LoginCallbackRequest) (*model.Identity, error) {
	// Handle OAuth callback
	multiReq := MultiCallbackRequest{
		Code:      req.Code,
		State:     req.State,
		SessionID: req.SessionID,
	}

	result, err := s.multiAuth.HandleCallback(ctx, multiReq)
	if err != nil {
		return nil, fmt.Errorf("OAuth callback failed: %w", err)
	}

	// Extract subject
	providerSubject := ""
	email := ""

	if result.Claims != nil {
		providerSubject = result.Claims.Subject
		if emailVal, ok := result.Claims.Extra["email"].(string); ok {
			email = emailVal
		}
	}
	if providerSubject == "" && result.UserInfo != nil {
		providerSubject = result.UserInfo.ID
	}
	if email == "" && result.UserInfo != nil {
		email = result.UserInfo.Email
	}

	if providerSubject == "" {
		return nil, fmt.Errorf("provider did not return a subject")
	}

	providerType, err := model.ParseIdentityProvider(result.Provider)
	if err != nil {
		return nil, fmt.Errorf("unknown provider: %s", result.Provider)
	}

	// Check if this identity is already linked to another account
	existing, err := s.identityRepo.GetByProvider(ctx, providerType, providerSubject)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing identity: %w", err)
	}

	if existing != nil {
		if existing.AccountID == accountID {
			return nil, fmt.Errorf("this provider is already linked to your account")
		}
		return nil, fmt.Errorf("this provider is already linked to another account")
	}

	// Create new identity
	identity := model.NewIdentity(accountID, providerType, providerSubject)
	if email != "" {
		identity.Email = &email
	}

	if err := s.identityRepo.Create(ctx, identity); err != nil {
		return nil, fmt.Errorf("failed to create identity: %w", err)
	}

	s.logAuditEvent(ctx, &accountID, model.AuditActionProviderLink, result.Provider, req.IPAddress, req.UserAgent, true, "")

	return identity, nil
}

// UnlinkProvider unlinks a provider from an account
func (s *AccountAuthService) UnlinkProvider(ctx context.Context, accountID uuid.UUID, provider string, ipAddress, userAgent string) error {
	providerType, err := model.ParseIdentityProvider(provider)
	if err != nil {
		return fmt.Errorf("invalid provider: %w", err)
	}

	// Get all identities for account
	identities, err := s.identityRepo.GetByAccountID(ctx, accountID)
	if err != nil {
		return fmt.Errorf("failed to get identities: %w", err)
	}

	// Find the identity to unlink
	var identityToUnlink *model.Identity
	for _, id := range identities {
		if id.Provider == providerType {
			identityToUnlink = &id
			break
		}
	}

	if identityToUnlink == nil {
		return fmt.Errorf("provider %s is not linked to this account", provider)
	}

	// Prevent unlinking the last identity
	if len(identities) <= 1 {
		return fmt.Errorf("cannot unlink the last authentication method")
	}

	// Delete the identity
	if err := s.identityRepo.Delete(ctx, identityToUnlink.ID); err != nil {
		return fmt.Errorf("failed to unlink provider: %w", err)
	}

	s.logAuditEvent(ctx, &accountID, model.AuditActionProviderUnlink, provider, ipAddress, userAgent, true, "")

	return nil
}

// RefreshAccessToken refreshes an access token using a refresh token
func (s *AccountAuthService) RefreshAccessToken(ctx context.Context, refreshTokenRaw, ipAddress, userAgent string) (*LoginCallbackResponse, error) {
	if s.refreshTokenRepo == nil {
		return nil, fmt.Errorf("refresh token storage not configured")
	}

	tokenHash := model.HashToken(refreshTokenRaw)
	token, err := s.refreshTokenRepo.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		return nil, fmt.Errorf("failed to lookup refresh token: %w", err)
	}

	if token == nil {
		return nil, fmt.Errorf("refresh token not found")
	}

	if !token.IsValid() {
		// Token is revoked or expired - potential token theft
		if token.RevokedAt != nil {
			// This might be a replay attack - revoke all tokens in the family
			s.logger.Warn("refresh token reuse detected", "account_id", token.AccountID)
			_ = s.refreshTokenRepo.RevokeAllByAccountID(ctx, token.AccountID, "token_reuse_detected")
		}
		return nil, fmt.Errorf("refresh token is invalid or expired")
	}

	// Get account
	account, err := s.accountRepo.GetByID(ctx, token.AccountID)
	if err != nil {
		return nil, fmt.Errorf("failed to get account: %w", err)
	}
	if account == nil {
		return nil, fmt.Errorf("account not found")
	}

	// Get primary identity for the account
	identities, err := s.identityRepo.GetByAccountID(ctx, token.AccountID)
	if err != nil || len(identities) == 0 {
		return nil, fmt.Errorf("no identity found for account")
	}
	primaryIdentity := identities[0]

	// Rotate the refresh token
	deviceInfo := model.DeviceInfo{
		UserAgent: userAgent,
		IPAddress: ipAddress,
	}

	newToken, newTokenRaw := token.Rotate(14*24*time.Hour, deviceInfo)

	// Update old token as rotated
	if err := s.refreshTokenRepo.Update(ctx, token); err != nil {
		return nil, fmt.Errorf("failed to update old refresh token: %w", err)
	}

	// Create new token
	if err := s.refreshTokenRepo.Create(ctx, newToken); err != nil {
		return nil, fmt.Errorf("failed to create new refresh token: %w", err)
	}

	// Generate new access token
	accessToken := ""
	expiresIn := 600 // 10 minutes default
	if s.jwtService != nil {
		accessToken, err = s.jwtService.GenerateAccessToken(account.ID, primaryIdentity.ID, primaryIdentity.Provider)
		if err != nil {
			return nil, fmt.Errorf("failed to generate access token: %w", err)
		}
		expiresIn = s.jwtService.GetAccessTokenTTL()
	}

	s.logAuditEvent(ctx, &account.ID, model.AuditActionTokenRefresh, string(primaryIdentity.Provider), ipAddress, userAgent, true, "")

	return &LoginCallbackResponse{
		AccessToken:  accessToken,
		TokenType:    "Bearer",
		ExpiresIn:    expiresIn,
		RefreshToken: newTokenRaw,
		Account:      account,
		IsNewAccount: false,
	}, nil
}

// Logout revokes refresh tokens and logs the logout
func (s *AccountAuthService) Logout(ctx context.Context, accountID uuid.UUID, refreshTokenRaw, ipAddress, userAgent string) error {
	if refreshTokenRaw != "" && s.refreshTokenRepo != nil {
		tokenHash := model.HashToken(refreshTokenRaw)
		token, err := s.refreshTokenRepo.GetByTokenHash(ctx, tokenHash)
		if err == nil && token != nil {
			_ = s.refreshTokenRepo.Revoke(ctx, token.ID, "logout")
		}
	}

	s.logAuditEvent(ctx, &accountID, model.AuditActionLogout, "", ipAddress, userAgent, true, "")

	return nil
}

// LogoutAll revokes all refresh tokens for an account
func (s *AccountAuthService) LogoutAll(ctx context.Context, accountID uuid.UUID, ipAddress, userAgent string) error {
	if s.refreshTokenRepo != nil {
		if err := s.refreshTokenRepo.RevokeAllByAccountID(ctx, accountID, "logout_all"); err != nil {
			return fmt.Errorf("failed to revoke all tokens: %w", err)
		}
	}

	s.logAuditEvent(ctx, &accountID, model.AuditActionLogout, "", ipAddress, userAgent, true, "logout_all")

	return nil
}

// GetAvailableProviders returns the list of available providers
func (s *AccountAuthService) GetAvailableProviders() []string {
	return s.multiAuth.GetAvailableProviders()
}

// logAuditEvent logs an authentication audit event
func (s *AccountAuthService) logAuditEvent(ctx context.Context, accountID *uuid.UUID, action, provider, ipAddress, userAgent string, success bool, failureReason string) {
	if s.auditLogRepo == nil {
		return
	}

	log := model.NewAuditLog(accountID, action, success)
	if provider != "" {
		log.Provider = &provider
	}
	if ipAddress != "" {
		log.IPAddress = &ipAddress
	}
	if userAgent != "" {
		log.UserAgent = &userAgent
	}
	if failureReason != "" {
		log.FailureReason = &failureReason
	}

	if err := s.auditLogRepo.Create(ctx, log); err != nil {
		s.logger.Warn("failed to create audit log", "error", err)
	}
}
