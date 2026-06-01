import { useState } from 'react'
import { useDocumentTitle } from '../hooks/useDocumentTitle'
import { NEWS } from '../lib/news'
import type { NewsCategory } from '../lib/news'
import NewsCategoryTabs from '../components/news/NewsCategoryTabs'
import FeaturedNewsCard from '../components/news/FeaturedNewsCard'
import NewsCard from '../components/news/NewsCard'

export default function NewsPage() {
  useDocumentTitle('News')
  const [active, setActive] = useState<NewsCategory | 'all'>('all')

  const featured = NEWS.find((n) => n.featured)

  const filtered = NEWS.filter((n) => {
    if (active !== 'all' && n.category !== active) return false
    // Exclude the featured item from the grid when viewing All
    if (active === 'all' && n.featured) return false
    return true
  })

  return (
    <main className="mx-auto max-w-5xl px-6 py-12">
      {/* Header */}
      <h1 className="mb-2 text-3xl font-bold text-gray-100">NEWS</h1>
      <p className="mb-8 text-gray-400">
        IT, AI, Dev, Cloud, Security และเทคโนโลยีที่น่าสนใจ
      </p>

      {/* Category tabs */}
      <div className="mb-8">
        <NewsCategoryTabs active={active} onChange={setActive} />
      </div>

      {/* Featured */}
      {featured && active === 'all' && (
        <section className="mb-10">
          <h2 className="mb-4 text-sm font-semibold tracking-wider text-gray-500 uppercase">
            Featured
          </h2>
          <FeaturedNewsCard item={featured} />
        </section>
      )}

      {/* Latest grid */}
      <section>
        <h2 className="mb-5 text-sm font-semibold tracking-wider text-gray-500 uppercase">
          {active === 'all' ? 'Latest' : 'Results'}
        </h2>
        {filtered.length === 0 ? (
          <p className="text-gray-500">No news in this category yet.</p>
        ) : (
          <div className="grid grid-cols-1 gap-5 sm:grid-cols-2 lg:grid-cols-3">
            {filtered.map((item) => (
              <NewsCard key={item.id} item={item} />
            ))}
          </div>
        )}
      </section>
    </main>
  )
}
