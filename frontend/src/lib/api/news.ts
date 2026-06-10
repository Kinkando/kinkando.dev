import { apiFetch } from './client'
import type { NewsItem, NewsCategory } from '../news'

// API shape (snake_case). Mapped to the app's NewsItem so the existing cards
// (which read `publishedAt`) stay unchanged.
export type ApiNewsItem = {
  id: string
  title: string
  summary: string
  category: NewsCategory
  source: string
  url: string
  published_at: string
  featured: boolean
}

/** Maps a single snake_case API response item to the camelCase app shape. */
export function mapNewsItem(it: ApiNewsItem): NewsItem {
  return {
    id: it.id,
    title: it.title,
    summary: it.summary,
    category: it.category,
    source: it.source,
    url: it.url,
    publishedAt: it.published_at,
    featured: it.featured,
  }
}

export async function fetchNews(): Promise<NewsItem[]> {
  const items = await apiFetch<ApiNewsItem[]>('/news')
  return (items ?? []).map(mapNewsItem)
}
