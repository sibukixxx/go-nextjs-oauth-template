package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/your-org/go-nextjs-oauth-template/backend/internal/config"
	"github.com/your-org/go-nextjs-oauth-template/backend/internal/handler"
	"github.com/your-org/go-nextjs-oauth-template/backend/internal/middleware"
	"github.com/your-org/go-nextjs-oauth-template/backend/internal/service"
)

func main() {
	// Initialize logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Error("Failed to load configuration", "error", err)
		os.Exit(1)
	}

	logger.Info("Application starting",
		"environment", cfg.Server.Environment,
		"port", cfg.Server.Port,
	)

	// Initialize authentication services
	var accountAuthHandler *handler.AccountAuthHandler
	var jwtAuthMiddleware *middleware.JWTAuthMiddleware

	if cfg.OAuthProviders.Google.Enabled || cfg.OAuthProviders.LINE.Enabled {
		// Initialize MultiAuthService
		multiAuthService, err := service.NewMultiAuthService(service.MultiAuthServiceConfig{
			Config: cfg.OAuthProviders,
			Logger: logger,
		})
		if err != nil {
			logger.Error("Failed to initialize MultiAuthService", "error", err)
			if cfg.Server.IsProduction {
				os.Exit(1)
			}
		} else {
			// Initialize JWTService
			jwtService, err := service.NewJWTService(service.JWTServiceConfig{
				Issuer:         cfg.JWT.Issuer,
				Audience:       cfg.JWT.Audience,
				AccessTokenTTL: time.Duration(cfg.JWT.AccessTokenTTL) * time.Second,
				SigningKey:     cfg.JWT.SigningKey,
			})
			if err != nil {
				logger.Error("Failed to initialize JWTService", "error", err)
				if cfg.Server.IsProduction {
					os.Exit(1)
				}
			} else {
				// Initialize AccountAuthService
				// Note: In production, you should inject actual repository implementations
				// For now, we create the service without repositories (in-memory mode)
				accountAuthService, err := service.NewAccountAuthService(service.AccountAuthServiceConfig{
					MultiAuthService: multiAuthService,
					// Add repository implementations here when database is available:
					// AccountRepo:      yourAccountRepo,
					// IdentityRepo:     yourIdentityRepo,
					// AuthSessionRepo:  yourAuthSessionRepo,
					// RefreshTokenRepo: yourRefreshTokenRepo,
					// AuditLogRepo:     yourAuditLogRepo,
					JWTService: jwtService,
					Logger:     logger,
				})
				if err != nil {
					logger.Error("Failed to initialize AccountAuthService", "error", err)
					if cfg.Server.IsProduction {
						os.Exit(1)
					}
				} else {
					accountAuthHandler = handler.NewAccountAuthHandler(accountAuthService, jwtService, logger, cfg.Server.IsProduction)
					jwtAuthMiddleware = middleware.NewJWTAuthMiddleware(jwtService, logger)

					// Start background session cleanup
					go func() {
						ticker := time.NewTicker(1 * time.Minute)
						defer ticker.Stop()
						for range ticker.C {
							multiAuthService.CleanupExpiredSessions()
						}
					}()

					logger.Info("Authentication initialized",
						"providers", cfg.OAuthProviders.EnabledProviders,
						"default_provider", cfg.OAuthProviders.DefaultProvider,
					)
				}
			}
		}
	} else {
		logger.Warn("No OAuth providers enabled. Authentication endpoints will not be registered")
	}

	// Setup router
	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	// Root endpoint
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"name":"go-nextjs-oauth-template","version":"1.0.0"}`))
	})

	// Register authentication endpoints
	if accountAuthHandler != nil {
		accountAuthHandler.RegisterRoutes(mux)
		logger.Info("Authentication endpoints registered",
			"endpoints", []string{
				"/api/v1/auth/providers",
				"/api/v1/auth/login",
				"/api/v1/auth/callback",
				"/api/v1/auth/refresh",
				"/api/v1/auth/logout",
				"/api/v1/auth/me",
				"/api/v1/auth/link",
			})
	}

	// Example protected endpoint
	if jwtAuthMiddleware != nil {
		mux.Handle("GET /api/v1/protected/example", jwtAuthMiddleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			accountID := middleware.GetAccountID(r.Context())
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(fmt.Sprintf(`{"message":"Hello, authenticated user!","account_id":"%s"}`, accountID.String())))
		})))
		logger.Info("Protected endpoint registered", "endpoint", "/api/v1/protected/example")
	}

	// Apply middleware
	var h http.Handler = mux
	h = middleware.CORS(cfg.Server.AllowedOrigins)(h)
	h = loggingMiddleware(logger)(h)
	h = recoveryMiddleware(logger)(h)

	// Create HTTP server
	server := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      h,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		logger.Info("HTTP server starting", "addr", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Server error", "error", err)
			os.Exit(1)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Server shutdown failed", "error", err)
		os.Exit(1)
	}

	logger.Info("Server shutdown complete")
}

// loggingMiddleware logs HTTP requests
func loggingMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Wrap response writer to capture status code
			wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
			next.ServeHTTP(wrapped, r)

			logger.Info("HTTP request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", wrapped.statusCode,
				"duration", time.Since(start).String(),
				"remote_addr", r.RemoteAddr,
			)
		})
	}
}

// recoveryMiddleware recovers from panics
func recoveryMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					logger.Error("Panic recovered",
						"error", err,
						"method", r.Method,
						"path", r.URL.Path,
					)
					http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
