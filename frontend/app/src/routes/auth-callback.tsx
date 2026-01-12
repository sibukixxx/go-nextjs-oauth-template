import { useEffect, useState } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { useAuthStore } from '@/stores/authStore';
import { api } from '@/lib/api';

export function AuthCallbackPage() {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const { setAuth } = useAuthStore();
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const code = searchParams.get('code');
    const state = searchParams.get('state');
    const errorParam = searchParams.get('error');

    if (errorParam) {
      setError(`Authentication failed: ${errorParam}`);
      return;
    }

    if (!code || !state) {
      setError('Invalid callback parameters');
      return;
    }

    // Exchange code for tokens
    api
      .post<{
        access_token: string;
        account: {
          id: string;
          display_name: string;
          primary_email: string;
          avatar_url?: string;
          providers: string[];
        };
      }>('/api/v1/auth/callback', { code, state })
      .then((data) => {
        setAuth(data.access_token, data.account);
        navigate('/', { replace: true });
      })
      .catch((err) => {
        setError(err instanceof Error ? err.message : 'Authentication failed');
      });
  }, [searchParams, setAuth, navigate]);

  if (error) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-background p-4">
        <div className="text-center">
          <div className="w-16 h-16 bg-error/10 rounded-full flex items-center justify-center mx-auto mb-4">
            <svg
              className="w-8 h-8 text-error"
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
          <h1 className="text-xl font-semibold text-foreground mb-2">
            Authentication Failed
          </h1>
          <p className="text-foreground-secondary mb-6">{error}</p>
          <button
            onClick={() => navigate('/login')}
            className="text-primary hover:underline"
          >
            Return to login
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-background">
      <div className="text-center">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary mx-auto mb-4" />
        <p className="text-foreground-secondary">Completing sign in...</p>
      </div>
    </div>
  );
}
