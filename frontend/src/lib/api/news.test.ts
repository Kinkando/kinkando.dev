import { describe, it, expect } from 'vitest'
import { mapNewsItem } from './news'
import type { ApiNewsItem } from './news'

const apiItem: ApiNewsItem = {
  id: 'abc123',
  title: 'New AI model released',
  summary: 'A brief description of the release.',
  category: 'ai',
  source: 'TechCrunch',
  url: 'https://techcrunch.com/article',
  published_at: '2026-06-10T08:00:00Z',
  featured: true,
}

describe('mapNewsItem', () => {
  it('maps published_at → publishedAt (anti-corruption mapping)', () => {
    const result = mapNewsItem(apiItem)
    expect(result.publishedAt).toBe('2026-06-10T08:00:00Z')
    expect(
      (result as unknown as Record<string, unknown>)['published_at'],
    ).toBeUndefined()
  })

  it('passes through all other fields unchanged', () => {
    const result = mapNewsItem(apiItem)
    expect(result.id).toBe('abc123')
    expect(result.title).toBe('New AI model released')
    expect(result.summary).toBe('A brief description of the release.')
    expect(result.category).toBe('ai')
    expect(result.source).toBe('TechCrunch')
    expect(result.url).toBe('https://techcrunch.com/article')
    expect(result.featured).toBe(true)
  })

  it('preserves featured: false', () => {
    const result = mapNewsItem({ ...apiItem, featured: false })
    expect(result.featured).toBe(false)
  })
})
