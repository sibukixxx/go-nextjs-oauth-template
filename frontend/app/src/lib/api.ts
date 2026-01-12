import { useAuthStore } from '@/stores/authStore';

class ApiError extends Error {
  constructor(
    public status: number,
    message: string
  ) {
    super(message);
    this.name = 'ApiError';
  }
}

async function request<T>(
  url: string,
  options: RequestInit = {}
): Promise<T> {
  const { accessToken, refresh, logout } = useAuthStore.getState();

  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...(options.headers as Record<string, string>),
  };

  if (accessToken) {
    headers['Authorization'] = `Bearer ${accessToken}`;
  }

  let response = await fetch(url, {
    ...options,
    headers,
    credentials: 'include',
  });

  // Try to refresh token on 401
  if (response.status === 401 && accessToken) {
    const refreshed = await refresh();
    if (refreshed) {
      const newToken = useAuthStore.getState().accessToken;
      headers['Authorization'] = `Bearer ${newToken}`;
      response = await fetch(url, {
        ...options,
        headers,
        credentials: 'include',
      });
    } else {
      await logout();
      throw new ApiError(401, 'Session expired');
    }
  }

  if (!response.ok) {
    const errorData = await response.json().catch(() => ({}));
    throw new ApiError(
      response.status,
      errorData.message || errorData.error || 'Request failed'
    );
  }

  // Handle empty responses
  const text = await response.text();
  if (!text) {
    return {} as T;
  }

  return JSON.parse(text) as T;
}

export const api = {
  get: <T>(url: string, options?: RequestInit) =>
    request<T>(url, { ...options, method: 'GET' }),

  post: <T>(url: string, body?: unknown, options?: RequestInit) =>
    request<T>(url, {
      ...options,
      method: 'POST',
      body: body ? JSON.stringify(body) : undefined,
    }),

  put: <T>(url: string, body?: unknown, options?: RequestInit) =>
    request<T>(url, {
      ...options,
      method: 'PUT',
      body: body ? JSON.stringify(body) : undefined,
    }),

  delete: <T>(url: string, options?: RequestInit) =>
    request<T>(url, { ...options, method: 'DELETE' }),
};

export { ApiError };
