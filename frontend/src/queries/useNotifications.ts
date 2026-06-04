import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import {
  checkPushToken,
  getNotificationSettings,
  registerPushToken,
  removePushToken,
  sendTestNotification,
  updateNotificationSettings,
} from '../lib/api/notifications'
import { getCurrentToken } from '../lib/messaging'
import type { UpsertNotificationSettingsInput } from '../lib/api/types'
import { keys } from './keys'

export function useNotificationSettings() {
  return useQuery({
    queryKey: keys.notificationSettings,
    queryFn: getNotificationSettings,
  })
}

export function useUpdateNotificationSettings() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (input: UpsertNotificationSettingsInput) =>
      updateNotificationSettings(input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: keys.notificationSettings })
    },
  })
}

/**
 * Queries whether the current device (browser) FCM token is registered
 * in the backend for the authenticated user.
 * Returns { token: string | null, registered: boolean }.
 */
export function useDeviceRegistration() {
  return useQuery({
    queryKey: keys.notificationDevice,
    queryFn: async () => {
      const token = await getCurrentToken()
      if (!token) return { token: null, registered: false }
      const result = await checkPushToken(token)
      return { token, registered: result?.registered ?? false }
    },
  })
}

export function useRegisterPushToken() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (token: string) => registerPushToken(token),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: keys.notificationDevice })
    },
  })
}

export function useRemovePushToken() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (token: string) => removePushToken(token),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: keys.notificationDevice })
    },
  })
}

export function useSendTestNotification() {
  return useMutation({
    mutationFn: sendTestNotification,
  })
}
