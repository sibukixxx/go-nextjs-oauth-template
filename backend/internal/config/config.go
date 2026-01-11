package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/joho/godotenv"
)

// Config holds application configuration
type Config struct {
	Server         ServerConfig
	Database       DatabaseConfig
	JWT            JWTConfig
	OAuthProviders OAuthProvidersConfig
	PasswordAuth   PasswordAuthConfig
	Email          EmailConfig
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port           string
	Environment    string
	AllowedOrigins []string
	IsProduction   bool
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	URL string
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	SigningKey     string
	Issuer         string
	Audience       string
	AccessTokenTTL int // seconds
}

// OAuthProvidersConfig holds multi-provider OAuth configuration
type OAuthProvidersConfig struct {
	// Enabled providers (google, line)
	EnabledProviders []string

	// Google OAuth settings
	Google OAuthProviderConfig

	// LINE OAuth settings
	LINE OAuthProviderConfig

	// Default provider
	DefaultProvider string

	// Common redirect URL (used if provider-specific URL not set)
	CommonRedirectURL string
}

// OAuthProviderConfig holds provider-specific OAuth configuration
type OAuthProviderConfig struct {
	// Enabled flag
	Enabled bool

	// Client credentials
	ClientID     string
	ClientSecret string
	RedirectURL  string // Provider-specific redirect URL (optional)

	// Scopes
	Scopes []string

	// Additional options
	Options map[string]string
}

// PasswordAuthConfig holds password authentication configuration
type PasswordAuthConfig struct {
	Enabled              bool
	MinPasswordLength    int // Minimum password length (default: 8)
	BcryptCost           int // Bcrypt cost factor (default: 12)
	EmailVerificationTTL int // Email verification token TTL in seconds (default: 86400 = 24h)
	PasswordResetTTL     int // Password reset token TTL in seconds (default: 3600 = 1h)
}

// EmailConfig holds email service configuration
type EmailConfig struct {
	Provider    string // Email provider: "stub", "smtp", "sendgrid"
	BaseURL     string // Base URL for email links (e.g., password reset)
	FromAddress string // Sender email address
	FromName    string // Sender name
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	// Load .env file
	loadEnvFile()

	config := &Config{
		Server: ServerConfig{
			Port:        getEnv("PORT", "8080"),
			Environment: getEnv("ENVIRONMENT", "development"),
			AllowedOrigins: parseCommaSeparated(
				getEnv("ALLOWED_ORIGINS", "http://localhost:3000"),
			),
		},
		Database: DatabaseConfig{
			URL: getEnv("DATABASE_URL", ""),
		},
		JWT: JWTConfig{
			SigningKey:     getEnv("JWT_SIGNING_KEY", "your-secret-key-change-in-production"),
			Issuer:         getEnv("JWT_ISSUER", "go-nextjs-oauth-template"),
			Audience:       getEnv("JWT_AUDIENCE", "go-nextjs-oauth-template-api"),
			AccessTokenTTL: getEnvAsInt("JWT_ACCESS_TOKEN_TTL", 600), // 10 minutes
		},
		OAuthProviders: OAuthProvidersConfig{
			EnabledProviders:  parseCommaSeparated(getEnv("OAUTH_PROVIDERS", "")),
			DefaultProvider:   getEnv("OAUTH_DEFAULT_PROVIDER", "google"),
			CommonRedirectURL: getEnv("OAUTH_REDIRECT_URL", ""),

			// Google OAuth
			Google: OAuthProviderConfig{
				Enabled:      getEnv("GOOGLE_OAUTH_ENABLED", "false") == "true",
				ClientID:     getEnv("GOOGLE_CLIENT_ID", ""),
				ClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
				RedirectURL:  getEnv("GOOGLE_REDIRECT_URL", ""),
				Scopes:       parseCommaSeparated(getEnv("GOOGLE_SCOPES", "openid,profile,email")),
			},

			// LINE OAuth
			LINE: OAuthProviderConfig{
				Enabled:      getEnv("LINE_OAUTH_ENABLED", "false") == "true",
				ClientID:     getEnv("LINE_CHANNEL_ID", ""),
				ClientSecret: getEnv("LINE_CHANNEL_SECRET", ""),
				RedirectURL:  getEnv("LINE_REDIRECT_URL", ""),
				Scopes:       parseCommaSeparated(getEnv("LINE_SCOPES", "profile,openid")),
				Options: map[string]string{
					"bot_prompt": getEnv("LINE_BOT_PROMPT", ""),
				},
			},
		},
		PasswordAuth: PasswordAuthConfig{
			Enabled:              getEnv("PASSWORD_AUTH_ENABLED", "true") == "true",
			MinPasswordLength:    getEnvAsInt("PASSWORD_MIN_LENGTH", 8),
			BcryptCost:           getEnvAsInt("BCRYPT_COST", 12),
			EmailVerificationTTL: getEnvAsInt("EMAIL_VERIFICATION_TTL", 86400), // 24 hours
			PasswordResetTTL:     getEnvAsInt("PASSWORD_RESET_TTL", 3600),       // 1 hour
		},
		Email: EmailConfig{
			Provider:    getEnv("EMAIL_PROVIDER", "stub"),
			BaseURL:     getEnv("APP_BASE_URL", "http://localhost:3000"),
			FromAddress: getEnv("EMAIL_FROM_ADDRESS", "noreply@example.com"),
			FromName:    getEnv("EMAIL_FROM_NAME", "Go OAuth Template"),
		},
	}

	// Set production flag
	config.Server.IsProduction = config.Server.Environment == "production"

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return config, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	// Production validation
	if c.Server.IsProduction {
		if c.Database.URL == "" {
			return fmt.Errorf("DATABASE_URL is required in production")
		}
		if c.JWT.SigningKey == "your-secret-key-change-in-production" {
			return fmt.Errorf("JWT_SIGNING_KEY must be changed in production")
		}
	}

	return nil
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt gets an environment variable as int or returns a default value
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		var result int
		if _, err := fmt.Sscanf(value, "%d", &result); err == nil {
			return result
		}
	}
	return defaultValue
}

// parseCommaSeparated parses a comma-separated string into a slice
func parseCommaSeparated(s string) []string {
	if s == "" {
		return []string{}
	}

	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))

	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result
}

// loadEnvFile loads the .env file
func loadEnvFile() {
	// Try explicit path first
	if envPath := os.Getenv("ENV_FILE_PATH"); envPath != "" {
		_ = godotenv.Load(envPath)
		return
	}

	// Try to find .env based on this source file's location
	_, currentFile, _, ok := runtime.Caller(0)
	if ok {
		configDir := filepath.Dir(currentFile)
		backendDir := filepath.Join(configDir, "..", "..")
		envPath := filepath.Join(backendDir, ".env")
		if _, err := os.Stat(envPath); err == nil {
			_ = godotenv.Load(envPath)
			return
		}
	}

	// Try common paths
	candidates := []string{
		".env",
		"backend/.env",
		"../backend/.env",
	}

	for _, path := range candidates {
		if _, err := os.Stat(path); err == nil {
			_ = godotenv.Load(path)
			return
		}
	}

	// Default behavior
	_ = godotenv.Load()
}
