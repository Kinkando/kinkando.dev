import {
  Brain,
  Code2,
  Cloud,
  ShieldCheck,
  Gamepad2,
  type LucideIcon,
} from 'lucide-react'

export type NewsCategory = 'ai' | 'webdev' | 'cloud' | 'security' | 'gametech'

export interface NewsItem {
  id: string
  title: string
  summary: string
  category: NewsCategory
  source: string
  url: string
  publishedAt: string // ISO date string
  featured?: boolean
}

export const CATEGORIES: { key: NewsCategory; label: string }[] = [
  { key: 'ai', label: 'AI' },
  { key: 'webdev', label: 'Web Dev' },
  { key: 'cloud', label: 'Cloud' },
  { key: 'security', label: 'Security' },
  { key: 'gametech', label: 'Game Tech' },
]

export const CATEGORY_STYLE: Record<
  NewsCategory,
  { gradient: string; icon: LucideIcon; label: string }
> = {
  ai: {
    gradient: 'from-violet-900/60 to-indigo-900/60',
    icon: Brain,
    label: 'AI',
  },
  webdev: {
    gradient: 'from-sky-900/60 to-cyan-900/60',
    icon: Code2,
    label: 'Web Dev',
  },
  cloud: {
    gradient: 'from-blue-900/60 to-slate-900/60',
    icon: Cloud,
    label: 'Cloud',
  },
  security: {
    gradient: 'from-rose-900/60 to-red-900/60',
    icon: ShieldCheck,
    label: 'Security',
  },
  gametech: {
    gradient: 'from-emerald-900/60 to-teal-900/60',
    icon: Gamepad2,
    label: 'Game Tech',
  },
}
