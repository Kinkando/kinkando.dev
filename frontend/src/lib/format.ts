/** Formats a number as Thai Baht currency (e.g. "฿1,234.56"). */
export function formatCurrency(amount: number): string {
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: 'THB',
  }).format(amount)
}

/** Formats a number with thousands separators (e.g. 12345 → "12,345"). */
export function formatNumber(n: number): string {
  return n.toLocaleString()
}
