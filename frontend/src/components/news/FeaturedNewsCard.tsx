import { CATEGORY_STYLE } from '../../lib/news'
import type { NewsItem } from '../../lib/news'
import { formatDate } from '../../lib/date'

interface Props {
  item: NewsItem
}

export default function FeaturedNewsCard({ item }: Props) {
  const style = CATEGORY_STYLE[item.category]
  const Icon = style.icon

  return (
    <div className="animate-fade-in overflow-hidden rounded-2xl border border-gray-800 bg-gray-900 shadow-lg shadow-black/20 transition-colors hover:border-indigo-700 md:flex">
      {/* Gradient panel */}
      <div
        className={`flex items-center justify-center bg-gradient-to-br ${style.gradient} p-10 md:w-56 md:shrink-0`}
      >
        <Icon className="h-16 w-16 text-white/70" strokeWidth={1.25} />
      </div>

      {/* Content */}
      <div className="flex flex-1 flex-col justify-between gap-4 p-6">
        <div className="flex flex-col gap-3">
          <div className="flex items-center gap-2">
            <span className="rounded-full bg-indigo-600 px-2.5 py-0.5 text-xs font-semibold text-white">
              Featured
            </span>
            <span className="rounded-full bg-indigo-950 px-2 py-0.5 text-xs text-indigo-300">
              {style.label}
            </span>
          </div>
          <h2 className="text-xl leading-snug font-bold text-gray-100">
            {item.title}
          </h2>
          <p className="text-sm leading-relaxed text-gray-400">
            {item.summary}
          </p>
        </div>

        <div className="flex items-center justify-between">
          <span className="text-xs text-gray-500">
            {item.source} · {formatDate(item.publishedAt)}
          </span>
          <a
            href={item.url}
            target="_blank"
            rel="noopener noreferrer"
            className="rounded-lg bg-indigo-600 px-4 py-1.5 text-sm font-medium text-white transition-colors hover:bg-indigo-500"
          >
            Read more →
          </a>
        </div>
      </div>
    </div>
  )
}
