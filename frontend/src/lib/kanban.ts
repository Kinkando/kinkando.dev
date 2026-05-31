import type { Card, Priority } from './api/types'

export const PRIORITY_META: Record<Priority, { label: string; color: string }> =
  {
    none: { label: 'None', color: '#6b7280' },
    low: { label: 'Low', color: '#3b82f6' },
    medium: { label: 'Medium', color: '#eab308' },
    high: { label: 'High', color: '#f97316' },
    urgent: { label: 'Urgent', color: '#ef4444' },
  }

export function isOverdue(card: Card): boolean {
  if (!card.due_date) return false
  return new Date(card.due_date) < new Date(new Date().toDateString())
}

export function isDueSoon(card: Card): boolean {
  if (!card.due_date) return false
  const due = new Date(card.due_date)
  const now = new Date(new Date().toDateString())
  const threeDays = new Date(now)
  threeDays.setDate(threeDays.getDate() + 3)
  return due >= now && due <= threeDays
}
