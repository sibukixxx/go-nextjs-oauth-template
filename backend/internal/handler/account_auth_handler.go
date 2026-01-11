// Package handler contains HTTP handlers for the API.
package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/your-org/go-nextjs-oauth-template/backend/internal/service"
)

// AccountAuthHandler handles authentication API endpoints
type AccountAuthHandler struct {
	authService  *service.AccountAuthService
	jwtService   *service.JWTService
	logger       *slog.Logger
	isProduction bool
}

// NewAccountAuthHandler creates a new AccountAuthHandler
func NewAccountAuthHandler(authService *service.AccountAuthService, jwtService *service.JWTService, logger *slog.Logger, isProduction bool) *AccountAuthHandler {
	if logger == nil {
		logger = slog.Default()
	}
	return &AccountAuthHandler{
		authService:  authService,
		jwtService:   jwtService,
		logger:       logger,
		isProduction: isProduction,
	}
}

// RegisterRoutes registers the handler routes to the given mux
func (h *AccountAuthHandler) RegisterRoutes(mux *http.ServeMux) {
	// Public endpoints
	mux.HandleFunc("GET /api/v1/auth/providers", h.GetProviders)
	mux.HandleFunc("POST /api/v1/auth/login", h.InitiateLogin)
	mux.HandleFunc("POST /api/v1/auth/callback", h.HandleCallback)
	mux.HandleFunc("POST /api/v1/auth/refresh", h.RefreshToken)

	// Protected endpoints (require authentication)
	mux.HandleFunc("POST /api/v1/auth/logout", h.Logout)
	mux.HandleFunc("POST /api/v1/auth/logout/all", h.LogoutAll)
	mux.HandleFunc("GET /api/v1/auth/me", h.GetMe)
	mux.HandleFunc("POST /api/v1/auth/link", h.InitiateLinkProvider)
	mux.HandleFunc("POST /api/v1/auth/link/callback", h.HandleLinkCallback)
	mux.HandleFunc("DELETE /api/v1/auth/link/{provider}", h.UnlinkProvider)
}

// GetProviders returns the list of available authentication providers
func (h *AccountAuthHandler) GetProviders(w http.ResponseWriter, r *http.Request) {
	providers := h.authService.GetAvailableProviders()

	resp := ProvidersResponse{
		Providers: providers,
	}

	h.writeJSON(w, http.StatusOK, resp)
}

// ProvidersResponse represents the response for available providers
type ProvidersResponse struct {
	Providers []string `json:"providers"`
}

// LoginRequest represents a login initiation request
type LoginRequest struct {
	Provider    string `json:"provider,omitempty"`
	RedirectURI string `json:"redirect_uri,omitempty"`
}

// LoginResponse represents the response for login initiation
type LoginResponse struct {
	AuthorizationURL string `json:"authorization_url"`
	State            string `json:"state"`
	Provider         string `json:"provider"`
}

// InitiateLogin starts the OAuth login flow
func (h *AccountAuthHandler) InitiateLogin(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	result, err := h.authService.InitiateLogin(r.Context(), service.LoginInitRequest{
		Provider:    req.Provider,
		RedirectURI: req.RedirectURI,
	})
	if err != nil {
		h.logger.Error("failed to initiate login", "error", err)
		h.writeError(w, http.StatusInternalServerError, "Failed to initiate login")
		return
	}

	// Set session_id in HTTPOnly cookie for security
	// This prevents XSS attacks from accessing the session data
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_session",
		Value:    result.SessionID,
		Path:     "/",
		MaxAge:   600, // 10 minutes - OAuth flow should complete within this time
		HttpOnly: true,
		Secure:   r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https",
		SameSite: http.SameSiteLaxMode, // Lax allows redirect back from OAuth provider
	})

	h.writeJSON(w, http.StatusOK, LoginResponse{
		AuthorizationURL: result.AuthorizationURL,
		State:            result.State,
		Provider:         result.Provider,
	})
}

// CallbackRequest represents the OAuth callback request
type CallbackRequest struct {
	Code      string `json:"code"`
	State     string `json:"state"`
	SessionID string `json:"session_id"`
}

// CallbackResponse represents the response after successful authentication
type CallbackResponse struct {
	AccessToken  string       `json:"access_token"`
	TokenType    string       `json:"token_type"`
	ExpiresIn    int          `json:"expires_in"`
	RefreshToken string       `json:"refresh_token,omitempty"`
	Account      *AccountInfo `json:"account"`
	IsNewAccount bool         `json:"is_new_account"`
}

// AccountInfo represents account information in the response
type AccountInfo struct {
	ID           string   `json:"id"`
	DisplayName  *string  `json:"display_name,omitempty"`
	PrimaryEmail *string  `json:"primary_email,omitempty"`
	AvatarURL    *string  `json:"avatar_url,omitempty"`
	Providers    []string `json:"providers,omitempty"`
}

