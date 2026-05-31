'use client';

import { Avatar } from '@/components/ui/Avatar';
import { Button } from '@/components/ui/Button';
import { useAuth } from '@/hooks/useAuth';

interface Props {
  onMenuClick: () => void;
}

export function DashboardHeader({ onMenuClick }: Props) {
  const { user, signOut } = useAuth();

  return (
    <header className="flex items-center border-b border-gray-200 bg-white px-4 py-3 sm:px-6">
      <button
        type="button"
        onClick={onMenuClick}
        aria-label="Open navigation menu"
        className="-ml-1 rounded p-1 text-2xl leading-none text-gray-600 hover:text-gray-900 md:hidden"
      >
        ☰
      </button>
      <div className="ml-auto flex items-center gap-3">
        <Avatar src={user?.photoURL} name={user?.displayName} email={user?.email} />
        <span className="max-w-[40vw] truncate text-sm text-gray-600 sm:max-w-none">{user?.displayName ?? user?.email}</span>
        <Button variant="secondary" onClick={signOut}>
          Sign Out
        </Button>
      </div>
    </header>
  );
}
