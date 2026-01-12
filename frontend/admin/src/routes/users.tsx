import { signal } from '@preact/signals';

// Mock users data
const users = signal([
  {
    id: '1',
    name: 'John Doe',
    email: 'john@example.com',
    provider: 'google',
    status: 'active',
    createdAt: '2024-01-15',
  },
  {
    id: '2',
    name: 'Jane Smith',
    email: 'jane@example.com',
    provider: 'line',
    status: 'active',
    createdAt: '2024-01-14',
  },
  {
    id: '3',
    name: 'Bob Johnson',
    email: 'bob@example.com',
    provider: 'google',
    status: 'inactive',
    createdAt: '2024-01-10',
  },
]);

export function UsersPage() {
  return (
    <div class="space-y-8">
      {/* Header */}
      <div class="flex items-center justify-between">
        <div>
          <h1 class="text-2xl font-display font-bold text-foreground">Users</h1>
          <p class="text-foreground-secondary mt-1">
            Manage user accounts and permissions
          </p>
        </div>
        <button class="px-4 py-2 bg-primary text-white rounded-lg font-medium hover:bg-primary-hover transition-colors">
          Add User
        </button>
      </div>

      {/* Search and filters */}
      <div class="flex items-center gap-4">
        <div class="flex-1 relative">
          <SearchIcon class="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-foreground-tertiary" />
          <input
            type="text"
            placeholder="Search users..."
            class="w-full pl-10 pr-4 py-2.5 bg-background-secondary border border-border rounded-lg text-foreground placeholder:text-foreground-tertiary focus:outline-none focus:ring-2 focus:ring-primary-light focus:border-primary"
          />
        </div>
        <select class="px-4 py-2.5 bg-background-secondary border border-border rounded-lg text-foreground focus:outline-none focus:ring-2 focus:ring-primary-light focus:border-primary">
          <option value="">All Status</option>
          <option value="active">Active</option>
          <option value="inactive">Inactive</option>
        </select>
      </div>

      {/* Users table */}
      <div class="bg-background-secondary border border-border rounded-xl overflow-hidden">
        <table class="w-full">
          <thead>
            <tr class="border-b border-border">
              <th class="px-6 py-4 text-left text-sm font-medium text-foreground-secondary">
                User
              </th>
              <th class="px-6 py-4 text-left text-sm font-medium text-foreground-secondary">
                Provider
              </th>
              <th class="px-6 py-4 text-left text-sm font-medium text-foreground-secondary">
                Status
              </th>
              <th class="px-6 py-4 text-left text-sm font-medium text-foreground-secondary">
                Created
              </th>
              <th class="px-6 py-4 text-right text-sm font-medium text-foreground-secondary">
                Actions
              </th>
            </tr>
          </thead>
          <tbody>
            {users.value.map((user) => (
              <tr key={user.id} class="border-b border-border last:border-0">
                <td class="px-6 py-4">
                  <div class="flex items-center gap-3">
                    <div class="w-10 h-10 bg-primary rounded-full flex items-center justify-center">
                      <span class="text-white font-medium">
                        {user.name.charAt(0).toUpperCase()}
                      </span>
                    </div>
                    <div>
                      <p class="font-medium text-foreground">{user.name}</p>
                      <p class="text-sm text-foreground-secondary">
                        {user.email}
                      </p>
                    </div>
                  </div>
                </td>
                <td class="px-6 py-4">
                  <span class="px-2 py-1 bg-background-tertiary text-foreground-secondary rounded text-sm capitalize">
                    {user.provider}
                  </span>
                </td>
                <td class="px-6 py-4">
                  <span
                    class={`px-2 py-1 rounded-full text-xs font-medium ${
                      user.status === 'active'
                        ? 'bg-emerald-100 text-emerald-700'
                        : 'bg-gray-100 text-gray-700'
                    }`}
                  >
                    {user.status}
                  </span>
                </td>
                <td class="px-6 py-4 text-sm text-foreground-secondary">
                  {user.createdAt}
                </td>
                <td class="px-6 py-4 text-right">
                  <button class="p-2 text-foreground-secondary hover:text-foreground rounded-lg hover:bg-background-tertiary transition-colors">
                    <MoreIcon class="w-5 h-5" />
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {/* Pagination */}
      <div class="flex items-center justify-between">
        <p class="text-sm text-foreground-secondary">
          Showing 1-{users.value.length} of {users.value.length} users
        </p>
        <div class="flex items-center gap-2">
          <button class="px-4 py-2 text-sm text-foreground-secondary border border-border rounded-lg hover:bg-background-tertiary transition-colors disabled:opacity-50">
            Previous
          </button>
          <button class="px-4 py-2 text-sm text-foreground-secondary border border-border rounded-lg hover:bg-background-tertiary transition-colors disabled:opacity-50">
            Next
          </button>
        </div>
      </div>
    </div>
  );
}

function SearchIcon({ class: className }: { class?: string }) {
  return (
    <svg class={className} fill="none" viewBox="0 0 24 24" stroke="currentColor">
      <path
        stroke-linecap="round"
        stroke-linejoin="round"
        stroke-width={2}
        d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"
      />
    </svg>
  );
}

function MoreIcon({ class: className }: { class?: string }) {
  return (
    <svg class={className} fill="none" viewBox="0 0 24 24" stroke="currentColor">
      <path
        stroke-linecap="round"
        stroke-linejoin="round"
        stroke-width={2}
        d="M12 5v.01M12 12v.01M12 19v.01M12 6a1 1 0 110-2 1 1 0 010 2zm0 7a1 1 0 110-2 1 1 0 010 2zm0 7a1 1 0 110-2 1 1 0 010 2z"
      />
    </svg>
  );
}
