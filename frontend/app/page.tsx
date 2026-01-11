import Link from 'next/link'

export default function HomePage() {
  return (
    <div className="flex min-h-screen flex-col items-center justify-center p-8">
      <div className="max-w-md text-center">
        <h1 className="mb-4 text-3xl font-bold">Go + Next.js OAuth Template</h1>
        <p className="mb-8 text-[var(--muted-foreground)]">
          OAuth 2.1 with PKCE, Multi-provider support (Google, LINE)
        </p>
        <div className="flex flex-col gap-4">
          <Link
            href="/login"
            className="rounded-lg bg-[var(--primary)] px-6 py-3 text-[var(--primary-foreground)] transition-opacity hover:opacity-90"
          >
            Login
          </Link>
          <Link
            href="/dashboard"
            className="rounded-lg border border-[var(--muted)] px-6 py-3 transition-colors hover:bg-[var(--muted)]"
          >
            Dashboard (Protected)
          </Link>
        </div>
      </div>
    </div>
  )
}
