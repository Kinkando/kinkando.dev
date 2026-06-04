import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { useDocumentTitle } from '../hooks/useDocumentTitle'
import { useMe } from '../queries/useUser'
import {
  useDeviceRegistration,
  useNotificationSettings,
  useRegisterPushToken,
  useRemovePushToken,
  useSendTestNotification,
  useUpdateNotificationSettings,
} from '../queries/useNotifications'
import { isPushSupported, requestPushToken } from '../lib/messaging'
import type { UpsertNotificationSettingsInput } from '../lib/api/types'

const inputClass =
  'w-full rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100 placeholder-gray-500 focus:border-indigo-500 focus:outline-none'
const labelClass = 'mb-1 block text-xs font-medium text-gray-400'

export default function NotificationsPage() {
  useDocumentTitle('Notifications')

  const { data: settings, isLoading } = useNotificationSettings()
  const { data: me } = useMe()
  const { data: deviceReg, isLoading: deviceLoading } = useDeviceRegistration()

  const updateSettings = useUpdateNotificationSettings()
  const registerToken = useRegisterPushToken()
  const removeToken = useRemovePushToken()
  const sendTest = useSendTestNotification()

  // ── Local form state ────────────────────────────────────────────────────────
  const [lineEnabled, setLineEnabled] = useState(false)
  const [discordEnabled, setDiscordEnabled] = useState(false)
  const [discordWebhookUrl, setDiscordWebhookUrl] = useState('')
  const [webPushEnabled, setWebPushEnabled] = useState(false)

  const [saveError, setSaveError] = useState('')
  const [saveSuccess, setSaveSuccess] = useState(false)
  const [testMessage, setTestMessage] = useState('')
  const [testIsError, setTestIsError] = useState(false)
  const [pushError, setPushError] = useState('')

  // Pre-fill form when server data arrives
  useEffect(() => {
    if (!settings) return
    setLineEnabled(settings.line_enabled)
    setDiscordEnabled(settings.discord_enabled)
    setDiscordWebhookUrl(settings.discord_webhook_url ?? '')
    setWebPushEnabled(settings.web_push_enabled)
  }, [settings])

  // ── Handlers ─────────────────────────────────────────────────────────────────

  async function handleSave() {
    setSaveError('')
    setSaveSuccess(false)
    const input: UpsertNotificationSettingsInput = {
      line_enabled: lineEnabled,
      discord_enabled: discordEnabled,
      discord_webhook_url: discordWebhookUrl.trim() || null,
      web_push_enabled: webPushEnabled,
    }
    try {
      await updateSettings.mutateAsync(input)
      setSaveSuccess(true)
      setTimeout(() => setSaveSuccess(false), 2500)
    } catch (err) {
      setSaveError(
        err instanceof Error ? err.message : 'Could not save settings.',
      )
    }
  }

  /**
   * Prompts for browser permission (if needed), obtains an FCM token, and
   * registers it with the backend. Designed to be called automatically when
   * the web-push toggle is enabled, and also as a manual retry.
   */
  async function handleEnablePush() {
    setPushError('')
    const token = await requestPushToken()
    if (!token) {
      setPushError(
        'Could not get a push token. Make sure notifications are allowed in your browser.',
      )
      return
    }
    try {
      await registerToken.mutateAsync(token)
    } catch (err) {
      setPushError(
        err instanceof Error
          ? err.message
          : 'Could not register this device for push.',
      )
    }
  }

  async function handleDisablePush() {
    if (!deviceReg?.token) return
    setPushError('')
    try {
      await removeToken.mutateAsync(deviceReg.token)
    } catch (err) {
      setPushError(
        err instanceof Error
          ? err.message
          : 'Could not unregister this device.',
      )
    }
  }

  /** Toggle web push. When turned on, immediately kick off device registration. */
  function handleWebPushToggle(checked: boolean) {
    setWebPushEnabled(checked)
    if (checked) {
      handleEnablePush()
    }
  }

  async function handleSendTest() {
    setTestMessage('')
    setTestIsError(false)
    try {
      const result = await sendTest.mutateAsync()
      if (result && result.delivered > 0) {
        setTestMessage(
          `Sent to ${result.delivered} channel${result.delivered !== 1 ? 's' : ''}.`,
        )
        setTestIsError(false)
      } else {
        const firstError = result?.errors?.[0]
        setTestMessage(
          firstError ?? 'No channel delivered the test. Check your settings.',
        )
        setTestIsError(true)
      }
      setTimeout(() => setTestMessage(''), 4000)
    } catch (err) {
      setTestMessage(
        err instanceof Error
          ? err.message
          : 'Could not send test notification.',
      )
      setTestIsError(true)
    }
  }

  const isLinked = Boolean(me?.line_id)
  const pushSupported = isPushSupported()

  return (
    <div className="mx-auto max-w-2xl px-4 py-8">
      <h1 className="mb-6 text-xl font-semibold text-gray-100">
        Notifications
      </h1>

      {isLoading && <p className="text-sm text-gray-500">Loading…</p>}

      {!isLoading && (
        <div className="space-y-4">
          {/* ── LINE card ────────────────────────────────────────────────────── */}
          <div className="rounded-xl border border-gray-800 bg-gray-900 p-5">
            <h2 className="mb-4 text-sm font-semibold tracking-wide text-gray-400 uppercase">
              LINE
            </h2>

            {!isLinked ? (
              <p className="text-sm text-gray-400">
                Your LINE account is not linked yet.{' '}
                <Link
                  to="/account"
                  className="text-indigo-400 underline hover:text-indigo-300"
                >
                  Link it on the Account page
                </Link>{' '}
                to receive LINE notifications.
              </p>
            ) : (
              <label className="flex items-center gap-3">
                <input
                  type="checkbox"
                  checked={lineEnabled}
                  onChange={(e) => setLineEnabled(e.target.checked)}
                  className="h-4 w-4 cursor-pointer rounded border-gray-600 bg-gray-800 accent-indigo-500"
                />
                <span className="text-sm text-gray-300">
                  Send notifications via LINE
                </span>
              </label>
            )}
          </div>

          {/* ── Discord card ─────────────────────────────────────────────────── */}
          <div className="rounded-xl border border-gray-800 bg-gray-900 p-5">
            <h2 className="mb-4 text-sm font-semibold tracking-wide text-gray-400 uppercase">
              Discord
            </h2>

            <div className="space-y-3">
              <label className="flex items-center gap-3">
                <input
                  type="checkbox"
                  checked={discordEnabled}
                  onChange={(e) => setDiscordEnabled(e.target.checked)}
                  className="h-4 w-4 cursor-pointer rounded border-gray-600 bg-gray-800 accent-indigo-500"
                />
                <span className="text-sm text-gray-300">
                  Send notifications via Discord webhook
                </span>
              </label>

              {discordEnabled && (
                <div>
                  <label className={labelClass} htmlFor="discord-webhook">
                    Webhook URL
                  </label>
                  <input
                    id="discord-webhook"
                    type="url"
                    value={discordWebhookUrl}
                    onChange={(e) => setDiscordWebhookUrl(e.target.value)}
                    placeholder="https://discord.com/api/webhooks/..."
                    className={inputClass}
                  />
                  <p className="mt-1 text-xs text-gray-500">
                    Create one via Discord → Server Settings → Integrations →
                    Webhooks.
                  </p>
                </div>
              )}
            </div>
          </div>

          {/* ── Web Push card ────────────────────────────────────────────────── */}
          <div className="rounded-xl border border-gray-800 bg-gray-900 p-5">
            <h2 className="mb-4 text-sm font-semibold tracking-wide text-gray-400 uppercase">
              Web Push
            </h2>

            {!pushSupported ? (
              <p className="text-sm text-gray-400">
                Web Push notifications are not supported in this browser.
              </p>
            ) : (
              <div className="space-y-3">
                <label className="flex items-center gap-3">
                  <input
                    type="checkbox"
                    checked={webPushEnabled}
                    onChange={(e) => handleWebPushToggle(e.target.checked)}
                    className="h-4 w-4 cursor-pointer rounded border-gray-600 bg-gray-800 accent-indigo-500"
                  />
                  <span className="text-sm text-gray-300">
                    Send push notifications to this browser
                  </span>
                </label>

                {webPushEnabled && (
                  <div className="space-y-2">
                    {/* Device registration status */}
                    {deviceLoading || registerToken.isPending ? (
                      <p className="text-sm text-gray-500">
                        Checking device registration…
                      </p>
                    ) : deviceReg?.registered ? (
                      <div className="flex items-center gap-3">
                        <p className="text-sm text-emerald-400">
                          This device is registered for push.
                        </p>
                        <button
                          onClick={handleDisablePush}
                          disabled={removeToken.isPending}
                          className="cursor-pointer rounded-lg bg-gray-700 px-3 py-1 text-xs font-medium text-gray-300 hover:bg-gray-600 disabled:opacity-50"
                        >
                          {removeToken.isPending
                            ? 'Removing…'
                            : 'Disable on this device'}
                        </button>
                      </div>
                    ) : (
                      <div className="flex items-center gap-3">
                        <p className="text-sm text-yellow-400">
                          This device is not registered.
                        </p>
                        <button
                          onClick={handleEnablePush}
                          disabled={registerToken.isPending}
                          className="cursor-pointer rounded-lg bg-indigo-600 px-3 py-1 text-xs font-medium text-white hover:bg-indigo-500 disabled:opacity-50"
                        >
                          {registerToken.isPending
                            ? 'Registering…'
                            : 'Enable on this device'}
                        </button>
                      </div>
                    )}

                    {pushError && (
                      <p className="text-sm text-red-400">{pushError}</p>
                    )}
                  </div>
                )}
              </div>
            )}
          </div>

          {/* ── Save + Test ──────────────────────────────────────────────────── */}
          <div className="rounded-xl border border-gray-800 bg-gray-900 p-5">
            <div className="flex justify-end gap-2">
              <button
                onClick={handleSendTest}
                disabled={sendTest.isPending}
                className="cursor-pointer rounded-lg bg-gray-700 px-4 py-2 text-sm font-medium text-gray-300 hover:bg-gray-600 disabled:opacity-50"
              >
                {sendTest.isPending ? 'Sending…' : 'Send test notification'}
              </button>
              <button
                onClick={handleSave}
                disabled={updateSettings.isPending}
                className="cursor-pointer rounded-lg bg-indigo-600 px-4 py-2 text-sm font-medium text-white hover:bg-indigo-500 disabled:opacity-50"
              >
                {updateSettings.isPending ? 'Saving…' : 'Save'}
              </button>
            </div>

            {saveSuccess && (
              <p className="mt-3 text-right text-sm text-emerald-400">
                Settings saved.
              </p>
            )}
            {saveError && (
              <p className="mt-3 text-right text-sm text-red-400">
                {saveError}
              </p>
            )}
            {testMessage && (
              <p
                className={`mt-3 text-right text-sm ${testIsError ? 'text-red-400' : 'text-emerald-400'}`}
              >
                {testMessage}
              </p>
            )}
          </div>
        </div>
      )}
    </div>
  )
}
