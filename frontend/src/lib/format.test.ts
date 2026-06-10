import { describe, it, expect } from 'vitest'
import { formatCurrency, formatNumber } from './format'

describe('formatCurrency', () => {
  // Node.js Intl renders THB as "THB"; browsers may render "฿". Both are valid.
  // Tests check the number portion and that a currency marker is present.
  it('formats a whole number', () => {
    const result = formatCurrency(1234)
    expect(result).toContain('1,234')
    expect(result).toMatch(/THB|฿/)
  })

  it('formats a decimal amount', () => {
    const result = formatCurrency(1234.56)
    expect(result).toContain('1,234.56')
    expect(result).toMatch(/THB|฿/)
  })

  it('formats zero', () => {
    const result = formatCurrency(0)
    expect(result).toContain('0')
    expect(result).toMatch(/THB|฿/)
  })

  it('formats a large number with commas', () => {
    const result = formatCurrency(1000000)
    expect(result).toContain('1,000,000')
  })
})

describe('formatNumber', () => {
  it('adds thousands separator', () => {
    expect(formatNumber(12345)).toBe('12,345')
  })

  it('handles numbers below 1000 without separator', () => {
    expect(formatNumber(999)).toBe('999')
  })

  it('handles zero', () => {
    expect(formatNumber(0)).toBe('0')
  })

  it('handles large numbers', () => {
    expect(formatNumber(1000000)).toBe('1,000,000')
  })
})
