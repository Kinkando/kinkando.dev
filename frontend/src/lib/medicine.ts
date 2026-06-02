import type { Medicine } from './api/types'

/**
 * Estimate the number of doses per day based on frequency settings.
 * Returns null when we can't determine a daily dose (as_needed, custom).
 */
export function dailyDoseEstimate(m: Medicine): number | null {
  const amount = m.dosage_amount
  if (amount <= 0) return null

  switch (m.frequency_type) {
    case 'daily':
      return amount * (m.frequency_value ?? 1)
    case 'weekly':
      return (amount * (m.frequency_value ?? 1)) / 7
    default:
      return null
  }
}

/**
 * Estimate how many days of stock remain.
 * Returns null when the daily dose cannot be calculated.
 */
export function estimatedDaysRemaining(m: Medicine): number | null {
  const dose = dailyDoseEstimate(m)
  if (dose === null || dose <= 0) return null
  return Math.floor(m.stock_quantity / dose)
}

/**
 * Returns true when the medicine is at or below its low-stock threshold.
 */
export function isLowStock(m: Medicine): boolean {
  return m.stock_quantity <= m.low_stock_threshold
}

/**
 * Returns the number of doses required per day, or null when daily gating
 * does not apply (weekly, as_needed, custom).
 */
export function requiredDailyDoses(m: Medicine): number | null {
  if (m.frequency_type !== 'daily') return null
  return m.frequency_value ?? 1
}

/**
 * Returns true when the ISO timestamp falls on the local calendar "today".
 * Uses local-timezone Date getters so midnight boundaries respect the user's TZ.
 */
export function isLocalToday(iso: string): boolean {
  const d = new Date(iso)
  const now = new Date()
  return (
    d.getFullYear() === now.getFullYear() &&
    d.getMonth() === now.getMonth() &&
    d.getDate() === now.getDate()
  )
}
