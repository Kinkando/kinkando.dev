import { describe, it, expect } from 'vitest'
import { isOverdue, isDueSoon } from './kanban'
import type { Card } from './api/types'

// Minimal Card stub.
function card(overrides: Partial<Card>): Card {
  return {
    id: 'c1',
    board_id: 'b1',
    column_id: 'col1',
    title: 'Test card',
    content: '',
    description: '',
    priority: 'none',
    tags: [],
    order: 0,
    created_at: '2026-01-01T00:00:00Z',
    ...overrides,
  } as Card
}

// A date string definitely in the past (year 2000).
const PAST = '2000-01-01'
// A date string definitely in the future (far out).
const FUTURE = '2099-12-31'

// Tomorrow: add 1 day to today.
function tomorrow(): string {
  const d = new Date()
  d.setDate(d.getDate() + 1)
  return d.toISOString().slice(0, 10)
}

describe('isOverdue', () => {
  it('returns false when due_date is absent', () => {
    expect(isOverdue(card({}))).toBe(false)
  })

  it('returns true for a due date in the past', () => {
    expect(isOverdue(card({ due_date: PAST }))).toBe(true)
  })

  it('returns false for a far-future due date', () => {
    expect(isOverdue(card({ due_date: FUTURE }))).toBe(false)
  })
})

describe('isDueSoon', () => {
  it('returns false when due_date is absent', () => {
    expect(isDueSoon(card({}))).toBe(false)
  })

  it('returns false for a past due date', () => {
    expect(isDueSoon(card({ due_date: PAST }))).toBe(false)
  })

  it('returns true for tomorrow', () => {
    expect(isDueSoon(card({ due_date: tomorrow() }))).toBe(true)
  })

  it('returns false for a far-future due date', () => {
    expect(isDueSoon(card({ due_date: FUTURE }))).toBe(false)
  })
})