// HandleCallback handles the OAuth callback
func (h *AccountAuthHandler) HandleCallback(w http.ResponseWriter, r *http.Request) {
	var req CallbackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Code == "" || req.State == "" {
		h.writeError(w, http.StatusBadRequest, "Missing required parameters")
		return
	}

	// Get session_id from cookie if not provided in request body
	// This is more secure as HTTPOnly cookies prevent XSS access
	sessionID := req.SessionID
	if sessionID == "" {
		if cookie, err := r.Cookie("oauth_session"); err == nil {
			sessionID = cookie.Value
		}
	}
	if sessionID == "" {
		h.writeError(w, http.StatusBadRequest, "Missing session ID")
		return
	}

	result, err := h.authService.HandleLoginCallback(r.Context(), service.LoginCallbackRequest{
		Code:      req.Code,
		State:     req.State,
		SessionID: sessionID,
		UserAgent: r.UserAgent(),
		IPAddress: getClientIP(r),
	})
	if err != nil {
		h.logger.Error("failed to handle callback", "error", err)
		h.writeError(w, http.StatusUnauthorized, "Authentication failed")
		return
	}

	// Build account info
	accountInfo := &AccountInfo{
		ID: result.Account.ID.String(),
	}
	if result.Account.DisplayName != nil {
		accountInfo.DisplayName = result.Account.DisplayName
	}
	if result.Account.PrimaryEmail != nil {
		accountInfo.PrimaryEmail = result.Account.PrimaryEmail
	}
	if result.Account.AvatarURL != nil {
		accountInfo.AvatarURL = result.Account.AvatarURL
	}

	// Clear the OAuth session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_session",
		Value:    "",
		Path:     "/",
		MaxAge:   -1, // Delete the cookie
		HttpOnly: true,
	})

	// Set refresh token as HTTP-only cookie (14 days)
	if result.RefreshToken != "" {
		sameSite := http.SameSiteLaxMode
		if h.isProduction {
			sameSite = http.SameSiteStrictMode
		}
		http.SetCookie(w, &http.Cookie{
			Name:     "refresh_token",
			Value:    result.RefreshToken,
			Path:     "/",
			HttpOnly: true,
			Secure:   h.isProduction,
			SameSite: sameSite,
			MaxAge:   14 * 24 * 60 * 60, // 14 days
		})
	}

	h.writeJSON(w, http.StatusOK, CallbackResponse{
		AccessToken:  result.AccessToken,
		TokenType:    result.TokenType,
		ExpiresIn:    result.ExpiresIn,
		RefreshToken: result.RefreshToken,
		Account:      accountInfo,
		IsNewAccount: result.IsNewAccount,
	})
}

// RefreshRequest represents a token refresh request
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token,omitempty"`
}

// RefreshToken refreshes an access token
func (h *AccountAuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var refreshToken string

	// Try to get refresh token from request body
	var req RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err == nil {
		refreshToken = req.RefreshToken
	}

	// Fallback to cookie
	if refreshToken == "" {
		cookie, err := r.Cookie("refresh_token")
		if err == nil {
			refreshToken = cookie.Value
		}
	}

	if refreshToken == "" {
		h.writeError(w, http.StatusBadRequest, "Refresh token is required")
		return
	}

	result, err := h.authService.RefreshAccessToken(r.Context(), refreshToken, getClientIP(r), r.UserAgent())
	if err != nil {
		h.logger.Error("failed to refresh token", "error", err)
		h.writeError(w, http.StatusUnauthorized, "Failed to refresh token")
		return
	}

	// Update refresh token cookie (14 days)
	if result.RefreshToken != "" {
		sameSite := http.SameSiteLaxMode
		if h.isProduction {
			sameSite = http.SameSiteStrictMode
		}
		http.SetCookie(w, &http.Cookie{
			Name:     "refresh_token",
			Value:    result.RefreshToken,
			Path:     "/",
			HttpOnly: true,
			Secure:   h.isProduction,
			SameSite: sameSite,
			MaxAge:   14 * 24 * 60 * 60, // 14 days
		})
	}

	accountInfo := &AccountInfo{
		ID: result.Account.ID.String(),
	}

	h.writeJSON(w, http.StatusOK, CallbackResponse{
		AccessToken:  result.AccessToken,
		TokenType:    result.TokenType,
		ExpiresIn:    result.ExpiresIn,
		RefreshToken: result.RefreshToken,
		Account:      accountInfo,
		IsNewAccount: false,
	})
}

