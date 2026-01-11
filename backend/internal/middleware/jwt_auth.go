// Package middleware contains HTTP middleware for the API.
package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/your-org/go-nextjs-oauth-template/backend/internal/service"
)

// Context keys for authentication data
type contextKey string

const (
	// ContextKeyAccountID is the context key for the authenticated account ID
	ContextKeyAccountID contextKey = "account_id"
	// ContextKeyIdentityID is the context key for the authenticated identity ID
	ContextKeyIdentityID contextKey = "identity_id"
	// ContextKeyProvider is the context key for the authentication provider
	ContextKeyProvider contextKey = "provider"
	// ContextKeyClaims is the context key for the full JWT claims
	ContextKeyClaims contextKey = "claims"
)

// JWTAuthMiddleware validates JWT tokens and injects account_id into context
type JWTAuthMiddleware struct {
	jwtService *service.JWTService
	logger     *slog.Logger
}

// NewJWTAuthMiddleware creates a new JWT authentication middleware
func NewJWTAuthMiddleware(jwtService *service.JWTService, logger *slog.Logger) *JWTAuthMiddleware {
	if logger == nil {
		logger = slog.Default()
	}
	return &JWTAuthMiddleware{
		jwtService: jwtService,
		logger:     logger,
	}
}

// Middleware returns the HTTP middleware function
func (m *JWTAuthMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			m.writeUnauthorized(w, "Authorization header is required")
			return
		}

		// Expect "Bearer <token>"
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			m.writeUnauthorized(w, "Invalid authorization header format")
			return
		}

		tokenString := parts[1]
		if tokenString == "" {
			m.writeUnauthorized(w, "Token is required")
			return
		}

		// Validate token
		claims, err := m.jwtService.ValidateAccessToken(tokenString)
		if err != nil {
			m.logger.Debug("token validation failed", "error", err)
			m.writeUnauthorized(w, "Invalid or expired token")
			return
		}

		// Inject claims into context
		ctx := r.Context()
		ctx = context.WithValue(ctx, ContextKeyAccountID, claims.AccountID)
		ctx = context.WithValue(ctx, ContextKeyIdentityID, claims.IdentityID)
		ctx = context.WithValue(ctx, ContextKeyProvider, claims.Provider)
		ctx = context.WithValue(ctx, ContextKeyClaims, claims)

		// Continue with the next handler
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireAuth wraps a handler function with JWT authentication
func (m *JWTAuthMiddleware) RequireAuth(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m.Middleware(http.HandlerFunc(handler)).ServeHTTP(w, r)
	}
}

// Optional returns a middleware that allows both authenticated and unauthenticated requests
// If a valid token is present, it will inject the claims into context
// If no token or invalid token, request continues without authentication
func (m *JWTAuthMiddleware) Optional(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			next.ServeHTTP(w, r)
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			next.ServeHTTP(w, r)
			return
		}

		tokenString := parts[1]
		if tokenString == "" {
			next.ServeHTTP(w, r)
			return
		}

		claims, err := m.jwtService.ValidateAccessToken(tokenString)
		if err != nil {
			// Token is invalid but we allow the request to continue
			next.ServeHTTP(w, r)
			return
		}

		// Inject claims into context
		ctx := r.Context()
		ctx = context.WithValue(ctx, ContextKeyAccountID, claims.AccountID)
		ctx = context.WithValue(ctx, ContextKeyIdentityID, claims.IdentityID)
		ctx = context.WithValue(ctx, ContextKeyProvider, claims.Provider)
		ctx = context.WithValue(ctx, ContextKeyClaims, claims)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// writeUnauthorized writes a 401 Unauthorized response
func (m *JWTAuthMiddleware) writeUnauthorized(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("WWW-Authenticate", `Bearer realm="go-nextjs-oauth-template-api"`)
	w.WriteHeader(http.StatusUnauthorized)
	_, _ = w.Write([]byte(`{"error":"` + message + `"}`))
}

// GetAccountID extracts the account ID from the request context
// Returns uuid.Nil if not authenticated
func GetAccountID(ctx context.Context) uuid.UUID {
	if id, ok := ctx.Value(ContextKeyAccountID).(uuid.UUID); ok {
		return id
	}
	return uuid.Nil
}

// GetIdentityID extracts the identity ID from the request context
// Returns uuid.Nil if not authenticated
func GetIdentityID(ctx context.Context) uuid.UUID {
	if id, ok := ctx.Value(ContextKeyIdentityID).(uuid.UUID); ok {
		return id
	}
	return uuid.Nil
}

// GetClaims extracts the full JWT claims from the request context
// Returns nil if not authenticated
func GetClaims(ctx context.Context) *service.AccessTokenClaims {
	if claims, ok := ctx.Value(ContextKeyClaims).(*service.AccessTokenClaims); ok {
		return claims
	}
	return nil
}

// IsAuthenticated returns true if the request has valid authentication
func IsAuthenticated(ctx context.Context) bool {
	return GetAccountID(ctx) != uuid.Nil
}
