# Go OAuth Template

Go authentication backend plus pnpm frontend workspaces for app/admin/marketing surfaces.

## Commands
- Backend: `make build`, `make test`, `make test-race`, `make fmt`, `make lint`
- Frontend install: `cd frontend && pnpm install`
- Frontend: `pnpm dev:app` / `pnpm dev:admin` / `pnpm dev:marketing`
- Frontend checks: `pnpm build && pnpm lint && pnpm typecheck`

## Shared rules
- Authentication/authorization decisions are server-side; never trust browser-supplied identity, role, or OAuth state.
- Never log tokens, passwords, OAuth codes, or provider secrets.
- Add OAuth providers through the existing provider abstraction rather than provider-specific branches throughout handlers.
- Frontend uses pnpm workspaces.
- Security-sensitive cookie/session/redirect changes require explicit negative-path tests.

## Change-dependent checks
- Auth/backend: race tests + lint.
- Frontend: build + lint + typecheck.
- Provider/session changes: add state/CSRF/redirect/error-path coverage.

## Done
- Affected checks pass.
- Secret/token material is absent from logs/tests.
- Authentication authority remains server-side.
