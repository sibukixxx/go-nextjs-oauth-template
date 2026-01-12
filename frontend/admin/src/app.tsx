import { useEffect } from 'preact/hooks';
import Router, { Route } from 'preact-router';
import { isAuthenticated, isLoading, refresh } from '@/signals/auth';
import { LoginPage } from '@/routes/login';
import { DashboardPage } from '@/routes/dashboard';
import { UsersPage } from '@/routes/users';
import { SettingsPage } from '@/routes/settings';
import { Layout } from '@/components/Layout';

function ProtectedRoute({
  component: Component,
}: {
  component: () => JSX.Element;
  path: string;
}) {
  if (isLoading.value) {
    return (
      <div class="min-h-screen flex items-center justify-center bg-background">
        <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-primary" />
      </div>
    );
  }

  if (!isAuthenticated.value) {
    // Redirect to login
    if (typeof window !== 'undefined') {
      window.location.href = '/login';
    }
    return null;
  }

  return (
    <Layout>
      <Component />
    </Layout>
  );
}

export function App() {
  useEffect(() => {
    // Try to restore session on app load
    refresh();
  }, []);

  return (
    <Router>
      <Route path="/login" component={LoginPage} />
      <ProtectedRoute path="/" component={DashboardPage} />
      <ProtectedRoute path="/users" component={UsersPage} />
      <ProtectedRoute path="/settings" component={SettingsPage} />
    </Router>
  );
}
