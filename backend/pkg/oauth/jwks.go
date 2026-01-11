// Package oauth provides OAuth 2.1 compliant authentication utilities.
// This package implements JWKs fetching with automatic rotation and JWT validation.
package oauth

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"math/big"
	"net/http"
	"sync"
	"time"
)

// JWK represents a JSON Web Key
type JWK struct {
	Kty string `json:"kty"` // Key Type (RSA, EC, etc.)
	Kid string `json:"kid"` // Key ID
	Use string `json:"use"` // Key Use (sig, enc)
	Alg string `json:"alg"` // Algorithm (RS256, etc.)
	N   string `json:"n"`   // RSA modulus
	E   string `json:"e"`   // RSA exponent
}

// JWKS represents a JSON Web Key Set
type JWKS struct {
	Keys []JWK `json:"keys"`
}

// JWKSClient handles fetching and caching of JWKs
type JWKSClient struct {
	jwksURL     string
	httpClient  *http.Client
	cache       *jwksCache
	refreshRate time.Duration
	mu          sync.RWMutex
	logger      *slog.Logger
}

// jwksCache stores cached JWKs with expiration
type jwksCache struct {
	jwks      *JWKS
	keys      map[string]*rsa.PublicKey
	fetchedAt time.Time
	expiresAt time.Time
}

// JWKSConfig holds configuration for JWKSClient
type JWKSConfig struct {
	JWKSURL     string        // URL to fetch JWKs from
	RefreshRate time.Duration // How often to refresh JWKs (default: 1 hour)
	CacheTTL    time.Duration // Cache TTL (default: 24 hours)
	HTTPTimeout time.Duration // HTTP request timeout (default: 10 seconds)
	Logger      *slog.Logger
}

// NewJWKSClient creates a new JWKs client with automatic rotation
func NewJWKSClient(cfg JWKSConfig) (*JWKSClient, error) {
	if cfg.JWKSURL == "" {
		return nil, fmt.Errorf("JWKS URL is required")
	}

	if cfg.RefreshRate == 0 {
		cfg.RefreshRate = 1 * time.Hour
	}
	if cfg.CacheTTL == 0 {
		cfg.CacheTTL = 24 * time.Hour
	}
	if cfg.HTTPTimeout == 0 {
		cfg.HTTPTimeout = 10 * time.Second
	}
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}

	client := &JWKSClient{
		jwksURL: cfg.JWKSURL,
		httpClient: &http.Client{
			Timeout: cfg.HTTPTimeout,
		},
		refreshRate: cfg.RefreshRate,
		logger:      cfg.Logger,
	}

	// Initial fetch
	if err := client.refresh(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to fetch initial JWKs: %w", err)
	}

	// Start background refresh goroutine
	go client.backgroundRefresh(cfg.CacheTTL)

	return client, nil
}

// GetPublicKey returns the RSA public key for the given key ID
func (c *JWKSClient) GetPublicKey(kid string) (*rsa.PublicKey, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.cache == nil || c.cache.keys == nil {
		return nil, fmt.Errorf("JWKs cache is not initialized")
	}

	key, ok := c.cache.keys[kid]
	if !ok {
		return nil, fmt.Errorf("key with kid %q not found in JWKs", kid)
	}

	return key, nil
}

// GetAllKeys returns all cached public keys
func (c *JWKSClient) GetAllKeys() map[string]*rsa.PublicKey {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.cache == nil {
		return nil
	}

	// Return a copy to prevent external modification
	result := make(map[string]*rsa.PublicKey, len(c.cache.keys))
	for k, v := range c.cache.keys {
		result[k] = v
	}
	return result
}

// ForceRefresh forces an immediate refresh of the JWKs cache
func (c *JWKSClient) ForceRefresh(ctx context.Context) error {
	return c.refresh(ctx)
}

// refresh fetches JWKs from the remote URL and updates the cache
func (c *JWKSClient) refresh(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.jwksURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to fetch JWKs: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var jwks JWKS
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return fmt.Errorf("failed to decode JWKs: %w", err)
	}

	// Parse keys
	keys := make(map[string]*rsa.PublicKey)
	for _, jwk := range jwks.Keys {
		if jwk.Kty != "RSA" {
			c.logger.Debug("skipping non-RSA key", "kid", jwk.Kid, "kty", jwk.Kty)
			continue
		}

		pubKey, err := parseRSAPublicKey(jwk)
		if err != nil {
			c.logger.Warn("failed to parse RSA key", "kid", jwk.Kid, "error", err)
			continue
		}

		keys[jwk.Kid] = pubKey
	}

	if len(keys) == 0 {
		return fmt.Errorf("no valid RSA keys found in JWKs")
	}

	// Update cache
	c.mu.Lock()
	c.cache = &jwksCache{
		jwks:      &jwks,
		keys:      keys,
		fetchedAt: time.Now(),
		expiresAt: time.Now().Add(24 * time.Hour),
	}
	c.mu.Unlock()

	c.logger.Info("JWKs cache refreshed", "keys_count", len(keys))
	return nil
}

// backgroundRefresh periodically refreshes the JWKs cache
func (c *JWKSClient) backgroundRefresh(cacheTTL time.Duration) {
	ticker := time.NewTicker(c.refreshRate)
	defer ticker.Stop()

	for range ticker.C {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		if err := c.refresh(ctx); err != nil {
			c.logger.Error("background JWKs refresh failed", "error", err)
		}
		cancel()
	}
}

// parseRSAPublicKey converts a JWK to an RSA public key
func parseRSAPublicKey(jwk JWK) (*rsa.PublicKey, error) {
	// Decode modulus (n)
	nBytes, err := base64.RawURLEncoding.DecodeString(jwk.N)
	if err != nil {
		return nil, fmt.Errorf("failed to decode modulus: %w", err)
	}

	// Decode exponent (e)
	eBytes, err := base64.RawURLEncoding.DecodeString(jwk.E)
	if err != nil {
		return nil, fmt.Errorf("failed to decode exponent: %w", err)
	}

	// Convert exponent bytes to int
	var e int
	for _, b := range eBytes {
		e = e<<8 + int(b)
	}

	return &rsa.PublicKey{
		N: new(big.Int).SetBytes(nBytes),
		E: e,
	}, nil
}
