'use client'

import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import { getStoredUser, logout, isAuthenticated, getAccessToken, API_BASE_URL } from '@/lib/auth'

interface User {
  id: string
  display_name?: string
  primary_email?: string
  avatar_url?: string
}

export default function DashboardPage() {
  const router = useRouter()
  const [user, setUser] = useState<User | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    // Check authentication
    if (!isAuthenticated()) {
      router.push('/login')
      return
    }

    // Get stored user info
    const storedUser = getStoredUser()
    if (storedUser) {
      setUser(storedUser)
    }

    setLoading(false)
  }, [router])

  const handleLogout = async () => {
    await logout()
    router.push('/login')
  }

  const testProtectedEndpoint = async () => {
    try {
      const token = getAccessToken()
      const response = await fetch(`${API_BASE_URL}/api/v1/protected/example`, {
        headers: {
          Authorization: `Bearer ${token}`,
        },
      })
      const data = await response.json()
      alert(JSON.stringify(data, null, 2))
    } catch (error) {
      alert('Error calling protected endpoint')
      console.error(error)
    }
  }

  if (loading) {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <div className="h-8 w-8 animate-spin rounded-full border-4 border-[var(--primary)] border-t-transparent" />
      </div>
    )
  }

  return (
    <div className="min-h-screen p-8">
      <div className="mx-auto max-w-4xl">
        <div className="mb-8 flex items-center justify-between">
          <h1 className="text-2xl font-bold">Dashboard</h1>
          <button
            onClick={handleLogout}
            className="rounded-lg border border-[var(--muted)] px-4 py-2 transition-colors hover:bg-[var(--muted)]"
          >
            Logout
          </button>
        </div>

        <div className="rounded-xl border border-[var(--muted)] bg-[var(--background)] p-6 shadow-sm">
          <h2 className="mb-4 text-lg font-medium">User Information</h2>

          {user ? (
            <div className="space-y-4">
              <div className="flex items-center gap-4">
                {user.avatar_url && (
                  <img
                    src={user.avatar_url}
                    alt="Avatar"
                    className="h-16 w-16 rounded-full"
                  />
                )}
                <div>
                  <p className="font-medium">{user.display_name || 'No name'}</p>
                  <p className="text-sm text-[var(--muted-foreground)]">
                    {user.primary_email || 'No email'}
                  </p>
                </div>
              </div>

              <div className="rounded-lg bg-[var(--muted)] p-4">
                <p className="text-sm font-mono">
                  <span className="text-[var(--muted-foreground)]">Account ID:</span>{' '}
                  {user.id}
                </p>
              </div>
            </div>
          ) : (
            <p className="text-[var(--muted-foreground)]">No user information available</p>
          )}
        </div>

        <div className="mt-6 rounded-xl border border-[var(--muted)] bg-[var(--background)] p-6 shadow-sm">
          <h2 className="mb-4 text-lg font-medium">Test Protected Endpoint</h2>
          <p className="mb-4 text-sm text-[var(--muted-foreground)]">
            Click the button below to test the protected API endpoint with your access token.
          </p>
          <button
            onClick={testProtectedEndpoint}
            className="rounded-lg bg-[var(--primary)] px-4 py-2 text-[var(--primary-foreground)] transition-opacity hover:opacity-90"
          >
            Call /api/v1/protected/example
          </button>
        </div>
      </div>
    </div>
  )
}
