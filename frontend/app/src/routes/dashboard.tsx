import { Card, CardHeader, CardTitle, CardDescription, CardContent, Badge } from '@app/ui';
import { useAuthStore } from '@/stores/authStore';

export function DashboardPage() {
  const { user } = useAuthStore();

  return (
    <div className="space-y-8">
      {/* Welcome section */}
      <div>
        <h1 className="text-2xl font-display font-bold text-foreground">
          Welcome back, {user?.display_name?.split(' ')[0]}!
        </h1>
        <p className="text-foreground-secondary mt-1">
          Here's what's happening with your account today.
        </p>
      </div>

      {/* Stats grid */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-foreground-secondary">Total Projects</p>
                <p className="text-3xl font-bold text-foreground mt-1">12</p>
              </div>
              <div className="w-12 h-12 bg-primary/10 rounded-xl flex items-center justify-center">
                <FolderIcon className="w-6 h-6 text-primary" />
              </div>
            </div>
            <div className="flex items-center gap-1 mt-4 text-sm">
              <Badge variant="success">+2</Badge>
              <span className="text-foreground-secondary">from last month</span>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-foreground-secondary">Active Tasks</p>
                <p className="text-3xl font-bold text-foreground mt-1">28</p>
              </div>
              <div className="w-12 h-12 bg-warning/10 rounded-xl flex items-center justify-center">
                <TaskIcon className="w-6 h-6 text-warning" />
              </div>
            </div>
            <div className="flex items-center gap-1 mt-4 text-sm">
              <Badge variant="warning">5</Badge>
              <span className="text-foreground-secondary">due this week</span>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-foreground-secondary">Team Members</p>
                <p className="text-3xl font-bold text-foreground mt-1">8</p>
              </div>
              <div className="w-12 h-12 bg-info/10 rounded-xl flex items-center justify-center">
                <UsersIcon className="w-6 h-6 text-info" />
              </div>
            </div>
            <div className="flex items-center gap-1 mt-4 text-sm">
              <Badge variant="info">+1</Badge>
              <span className="text-foreground-secondary">new this week</span>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Recent activity */}
      <Card>
        <CardHeader>
          <CardTitle>Recent Activity</CardTitle>
          <CardDescription>Your latest actions and updates</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            {recentActivity.map((activity, index) => (
              <div
                key={index}
                className="flex items-start gap-4 pb-4 border-b border-border last:border-0 last:pb-0"
              >
                <div
                  className={`w-10 h-10 rounded-full flex items-center justify-center ${activity.iconBg}`}
                >
                  {activity.icon}
                </div>
                <div className="flex-1 min-w-0">
                  <p className="text-sm text-foreground">{activity.title}</p>
                  <p className="text-sm text-foreground-secondary mt-0.5">
                    {activity.description}
                  </p>
                </div>
                <span className="text-xs text-foreground-tertiary whitespace-nowrap">
                  {activity.time}
                </span>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>
    </div>
  );
}

const recentActivity = [
  {
    title: 'Project "Website Redesign" created',
    description: 'You created a new project',
    time: '2 hours ago',
    icon: <FolderIcon className="w-5 h-5 text-primary" />,
    iconBg: 'bg-primary/10',
  },
  {
    title: 'Task "Update homepage" completed',
    description: 'Marked as done in Website Redesign',
    time: '5 hours ago',
    icon: <CheckIcon className="w-5 h-5 text-success" />,
    iconBg: 'bg-success/10',
  },
  {
    title: 'New team member joined',
    description: 'Sarah joined the Marketing team',
    time: '1 day ago',
    icon: <UsersIcon className="w-5 h-5 text-info" />,
    iconBg: 'bg-info/10',
  },
];

// Icons
function FolderIcon({ className }: { className?: string }) {
  return (
    <svg className={className} fill="none" viewBox="0 0 24 24" stroke="currentColor">
      <path
        strokeLinecap="round"
        strokeLinejoin="round"
        strokeWidth={2}
        d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z"
      />
    </svg>
  );
}

function TaskIcon({ className }: { className?: string }) {
  return (
    <svg className={className} fill="none" viewBox="0 0 24 24" stroke="currentColor">
      <path
        strokeLinecap="round"
        strokeLinejoin="round"
        strokeWidth={2}
        d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4"
      />
    </svg>
  );
}

function UsersIcon({ className }: { className?: string }) {
  return (
    <svg className={className} fill="none" viewBox="0 0 24 24" stroke="currentColor">
      <path
        strokeLinecap="round"
        strokeLinejoin="round"
        strokeWidth={2}
        d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z"
      />
    </svg>
  );
}

function CheckIcon({ className }: { className?: string }) {
  return (
    <svg className={className} fill="none" viewBox="0 0 24 24" stroke="currentColor">
      <path
        strokeLinecap="round"
        strokeLinejoin="round"
        strokeWidth={2}
        d="M5 13l4 4L19 7"
      />
    </svg>
  );
}
