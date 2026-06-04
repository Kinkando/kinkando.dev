import { apiFetch } from './client'
import type {
  DeliveryResult,
  NotificationSettings,
  UpsertNotificationSettingsInput,
} from './types'

export function getNotificationSettings(): Promise<
  NotificationSettings | undefined
> {
  return apiFetch<NotificationSettings>('/notifications/settings', {
    auth: true,
  })
}

export function updateNotificationSettings(
  input: UpsertNotificationSettingsInput,
): Promise<NotificationSettings | undefined> {
  return apiFetch<NotificationSettings>('/notifications/settings', {
    method: 'PUT',
    body: input,
    auth: true,
  })
}

export function registerPushToken(token: string): Promise<undefined> {
  return apiFetch('/notifications/tokens', {
    method: 'POST',
    body: { token },
    auth: true,
  }) as Promise<undefined>
}

export function checkPushToken(
  token: string,
): Promise<{ registered: boolean } | undefined> {
  return apiFetch<{ registered: boolean }>('/notifications/tokens/check', {
    method: 'POST',
    body: { token },
    auth: true,
  })
}

export function removePushToken(token: string): Promise<undefined> {
  return apiFetch('/notifications/tokens', {
    method: 'DELETE',
    body: { token },
    auth: true,
  }) as Promise<undefined>
}

export function sendTestNotification(): Promise<DeliveryResult | undefined> {
  return apiFetch<DeliveryResult>('/notifications/test', {
    method: 'POST',
    auth: true,
  })
}
