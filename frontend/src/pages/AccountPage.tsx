import { useEffect, useState } from 'react'
import { useDocumentTitle } from '../hooks/useDocumentTitle'
import { useCreateLineLinkCode, useMe, useUnlinkLine } from '../queries/useUser'

export default function AccountPage() {
  useDocumentTitle('Account')

  // Poll every 5 s while a link code is displayed — stops once line_id is set.
  const [polling, setPolling] = useState(false)
  const { data: me, isLoading } = useMe({
    refetchInterval: polling ? 5000 : false,
  })

  const createCode = useCreateLineLinkCode()
  const unlink = useUnlinkLine()

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
      await unlink.mutateAsync()
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
                <svg
                  className="h-3 w-3"
                  fill="currentColor"
                  viewBox="0 0 20 20"
                >
                  <path
                    fillRule="evenodd"
                    d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z"
                    clipRule="evenodd"
                  />
                </svg>
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
                    disabled={unlink.isPending}
                    className="cursor-pointer rounded-lg bg-red-600 px-4 py-2 text-sm font-medium text-white hover:bg-red-500 disabled:opacity-50"
                  >
                    {unlink.isPending ? 'Unlinking…' : 'Yes, unlink'}
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
              Link your LINE account to use the LINE bot as your personal
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
              <div className="space-y-3 rounded-lg border border-indigo-600/30 bg-indigo-900/20 p-4">
                <p className="text-sm text-gray-300">
                  Open the LINE bot and send the following message:
                </p>
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
                    <svg
                      xmlns="http://www.w3.org/2000/svg"
                      className="h-4 w-4"
                      fill="none"
                      viewBox="0 0 24 24"
                      stroke="currentColor"
                      strokeWidth={1.5}
                    >
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z"
                      />
                    </svg>
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
