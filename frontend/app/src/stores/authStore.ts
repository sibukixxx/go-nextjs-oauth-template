import { create } from 'zustand';
import { api } from '@/lib/api';

export interface User {
  id: string;
  display_name: string;
  primary_email: string;
  avatar_url?: string;
  providers: string[];
}

interface AuthState {
  accessToken: string | null;
  user: User | null;
  isLoading: boolean;
  isAuthenticated: boolean;

  setAuth: (token: string, user: User) => void;
  setLoading: (loading: boolean) => void;
  logout: () => Promise<void>;
  refresh: () => Promise<boolean>;
}

export const useAuthStore = create<AuthState>((set, get) => ({
  accessToken: null,
  user: null,
  isLoading: true,
  isAuthenticated: false,

  setAuth: (accessToken, user) =>
    set({ accessToken, user, isAuthenticated: true, isLoading: false }),

  setLoading: (isLoading) => set({ isLoading }),

  logout: async () => {
    try {
      await api.post('/api/v1/auth/logout');
    } catch {
      // Ignore logout errors
    }
    set({ accessToken: null, user: null, isAuthenticated: false });
  },

  refresh: async () => {
    try {
      const data = await api.post<{
        access_token: string;
        account: User;
      }>('/api/v1/auth/refresh');

      set({
        accessToken: data.access_token,
        user: data.account,
        isAuthenticated: true,
        isLoading: false,
      });
      return true;
    } catch {
      set({
        accessToken: null,
        user: null,
        isAuthenticated: false,
        isLoading: false,
      });
      return false;
    }
  },
}));

// Helper to get auth header
export function getAuthHeader(): Record<string, string> {
  const token = useAuthStore.getState().accessToken;
  return token ? { Authorization: `Bearer ${token}` } : {};
}