// Logout logs out the current session
func (h *AccountAuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	accountID, err := h.getAuthenticatedAccountID(r)
	if err != nil {
		h.writeError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	// Get refresh token from cookie
	var refreshToken string
	cookie, err := r.Cookie("refresh_token")
	if err == nil {
		refreshToken = cookie.Value
	}

	if err := h.authService.Logout(r.Context(), accountID, refreshToken, getClientIP(r), r.UserAgent()); err != nil {
		h.logger.Error("failed to logout", "error", err)
	}

	// Clear refresh token cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   h.isProduction,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})

	w.WriteHeader(http.StatusNoContent)
}

// LogoutAll logs out all sessions for the current user
func (h *AccountAuthHandler) LogoutAll(w http.ResponseWriter, r *http.Request) {
	accountID, err := h.getAuthenticatedAccountID(r)
	if err != nil {
		h.writeError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	if err := h.authService.LogoutAll(r.Context(), accountID, getClientIP(r), r.UserAgent()); err != nil {
		h.logger.Error("failed to logout all sessions", "error", err)
		h.writeError(w, http.StatusInternalServerError, "Failed to logout all sessions")
		return
	}

	// Clear refresh token cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   h.isProduction,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})

	w.WriteHeader(http.StatusNoContent)
}

// GetMe returns the current authenticated user's information
func (h *AccountAuthHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	claims, err := h.getAuthenticatedClaims(r)
	if err != nil {
		h.writeError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	accountInfo := &AccountInfo{
		ID: claims.AccountID.String(),
	}

	h.writeJSON(w, http.StatusOK, accountInfo)
}

// LinkProviderRequest represents a request to link a new provider
type LinkProviderRequest struct {
	Provider    string `json:"provider"`
	RedirectURI string `json:"redirect_uri,omitempty"`
}

// InitiateLinkProvider starts the OAuth flow to link a new provider
func (h *AccountAuthHandler) InitiateLinkProvider(w http.ResponseWriter, r *http.Request) {
	accountID, err := h.getAuthenticatedAccountID(r)
	if err != nil {
		h.writeError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	var req LinkProviderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Provider == "" {
		h.writeError(w, http.StatusBadRequest, "Provider is required")
		return
	}

	result, err := h.authService.InitiateLinkProvider(r.Context(), service.LinkProviderRequest{
		AccountID:   accountID,
		Provider:    req.Provider,
		RedirectURI: req.RedirectURI,
	})
	if err != nil {
		h.logger.Error("failed to initiate link", "error", err)
		h.writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.writeJSON(w, http.StatusOK, LoginResponse{
		AuthorizationURL: result.AuthorizationURL,
		State:            result.State,
		Provider:         result.Provider,
	})
}

// HandleLinkCallback handles the OAuth callback for linking
func (h *AccountAuthHandler) HandleLinkCallback(w http.ResponseWriter, r *http.Request) {
	accountID, err := h.getAuthenticatedAccountID(r)
	if err != nil {
		h.writeError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	var req CallbackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	identity, err := h.authService.HandleLinkCallback(r.Context(), accountID, service.LoginCallbackRequest{
		Code:      req.Code,
		State:     req.State,
		SessionID: req.SessionID,
		UserAgent: r.UserAgent(),
		IPAddress: getClientIP(r),
	})
	if err != nil {
		h.logger.Error("failed to link provider", "error", err)
		h.writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.writeJSON(w, http.StatusOK, IdentityInfo{
		ID:       identity.ID.String(),
		Provider: string(identity.Provider),
		Email:    identity.Email,
	})
}

// IdentityInfo represents identity information in the response
type IdentityInfo struct {
	ID       string  `json:"id"`
	Provider string  `json:"provider"`
	Email    *string `json:"email,omitempty"`
}

// UnlinkProvider unlinks a provider from the account
func (h *AccountAuthHandler) UnlinkProvider(w http.ResponseWriter, r *http.Request) {
	accountID, err := h.getAuthenticatedAccountID(r)
	if err != nil {
		h.writeError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	provider := r.PathValue("provider")
	if provider == "" {
		h.writeError(w, http.StatusBadRequest, "Provider is required")
		return
	}

	if err := h.authService.UnlinkProvider(r.Context(), accountID, provider, getClientIP(r), r.UserAgent()); err != nil {
		h.logger.Error("failed to unlink provider", "error", err)
		h.writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Helper methods

func (h *AccountAuthHandler) getAuthenticatedAccountID(r *http.Request) (uuid.UUID, error) {
	claims, err := h.getAuthenticatedClaims(r)
	if err != nil {
		return uuid.Nil, err
	}
	return claims.AccountID, nil
}

func (h *AccountAuthHandler) getAuthenticatedClaims(r *http.Request) (*service.AccessTokenClaims, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return nil, http.ErrNoCookie
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return nil, http.ErrNoCookie
	}

	return h.jwtService.ValidateAccessToken(parts[1])
}

func (h *AccountAuthHandler) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("failed to encode response", "error", err)
	}
}

func (h *AccountAuthHandler) writeError(w http.ResponseWriter, status int, message string) {
	h.writeJSON(w, status, map[string]string{"error": message})
}

func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}

	// Check X-Real-IP header
	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	return r.RemoteAddr
}
