import type { RecordType } from './api/types'

export const RECORD_TYPES: RecordType[] = ['income', 'expense']

export const RECORD_TYPE_META: Record<
  RecordType,
  {
    label: string
    sign: string
    activeClass: string
    textClass: string
    cardClass: string
    bgColor: string
    iconColor: string
  }
> = {
  income: {
    label: 'Income',
    sign: '+',
    activeClass: 'bg-green-700 text-white',
    textClass: 'text-green-400',
    cardClass: 'border-green-900/40 bg-green-950/30',
    bgColor: '#16a34a26',
    iconColor: '#4ade80',
  },
  expense: {
    label: 'Expense',
    sign: '-',
    activeClass: 'bg-red-700 text-white',
    textClass: 'text-red-400',
    cardClass: 'border-red-900/40 bg-red-950/30',
    bgColor: '#dc262626',
    iconColor: '#f87171',
  },
}
