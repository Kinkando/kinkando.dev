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
