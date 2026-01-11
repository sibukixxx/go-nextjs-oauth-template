package oauth

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// Claims represents the standard JWT claims plus OAuth 2.1 specific claims
type Claims struct {
	// Standard claims
	Issuer    string   `json:"iss"`
	Subject   string   `json:"sub"`
	Audience  Audience `json:"aud"`
	ExpiresAt int64    `json:"exp"`
	NotBefore int64    `json:"nbf,omitempty"`
	IssuedAt  int64    `json:"iat,omitempty"`
	JWTID     string   `json:"jti,omitempty"`

	// OAuth 2.1 specific claims
	Scope           string `json:"scope,omitempty"`     // Space-separated scopes
	AuthorizedParty string `json:"azp,omitempty"`       // Authorized Party (client_id)
	ClientID        string `json:"client_id,omitempty"` // Client ID (alternative to azp)

	// OpenID Connect claims (optional)
	Nonce  string `json:"nonce,omitempty"`
	AtHash string `json:"at_hash,omitempty"`
	CHash  string `json:"c_hash,omitempty"`

	// Custom claims storage
	Extra map[string]interface{} `json:"-"`
}

// Audience can be a string or array of strings
type Audience []string

// UnmarshalJSON handles both string and array audience claims
func (a *Audience) UnmarshalJSON(data []byte) error {
	var single string
	if err := json.Unmarshal(data, &single); err == nil {
		*a = Audience{single}
		return nil
	}

	var multi []string
	if err := json.Unmarshal(data, &multi); err != nil {
		return err
	}
	*a = Audience(multi)
	return nil
}

// Contains checks if the audience contains the given value
func (a Audience) Contains(aud string) bool {
	for _, v := range a {
		if v == aud {
			return true
		}
	}
	return false
}

// JWTHeader represents the JWT header
type JWTHeader struct {
	Algorithm string `json:"alg"`
	Type      string `json:"typ"`
	KeyID     string `json:"kid"`
}

// JWTValidator validates JWTs according to OAuth 2.1 specification
type JWTValidator struct {
	jwksClient        *JWKSClient
	issuer            string
	audiences         []string
	authorizedParties []string // Valid azp values (client_ids)
	requiredScopes    []string
	clockSkew         time.Duration
}

// JWTValidatorConfig holds configuration for JWTValidator
type JWTValidatorConfig struct {
	JWKSClient        *JWKSClient
	Issuer            string        // Expected issuer
	Audiences         []string      // Expected audiences
	AuthorizedParties []string      // Valid azp values (client_ids)
	RequiredScopes    []string      // Required scopes (at least one must be present)
	ClockSkew         time.Duration // Allowed clock skew (default: 1 minute)
}

// NewJWTValidator creates a new JWT validator
func NewJWTValidator(cfg JWTValidatorConfig) (*JWTValidator, error) {
	if cfg.JWKSClient == nil {
		return nil, fmt.Errorf("JWKSClient is required")
	}
	if cfg.Issuer == "" {
		return nil, fmt.Errorf("issuer is required")
	}
	if len(cfg.Audiences) == 0 {
		return nil, fmt.Errorf("at least one audience is required (OAuth 2.1)")
	}

	if cfg.ClockSkew == 0 {
		cfg.ClockSkew = 1 * time.Minute
	}

	return &JWTValidator{
		jwksClient:        cfg.JWKSClient,
		issuer:            cfg.Issuer,
		audiences:         cfg.Audiences,
		authorizedParties: cfg.AuthorizedParties,
		requiredScopes:    cfg.RequiredScopes,
		clockSkew:         cfg.ClockSkew,
	}, nil
}

// ValidationResult contains the validated claims and any warnings
type ValidationResult struct {
	Claims   *Claims
	Header   *JWTHeader
	Warnings []string
}

// Validate validates a JWT token and returns the claims
func (v *JWTValidator) Validate(tokenString string) (*ValidationResult, error) {
	// Split the token
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid JWT format: expected 3 parts, got %d", len(parts))
	}

	// Decode header
	header, err := decodeHeader(parts[0])
	if err != nil {
		return nil, fmt.Errorf("failed to decode header: %w", err)
	}

	// Validate algorithm (OAuth 2.1 recommends RS256 or ES256)
	if header.Algorithm != "RS256" && header.Algorithm != "RS384" && header.Algorithm != "RS512" {
		return nil, fmt.Errorf("unsupported algorithm: %s (expected RS256, RS384, or RS512)", header.Algorithm)
	}

	// Get public key
	var publicKey *rsa.PublicKey
	if header.KeyID != "" {
		publicKey, err = v.jwksClient.GetPublicKey(header.KeyID)
		if err != nil {
			return nil, fmt.Errorf("failed to get public key: %w", err)
		}
	} else {
		// If no kid, try all keys
		keys := v.jwksClient.GetAllKeys()
		for _, key := range keys {
			if err := verifySignature(parts[0]+"."+parts[1], parts[2], key, header.Algorithm); err == nil {
				publicKey = key
				break
			}
		}
		if publicKey == nil {
			return nil, fmt.Errorf("no matching key found for token")
		}
	}

	// Verify signature
	if err := verifySignature(parts[0]+"."+parts[1], parts[2], publicKey, header.Algorithm); err != nil {
		return nil, fmt.Errorf("signature verification failed: %w", err)
	}

	// Decode claims
	claims, err := decodeClaims(parts[1])
	if err != nil {
		return nil, fmt.Errorf("failed to decode claims: %w", err)
	}

	// Validate claims
	warnings, err := v.validateClaims(claims)
	if err != nil {
		return nil, err
	}

	return &ValidationResult{
		Claims:   claims,
		Header:   header,
		Warnings: warnings,
	}, nil
}

