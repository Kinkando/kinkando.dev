'use client';

import Link from 'next/link';

import { Avatar } from '@/components/ui/Avatar';
import { useAuth } from '@/hooks/useAuth';

export function Navbar() {
  const { user } = useAuth();

  return (
    <header className="border-b border-gray-200 bg-white">
      <nav className="mx-auto flex max-w-6xl items-center justify-between px-4 py-3">
        <Link href="/" className="text-xl font-bold text-gray-900">
          kinkando.dev
        </Link>

        <div className="flex items-center gap-4">
          {user ? (
            <Link
              href="/dashboard"
              className="flex items-center gap-2 rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700"
            >
              <Avatar src={user.photoURL} name={user.displayName} email={user.email} size={22} />
              Dashboard
            </Link>
          ) : (
            <>
              <Link href="/login" className="text-sm font-medium text-gray-600 hover:text-gray-900">
                Sign In
              </Link>
              <Link href="/register" className="rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700">
                Register
              </Link>
            </>
          )}
        </div>
      </nav>
    </header>
  );
}
