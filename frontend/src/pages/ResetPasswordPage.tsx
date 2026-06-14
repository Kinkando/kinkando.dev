import { useEffect, useState } from 'react'
import type { FormEvent } from 'react'
import { Link, useSearchParams } from 'react-router-dom'
import { confirmPasswordReset, verifyPasswordResetCode } from 'firebase/auth'
import { useDocumentTitle } from '../hooks/useDocumentTitle'
import { auth, friendlyAuthError } from '../lib/firebase'

type Phase = 'verifying' | 'form' | 'success' | 'error'

export default function ResetPasswordPage() {
  useDocumentTitle('Reset password')

  const [params] = useSearchParams()
  const mode = params.get('mode')
  const oobCode = params.get('oobCode')

  const [phase, setPhase] = useState<Phase>('verifying')
  const [email, setEmail] = useState('')
  const [error, setError] = useState('')

  const [password, setPassword] = useState('')
  const [confirm, setConfirm] = useState('')
  const [submitting, setSubmitting] = useState(false)

  useEffect(() => {
    if (mode !== 'resetPassword' || !oobCode) {
      setError('This password reset link is invalid or has already been used.')
      setPhase('error')
      return
    }
    verifyPasswordResetCode(auth, oobCode)
      .then((verifiedEmail) => {
        setEmail(verifiedEmail)
        setPhase('form')
      })
      .catch((err: unknown) => {
        setError(friendlyAuthError((err as { code?: string }).code))
        setPhase('error')
      })
  }, [mode, oobCode])

  async function handleSubmit(e: FormEvent) {
    e.preventDefault()
    if (!oobCode) return
    setError('')
    if (password.length < 6) {
      setError('Password must be at least 6 characters.')
      return
    }
    if (password !== confirm) {
      setError('Passwords do not match.')
      return
    }
    setSubmitting(true)
    try {
      await confirmPasswordReset(auth, oobCode, password)
      setPhase('success')
    } catch (err: unknown) {
      setError(friendlyAuthError((err as { code?: string }).code))
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <div className="mx-auto mt-20 max-w-sm rounded-xl border border-gray-800 bg-gray-900 p-8">
      <h1 className="mb-6 text-xl font-semibold text-gray-100">
        Reset password
      </h1>

      {phase === 'verifying' && (
        <p className="text-sm text-gray-400">Verifying reset link…</p>
      )}

      {phase === 'error' && (
        <div className="space-y-4">
          <p className="text-sm text-red-400">{error}</p>
          <Link
            to="/login"
            className="inline-block text-sm text-indigo-400 hover:text-indigo-300"
          >
            Back to sign in
          </Link>
        </div>
      )}

      {phase === 'form' && (
        <form onSubmit={handleSubmit} className="flex flex-col gap-4">
          <p className="text-sm text-gray-400">
            Set a new password for{' '}
            <span className="text-gray-200">{email}</span>.
          </p>
          {error && <p className="text-sm text-red-400">{error}</p>}
          <input
            type="password"
            placeholder="New password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            required
            autoComplete="new-password"
            className="rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100 placeholder-gray-500 focus:border-indigo-500 focus:outline-none"
          />
          <input
            type="password"
            placeholder="Confirm new password"
            value={confirm}
            onChange={(e) => setConfirm(e.target.value)}
            required
            autoComplete="new-password"
            className="rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100 placeholder-gray-500 focus:border-indigo-500 focus:outline-none"
          />
          <button
            type="submit"
            disabled={submitting}
            className="cursor-pointer rounded-lg bg-indigo-600 py-2 text-sm font-medium text-white hover:bg-indigo-500 disabled:opacity-50"
          >
            {submitting ? 'Updating…' : 'Update password'}
          </button>
        </form>
      )}

      {phase === 'success' && (
        <div className="space-y-4">
          <p className="text-sm text-emerald-400">
            Password updated. You can now sign in with your new password.
          </p>
          <Link
            to="/login"
            className="inline-block rounded-lg bg-indigo-600 px-4 py-2 text-sm font-medium text-white hover:bg-indigo-500"
          >
            Sign in
          </Link>
        </div>
      )}
    </div>
  )
}
