import { describe, it, expect } from 'vitest'
import {
  dailyDoseEstimate,
  estimatedDaysRemaining,
  isLowStock,
  requiredDailyDoses,
} from './medicine'
import type { Medicine } from './api/types'

// Minimal Medicine stub — only fill fields each test cares about.
function med(overrides: Partial<Medicine>): Medicine {
  return {
    id: 'test-id',
    user_id: 'user-id',
    name: 'Test Med',
    source_type: 'medication',
    generic_name: null,
    description: null,
    stock_quantity: 10,
    stock_unit: 'tablet',
    dosage_amount: 1,
    dosage_unit: null,
    frequency_type: 'daily',
    frequency_value: 1,
    timing: null,
    start_date: null,
    end_date: null,
    low_stock_threshold: 3,
    note: null,
    created_at: '2026-01-01T00:00:00Z',
    updated_at: '2026-01-01T00:00:00Z',
    archived_at: null,
    reminder_enabled: false,
    reminder_times: [],
    ...overrides,
  } satisfies Medicine
}

describe('dailyDoseEstimate', () => {
  it('returns dosage_amount × frequency_value for daily', () => {
    expect(
      dailyDoseEstimate(
        med({ frequency_type: 'daily', dosage_amount: 2, frequency_value: 3 }),
      ),
    ).toBe(6)
  })

  it('defaults frequency_value to 1 when null (daily)', () => {
    expect(
      dailyDoseEstimate(
        med({
          frequency_type: 'daily',
          dosage_amount: 2,
          frequency_value: null,
        }),
      ),
    ).toBe(2)
  })

  it('returns fractional dose per day for weekly', () => {
    // 1 tablet × 7 times per week = 1 per day
    expect(
      dailyDoseEstimate(
        med({ frequency_type: 'weekly', dosage_amount: 1, frequency_value: 7 }),
      ),
    ).toBeCloseTo(1)
  })

  it('returns null for as_needed', () => {
    expect(dailyDoseEstimate(med({ frequency_type: 'as_needed' }))).toBeNull()
  })

  it('returns null for custom frequency', () => {
    expect(dailyDoseEstimate(med({ frequency_type: 'custom' }))).toBeNull()
  })

  it('returns null when dosage_amount is 0', () => {
    expect(
      dailyDoseEstimate(med({ frequency_type: 'daily', dosage_amount: 0 })),
    ).toBeNull()
  })
})

describe('estimatedDaysRemaining', () => {
  it('floors stock / daily dose', () => {
    // 10 tablets, 3 per day → floor(10/3) = 3
    const m = med({
      frequency_type: 'daily',
      dosage_amount: 3,
      frequency_value: 1,
      stock_quantity: 10,
    })
    expect(estimatedDaysRemaining(m)).toBe(3)
  })

  it('returns null when daily dose is null (as_needed)', () => {
    expect(
      estimatedDaysRemaining(med({ frequency_type: 'as_needed' })),
    ).toBeNull()
  })

  it('returns null when stock is 0 and dose is 0', () => {
    expect(estimatedDaysRemaining(med({ dosage_amount: 0 }))).toBeNull()
  })
})

describe('isLowStock', () => {
  it('returns true when stock equals threshold', () => {
    expect(isLowStock(med({ stock_quantity: 3, low_stock_threshold: 3 }))).toBe(
      true,
    )
  })

  it('returns true when stock is below threshold', () => {
    expect(isLowStock(med({ stock_quantity: 1, low_stock_threshold: 3 }))).toBe(
      true,
    )
  })

  it('returns false when stock is above threshold', () => {
    expect(
      isLowStock(med({ stock_quantity: 10, low_stock_threshold: 3 })),
    ).toBe(false)
  })
})

describe('requiredDailyDoses', () => {
  it('returns frequency_value for daily', () => {
    expect(
      requiredDailyDoses(med({ frequency_type: 'daily', frequency_value: 3 })),
    ).toBe(3)
  })

  it('defaults to 1 when frequency_value is null (daily)', () => {
    expect(
      requiredDailyDoses(
        med({ frequency_type: 'daily', frequency_value: null }),
      ),
    ).toBe(1)
  })

  it('returns null for weekly', () => {
    expect(requiredDailyDoses(med({ frequency_type: 'weekly' }))).toBeNull()
  })

  it('returns null for as_needed', () => {
    expect(requiredDailyDoses(med({ frequency_type: 'as_needed' }))).toBeNull()
  })

  it('returns null for custom', () => {
    expect(requiredDailyDoses(med({ frequency_type: 'custom' }))).toBeNull()
  })
})
