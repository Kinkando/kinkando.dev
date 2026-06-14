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
