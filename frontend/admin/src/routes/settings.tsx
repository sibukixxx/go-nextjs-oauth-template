import { currentTheme, setTheme, themes, type Theme } from '@/signals/theme';
import { user } from '@/signals/auth';

const themePreviewColors: Record<Theme, { bg: string; accent: string }> = {
  mint: { bg: '#E8F5F3', accent: '#2DD4BF' },
  cream: { bg: '#FDF8F3', accent: '#F59E0B' },
  lavender: { bg: '#F5F3FF', accent: '#A78BFA' },
};

export function SettingsPage() {
  return (
    <div class="space-y-8 max-w-3xl">
      <div>
        <h1 class="text-2xl font-display font-bold text-foreground">Settings</h1>
        <p class="text-foreground-secondary mt-1">
          Manage your admin preferences
        </p>
      </div>

      {/* Profile section */}
      <div class="bg-background-secondary border border-border rounded-xl p-6">
        <h2 class="text-lg font-semibold text-foreground mb-4">Profile</h2>
        <div class="space-y-4">
          <div>
            <label class="text-sm font-medium text-foreground-secondary">
              Display Name
            </label>
            <p class="text-foreground mt-1">{user.value?.display_name}</p>
          </div>
          <div>
            <label class="text-sm font-medium text-foreground-secondary">
              Email
            </label>
            <p class="text-foreground mt-1">{user.value?.primary_email}</p>
          </div>
          <div>
            <label class="text-sm font-medium text-foreground-secondary">
              Connected Accounts
            </label>
            <div class="flex flex-wrap gap-2 mt-2">
              {user.value?.providers.map((provider) => (
                <span
                  key={provider}
                  class="px-2 py-1 bg-background-tertiary text-foreground-secondary rounded text-sm capitalize"
                >
                  {provider}
                </span>
              ))}
            </div>
          </div>
        </div>
      </div>

      {/* Theme section */}
      <div class="bg-background-secondary border border-border rounded-xl p-6">
        <h2 class="text-lg font-semibold text-foreground mb-4">Appearance</h2>
        <div class="space-y-4">
          <div>
            <label class="text-sm font-medium text-foreground-secondary mb-3 block">
              Theme
            </label>
            <div class="grid grid-cols-1 sm:grid-cols-3 gap-4">
              {themes.map((theme) => {
                const isSelected = currentTheme.value === theme.value;
                const colors = themePreviewColors[theme.value];

                return (
                  <button
                    key={theme.value}
                    onClick={() => setTheme(theme.value)}
                    class={`p-4 rounded-xl text-left transition-all duration-200 border-2 ${
                      isSelected
                        ? 'border-primary shadow-md'
                        : 'border-border hover:border-border-hover'
                    }`}
                  >
                    {/* Theme preview */}
                    <div
                      class="w-full h-16 rounded-lg mb-3 flex items-end justify-end p-2"
                      style={{ backgroundColor: colors.bg }}
                    >
                      <div
                        class="w-8 h-8 rounded-md"
                        style={{ backgroundColor: colors.accent }}
                      />
                    </div>

                    {/* Theme info */}
                    <div class="flex items-center gap-2 mb-1">
                      <span class="font-medium text-foreground">
                        {theme.label}
                      </span>
                      {isSelected && (
                        <svg
                          class="w-4 h-4 text-primary"
                          fill="currentColor"
                          viewBox="0 0 20 20"
                        >
                          <path
                            fill-rule="evenodd"
                            d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z"
                            clip-rule="evenodd"
                          />
                        </svg>
                      )}
                    </div>
                    <p class="text-sm text-foreground-secondary">
                      {theme.description}
                    </p>
                  </button>
                );
              })}
            </div>
          </div>
        </div>
      </div>

      {/* System section */}
      <div class="bg-background-secondary border border-border rounded-xl p-6">
        <h2 class="text-lg font-semibold text-foreground mb-4">System</h2>
        <div class="space-y-4">
          <div class="flex items-center justify-between">
            <div>
              <p class="font-medium text-foreground">Clear Cache</p>
              <p class="text-sm text-foreground-secondary">
                Clear all cached data and refresh
              </p>
            </div>
            <button class="px-4 py-2 text-sm font-medium text-foreground border border-border rounded-lg hover:bg-background-tertiary transition-colors">
              Clear
            </button>
          </div>
          <div class="flex items-center justify-between border-t border-border pt-4">
            <div>
              <p class="font-medium text-foreground">Export Data</p>
              <p class="text-sm text-foreground-secondary">
                Download all admin data as JSON
              </p>
            </div>
            <button class="px-4 py-2 text-sm font-medium text-foreground border border-border rounded-lg hover:bg-background-tertiary transition-colors">
              Export
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}
