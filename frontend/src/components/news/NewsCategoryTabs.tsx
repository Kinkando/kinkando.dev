import { CATEGORIES } from '../../lib/news'
import type { NewsCategory } from '../../lib/news'

interface Props {
  active: NewsCategory | 'all'
  onChange: (category: NewsCategory | 'all') => void
}

export default function NewsCategoryTabs({ active, onChange }: Props) {
  const tabClass = (key: NewsCategory | 'all') =>
    `rounded-full px-3 py-1 text-sm transition-colors cursor-pointer ${
      active === key
        ? 'bg-indigo-600 text-white'
        : 'bg-gray-800 text-gray-300 hover:bg-gray-700 hover:text-gray-100'
    }`

  return (
    <div className="flex flex-wrap gap-2">
      <button className={tabClass('all')} onClick={() => onChange('all')}>
        All
      </button>
      {CATEGORIES.map(({ key, label }) => (
        <button
          key={key}
          className={tabClass(key)}
          onClick={() => onChange(key)}
        >
          {label}
        </button>
      ))}
    </div>
  )
}
