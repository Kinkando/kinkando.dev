'use client';

import { useState } from 'react';

import { AuthGuard } from '@/components/auth/AuthGuard';
import { DashboardHeader } from '@/components/layout/DashboardHeader';
import { Sidebar } from '@/components/layout/Sidebar';

export default function DashboardLayout({ children }: { children: React.ReactNode }) {
  const [navOpen, setNavOpen] = useState(false);

  return (
    <AuthGuard>
      <div className="flex h-dvh min-h-screen">
        <Sidebar open={navOpen} onClose={() => setNavOpen(false)} />
        <div className="flex flex-1 flex-col overflow-hidden">
          <DashboardHeader onMenuClick={() => setNavOpen(true)} />
          <main className="flex-1 overflow-y-auto bg-gray-50 p-4 sm:p-6">{children}</main>
        </div>
      </div>
    </AuthGuard>
  );
}
