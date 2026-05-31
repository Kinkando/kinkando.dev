import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import {
  fetchRecords,
  createRecord,
  deleteRecord,
  fetchSummary,
} from '../lib/api/finance'
import type { CreateRecordInput } from '../lib/api/types'
import { keys } from './keys'

export function useRecords(month: string) {
  return useQuery({
    queryKey: keys.financeRecords(month),
    queryFn: () => fetchRecords(month),
  })
}

export function useSummary(month: string) {
  return useQuery({
    queryKey: keys.financeSummary(month),
    queryFn: () => fetchSummary(month),
  })
}

export function useCreateRecord(month: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (input: CreateRecordInput) => createRecord(input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: keys.financeRecords(month) })
      queryClient.invalidateQueries({ queryKey: keys.financeSummary(month) })
    },
  })
}

export function useDeleteRecord(month: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => deleteRecord(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: keys.financeRecords(month) })
      queryClient.invalidateQueries({ queryKey: keys.financeSummary(month) })
    },
  })
}
