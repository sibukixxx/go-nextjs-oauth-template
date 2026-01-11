'use client'

import { useState } from 'react'
import Link from 'next/link'
import { initiateLogin, OAuthProvider } from '@/lib/auth'

export default function LoginPage() {
  const [loading, setLoading] = useState(false)
  const [loadingProvider, setLoadingProvider] = useState<OAuthProvider | null>(null)
  const [error, setError] = useState<string | null>(null)

  const handleOAuthLogin = async (provider: OAuthProvider) => {
    setLoading(true)
    setLoadingProvider(provider)
    setError(null)

    try {
      const data = await initiateLogin(provider)

      // Store state for client-side CSRF verification (optional extra layer)
      if (data.state) {
        sessionStorage.setItem(`${provider}_oauth_state`, data.state)
      }

      // Redirect to authorization URL
      if (data.authorization_url) {
        window.location.href = data.authorization_url
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : `Failed to login with ${provider}`)
      setLoading(false)
      setLoadingProvider(null)
    }
  }

  return (
    <div className="flex min-h-screen flex-col items-center justify-center px-4 py-12">
      <div className="w-full max-w-md">
        {/* Logo & Brand */}
        <div className="mb-10 text-center">
          <div className="mb-4 inline-flex h-16 w-16 items-center justify-center rounded-2xl bg-[var(--primary)]">
            <svg
              className="h-8 w-8 text-white"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={1.5}
                d="M15 7a2 2 0 012 2m4 0a6 6 0 01-7.743 5.743L11 17H9v2H7v2H4a1 1 0 01-1-1v-2.586a1 1 0 01.293-.707l5.964-5.964A6 6 0 1121 9z"
              />
            </svg>
          </div>
          <h1 className="text-2xl font-medium">OAuth Template</h1>
          <p className="mt-2 text-sm text-[var(--muted-foreground)]">
            Sign in with your account
          </p>
        </div>

        {/* Login Card */}
        <div className="rounded-xl border border-[var(--muted)] bg-[var(--background)] p-8 shadow-sm">
          <h2 className="mb-6 text-center text-lg font-medium">Login</h2>

          {error && (
            <div className="mb-6 rounded-lg bg-[var(--danger-light)] px-4 py-3 text-sm text-[var(--danger)]">
              {error}
            </div>
          )}

          <div className="space-y-4">
            {/* Google OAuth Button */}
            <button
              onClick={() => handleOAuthLogin('google')}
              disabled={loading}
              className="flex w-full items-center justify-center gap-3 rounded-lg border border-[var(--muted)] bg-white px-4 py-3 text-gray-800 transition-colors hover:bg-gray-50 disabled:opacity-50"
            >
              <svg className="h-5 w-5" viewBox="0 0 24 24">
                <path
                  fill="#4285F4"
                  d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"
                />
                <path
                  fill="#34A853"
                  d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"
                />
                <path
                  fill="#FBBC05"
                  d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z"
                />
                <path
                  fill="#EA4335"
                  d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"
                />
              </svg>
              <span className="font-medium">
                {loadingProvider === 'google' ? 'Loading...' : 'Continue with Google'}
              </span>
            </button>

            {/* LINE OAuth Button */}
            <button
              onClick={() => handleOAuthLogin('line')}
              disabled={loading}
              className="flex w-full items-center justify-center gap-3 rounded-lg bg-[#06C755] px-4 py-3 text-white transition-colors hover:bg-[#05B54C] disabled:opacity-50"
            >
              <svg className="h-5 w-5" viewBox="0 0 24 24" fill="currentColor">
                <path d="M19.365 9.863c.349 0 .63.285.63.631 0 .345-.281.63-.63.63H17.61v1.125h1.755c.349 0 .63.283.63.63 0 .344-.281.629-.63.629h-2.386c-.345 0-.627-.285-.627-.629V8.108c0-.345.282-.63.63-.63h2.386c.346 0 .627.285.627.63 0 .349-.281.63-.63.63H17.61v1.125h1.755zm-3.855 3.016c0 .27-.174.51-.432.596-.064.021-.133.031-.199.031-.211 0-.391-.09-.51-.25l-2.443-3.317v2.94c0 .344-.279.629-.631.629-.346 0-.626-.285-.626-.629V8.108c0-.27.173-.51.43-.595.06-.023.136-.033.194-.033.195 0 .375.104.495.254l2.462 3.33V8.108c0-.345.282-.63.63-.63.345 0 .63.285.63.63v4.771zm-5.741 0c0 .344-.282.629-.631.629-.345 0-.627-.285-.627-.629V8.108c0-.345.282-.63.63-.63.346 0 .628.285.628.63v4.771zm-2.466.629H4.917c-.345 0-.63-.285-.63-.629V8.108c0-.345.285-.63.63-.63.348 0 .63.285.63.63v4.141h1.756c.348 0 .629.283.629.63 0 .344-.282.629-.629.629M24 10.314C24 4.943 18.615.572 12 .572S0 4.943 0 10.314c0 4.811 4.27 8.842 10.035 9.608.391.082.923.258 1.058.59.12.301.079.766.038 1.08l-.164 1.02c-.045.301-.24 1.186 1.049.645 1.291-.539 6.916-4.078 9.436-6.975C23.176 14.393 24 12.458 24 10.314" />
              </svg>
              <span className="font-medium">
                {loadingProvider === 'line' ? 'Loading...' : 'Continue with LINE'}
              </span>
            </button>
          </div>

          {/* Footer */}
          <div className="mt-6 text-center">
            <p className="text-sm text-[var(--muted-foreground)]">
              New users can also sign in using the buttons above
            </p>
          </div>
        </div>

        {/* Back to home */}
        <p className="mt-8 text-center text-sm text-[var(--muted-foreground)]">
          <Link href="/" className="underline hover:text-[var(--foreground)]">
            Back to home
          </Link>
        </p>
      </div>
    </div>
  )
}