// validateClaims validates the JWT claims according to OAuth 2.1
func (v *JWTValidator) validateClaims(claims *Claims) ([]string, error) {
	var warnings []string
	now := time.Now()

	// 1. Validate issuer (REQUIRED)
	if claims.Issuer != v.issuer {
		return nil, fmt.Errorf("invalid issuer: expected %q, got %q", v.issuer, claims.Issuer)
	}

	// 2. Validate audience (REQUIRED in OAuth 2.1)
	audienceValid := false
	for _, expectedAud := range v.audiences {
		if claims.Audience.Contains(expectedAud) {
			audienceValid = true
			break
		}
	}
	if !audienceValid {
		return nil, fmt.Errorf("invalid audience: token audience %v does not contain any of %v",
			claims.Audience, v.audiences)
	}

	// 3. Validate expiration (REQUIRED)
	if claims.ExpiresAt == 0 {
		return nil, fmt.Errorf("expiration claim (exp) is required")
	}
	expTime := time.Unix(claims.ExpiresAt, 0)
	if now.After(expTime.Add(v.clockSkew)) {
		return nil, fmt.Errorf("token has expired at %v", expTime)
	}

	// 4. Validate not before (if present)
	if claims.NotBefore != 0 {
		nbfTime := time.Unix(claims.NotBefore, 0)
		if now.Before(nbfTime.Add(-v.clockSkew)) {
			return nil, fmt.Errorf("token is not yet valid (nbf: %v)", nbfTime)
		}
	}

	// 5. Validate issued at (if present)
	if claims.IssuedAt != 0 {
		iatTime := time.Unix(claims.IssuedAt, 0)
		if iatTime.After(now.Add(v.clockSkew)) {
			warnings = append(warnings, fmt.Sprintf("token issued in the future (iat: %v)", iatTime))
		}
	}

	// 6. Validate authorized party (azp) - OAuth 2.1 requirement
	if len(claims.Audience) > 1 && claims.AuthorizedParty == "" {
		return nil, fmt.Errorf("azp claim is required when audience contains multiple values")
	}

	// Get the effective client ID (azp or client_id)
	clientID := claims.AuthorizedParty
	if clientID == "" {
		clientID = claims.ClientID
	}

	if len(v.authorizedParties) > 0 && clientID != "" {
		azpValid := false
		for _, validAzp := range v.authorizedParties {
			if clientID == validAzp {
				azpValid = true
				break
			}
		}
		if !azpValid {
			return nil, fmt.Errorf("unauthorized party: %q is not in the list of authorized parties", clientID)
		}
	}

	// 7. Validate scope (OAuth 2.1 - scope is important for resource access)
	if len(v.requiredScopes) > 0 {
		tokenScopes := strings.Fields(claims.Scope)
		if len(tokenScopes) == 0 {
			return nil, fmt.Errorf("scope claim is required but not present")
		}

		scopeFound := false
		for _, required := range v.requiredScopes {
			for _, tokenScope := range tokenScopes {
				if required == tokenScope {
					scopeFound = true
					break
				}
			}
			if scopeFound {
				break
			}
		}
		if !scopeFound {
			return nil, fmt.Errorf("insufficient scope: token has %v, requires one of %v",
				tokenScopes, v.requiredScopes)
		}
	}

	return warnings, nil
}

// HasScope checks if the claims contain a specific scope
func (c *Claims) HasScope(scope string) bool {
	scopes := strings.Fields(c.Scope)
	for _, s := range scopes {
		if s == scope {
			return true
		}
	}
	return false
}

// GetScopes returns all scopes as a slice
func (c *Claims) GetScopes() []string {
	if c.Scope == "" {
		return nil
	}
	return strings.Fields(c.Scope)
}

// decodeHeader decodes the JWT header
func decodeHeader(encoded string) (*JWTHeader, error) {
	decoded, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return nil, err
	}

	var header JWTHeader
	if err := json.Unmarshal(decoded, &header); err != nil {
		return nil, err
	}

	return &header, nil
}

// decodeClaims decodes the JWT claims
func decodeClaims(encoded string) (*Claims, error) {
	decoded, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return nil, err
	}

	var claims Claims
	if err := json.Unmarshal(decoded, &claims); err != nil {
		return nil, err
	}

	// Also unmarshal into a map to capture extra claims
	var extra map[string]interface{}
	if err := json.Unmarshal(decoded, &extra); err == nil {
		// Remove standard claims from extra
		delete(extra, "iss")
		delete(extra, "sub")
		delete(extra, "aud")
		delete(extra, "exp")
		delete(extra, "nbf")
		delete(extra, "iat")
		delete(extra, "jti")
		delete(extra, "scope")
		delete(extra, "azp")
		delete(extra, "client_id")
		delete(extra, "nonce")
		delete(extra, "at_hash")
		delete(extra, "c_hash")
		claims.Extra = extra
	}

	return &claims, nil
}
