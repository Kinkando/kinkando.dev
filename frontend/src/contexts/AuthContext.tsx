'use client';

import type { User } from 'firebase/auth';
import { createContext, type ReactNode, useCallback, useEffect, useState } from 'react';

import { ensureUser } from '@/lib/api/user';
import { getFirebaseAuth } from '@/lib/firebase';

interface AuthContextValue {
  user: User | null;
  loading: boolean;
  signIn: (email: string, password: string) => Promise<void>;
  signUp: (email: string, password: string) => Promise<void>;
  signInWithGoogle: () => Promise<void>;
  signOut: () => Promise<void>;
}

export const AuthContext = createContext<AuthContextValue | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let unsubscribe: (() => void) | undefined;

    (async () => {
      const { onAuthStateChanged } = await import('firebase/auth');
      const auth = await getFirebaseAuth();
      unsubscribe = onAuthStateChanged(auth, (u) => {
        setUser(u);
        setLoading(false);
      });
    })();

    return () => unsubscribe?.();
  }, []);

  const signIn = useCallback(async (email: string, password: string) => {
    const { signInWithEmailAndPassword } = await import('firebase/auth');
    const auth = await getFirebaseAuth();
    await signInWithEmailAndPassword(auth, email, password);
    // Provision (or update) the backend users row so finance queries resolve correctly.
    try {
      await ensureUser();
    } catch (err) {
      console.error('ensureUser failed on sign-in:', err);
    }
  }, []);

  const signUp = useCallback(async (email: string, password: string) => {
    const { createUserWithEmailAndPassword } = await import('firebase/auth');
    const auth = await getFirebaseAuth();
    await createUserWithEmailAndPassword(auth, email, password);
    // Provision the backend users row for the newly created account.
    try {
      await ensureUser();
    } catch (err) {
      console.error('ensureUser failed on sign-up:', err);
    }
  }, []);

  const signInWithGoogle = useCallback(async () => {
    const { GoogleAuthProvider, signInWithPopup } = await import('firebase/auth');
    const auth = await getFirebaseAuth();
    await signInWithPopup(auth, new GoogleAuthProvider());
    // Provision (or update) the backend users row so finance queries resolve correctly.
    try {
      await ensureUser();
    } catch (err) {
      console.error('ensureUser failed on Google sign-in:', err);
    }
  }, []);

  const signOut = useCallback(async () => {
    const { signOut: firebaseSignOut } = await import('firebase/auth');
    const auth = await getFirebaseAuth();
    await firebaseSignOut(auth);
  }, []);

  return <AuthContext.Provider value={{ user, loading, signIn, signUp, signInWithGoogle, signOut }}>{children}</AuthContext.Provider>;
}
