'use client';

import Link from 'next/link';
import { usePathname } from 'next/navigation';

const navItems = [
  { href: '/dashboard', label: 'Overview' },
  { href: '/dashboard/finance', label: 'Finance' },
  { href: '/dashboard/kanban', label: 'Kanban' }
];

interface Props {
  open: boolean;
  onClose: () => void;
}

export function Sidebar({ open, onClose }: Props) {
  const pathname = usePathname();

  return (
    <>
      {/* Backdrop — only below md when the drawer is open */}
      {open && <div className="fixed inset-0 z-40 bg-black/40 md:hidden" onClick={onClose} />}

      <aside
        className={`fixed inset-y-0 left-0 z-50 flex w-64 transform flex-col border-r border-gray-200 bg-gray-50 transition-transform md:static md:z-auto md:w-56 md:translate-x-0 ${
          open ? 'translate-x-0' : '-translate-x-full'
        }`}
      >
        <div className="p-4">
          <Link href="/" className="text-lg font-bold text-gray-900" onClick={onClose}>
            kinkando.dev
          </Link>
        </div>

        <nav className="flex-1 space-y-1 px-2 py-4">
          {navItems.map((item) => {
            const active = pathname === item.href;
            return (
              <Link
                key={item.href}
                href={item.href}
                onClick={onClose}
                className={`block rounded-lg px-3 py-2 text-sm font-medium transition-colors ${
                  active ? 'bg-blue-50 text-blue-700' : 'text-gray-600 hover:bg-gray-100 hover:text-gray-900'
                }`}
              >
                {item.label}
              </Link>
            );
          })}
        </nav>
      </aside>
    </>
  );
}
