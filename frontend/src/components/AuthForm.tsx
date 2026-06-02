import { useState } from 'react'
import type { FormEvent } from 'react'
import {
  signInWithEmailAndPassword,
  createUserWithEmailAndPassword,
  signInWithPopup,
  getAdditionalUserInfo,
  GoogleAuthProvider,
} from 'firebase/auth'
import { useNavigate, useSearchParams } from 'react-router-dom'
import { auth } from '../lib/firebase'

type Props = {
  mode: 'login' | 'register'
}

function friendlyError(code: string): string {
  const map: Record<string, string> = {
    'auth/invalid-email': 'Invalid email address.',
    'auth/user-not-found': 'No account with that email.',
    'auth/wrong-password': 'Incorrect password.',
    'auth/invalid-credential': 'Incorrect email or password.',
    'auth/email-already-in-use': 'An account with that email already exists.',
    'auth/weak-password': 'Password must be at least 6 characters.',
    'auth/too-many-requests': 'Too many attempts. Try again later.',
    'auth/popup-closed-by-user': 'Sign-in popup was closed.',
  }
  return map[code] ?? 'Authentication failed. Please try again.'
}

export default function AuthForm({ mode }: Props) {
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  const navigate = useNavigate()
  const [params] = useSearchParams()
  const redirect = params.get('redirect') ?? '/kanban'

  async function handleSubmit(e: FormEvent) {
    e.preventDefault()
    setError('')
    setLoading(true)
    try {
      if (mode === 'login') {
        await signInWithEmailAndPassword(auth, email, password)
      } else {
        await createUserWithEmailAndPassword(auth, email, password)
      }
      navigate(redirect)
    } catch (err: unknown) {
      setError(friendlyError((err as { code?: string }).code ?? ''))
    } finally {
      setLoading(false)
    }
  }

  async function handleGoogle() {
    setError('')
    setLoading(true)
    try {
      const result = await signInWithPopup(auth, new GoogleAuthProvider())
      if (getAdditionalUserInfo(result)?.isNewUser) {
        await result.user.delete()
        setError('No account found for this Google account.')
        return
      }
      navigate(redirect)
    } catch (err: unknown) {
      setError(friendlyError((err as { code?: string }).code ?? ''))
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="mx-auto mt-20 max-w-sm rounded-xl border border-gray-800 bg-gray-900 p-8">
      <h1 className="mb-6 text-xl font-semibold text-gray-100">
        {mode === 'login' ? 'Sign in' : 'Create account'}
      </h1>
      {error && <p className="mb-4 text-sm text-red-400">{error}</p>}
      <form onSubmit={handleSubmit} className="flex flex-col gap-4">
        <input
          type="email"
          placeholder="Email"
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          required
          className="rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100 placeholder-gray-500 focus:border-indigo-500 focus:outline-none"
        />
        <input
          type="password"
          placeholder="Password"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          required
          className="rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100 placeholder-gray-500 focus:border-indigo-500 focus:outline-none"
        />
        <button
          type="submit"
          disabled={loading}
          className="cursor-pointer rounded-lg bg-indigo-600 py-2 text-sm font-medium text-white hover:bg-indigo-500 disabled:opacity-50"
        >
          {mode === 'login' ? 'Sign in' : 'Register'}
        </button>
      </form>
      <div className="my-4 border-t border-gray-800" />
      <button
        onClick={handleGoogle}
        disabled={loading}
        className="flex w-full cursor-pointer items-center justify-center gap-3 rounded-lg border border-gray-700 bg-gray-800 py-2 text-sm text-gray-200 hover:bg-gray-700 disabled:opacity-50"
      >
        <svg
          xmlns="http://www.w3.org/2000/svg"
          viewBox="0 0 24 24"
          className="h-4 w-4 shrink-0"
        >
          <path
            fill="#4285F4"
            d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"
          />
          <path
            fill="#34A853"
            d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"
          />
          <path
            fill="#FBBC05"
            d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z"
          />
          <path
            fill="#EA4335"
            d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"
          />
        </svg>
        Sign in with Google
      </button>
    </div>
  )
}
