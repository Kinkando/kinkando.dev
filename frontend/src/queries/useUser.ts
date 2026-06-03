import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { createLineLinkCode, getMe, unlinkLine } from '../lib/api/users'
import { keys } from './keys'

export function useMe(options?: { refetchInterval?: number | false }) {
  return useQuery({
    queryKey: keys.userMe,
    queryFn: getMe,
    refetchInterval: options?.refetchInterval,
  })
}

export function useCreateLineLinkCode() {
  return useMutation({
    mutationFn: createLineLinkCode,
  })
}

export function useUnlinkLine() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: unlinkLine,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: keys.userMe })
    },
  })
}
