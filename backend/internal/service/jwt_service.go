// Package service contains business logic for the application.
package service

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/your-org/go-nextjs-oauth-template/backend/internal/domain/model"
)

// JWTService handles JWT token generation and validation
type JWTService struct {
	issuer         string
	audience       string
	accessTokenTTL time.Duration
	signingKey     []byte
	signingMethod  jwt.SigningMethod
}

// JWTServiceConfig holds configuration for JWTService
type JWTServiceConfig struct {
	Issuer         string        // JWT issuer claim
	Audience       string        // JWT audience claim
	AccessTokenTTL time.Duration // Access token TTL (default: 10 minutes)
	SigningKey     string        // HMAC signing key (for HS256)
}

// AccessTokenClaims represents the claims in an access token
type AccessTokenClaims struct {
	jwt.RegisteredClaims
	AccountID  uuid.UUID              `json:"account_id"`
	IdentityID uuid.UUID              `json:"identity_id"`
	Provider   model.IdentityProvider `json:"provider"`
	TokenType  string                 `json:"token_type"`
}

// NewJWTService creates a new JWT service
func NewJWTService(cfg JWTServiceConfig) (*JWTService, error) {
	if cfg.SigningKey == "" {
		return nil, fmt.Errorf("signing key is required")
	}

	if cfg.Issuer == "" {
		cfg.Issuer = "go-nextjs-oauth-template"
	}

	if cfg.Audience == "" {
		cfg.Audience = "go-nextjs-oauth-template-api"
	}

	if cfg.AccessTokenTTL == 0 {
		cfg.AccessTokenTTL = 10 * time.Minute
	}

	return &JWTService{
		issuer:         cfg.Issuer,
		audience:       cfg.Audience,
		accessTokenTTL: cfg.AccessTokenTTL,
		signingKey:     []byte(cfg.SigningKey),
		signingMethod:  jwt.SigningMethodHS256,
	}, nil
}

// GenerateAccessToken generates a new access token for an account
func (s *JWTService) GenerateAccessToken(accountID, identityID uuid.UUID, provider model.IdentityProvider) (string, error) {
	now := time.Now()
	jti, err := generateJTI()
	if err != nil {
		return "", fmt.Errorf("failed to generate token ID: %w", err)
	}

	claims := AccessTokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.issuer,
			Subject:   accountID.String(),
			Audience:  jwt.ClaimStrings{s.audience},
			ExpiresAt: jwt.NewNumericDate(now.Add(s.accessTokenTTL)),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        jti,
		},
		AccountID:  accountID,
		IdentityID: identityID,
		Provider:   provider,
		TokenType:  "access",
	}

	token := jwt.NewWithClaims(s.signingMethod, claims)
	signedToken, err := token.SignedString(s.signingKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return signedToken, nil
}

// ValidateAccessToken validates an access token and returns its claims
func (s *JWTService) ValidateAccessToken(tokenString string) (*AccessTokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &AccessTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if token.Method.Alg() != s.signingMethod.Alg() {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.signingKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*AccessTokenClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	// Verify issuer
	if claims.Issuer != s.issuer {
		return nil, fmt.Errorf("invalid issuer")
	}

	// Verify audience
	audienceValid := false
	for _, aud := range claims.Audience {
		if aud == s.audience {
			audienceValid = true
			break
		}
	}
	if !audienceValid {
		return nil, fmt.Errorf("invalid audience")
	}

	// Verify token type
	if claims.TokenType != "access" {
		return nil, fmt.Errorf("invalid token type")
	}

	return claims, nil
}

// GetAccessTokenTTL returns the access token TTL in seconds
func (s *JWTService) GetAccessTokenTTL() int {
	return int(s.accessTokenTTL.Seconds())
}

// generateJTI generates a unique token ID
func generateJTI() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
