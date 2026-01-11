# Go + Next.js OAuth Template

OAuth 2.1 準拠の認証機能を備えた Go バックエンド + Next.js フロントエンドのモノレポテンプレートです。

## プロジェクトの目的

このテンプレートは、以下の目的で作成されました：

1. **OAuth 2.1 準拠の認証基盤** - 最新の OAuth 2.1 仕様に準拠したセキュアな認証フローを提供
2. **マルチプロバイダー対応** - Google、LINE など複数の OAuth プロバイダーを統一的なインターフェースで扱える
3. **再利用可能な OAuth ライブラリ** - `pkg/oauth` パッケージは他のプロジェクトでも利用可能
4. **クリーンアーキテクチャ** - 保守性・テスト容易性を考慮した設計

## 主な機能

### 認証方式

このテンプレートは以下の認証方式をサポートしています。環境変数で有効/無効を切り替えられます。

| 認証方式 | 環境変数 | 説明 |
|---------|---------|------|
| パスワード認証 | `PASSWORD_AUTH_ENABLED=true` | メール/パスワードによる登録・ログイン |
| Google OAuth | `GOOGLE_OAUTH_ENABLED=true` | Google アカウントでログイン |
| LINE OAuth | `LINE_OAUTH_ENABLED=true` | LINE アカウントでログイン |

### パスワード認証機能

- **ユーザー登録** - メールアドレスとパスワードで新規登録
- **ログイン** - メール/パスワードでログイン
- **メール確認** - 登録時にメール確認リンクを送信
- **パスワードリセット** - メールでリセットリンクを送信
- **bcrypt ハッシュ** - パスワードは bcrypt でセキュアに保存

### OAuth 2.1 セキュリティ機能

- **PKCE (Proof Key for Code Exchange)** - 認可コード横取り攻撃を防止
- **State パラメータ** - CSRF 攻撃を防止
- **Nonce** - リプレイ攻撃を防止
- **JWT 検証** - JWKS による署名検証（RS256, RS384, RS512）
- **トークンリフレッシュ / 失効** - 完全なトークンライフサイクル管理

### 対応 OAuth プロバイダー

| プロバイダー | ID トークン検証 | ユーザー情報取得 | ログアウト |
|-------------|----------------|-----------------|-----------|
| Google      | JWKS           | ✓               | ✓         |
| LINE        | Verify API     | ✓               | -         |

## アーキテクチャ

```
├── backend/
│   ├── cmd/api/              # アプリケーションエントリーポイント
│   ├── internal/
│   │   ├── domain/           # ドメインモデル・リポジトリインターフェース
│   │   ├── service/          # ビジネスロジック
│   │   ├── handler/          # HTTP ハンドラー
│   │   ├── infra/            # インフラ実装（リポジトリ等）
│   │   ├── middleware/       # HTTP ミドルウェア
│   │   └── config/           # 設定
│   ├── pkg/oauth/            # 再利用可能な OAuth 2.1 ライブラリ
│   └── migrations/           # データベースマイグレーション
│
└── frontend/                 # Next.js フロントエンド
```

## クイックスタート

### 必要条件

- Go 1.24+
- Node.js 18+
- PostgreSQL（データベース使用時）

### バックエンド

```bash
cd backend

# 環境変数を設定
cp .env.example .env
# .env を編集して OAuth クレデンシャルを設定

# 実行
go run ./cmd/api

# テスト
go test ./...
```

### 環境変数

詳細は `backend/.env.example` を参照してください。

```env
# =============================================================================
# 認証方式の有効化
# =============================================================================
PASSWORD_AUTH_ENABLED=true      # パスワード認証
GOOGLE_OAUTH_ENABLED=false      # Google OAuth
LINE_OAUTH_ENABLED=false        # LINE OAuth

# =============================================================================
# パスワード認証設定
# =============================================================================
PASSWORD_MIN_LENGTH=8           # 最小パスワード長
EMAIL_PROVIDER=stub             # メール送信: stub, smtp, sendgrid
APP_BASE_URL=http://localhost:3000

# =============================================================================
# Google OAuth
# =============================================================================
GOOGLE_CLIENT_ID=your-client-id
GOOGLE_CLIENT_SECRET=your-client-secret
GOOGLE_REDIRECT_URL=http://localhost:3000/auth/google/callback

# =============================================================================
# LINE OAuth
# =============================================================================
LINE_CHANNEL_ID=your-channel-id
LINE_CHANNEL_SECRET=your-channel-secret
LINE_REDIRECT_URL=http://localhost:3000/auth/line/callback

# =============================================================================
# Server
# =============================================================================
PORT=8080
```

## パスワード認証 API

パスワード認証が有効な場合、以下のエンドポイントが利用可能です。

| エンドポイント | メソッド | 説明 |
|--------------|--------|------|
| `/api/auth/register` | POST | 新規ユーザー登録 |
| `/api/auth/login` | POST | ログイン |
| `/api/auth/forgot-password` | POST | パスワードリセット要求 |
| `/api/auth/reset-password` | POST | パスワードリセット実行 |
| `/api/auth/verify-email` | POST | メール確認 |
| `/api/auth/resend-verification` | POST | 確認メール再送 |

### リクエスト例

```bash
# ユーザー登録
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password123"}'

# ログイン
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password123"}'

# パスワードリセット要求
curl -X POST http://localhost:8080/api/auth/forgot-password \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com"}'
```

## OAuth ライブラリの使い方

### プロバイダーの初期化

```go
import "github.com/your-org/go-nextjs-oauth-template/backend/pkg/oauth"

// Google プロバイダー
googleProvider, err := oauth.NewGoogleProvider(oauth.GoogleProviderOptions{
    ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
    ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
    RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
})

// プロバイダーレジストリに登録
registry := oauth.NewProviderRegistry()
registry.Register(googleProvider)
```

### 認可フローの開始

```go
// PKCE とステートを生成
pkce, _ := oauth.GeneratePKCE()
state := oauth.GenerateState()
nonce := oauth.GenerateNonce()

// 認可 URL を構築
authURL, err := provider.BuildAuthorizationURL(&oauth.AuthorizationRequest{
    State: state,
    PKCE:  pkce,
    Nonce: nonce,
})
```

### トークン交換

```go
// コールバックで認可コードを受け取り、トークンと交換
tokens, err := provider.ExchangeCode(ctx, code, pkce.CodeVerifier, redirectURL)

// ID トークンを検証
claims, err := provider.ValidateIDToken(ctx, tokens.IDToken, nonce)
```

## 新しい OAuth プロバイダーの追加

1. `backend/pkg/oauth/` に新しいファイルを作成（例: `github.go`）
2. `Provider` インターフェースを実装
3. プロバイダーオプション構造体を定義
4. ファクトリ関数 `NewGitHubProvider()` を実装
5. テストを追加

```go
type GitHubProvider struct {
    client *Client
    config ProviderConfig
}

func (p *GitHubProvider) Type() ProviderType {
    return ProviderGitHub
}

// ... 他のインターフェースメソッドを実装
```

## 技術スタック

### バックエンド

- **Go 1.24** - メイン言語
- **golang-jwt/jwt/v5** - JWT 処理
- **google/uuid** - UUID 生成
- **godotenv** - 環境変数管理

### フロントエンド

- **Next.js** - React フレームワーク
- **TypeScript** - 型安全性

## ライセンス

MIT License
