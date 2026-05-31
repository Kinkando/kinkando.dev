'use client';

import { Avatar } from '@/components/ui/Avatar';
import { Button } from '@/components/ui/Button';
import { useAuth } from '@/hooks/useAuth';

export function DashboardHeader() {
  const { user, signOut } = useAuth();

  return (
    <header className="flex items-center justify-between border-b border-gray-200 bg-white px-6 py-3">
      <div />
      <div className="flex items-center gap-3">
        <Avatar src={user?.photoURL} name={user?.displayName} email={user?.email} />
        <span className="text-sm text-gray-600">{user?.displayName ?? user?.email}</span>
        <Button variant="secondary" onClick={signOut}>
          Sign Out
        </Button>
      </div>
    </header>
  );
}
