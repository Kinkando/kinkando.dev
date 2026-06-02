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
