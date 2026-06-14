import { useEffect, useState } from 'react'
import { Check, Copy } from 'lucide-react'
import type { UserInfo } from 'firebase/auth'
import {
  GoogleAuthProvider,
  linkWithPopup,
  sendPasswordResetEmail,
  unlink,
} from 'firebase/auth'
import { useDocumentTitle } from '../hooks/useDocumentTitle'
import { useCreateLineLinkCode, useMe, useUnlinkLine } from '../queries/useUser'
import { useAuth } from '../auth/AuthContext'
import { auth, friendlyAuthError } from '../lib/firebase'
import GoogleIcon from '../components/icons/GoogleIcon'

const PROVIDER_GOOGLE = 'google.com'
const PROVIDER_PASSWORD = 'password'

const lineBotLink = <a href="http://lin.ee/r5qOWAg" target="_blank" className="font-bold underline hover:text-gray-100">LINE bot</a>

export default function AccountPage() {
  useDocumentTitle('Account')

  const { user } = useAuth()

  // Poll every 5 s while a link code is displayed — stops once line_id is set.
  const [polling, setPolling] = useState(false)
  const { data: me, isLoading } = useMe({
    refetchInterval: polling ? 5000 : false,
  })

  const createCode = useCreateLineLinkCode()
  const unlinkLine = useUnlinkLine()

  // Track Firebase auth providerData locally so link/unlink can update the UI
  // without waiting for an onAuthStateChanged event.
  const [providerData, setProviderData] = useState<UserInfo[]>(
    () => user?.providerData ?? [],
  )
  useEffect(() => {
    setProviderData(user?.providerData ?? [])
  }, [user])
  const hasGoogle = providerData.some((p) => p.providerId === PROVIDER_GOOGLE)
  const hasPassword = providerData.some((p) => p.providerId === PROVIDER_PASSWORD)
  // Block unlink when Google is the sole Firebase auth provider — otherwise
  // the user would lose every way to sign back in.
  const canUnlinkGoogle = hasGoogle && providerData.length > 1
  const googleEmail = providerData.find(
    (p) => p.providerId === PROVIDER_GOOGLE,
  )?.email

  // Google link/unlink state
  const [googleLoading, setGoogleLoading] = useState(false)
  const [googleError, setGoogleError] = useState('')
  const [googleSuccess, setGoogleSuccess] = useState('')
  const [confirmUnlinkGoogle, setConfirmUnlinkGoogle] = useState(false)

  async function handleLinkGoogle() {
    if (!auth.currentUser) return
    setGoogleError('')
    setGoogleSuccess('')
    setGoogleLoading(true)
    try {
      const result = await linkWithPopup(
        auth.currentUser,
        new GoogleAuthProvider(),
      )
      setProviderData([...result.user.providerData])
      setGoogleSuccess('Google account linked.')
      setTimeout(() => setGoogleSuccess(''), 2500)
    } catch (err: unknown) {
      setGoogleError(friendlyAuthError((err as { code?: string }).code))
    } finally {
      setGoogleLoading(false)
    }
  }

  async function handleUnlinkGoogle() {
    if (!auth.currentUser) return
    setGoogleError('')
    setGoogleSuccess('')
    setGoogleLoading(true)
    try {
      const u = await unlink(auth.currentUser, PROVIDER_GOOGLE)
      setProviderData([...u.providerData])
      setConfirmUnlinkGoogle(false)
      setGoogleSuccess('Google account unlinked.')
      setTimeout(() => setGoogleSuccess(''), 2500)
    } catch (err: unknown) {
      setGoogleError(friendlyAuthError((err as { code?: string }).code))
    } finally {
      setGoogleLoading(false)
    }
  }

  // Password reset state
  const [resetLoading, setResetLoading] = useState(false)
  const [resetError, setResetError] = useState('')
  const [resetSuccess, setResetSuccess] = useState(false)

  async function handleResetPassword() {
    const email = auth.currentUser?.email
    if (!email) return
    setResetError('')
    setResetSuccess(false)
    setResetLoading(true)
    try {
      await sendPasswordResetEmail(auth, email)
      setResetSuccess(true)
      setTimeout(() => setResetSuccess(false), 4000)
    } catch (err: unknown) {
      setResetError(friendlyAuthError((err as { code?: string }).code))
    } finally {
      setResetLoading(false)
    }
  }

  // Pending code state
  const [pendingCode, setPendingCode] = useState<string | null>(null)
  const [pendingExpiry, setPendingExpiry] = useState<string | null>(null)

  // Unlink confirm state
  const [confirmUnlink, setConfirmUnlink] = useState(false)
  const [unlinkError, setUnlinkError] = useState('')
  const [unlinkSuccess, setUnlinkSuccess] = useState(false)

  // Stop polling once the account becomes linked.
  useEffect(() => {
    if (me?.line_id) {
      setPolling(false)
      setPendingCode(null)
      setPendingExpiry(null)
    }
  }, [me?.line_id])

  async function handleLinkClick() {
    createCode.reset()
    try {
      const result = await createCode.mutateAsync()
      if (result) {
        setPendingCode(result.code)
        setPendingExpiry(result.expires_at)
        setPolling(true)
      }
    } catch {
      // error shown via createCode.error
    }
  }

  async function handleUnlinkConfirm() {
    setUnlinkError('')
    setUnlinkSuccess(false)
    try {
      await unlinkLine.mutateAsync()
      setConfirmUnlink(false)
      setUnlinkSuccess(true)
      setTimeout(() => setUnlinkSuccess(false), 2500)
    } catch (err) {
      setUnlinkError(
        err instanceof Error ? err.message : 'Something went wrong.',
      )
    }
  }

  const isLinked = Boolean(me?.line_id)

  return (
    <div className="mx-auto max-w-2xl px-4 py-8">
      <h1 className="mb-6 text-xl font-semibold text-gray-100">Account</h1>

      {/* Google account card */}
      <div className="mb-4 rounded-xl border border-gray-800 bg-gray-900 p-5">
        <h2 className="mb-4 text-sm font-semibold tracking-wide text-gray-400 uppercase">
          Google Account
        </h2>

        {hasGoogle ? (
          <div className="space-y-3">
            <div className="flex items-center gap-2">
              <span className="inline-flex items-center gap-1.5 rounded-full bg-emerald-900/40 px-2.5 py-0.5 text-xs font-medium text-emerald-400">
                <Check className="h-3 w-3" strokeWidth={3} />
                Linked
              </span>
              <span className="text-xs text-gray-500">{googleEmail ?? ''}</span>
            </div>

            {googleSuccess && (
              <p className="text-sm text-emerald-400">{googleSuccess}</p>
            )}

            {!confirmUnlinkGoogle ? (
              <>
                <button
                  onClick={() => {
                    setConfirmUnlinkGoogle(true)
                    setGoogleError('')
                  }}
                  disabled={!canUnlinkGoogle}
                  className="cursor-pointer rounded-lg bg-gray-800 px-4 py-2 text-sm font-medium text-red-400 hover:bg-gray-700 hover:text-red-300 disabled:cursor-not-allowed disabled:opacity-50"
                >
                  Unlink Google
                </button>
                {!canUnlinkGoogle && (
                  <p className="text-xs text-gray-500">
                    Google is your only sign-in method. Set a password before
                    unlinking so you don't lose access.
                  </p>
                )}
              </>
            ) : (
              <div className="space-y-3 rounded-lg border border-yellow-600/40 bg-yellow-900/20 p-4">
                <p className="text-sm text-yellow-300">
                  Are you sure you want to unlink your Google account?
                </p>
                {googleError && (
                  <p className="text-sm text-red-400">{googleError}</p>
                )}
                <div className="flex justify-end gap-2">
                  <button
                    onClick={() => setConfirmUnlinkGoogle(false)}
                    className="cursor-pointer rounded-lg bg-gray-800 px-4 py-2 text-sm font-medium text-gray-300 hover:bg-gray-700"
                  >
                    Cancel
                  </button>
                  <button
                    onClick={handleUnlinkGoogle}
                    disabled={googleLoading}
                    className="cursor-pointer rounded-lg bg-red-600 px-4 py-2 text-sm font-medium text-white hover:bg-red-500 disabled:opacity-50"
                  >
                    {googleLoading ? 'Unlinking…' : 'Yes, unlink'}
                  </button>
                </div>
              </div>
            )}
          </div>
        ) : (
          <div className="space-y-3">
            <p className="text-sm text-gray-400">
              Link Google so you can sign in with one tap.
            </p>
            <button
              onClick={handleLinkGoogle}
              disabled={googleLoading}
              className="flex cursor-pointer items-center gap-2 rounded-lg border border-gray-700 bg-gray-800 px-4 py-2 text-sm font-medium text-gray-200 hover:bg-gray-700 disabled:opacity-50"
            >
              <GoogleIcon className="h-4 w-4 shrink-0" />
              {googleLoading ? 'Linking…' : 'Link Google'}
            </button>
            {googleError && (
              <p className="text-sm text-red-400">{googleError}</p>
            )}
            {googleSuccess && (
              <p className="text-sm text-emerald-400">{googleSuccess}</p>
            )}
          </div>
        )}
      </div>

      {/* Password card */}
      <div className="mb-4 rounded-xl border border-gray-800 bg-gray-900 p-5">
        <h2 className="mb-4 text-sm font-semibold tracking-wide text-gray-400 uppercase">
          Password
        </h2>

        {hasPassword ? (
          <div className="space-y-3">
            <p className="text-sm text-gray-400">
              Send a password reset link to{' '}
              <span className="text-gray-300">{user?.email}</span>.
            </p>
            <button
              onClick={handleResetPassword}
              disabled={resetLoading || resetSuccess}
              className="cursor-pointer rounded-lg bg-indigo-600 px-4 py-2 text-sm font-medium text-white hover:bg-indigo-500 disabled:opacity-50"
            >
              {resetLoading ? 'Sending…' : 'Reset password'}
            </button>
            {resetSuccess && (
              <p className="text-sm text-emerald-400">
                Reset email sent. Check your inbox.
              </p>
            )}
            {resetError && (
              <p className="text-sm text-red-400">{resetError}</p>
            )}
          </div>
        ) : (
          <p className="text-sm text-gray-400">
            This account has no password — sign in with Google.
          </p>
        )}
      </div>

      {/* LINE account card */}
      <div className="rounded-xl border border-gray-800 bg-gray-900 p-5">
        <h2 className="mb-4 text-sm font-semibold tracking-wide text-gray-400 uppercase">
          LINE Account
        </h2>

        {isLoading && <p className="text-sm text-gray-500">Loading…</p>}

        {!isLoading && isLinked && (
          <div className="space-y-3">
            <div className="flex items-center gap-2">
              <span className="inline-flex items-center gap-1.5 rounded-full bg-emerald-900/40 px-2.5 py-0.5 text-xs font-medium text-emerald-400">
                <Check className="h-3 w-3" strokeWidth={3} />
                Linked
              </span>
              <span className="font-mono text-xs text-gray-500">
                {maskLineId(me!.line_id!)}
              </span>
            </div>

            {unlinkSuccess && (
              <p className="text-sm text-emerald-400">LINE account unlinked.</p>
            )}

            {!confirmUnlink ? (
              <button
                onClick={() => {
                  setConfirmUnlink(true)
                  setUnlinkError('')
                }}
                className="cursor-pointer rounded-lg bg-gray-800 px-4 py-2 text-sm font-medium text-red-400 hover:bg-gray-700 hover:text-red-300"
              >
                Unlink LINE
              </button>
            ) : (
              <div className="space-y-3 rounded-lg border border-yellow-600/40 bg-yellow-900/20 p-4">
                <p className="text-sm text-yellow-300">
                  Are you sure you want to unlink your LINE account? You can
                  always re-link it later.
                </p>
                {unlinkError && (
                  <p className="text-sm text-red-400">{unlinkError}</p>
                )}
                <div className="flex gap-2">
                  <button
                    onClick={handleUnlinkConfirm}
                    disabled={unlinkLine.isPending}
                    className="cursor-pointer rounded-lg bg-red-600 px-4 py-2 text-sm font-medium text-white hover:bg-red-500 disabled:opacity-50"
                  >
                    {unlinkLine.isPending ? 'Unlinking…' : 'Yes, unlink'}
                  </button>
                  <button
                    onClick={() => setConfirmUnlink(false)}
                    className="cursor-pointer rounded-lg bg-gray-800 px-4 py-2 text-sm font-medium text-gray-300 hover:bg-gray-700"
                  >
                    Cancel
                  </button>
                </div>
              </div>
            )}
          </div>
        )}

        {!isLoading && !isLinked && (
          <div className="space-y-3">
            <p className="text-sm text-gray-400">
              Link your LINE account to use the {lineBotLink} as your personal
              assistant.
            </p>

            {!pendingCode ? (
              <>
                <button
                  onClick={handleLinkClick}
                  disabled={createCode.isPending}
                  className="cursor-pointer rounded-lg bg-indigo-600 px-4 py-2 text-sm font-medium text-white hover:bg-indigo-500 disabled:opacity-50"
                >
                  {createCode.isPending ? 'Generating…' : 'Link LINE'}
                </button>
                {createCode.isError && (
                  <p className="text-sm text-red-400">
                    {createCode.error instanceof Error
                      ? createCode.error.message
                      : 'Could not generate a code. Please try again.'}
                  </p>
                )}
              </>
            ) : (
              <div className="space-y-3 flex flex-col items-center rounded-lg border border-indigo-600/30 bg-indigo-900/20 p-4">
                <p className="text-sm text-gray-300">
                  Open the {lineBotLink} and send the following code to link your account:
                </p>
                <img src="images/line_qr.png" />
                <div className="flex items-center gap-2">
                  <code className="rounded bg-gray-800 px-3 py-1.5 font-mono text-base tracking-widest text-indigo-300">
                    LINK {pendingCode}
                  </code>
                  <button
                    onClick={() =>
                      navigator.clipboard.writeText(`LINK ${pendingCode}`)
                    }
                    title="Copy"
                    className="cursor-pointer rounded p-1 text-gray-400 hover:text-gray-200"
                  >
                    <Copy className="h-4 w-4" strokeWidth={1.5} />
                  </button>
                </div>
                {pendingExpiry && (
                  <p className="text-xs text-gray-500">
                    Expires at{' '}
                    {new Date(pendingExpiry).toLocaleTimeString([], {
                      hour: '2-digit',
                      minute: '2-digit',
                    })}{' '}
                    — waiting for confirmation
                    <span className="ml-1 inline-block animate-pulse">…</span>
                  </p>
                )}
                <button
                  onClick={() => {
                    setPendingCode(null)
                    setPendingExpiry(null)
                    setPolling(false)
                    createCode.reset()
                  }}
                  className="cursor-pointer text-xs text-gray-500 underline hover:text-gray-300"
                >
                  Generate a new code
                </button>
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  )
}

/** Show the first 4 and last 4 chars of a LINE user ID, masking the middle. */
function maskLineId(id: string): string {
  if (id.length <= 8) return id
  return `${id.slice(0, 4)}${'•'.repeat(id.length - 8)}${id.slice(-4)}`
}
