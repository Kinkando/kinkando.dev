import { auth } from './client';

/**
 * Returns the current user's Firebase ID token, or null if not signed in.
 * getIdToken() returns a cached token and auto-refreshes near expiry.
 */
export async function getToken(): Promise<string | null> {
  const user = auth.currentUser;
  if (!user) return null;
  return user.getIdToken();
}
