import type { FirebaseApp } from 'firebase/app';
import type { Auth } from 'firebase/auth';

import config from '@/config/config';

let _app: FirebaseApp | null = null;
let _auth: Auth | null = null;

/**
 * Returns the Firebase Auth instance. Client-side only.
 * Uses dynamic import so firebase modules are never pulled into
 * the SSR / static-generation bundle.
 */
export async function getFirebaseAuth(): Promise<Auth> {
  if (_auth) return _auth;

  const { initializeApp, getApps, getApp } = await import('firebase/app');
  const { getAuth } = await import('firebase/auth');

  if (!_app) {
    _app = getApps().length ? getApp() : initializeApp(config.firebase);
  }

  _auth = getAuth(_app);
  return _auth;
}
