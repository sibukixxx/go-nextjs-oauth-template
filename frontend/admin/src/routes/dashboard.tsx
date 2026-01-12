import { user } from '@/signals/auth';

export function DashboardPage() {
  return (
    <div class="space-y-8">
      {/* Welcome section */}
      <div>
        <h1 class="text-2xl font-display font-bold text-foreground">
          Dashboard
        </h1>
        <p class="text-foreground-secondary mt-1">
          Welcome back, {user.value?.display_name?.split(' ')[0]}!
        </p>
      </div>

      {/* Stats grid */}
      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        <StatCard
          title="Total Users"
          value="1,234"
          change="+12%"
          changeType="positive"
          icon={<UsersIcon />}
        />
        <StatCard
          title="Active Sessions"
          value="89"
          change="+5%"
          changeType="positive"
          icon={<SessionIcon />}
        />
        <StatCard
          title="API Calls (24h)"
          value="45.2K"
          change="-3%"
          changeType="negative"
          icon={<ApiIcon />}
        />
        <StatCard
          title="Error Rate"
          value="0.12%"
          change="-0.05%"
          changeType="positive"
          icon={<ErrorIcon />}
        />
      </div>

      {/* Recent activity */}
      <div class="bg-background-secondary border border-border rounded-xl p-6">
        <h2 class="text-lg font-semibold text-foreground mb-4">
          Recent Activity
        </h2>
        <div class="space-y-4">
          {recentActivity.map((activity, index) => (
            <div
              key={index}
              class="flex items-start gap-4 pb-4 border-b border-border last:border-0 last:pb-0"
            >
              <div
                class={`w-10 h-10 rounded-full flex items-center justify-center ${activity.iconBg}`}
              >
                {activity.icon}
              </div>
              <div class="flex-1 min-w-0">
                <p class="text-sm text-foreground">{activity.title}</p>
                <p class="text-sm text-foreground-secondary mt-0.5">
                  {activity.description}
                </p>
              </div>
              <span class="text-xs text-foreground-tertiary whitespace-nowrap">
                {activity.time}
              </span>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}

interface StatCardProps {
  title: string;
  value: string;
  change: string;
  changeType: 'positive' | 'negative';
  icon: JSX.Element;
}

function StatCard({ title, value, change, changeType, icon }: StatCardProps) {
  return (
    <div class="bg-background-secondary border border-border rounded-xl p-6">
      <div class="flex items-center justify-between">
        <div>
          <p class="text-sm text-foreground-secondary">{title}</p>
          <p class="text-3xl font-bold text-foreground mt-1">{value}</p>
        </div>
        <div class="w-12 h-12 bg-primary/10 rounded-xl flex items-center justify-center text-primary">
          {icon}
        </div>
      </div>
      <div class="flex items-center gap-1 mt-4 text-sm">
        <span
          class={`px-2 py-0.5 rounded-full text-xs font-medium ${
            changeType === 'positive'
              ? 'bg-emerald-100 text-emerald-700'
              : 'bg-red-100 text-red-700'
          }`}
        >
          {change}
        </span>
        <span class="text-foreground-secondary">from last month</span>
      </div>
    </div>
  );
}

const recentActivity = [
  {
    title: 'New user registered',
    description: 'john@example.com signed up via Google',
    time: '2 min ago',
    icon: <UserPlusIcon />,
    iconBg: 'bg-primary/10',
  },
  {
    title: 'Failed login attempt',
    description: 'Invalid credentials for admin@example.com',
    time: '15 min ago',
    icon: <AlertIcon />,
    iconBg: 'bg-warning/10',
  },
  {
    title: 'API rate limit reached',
    description: 'Client ID: abc123 exceeded 1000 req/min',
    time: '1 hour ago',
    icon: <AlertIcon />,
    iconBg: 'bg-error/10',
  },
];

// Icons
function UsersIcon() {
  return (
    <svg class="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
      <path
        stroke-linecap="round"
        stroke-linejoin="round"
        stroke-width={2}
        d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z"
      />
    </svg>
  );
}

function SessionIcon() {
  return (
    <svg class="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
      <path
        stroke-linecap="round"
        stroke-linejoin="round"
        stroke-width={2}
        d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"
      />
    </svg>
  );
}

function ApiIcon() {
  return (
    <svg class="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
      <path
        stroke-linecap="round"
        stroke-linejoin="round"
        stroke-width={2}
        d="M8 9l3 3-3 3m5 0h3M5 20h14a2 2 0 002-2V6a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z"
      />
    </svg>
  );
}

function ErrorIcon() {
  return (
    <svg class="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
      <path
        stroke-linecap="round"
        stroke-linejoin="round"
        stroke-width={2}
        d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"
      />
    </svg>
  );
}

function UserPlusIcon() {
  return (
    <svg class="w-5 h-5 text-primary" fill="none" viewBox="0 0 24 24" stroke="currentColor">
      <path
        stroke-linecap="round"
        stroke-linejoin="round"
        stroke-width={2}
        d="M18 9v3m0 0v3m0-3h3m-3 0h-3m-2-5a4 4 0 11-8 0 4 4 0 018 0zM3 20a6 6 0 0112 0v1H3v-1z"
      />
    </svg>
  );
}

function AlertIcon() {
  return (
    <svg class="w-5 h-5 text-warning" fill="none" viewBox="0 0 24 24" stroke="currentColor">
      <path
        stroke-linecap="round"
        stroke-linejoin="round"
        stroke-width={2}
        d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"
      />
    </svg>
  );
}
