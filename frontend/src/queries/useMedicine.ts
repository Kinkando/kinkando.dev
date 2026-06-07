import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import {
  fetchMedicines,
  createMedicine,
  updateMedicine,
  archiveMedicine,
  unarchiveMedicine,
  takeMedicine,
  adjustStock,
  fetchIntakes,
  fetchStockAdjustments,
} from '../lib/api/medicine'
import type {
  CreateMedicineInput,
  UpdateMedicineInput,
  TakeMedicineInput,
  AdjustStockInput,
  MedicineSourceType,
} from '../lib/api/types'
import { keys } from './keys'

export function useMedicines(
  sourceType: MedicineSourceType,
  includeArchived = false,
) {
  return useQuery({
    queryKey: keys.medicines(sourceType, includeArchived),
    queryFn: () => fetchMedicines(sourceType, includeArchived),
  })
}

export function useCreateMedicine() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (input: CreateMedicineInput) => createMedicine(input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['medicine', 'list'] })
    },
  })
}

export function useUpdateMedicine() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ id, input }: { id: string; input: UpdateMedicineInput }) =>
      updateMedicine(id, input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['medicine', 'list'] })
    },
  })
}

export function useArchiveMedicine() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => archiveMedicine(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['medicine', 'list'] })
    },
  })
}

export function useUnarchiveMedicine() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => unarchiveMedicine(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['medicine', 'list'] })
    },
  })
}

export function useTakeMedicine() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ id, input }: { id: string; input: TakeMedicineInput }) =>
      takeMedicine(id, input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['medicine', 'list'] })
      queryClient.invalidateQueries({ queryKey: ['medicine', 'intakes'] })
      // A supplement or medicine take may auto-complete a quest; refresh quest overview so
      // the UI reflects the updated state without a manual page reload.
      queryClient.invalidateQueries({ queryKey: keys.questOverview })
    },
  })
}

export function useAdjustStock() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ id, input }: { id: string; input: AdjustStockInput }) =>
      adjustStock(id, input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['medicine', 'list'] })
      queryClient.invalidateQueries({ queryKey: ['medicine', 'adjustments'] })
    },
  })
}

export function useMedicineIntakes(
  sourceType: MedicineSourceType,
  date?: string,
) {
  return useQuery({
    queryKey: keys.medicineIntakes(sourceType, date),
    queryFn: () => fetchIntakes(sourceType, { date }),
  })
}

export function useStockAdjustments(
  sourceType: MedicineSourceType,
  date?: string,
) {
  return useQuery({
    queryKey: keys.medicineAdjustments(sourceType, date),
    queryFn: () => fetchStockAdjustments(sourceType, { date }),
  })
}
