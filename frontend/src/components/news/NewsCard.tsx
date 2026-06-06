import { CATEGORY_STYLE } from '../../lib/news'
import type { NewsItem } from '../../lib/news'
import { formatDate } from '../../lib/date'

interface Props {
  item: NewsItem
}

export default function NewsCard({ item }: Props) {
  const style = CATEGORY_STYLE[item.category]
  const Icon = style.icon

  return (
    <div className="animate-fade-in flex flex-col overflow-hidden rounded-xl border border-gray-800 bg-gray-900 transition-all hover:-translate-y-0.5 hover:border-indigo-700">
      {/* Gradient header strip */}
      <div
        className={`flex h-20 items-center justify-center bg-gradient-to-br ${style.gradient}`}
      >
        <Icon className="h-8 w-8 text-white/70" strokeWidth={1.5} />
      </div>

      {/* Body */}
      <div className="flex flex-1 flex-col gap-2 p-4">
        <span className="w-fit rounded-full bg-indigo-950 px-2 py-0.5 text-xs text-indigo-300">
          {style.label}
        </span>
        <h3 className="leading-snug font-semibold text-gray-100">
          {item.title}
        </h3>
        <p className="line-clamp-3 text-sm text-gray-400">{item.summary}</p>
      </div>

      {/* Footer */}
      <div className="flex items-center justify-between border-t border-gray-800 px-4 py-3">
        <span className="text-xs text-gray-500">
          {item.source} · {formatDate(item.publishedAt)}
        </span>
        <a
          href={item.url}
          target="_blank"
          rel="noopener noreferrer"
          className="text-xs font-medium text-indigo-400 transition-colors hover:text-indigo-300"
        >
          Read more →
        </a>
      </div>
    </div>
  )
}
