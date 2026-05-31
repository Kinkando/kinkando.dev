'use client';

import { useAuth } from '@/hooks/useAuth';

export default function DashboardPage() {
  const { user } = useAuth();

  return (
    <div>
      <h1 className="text-2xl font-bold text-gray-900">Dashboard</h1>
      <p className="mt-2 text-gray-600">Welcome back, {user?.email ?? 'user'}!</p>

      <div className="mt-8 grid grid-cols-1 gap-6 sm:grid-cols-2">
        <a href="/dashboard/finance" className="rounded-lg border border-gray-200 bg-white p-6 shadow-sm transition-shadow hover:shadow-md">
          <h2 className="text-lg font-semibold text-gray-900">Finance</h2>
          <p className="mt-1 text-sm text-gray-500">Track your income and expenses.</p>
        </a>
        <a href="/dashboard/kanban" className="rounded-lg border border-gray-200 bg-white p-6 shadow-sm transition-shadow hover:shadow-md">
          <h2 className="text-lg font-semibold text-gray-900">Kanban</h2>
          <p className="mt-1 text-sm text-gray-500">Manage your tasks with boards.</p>
        </a>
      </div>
    </div>
  );
}
