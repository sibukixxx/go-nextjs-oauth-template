'use client'

import { Suspense, useEffect, useState } from 'react'
import { useRouter, useSearchParams } from 'next/navigation'
import { handleCallback, storeTokens } from '@/lib/auth'

function GoogleCallbackContent() {
  const router = useRouter()
  const searchParams = useSearchParams()
  const [error, setError] = useState<string | null>(null)
  const [status, setStatus] = useState<'processing' | 'success' | 'error'>('processing')

  useEffect(() => {
    const processCallback = async () => {
      try {
        const code = searchParams.get('code')
        const state = searchParams.get('state')
        const errorParam = searchParams.get('error')
        const errorDescription = searchParams.get('error_description')

        if (errorParam) {
          throw new Error(errorDescription || `Google login error: ${errorParam}`)
        }

        if (!code) {
          throw new Error('Authorization code is missing')
        }

        if (!state) {
          throw new Error('State parameter is missing')
        }

        // Optional: Validate state client-side
        const storedState = sessionStorage.getItem('google_oauth_state')
        if (storedState && storedState !== state) {
          throw new Error('Session verification failed. Please try again.')
        }

        // Exchange code for tokens
        const data = await handleCallback(code, state)

        // Store tokens
        storeTokens(data)

        // Clear session storage
        sessionStorage.removeItem('google_oauth_state')

        setStatus('success')

        // Redirect to dashboard
        setTimeout(() => {
          router.push('/dashboard')
        }, 1000)
      } catch (err) {
        console.error('Google callback error:', err)
        setError(err instanceof Error ? err.message : 'An error occurred during authentication')
        setStatus('error')
        sessionStorage.removeItem('google_oauth_state')
      }
    }

    processCallback()
  }, [searchParams, router])

  return (
    <div className="flex min-h-screen flex-col items-center justify-center px-4 py-12">
      <div className="w-full max-w-md">
        <div className="rounded-xl border border-[var(--muted)] bg-[var(--background)] p-8 text-center shadow-sm">
          {status === 'processing' && (
            <>
              <div className="mb-4 flex justify-center">
                <svg
                  className="h-12 w-12 animate-spin text-[#4285F4]"
                  viewBox="0 0 24 24"
                  fill="none"
                >
                  <circle
                    className="opacity-25"
                    cx="12"
                    cy="12"
                    r="10"
                    stroke="currentColor"
                    strokeWidth="4"
                  />
                  <path
                    className="opacity-75"
                    fill="currentColor"
                    d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
                  />
                </svg>
              </div>
              <h2 className="text-lg font-medium">Authenticating with Google...</h2>
              <p className="mt-2 text-sm text-[var(--muted-foreground)]">Please wait</p>
            </>
          )}

          {status === 'success' && (
            <>
              <div className="mb-4 flex justify-center">
                <div className="flex h-12 w-12 items-center justify-center rounded-full bg-[#4285F4]">
                  <svg
                    className="h-6 w-6 text-white"
                    fill="none"
                    viewBox="0 0 24 24"
                    stroke="currentColor"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d="M5 13l4 4L19 7"
                    />
                  </svg>
                </div>
              </div>
              <h2 className="text-lg font-medium">Login successful!</h2>
              <p className="mt-2 text-sm text-[var(--muted-foreground)]">
                Redirecting to dashboard...
              </p>
            </>
          )}

          {status === 'error' && (
            <>
              <div className="mb-4 flex justify-center">
                <div className="flex h-12 w-12 items-center justify-center rounded-full bg-[var(--danger)]">
                  <svg
                    className="h-6 w-6 text-white"
                    fill="none"
                    viewBox="0 0 24 24"
                    stroke="currentColor"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d="M6 18L18 6M6 6l12 12"
                    />
                  </svg>
                </div>
              </div>
              <h2 className="text-lg font-medium">Login failed</h2>
              <p className="mt-2 text-sm text-[var(--danger)]">{error}</p>
              <button
                onClick={() => router.push('/login')}
                className="mt-6 w-full rounded-lg bg-[var(--primary)] px-4 py-3 text-[var(--primary-foreground)] transition-opacity hover:opacity-90"
              >
                Back to login
              </button>
            </>
          )}
        </div>
      </div>
    </div>
  )
}

function LoadingFallback() {
  return (
    <div className="flex min-h-screen flex-col items-center justify-center px-4 py-12">
      <div className="w-full max-w-md">
        <div className="rounded-xl border border-[var(--muted)] bg-[var(--background)] p-8 text-center shadow-sm">
          <div className="mb-4 flex justify-center">
            <svg className="h-12 w-12 animate-spin text-[#4285F4]" viewBox="0 0 24 24" fill="none">
              <circle
                className="opacity-25"
                cx="12"
                cy="12"
                r="10"
                stroke="currentColor"
                strokeWidth="4"
              />
              <path
                className="opacity-75"
                fill="currentColor"
                d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
              />
            </svg>
          </div>
          <h2 className="text-lg font-medium">Loading...</h2>
        </div>
      </div>
    </div>
  )
}

export default function GoogleCallbackPage() {
  return (
    <Suspense fallback={<LoadingFallback />}>
      <GoogleCallbackContent />
    </Suspense>
  )
}
