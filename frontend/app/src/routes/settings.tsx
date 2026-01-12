import { useState } from 'react';
import {
  Card,
  CardHeader,
  CardTitle,
  CardDescription,
  CardContent,
  ThemeSelector,
  Badge,
  Button,
  type Theme,
  getStoredTheme,
} from '@app/ui';
import { useAuthStore } from '@/stores/authStore';

export function SettingsPage() {
  const { user } = useAuthStore();
  const [currentTheme, setCurrentTheme] = useState<Theme>(
    getStoredTheme() ?? 'mint'
  );

  return (
    <div className="space-y-8 max-w-3xl">
      <div>
        <h1 className="text-2xl font-display font-bold text-foreground">
          Settings
        </h1>
        <p className="text-foreground-secondary mt-1">
          Manage your account settings and preferences
        </p>
      </div>

      {/* Profile section */}
      <Card>
        <CardHeader>
          <CardTitle>Profile</CardTitle>
          <CardDescription>Your account information</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            <div>
              <label className="text-sm font-medium text-foreground-secondary">
                Display Name
              </label>
              <p className="text-foreground mt-1">{user?.display_name}</p>
            </div>
            <div>
              <label className="text-sm font-medium text-foreground-secondary">
                Email
              </label>
              <p className="text-foreground mt-1">{user?.primary_email}</p>
            </div>
            <div>
              <label className="text-sm font-medium text-foreground-secondary">
                Connected Accounts
              </label>
              <div className="flex flex-wrap gap-2 mt-2">
                {user?.providers.map((provider) => (
                  <Badge key={provider} variant="default">
                    {provider.charAt(0).toUpperCase() + provider.slice(1)}
                  </Badge>
                ))}
              </div>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Theme section */}
      <Card>
        <CardHeader>
          <CardTitle>Appearance</CardTitle>
          <CardDescription>
            Customize how the app looks and feels
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            <div>
              <label className="text-sm font-medium text-foreground-secondary mb-3 block">
                Theme
              </label>
              <ThemeSelector
                currentTheme={currentTheme}
                onThemeChange={setCurrentTheme}
              />
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Danger zone */}
      <Card>
        <CardHeader>
          <CardTitle className="text-error">Danger Zone</CardTitle>
          <CardDescription>
            Irreversible actions for your account
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="flex items-center justify-between">
            <div>
              <p className="font-medium text-foreground">Delete Account</p>
              <p className="text-sm text-foreground-secondary">
                Permanently delete your account and all data
              </p>
            </div>
            <Button variant="danger" size="sm">
              Delete Account
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
