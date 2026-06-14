import { useState } from 'react'
import type { FormEvent } from 'react'
import {
  signInWithEmailAndPassword,
  createUserWithEmailAndPassword,
  signInWithPopup,
  getAdditionalUserInfo,
  GoogleAuthProvider,
  sendPasswordResetEmail,
} from 'firebase/auth'
import { useNavigate, useSearchParams } from 'react-router-dom'
import { auth, friendlyAuthError } from '../lib/firebase'
import GoogleIcon from './icons/GoogleIcon'

type Props = {
  mode: 'login' | 'register'
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
      setError(friendlyAuthError((err as { code?: string }).code))
    } finally {
      setLoading(false)
    }
  }

  // Forgot password (login mode only)
  const [forgotOpen, setForgotOpen] = useState(false)
  const [forgotEmail, setForgotEmail] = useState('')
  const [forgotSending, setForgotSending] = useState(false)
  const [forgotSent, setForgotSent] = useState(false)
  const [forgotError, setForgotError] = useState('')

  async function handleForgot(e: FormEvent) {
    e.preventDefault()
    setForgotError('')
    setForgotSending(true)
    try {
      await sendPasswordResetEmail(auth, forgotEmail, {
        url: `${window.location.origin}/reset-password`,
        handleCodeInApp: false,
      })
      setForgotSent(true)
    } catch (err: unknown) {
      setForgotError(friendlyAuthError((err as { code?: string }).code))
    } finally {
      setForgotSending(false)
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
      setError(friendlyAuthError((err as { code?: string }).code))
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
      {mode === 'login' && (
        <div className="mt-3">
          {!forgotOpen ? (
            <button
              type="button"
              onClick={() => {
                setForgotOpen(true)
                setForgotEmail(email)
                setForgotSent(false)
                setForgotError('')
              }}
              className="cursor-pointer text-xs text-indigo-400 hover:text-indigo-300"
            >
              Forgot password?
            </button>
          ) : forgotSent ? (
            <div className="space-y-2 rounded-lg border border-emerald-700/40 bg-emerald-900/20 p-3">
              <p className="text-sm text-emerald-400">
                Reset email sent. Check your inbox.
              </p>
              <button
                type="button"
                onClick={() => setForgotOpen(false)}
                className="cursor-pointer text-xs text-gray-400 hover:text-gray-200"
              >
                Close
              </button>
            </div>
          ) : (
            <form
              onSubmit={handleForgot}
              className="flex flex-col gap-2 rounded-lg border border-gray-800 bg-gray-950/50 p-3"
            >
              <p className="text-xs text-gray-400">
                Enter your email and we'll send a reset link.
              </p>
              <input
                type="email"
                placeholder="Email"
                value={forgotEmail}
                onChange={(e) => setForgotEmail(e.target.value)}
                required
                className="rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100 placeholder-gray-500 focus:border-indigo-500 focus:outline-none"
              />
              {forgotError && (
                <p className="text-xs text-red-400">{forgotError}</p>
              )}
              <div className="flex justify-end gap-2">
                <button
                  type="button"
                  onClick={() => setForgotOpen(false)}
                  className="cursor-pointer rounded-lg bg-gray-800 px-3 py-1.5 text-xs font-medium text-gray-300 hover:bg-gray-700"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  disabled={forgotSending}
                  className="cursor-pointer rounded-lg bg-indigo-600 px-3 py-1.5 text-xs font-medium text-white hover:bg-indigo-500 disabled:opacity-50"
                >
                  {forgotSending ? 'Sending…' : 'Send reset link'}
                </button>
              </div>
            </form>
          )}
        </div>
      )}
      <div className="my-4 border-t border-gray-800" />
      <button
        onClick={handleGoogle}
        disabled={loading}
        className="flex w-full cursor-pointer items-center justify-center gap-3 rounded-lg border border-gray-700 bg-gray-800 py-2 text-sm text-gray-200 hover:bg-gray-700 disabled:opacity-50"
      >
        <GoogleIcon className="h-4 w-4 shrink-0" />
        Sign in with Google
      </button>
    </div>
  )
}
