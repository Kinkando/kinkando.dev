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

export const NEWS: NewsItem[] = [
  {
    id: '1',
    title: 'Claude 4 Brings Extended Thinking and Stronger Code Generation',
    summary:
      'Anthropic releases Claude 4 with new extended thinking mode, outperforming previous models on SWE-bench and achieving state-of-the-art results on math and science reasoning benchmarks.',
    category: 'ai',
    source: 'Anthropic Blog',
    url: 'https://www.anthropic.com/news',
    publishedAt: '2026-05-28',
    featured: true,
  },
  {
    id: '2',
    title: 'React 20 Compiler Goes Stable — Auto-Memoization Is Here',
    summary:
      'The React team announces the stable release of the React Compiler, which automatically optimizes components without manual useMemo and useCallback calls, drastically reducing re-renders.',
    category: 'webdev',
    source: 'React Blog',
    url: 'https://react.dev/blog',
    publishedAt: '2026-05-22',
  },
  {
    id: '3',
    title:
      'AWS Introduces Graviton 5 Instances with 40% Better Price-Performance',
    summary:
      'Amazon Web Services launches EC2 instances powered by the new Graviton 5 chip, delivering significant improvements in throughput for compute-intensive workloads at lower cost.',
    category: 'cloud',
    source: 'AWS News',
    url: 'https://aws.amazon.com/blogs/aws/',
    publishedAt: '2026-05-20',
  },
  {
    id: '4',
    title: 'Critical Zero-Day in OpenSSH Patched — Update Immediately',
    summary:
      'A remotely exploitable vulnerability in OpenSSH versions prior to 9.9p1 allows unauthenticated root access. All Linux distributions have issued emergency patches.',
    category: 'security',
    source: 'The Hacker News',
    url: 'https://thehackernews.com',
    publishedAt: '2026-05-18',
  },
  {
    id: '5',
    title:
      'Unreal Engine 6 Previewed: Nanite 2 and Real-Time Path Tracing at 4K',
    summary:
      'Epic Games showcases Unreal Engine 6 at the Game Developers Summit, demonstrating fully dynamic global illumination and sub-millisecond nanite geometry streaming on consumer GPUs.',
    category: 'gametech',
    source: 'Gamasutra',
    url: 'https://www.gamedeveloper.com',
    publishedAt: '2026-05-15',
  },
  {
    id: '6',
    title: 'Gemini Ultra 2 Achieves Human-Level Performance on MMLU Pro',
    summary:
      "Google DeepMind's Gemini Ultra 2 surpasses human expert scores on the MMLU Pro benchmark across medicine, law, and engineering disciplines.",
    category: 'ai',
    source: 'Google DeepMind',
    url: 'https://deepmind.google/discover/blog/',
    publishedAt: '2026-05-12',
  },
  {
    id: '7',
    title: 'Vite 7 Released: Native ESM-First Bundling and 2x Faster HMR',
    summary:
      'Vite 7 drops CommonJS transform fallbacks and ships a rewritten HMR engine, cutting hot module replacement times in half for large TypeScript projects.',
    category: 'webdev',
    source: 'Vite Docs',
    url: 'https://vitejs.dev/blog/',
    publishedAt: '2026-05-10',
  },
  {
    id: '8',
    title: 'Google Cloud Announces Serverless GPU Instances for AI Inference',
    summary:
      'Google Cloud Run now supports GPU backends with per-request billing, enabling developers to serve large language models without managing persistent GPU infrastructure.',
    category: 'cloud',
    source: 'Google Cloud Blog',
    url: 'https://cloud.google.com/blog',
    publishedAt: '2026-05-08',
  },
  {
    id: '9',
    title: 'Supply Chain Attack Hits npm Ecosystem: 200+ Packages Compromised',
    summary:
      'Security researchers uncover a coordinated supply chain attack targeting popular npm packages, injecting data-exfiltration code affecting millions of weekly downloads.',
    category: 'security',
    source: 'Snyk Security',
    url: 'https://snyk.io/blog/',
    publishedAt: '2026-05-05',
  },
  {
    id: '10',
    title:
      'Steam Announces Native Linux Support for Anti-Cheat via Kernel Module',
    summary:
      'Valve partners with major anti-cheat vendors to ship a unified Linux kernel module, unblocking hundreds of previously Windows-only multiplayer titles on SteamOS.',
    category: 'gametech',
    source: 'Steam Blog',
    url: 'https://store.steampowered.com/news/',
    publishedAt: '2026-05-02',
  },
  {
    id: '11',
    title:
      'TypeScript 6.0 Introduces Structural Enum Types and Faster Incremental Builds',
    summary:
      'Microsoft ships TypeScript 6.0 with opt-in structural enum comparisons, isolated declaration emit GA, and a 35% improvement in incremental build times for monorepos.',
    category: 'webdev',
    source: 'TypeScript Blog',
    url: 'https://devblogs.microsoft.com/typescript/',
    publishedAt: '2026-04-28',
  },
  {
    id: '12',
    title: 'NVIDIA Blackwell Ultra: 1.5x H100 Throughput for LLM Training',
    summary:
      'NVIDIA begins shipping Blackwell Ultra B300 GPUs to cloud providers, offering improved memory bandwidth and NVLink interconnect speeds specifically tuned for 100B+ parameter model training.',
    category: 'ai',
    source: 'NVIDIA Blog',
    url: 'https://blogs.nvidia.com',
    publishedAt: '2026-04-25',
  },
]
