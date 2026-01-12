import { ComponentChildren } from 'preact';
import { user, logout } from '@/signals/auth';

const navItems = [
  { path: '/', label: 'Dashboard', icon: DashboardIcon },
  { path: '/users', label: 'Users', icon: UsersIcon },
  { path: '/settings', label: 'Settings', icon: SettingsIcon },
];

interface LayoutProps {
  children: ComponentChildren;
}

export function Layout({ children }: LayoutProps) {
  const currentPath = typeof window !== 'undefined' ? window.location.pathname : '/';

  return (
    <div class="min-h-screen bg-background flex">
      {/* Sidebar */}
      <aside class="w-64 bg-background-secondary border-r border-border flex flex-col">
        {/* Logo */}
        <div class="h-16 flex items-center px-6 border-b border-border">
          <div class="w-8 h-8 bg-primary rounded-lg flex items-center justify-center">
            <span class="text-white font-bold text-sm">A</span>
          </div>
          <span class="ml-3 font-display font-semibold text-foreground">
            Admin
          </span>
        </div>

        {/* Navigation */}
        <nav class="flex-1 p-4 space-y-1">
          {navItems.map((item) => {
            const isActive = currentPath === item.path;
            const Icon = item.icon;
            return (
              <a
                key={item.path}
                href={item.path}
                class={`flex items-center gap-3 px-4 py-2.5 rounded-lg text-sm font-medium transition-colors ${
                  isActive
                    ? 'bg-primary/10 text-primary'
                    : 'text-foreground-secondary hover:text-foreground hover:bg-background-tertiary'
                }`}
              >
                <Icon class="w-5 h-5" />
                {item.label}
              </a>
            );
          })}
        </nav>

        {/* User section */}
        <div class="p-4 border-t border-border">
          <div class="flex items-center gap-3">
            {user.value?.avatar_url ? (
              <img
                src={user.value.avatar_url}
                alt={user.value.display_name}
                class="w-10 h-10 rounded-full"
              />
            ) : (
              <div class="w-10 h-10 bg-primary rounded-full flex items-center justify-center">
                <span class="text-white font-medium">
                  {user.value?.display_name?.charAt(0).toUpperCase()}
                </span>
              </div>
            )}
            <div class="flex-1 min-w-0">
              <p class="text-sm font-medium text-foreground truncate">
                {user.value?.display_name}
              </p>
              <p class="text-xs text-foreground-secondary truncate">
                {user.value?.primary_email}
              </p>
            </div>
            <button
              onClick={() => logout()}
              class="p-2 text-foreground-secondary hover:text-foreground rounded-lg hover:bg-background-tertiary transition-colors"
              title="Logout"
            >
              <LogoutIcon class="w-5 h-5" />
            </button>
          </div>
        </div>
      </aside>

      {/* Main content */}
      <main class="flex-1 p-8 overflow-auto">{children}</main>
    </div>
  );
}

// Icons
function DashboardIcon({ class: className }: { class?: string }) {
  return (
    <svg class={className} fill="none" viewBox="0 0 24 24" stroke="currentColor">
      <path
        stroke-linecap="round"
        stroke-linejoin="round"
        stroke-width={2}
        d="M4 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2V6zM14 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2V6zM4 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2v-2zM14 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2v-2z"
      />
    </svg>
  );
}

function UsersIcon({ class: className }: { class?: string }) {
  return (
    <svg class={className} fill="none" viewBox="0 0 24 24" stroke="currentColor">
      <path
        stroke-linecap="round"
        stroke-linejoin="round"
        stroke-width={2}
        d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z"
      />
    </svg>
  );
}

function SettingsIcon({ class: className }: { class?: string }) {
  return (
    <svg class={className} fill="none" viewBox="0 0 24 24" stroke="currentColor">
      <path
        stroke-linecap="round"
        stroke-linejoin="round"
        stroke-width={2}
        d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z"
      />
      <path
        stroke-linecap="round"
        stroke-linejoin="round"
        stroke-width={2}
        d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"
      />
    </svg>
  );
}

function LogoutIcon({ class: className }: { class?: string }) {
  return (
    <svg class={className} fill="none" viewBox="0 0 24 24" stroke="currentColor">
      <path
        stroke-linecap="round"
        stroke-linejoin="round"
        stroke-width={2}
        d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1"
      />
    </svg>
  );
}
