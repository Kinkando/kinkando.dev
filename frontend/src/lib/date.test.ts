import { describe, it, expect } from 'vitest'
import { addDays, dayOfWeek, calculateAge, todayDate, todayMonth } from './date'

describe('addDays', () => {
  it('adds positive days', () => {
    expect(addDays('2026-01-01', 1)).toBe('2026-01-02')
    expect(addDays('2026-01-01', 30)).toBe('2026-01-31')
  })

  it('crosses month boundary', () => {
    expect(addDays('2026-01-31', 1)).toBe('2026-02-01')
  })

  it('crosses year boundary', () => {
    expect(addDays('2025-12-31', 1)).toBe('2026-01-01')
  })

  it('subtracts days with negative offset', () => {
    expect(addDays('2026-06-10', -1)).toBe('2026-06-09')
    expect(addDays('2026-03-01', -1)).toBe('2026-02-28')
  })

  it('returns the same date for offset 0', () => {
    expect(addDays('2026-06-10', 0)).toBe('2026-06-10')
  })
})

describe('dayOfWeek', () => {
  // Known anchor: 2026-06-08 is a Monday (1), 2026-06-14 is Sunday (0)
  it('returns 1 for Monday', () => {
    expect(dayOfWeek('2026-06-08')).toBe(1)
  })

  it('returns 0 for Sunday', () => {
    expect(dayOfWeek('2026-06-14')).toBe(0)
  })

  it('returns 6 for Saturday', () => {
    expect(dayOfWeek('2026-06-13')).toBe(6)
  })

  it('returns 3 for Wednesday', () => {
    expect(dayOfWeek('2026-06-10')).toBe(3)
  })
})

describe('calculateAge', () => {
  // Build dates relative to "today" so these tests remain valid any day they run.
  function relativeDate(
    yearsAgo: number,
    monthOffset = 0,
    dayOffset = 0,
  ): string {
    const d = new Date()
    d.setFullYear(
      d.getFullYear() - yearsAgo,
      d.getMonth() + monthOffset,
      d.getDate() + dayOffset,
    )
    return d.toISOString().slice(0, 10)
  }

  it('returns N for a birthday exactly N years ago (same calendar day)', () => {
    expect(calculateAge(relativeDate(30))).toBe(30)
  })

  it('returns N-1 when the birthday falls later this year (not yet reached)', () => {
    // Birthday 30 years ago but 1 month in the future relative to today.
    // As of today, the person has not yet turned 30.
    expect(calculateAge(relativeDate(30, 1))).toBe(29)
  })

  it('returns N when the birthday already passed this year', () => {
    // Birthday 30 years ago but 1 month in the past relative to today.
    expect(calculateAge(relativeDate(30, -1))).toBe(30)
  })
})

describe('todayDate and todayMonth', () => {
  it('todayDate returns a YYYY-MM-DD string', () => {
    expect(todayDate()).toMatch(/^\d{4}-\d{2}-\d{2}$/)
  })

  it('todayMonth is the first 7 chars of todayDate', () => {
    expect(todayMonth()).toBe(todayDate().slice(0, 7))
  })

  it('todayMonth returns a YYYY-MM string', () => {
    expect(todayMonth()).toMatch(/^\d{4}-\d{2}$/)
  })
})
