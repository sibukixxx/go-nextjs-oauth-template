# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

A monorepo implementing a Go backend with multiple authentication methods and a Next.js frontend. The backend provides:
- **Password authentication** - Email/password registration, login, reset
- **OAuth 2.1 authentication** - Google, LINE providers via reusable `pkg/oauth` library

## Architecture

### Backend Structure (Clean Architecture)

```
backend/
├── cmd/api/              # Application entry point
├── internal/
│   ├── domain/
│   │   ├── model/       # Domain entities (Account, Identity, PasswordCredential)
│   │   └── repository/  # Repository interfaces
│   ├── service/         # Business logic (PasswordAuthService, email templates)
│   ├── handler/         # HTTP handlers (PasswordAuthHandler)
│   ├── infra/           # Infrastructure implementations
│   │   └── email/       # Email sender implementations (stub, smtp)
│   ├── middleware/      # HTTP middleware
│   └── config/          # Configuration
├── pkg/
│   ├── oauth/           # Reusable OAuth 2.1 library
│   └── crypto/          # Password hashing utilities (bcrypt)
└── migrations/          # Database migrations
```

### Key Components

**Password Authentication:**
- `pkg/crypto/password.go` - bcrypt password hashing
- `internal/domain/model/password_credential.go` - Password credential model
- `internal/service/password_auth_service.go` - Registration, login, reset logic
- `internal/service/email.go` - EmailSender interface and templates
- `internal/handler/password_auth_handler.go` - HTTP endpoints

**OAuth Library (`pkg/oauth`):**
- `Provider` interface - extensible provider support
- `ProviderRegistry` - registry pattern for multiple providers
- `Client` - base OAuth 2.1 client with PKCE
- `JWTValidator` - OAuth 2.1 compliant JWT validation

### Design Patterns
- Provider Pattern + Registry Pattern for OAuth providers
- Repository Pattern for data access abstraction
- Functional Options for configuration
- Dependency Injection via constructors

## Authentication Configuration

Enable/disable authentication methods via environment variables:

```env
# Password Authentication
PASSWORD_AUTH_ENABLED=true
PASSWORD_MIN_LENGTH=8
BCRYPT_COST=12
EMAIL_VERIFICATION_TTL=86400  # 24 hours
PASSWORD_RESET_TTL=3600       # 1 hour

# Email Service
EMAIL_PROVIDER=stub           # stub, smtp, sendgrid
APP_BASE_URL=http://localhost:3000

# OAuth Providers
GOOGLE_OAUTH_ENABLED=false
LINE_OAUTH_ENABLED=false
```

## Build and Test Commands

### Backend

`backend/Makefile` 経由で実行する（グローバルの Makefile 優先ルール）:

```bash
cd backend

make build      # Build (bin/api)
make run        # Build して起動
make dev        # 開発モードで起動
make test       # Test
make test-race  # Test with race detector
make fmt        # Format
make lint       # Lint (requires golangci-lint)
```

### Frontend

`frontend/` は pnpm workspace のモノレポ（`pnpm-workspace.yaml`）:

- `frontend/app` - アプリ本体 `@app/web`（Vite）
- `frontend/admin` - 管理画面 `@app/admin`（Vite）
- `frontend/marketing` - マーケティングサイト `@app/marketing`（Astro）
- `frontend/packages/ui` - 共有 UI パッケージ `@app/ui`

```bash
cd frontend && pnpm install
pnpm dev:app / dev:admin / dev:marketing   # 各 workspace の開発サーバー
pnpm build / lint / typecheck              # workspace 一括実行
```

## Password Auth API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/auth/register` | POST | Register new user |
| `/api/auth/login` | POST | Login |
| `/api/auth/forgot-password` | POST | Request password reset |
| `/api/auth/reset-password` | POST | Reset password with token |
| `/api/auth/verify-email` | POST | Verify email with token |
| `/api/auth/resend-verification` | POST | Resend verification email |

## Adding New OAuth Providers

1. Create `backend/pkg/oauth/provider_name.go`
2. Implement `Provider` interface
3. Create provider options struct with functional options pattern
4. Add factory function: `NewProviderNameProvider()`
5. Write tests for the new provider

## Security Requirements

- Passwords hashed with bcrypt (cost=12)
- PKCE mandatory for OAuth 2.1
- State parameter for CSRF protection
- Token expiration for password reset (1h) and email verification (24h)
- Generic error messages to prevent email enumeration
- Never log tokens or secrets
