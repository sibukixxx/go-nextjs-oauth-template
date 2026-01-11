// Backend API base URL
export const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'

export type OAuthProvider = 'google' | 'line'

export interface LoginResponse {
  authorization_url: string
  state: string
  provider: string
}

export interface CallbackResponse {
  access_token: string
  token_type: string
  expires_in: number
  refresh_token?: string
  account: {
    id: string
    display_name?: string
    primary_email?: string
    avatar_url?: string
  }
  is_new_account: boolean
}

// Initiate OAuth login
export async function initiateLogin(provider: OAuthProvider, redirectUri?: string): Promise<LoginResponse> {
  const response = await fetch(`${API_BASE_URL}/api/v1/auth/login`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    credentials: 'include',
    body: JSON.stringify({
      provider,
      redirect_uri: redirectUri || `${window.location.origin}/auth/callback/${provider}`,
    }),
  })

  if (!response.ok) {
    const errorData = await response.json()
    throw new Error(errorData.error || 'Failed to initiate login')
  }

  return response.json()
}

// Handle OAuth callback
export async function handleCallback(code: string, state: string): Promise<CallbackResponse> {
  const response = await fetch(`${API_BASE_URL}/api/v1/auth/callback`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    credentials: 'include',
    body: JSON.stringify({
      code,
      state,
    }),
  })

  if (!response.ok) {
    const errorData = await response.json()
    throw new Error(errorData.error || 'Authentication failed')
  }

  return response.json()
}

// Refresh access token
export async function refreshToken(): Promise<CallbackResponse> {
  const response = await fetch(`${API_BASE_URL}/api/v1/auth/refresh`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    credentials: 'include',
  })

  if (!response.ok) {
    throw new Error('Failed to refresh token')
  }

  return response.json()
}

// Logout
export async function logout(): Promise<void> {
  const accessToken = localStorage.getItem('access_token')

  await fetch(`${API_BASE_URL}/api/v1/auth/logout`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      ...(accessToken ? { Authorization: `Bearer ${accessToken}` } : {}),
    },
    credentials: 'include',
  })

  // Clear local storage
  localStorage.removeItem('access_token')
  localStorage.removeItem('user')
}

// Get stored access token
export function getAccessToken(): string | null {
  if (typeof window === 'undefined') return null
  return localStorage.getItem('access_token')
}

// Store tokens after login
export function storeTokens(data: CallbackResponse): void {
  if (data.access_token) {
    localStorage.setItem('access_token', data.access_token)
  }
  if (data.account) {
    localStorage.setItem('user', JSON.stringify(data.account))
  }
}

// Clear stored tokens
export function clearTokens(): void {
  localStorage.removeItem('access_token')
  localStorage.removeItem('user')
}

// Get stored user
export function getStoredUser(): CallbackResponse['account'] | null {
  if (typeof window === 'undefined') return null
  const user = localStorage.getItem('user')
  return user ? JSON.parse(user) : null
}

// Check if user is authenticated
export function isAuthenticated(): boolean {
  return !!getAccessToken()
}
