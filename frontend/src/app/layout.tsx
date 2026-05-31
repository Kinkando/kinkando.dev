import './globals.css';

import type { Metadata } from 'next';
import { PublicEnvScript } from 'next-runtime-env';

import { AuthProvider } from '@/contexts/AuthContext';

export const metadata: Metadata = {
  title: 'kinkando.dev',
  description: 'Personal dashboard — portfolio, finance, kanban'
};

export default function RootLayout({
  children
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en">
      <head>
        <PublicEnvScript />
      </head>
      <body className="antialiased">
        <AuthProvider>{children}</AuthProvider>
      </body>
    </html>
  );
}
