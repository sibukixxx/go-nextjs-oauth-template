package oauth

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"fmt"
	"hash"
)

// verifySignature verifies the JWT signature using RSA
func verifySignature(signingInput, signature string, publicKey *rsa.PublicKey, algorithm string) error {
	sigBytes, err := base64.RawURLEncoding.DecodeString(signature)
	if err != nil {
		return fmt.Errorf("failed to decode signature: %w", err)
	}

	// Select hash algorithm based on JWT algorithm
	var hashFunc crypto.Hash
	var hasher hash.Hash

	switch algorithm {
	case "RS256":
		hashFunc = crypto.SHA256
		hasher = sha256.New()
	case "RS384":
		hashFunc = crypto.SHA384
		hasher = sha512.New384()
	case "RS512":
		hashFunc = crypto.SHA512
		hasher = sha512.New()
	default:
		return fmt.Errorf("unsupported algorithm: %s", algorithm)
	}

	hasher.Write([]byte(signingInput))
	hashed := hasher.Sum(nil)

	err = rsa.VerifyPKCS1v15(publicKey, hashFunc, hashed, sigBytes)
	if err != nil {
		return fmt.Errorf("signature verification failed: %w", err)
	}

	return nil
}

// PKCE implements Proof Key for Code Exchange (OAuth 2.1 requirement)
type PKCE struct {
	CodeVerifier  string
	CodeChallenge string
	Method        string // S256 (recommended) or plain
}

// GeneratePKCE generates a new PKCE code verifier and challenge
// OAuth 2.1 requires PKCE for all clients
func GeneratePKCE() (*PKCE, error) {
	// Generate a random code verifier (43-128 characters)
	verifierBytes := make([]byte, 32)
	if _, err := rand.Read(verifierBytes); err != nil {
		return nil, fmt.Errorf("failed to generate random bytes: %w", err)
	}

	codeVerifier := base64.RawURLEncoding.EncodeToString(verifierBytes)

	// Generate code challenge using S256 method
	h := sha256.New()
	h.Write([]byte(codeVerifier))
	codeChallenge := base64.RawURLEncoding.EncodeToString(h.Sum(nil))

	return &PKCE{
		CodeVerifier:  codeVerifier,
		CodeChallenge: codeChallenge,
		Method:        "S256",
	}, nil
}

// VerifyPKCE verifies a PKCE code verifier against a code challenge
func VerifyPKCE(codeVerifier, codeChallenge, method string) bool {
	switch method {
	case "S256":
		h := sha256.New()
		h.Write([]byte(codeVerifier))
		expected := base64.RawURLEncoding.EncodeToString(h.Sum(nil))
		return expected == codeChallenge
	case "plain":
		return codeVerifier == codeChallenge
	default:
		return false
	}
}

// GenerateState generates a cryptographically secure state parameter
// OAuth 2.1 requires state for CSRF protection
func GenerateState() (string, error) {
	stateBytes := make([]byte, 32)
	if _, err := rand.Read(stateBytes); err != nil {
		return "", fmt.Errorf("failed to generate state: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(stateBytes), nil
}

// GenerateNonce generates a cryptographically secure nonce for OpenID Connect
func GenerateNonce() (string, error) {
	nonceBytes := make([]byte, 32)
	if _, err := rand.Read(nonceBytes); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(nonceBytes), nil
}
