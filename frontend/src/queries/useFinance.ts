import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import {
  fetchRecords,
  createRecord,
  deleteRecord,
  fetchSummary,
  fetchCategories,
  createCategory,
  updateCategory,
  deleteCategory,
  fetchNotes,
} from '../lib/api/finance'
import type {
  CreateRecordInput,
  CreateCategoryInput,
  UpdateCategoryInput,
} from '../lib/api/types'
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

export function useCategories() {
  return useQuery({
    queryKey: keys.financeCategories,
    queryFn: fetchCategories,
  })
}

export function useCreateRecord(month: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (input: CreateRecordInput) => createRecord(input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: keys.financeRecords(month) })
      queryClient.invalidateQueries({ queryKey: keys.financeSummary(month) })
      queryClient.invalidateQueries({ queryKey: keys.financeNotes })
    },
  })
}

export function useFinanceNotes() {
  return useQuery({
    queryKey: keys.financeNotes,
    queryFn: fetchNotes,
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

export function useCreateCategory() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (input: CreateCategoryInput) => createCategory(input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: keys.financeCategories })
    },
  })
}

export function useUpdateCategory() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ id, input }: { id: string; input: UpdateCategoryInput }) =>
      updateCategory(id, input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: keys.financeCategories })
    },
  })
}

export function useDeleteCategory() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => deleteCategory(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: keys.financeCategories })
    },
  })
}
