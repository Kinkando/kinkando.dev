import { getApps, getApp, initializeApp } from 'firebase/app'
import { getAuth } from 'firebase/auth'
import env from '../config/env'

export const app = getApps().length
  ? getApp()
  : initializeApp(env.firebaseConfig)
export const auth = getAuth(app)

export async function getIdToken(): Promise<string | null> {
  const user = auth.currentUser
  if (!user) return null
  return user.getIdToken()
}

export function stateReady(): Promise<void> {
  return auth.authStateReady()
}

const AUTH_ERROR_MESSAGES: Record<string, string> = {
  'auth/invalid-email': 'Invalid email address.',
  'auth/user-not-found': 'No account with that email.',
  'auth/wrong-password': 'Incorrect password.',
  'auth/invalid-credential': 'Incorrect email or password.',
  'auth/email-already-in-use': 'An account with that email already exists.',
  'auth/weak-password': 'Password must be at least 6 characters.',
  'auth/too-many-requests': 'Too many attempts. Try again later.',
  'auth/popup-closed-by-user': 'Sign-in popup was closed.',
  'auth/cancelled-popup-request': 'Popup cancelled.',
  'auth/popup-blocked': 'Popup was blocked by the browser.',
  'auth/credential-already-in-use':
    'That Google account is already linked to another user.',
  'auth/provider-already-linked': 'This account is already linked.',
  'auth/no-such-provider': 'No such provider to unlink.',
  'auth/requires-recent-login':
    'Please sign in again before changing this setting.',
}

export function friendlyAuthError(code?: string): string {
  return AUTH_ERROR_MESSAGES[code ?? ''] ?? 'Authentication failed. Please try again.'
}
