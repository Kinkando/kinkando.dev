import { browser } from '$app/environment';
import { auth } from '$lib/firebase/client';
import type { User } from 'firebase/auth';
import { onAuthStateChanged } from 'firebase/auth';

export const authState = $state<{ user: User | null; loading: boolean }>({
  user: null,
  loading: true
});

// Module-level memoized promise so POST /users is called at most once per session.
let ensureUserPromise: Promise<void> | null = null;

export async function ensureUserOnce(): Promise<void> {
  if (!ensureUserPromise) {
    ensureUserPromise = (async () => {
      const { ensureUser } = await import('$lib/api/users');
      await ensureUser();
    })();
  }
  return ensureUserPromise;
}

/** Reset the memoized ensureUser promise (called on sign-out so next sign-in re-provisions). */
function resetEnsureUser() {
  ensureUserPromise = null;
}

let started = false;

/** Idempotent — safe to call multiple times; only subscribes once. Must run in browser. */
export function initAuth() {
  if (!browser || started) return;
  started = true;

  onAuthStateChanged(auth, async (user) => {
    authState.user = user;
    authState.loading = false;

    if (user) {
      await ensureUserOnce();
    } else {
      resetEnsureUser();
    }
  });
}
