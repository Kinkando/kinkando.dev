import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import {
  getNotificationSettings,
  registerPushToken,
  removePushToken,
  sendTestNotification,
  updateNotificationSettings,
} from '../lib/api/notifications'
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

export function useRegisterPushToken() {
  return useMutation({
    mutationFn: (token: string) => registerPushToken(token),
  })
}

export function useRemovePushToken() {
  return useMutation({
    mutationFn: (token: string) => removePushToken(token),
  })
}

export function useSendTestNotification() {
  return useMutation({
    mutationFn: sendTestNotification,
  })
}
