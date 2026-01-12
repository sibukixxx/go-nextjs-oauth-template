import { signal, computed } from '@preact/signals';

export interface User {
  id: string;
  display_name: string;
  primary_email: string;
  avatar_url?: string;
  providers: string[];
}

// Auth state signals
export const accessToken = signal<string | null>(null);
export const user = signal<User | null>(null);
export const isLoading = signal(true);

// Computed values
export const isAuthenticated = computed(() => !!accessToken.value && !!user.value);

// Actions
export function setAuth(token: string, userData: User) {
  accessToken.value = token;
  user.value = userData;
  isLoading.value = false;
}

export function clearAuth() {
  accessToken.value = null;
  user.value = null;
}

export async function logout() {
  try {
    await fetch('/api/v1/auth/logout', {
      method: 'POST',
      credentials: 'include',
      headers: accessToken.value
        ? { Authorization: `Bearer ${accessToken.value}` }
        : {},
    });
  } catch {
    // Ignore logout errors
  }
  clearAuth();
}

export async function refresh(): Promise<boolean> {
  try {
    const res = await fetch('/api/v1/auth/refresh', {
      method: 'POST',
      credentials: 'include',
    });

    if (res.ok) {
      const data = await res.json();
      setAuth(data.access_token, data.account);
      return true;
    }
  } catch {
    // Ignore refresh errors
  }

  clearAuth();
  isLoading.value = false;
  return false;
}

// Get auth header for API calls
export function getAuthHeader(): Record<string, string> {
  return accessToken.value
    ? { Authorization: `Bearer ${accessToken.value}` }
    : {};
}
