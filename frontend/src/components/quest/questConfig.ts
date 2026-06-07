import type {
  QuestType,
  SourceType,
  DailyQuestStatus,
} from '../../lib/api/types'

export type FormState = {
  title: string
  description: string
  xp_reward: string
  target_count: string
  source_type: SourceType
}

export function questToForm(q: DailyQuestStatus): FormState {
  return {
    title: q.title,
    description: q.description,
    xp_reward: String(q.xp_reward),
    target_count: String(q.target_count),
    source_type: q.source_type,
  }
}

export const SOURCE_LABELS: Record<SourceType, string> = {
  manual: 'Manual',
  medicine: 'Medicine (auto)',
  workout: 'Workout (auto)',
  supplement: 'Supplement (auto)',
  weight: 'Weight (auto)',
  sleep: 'Sleep (auto)',
}

export const QUEST_SOURCE_ROUTES: Record<SourceType, string | null> = {
  manual: null,
  medicine: '/health/medicine',
  workout: '/health/workout',
  supplement: '/health/supplements',
  weight: '/health?tab=progress',
  sleep: '/health/sleep',
}

/** Returns the navigation route for an auto-linked quest, or null for manual. */
export function questSourceRoute(source: SourceType): string | null {
  return QUEST_SOURCE_ROUTES[source]
}

type QuestTypeConfig = {
  titleNoun: string
  period: 'day' | 'week'
  accentBar: string
  accentBadge: string
  manualOptionLabel: string
  autoLead: string
  autoTail: string
  emptyText: string
  deleteText: string
  defaultForm: FormState
  titlePlaceholder: string
  targetPlaceholder: string
}

export const QUEST_TYPE_CONFIG: Record<QuestType, QuestTypeConfig> = {
  daily: {
    titleNoun: 'Daily Quest',
    period: 'day',
    accentBar: 'bg-sky-500',
    accentBadge: 'bg-sky-900/60 text-sky-400',
    manualOptionLabel: 'Manual (check off yourself)',
    autoLead: 'auto-complete',
    autoTail: 'No manual ±buttons will be shown.',
    emptyText: 'No daily quests yet. Add one to get started.',
    deleteText:
      'Delete this daily quest? All completions and XP events will be removed.',
    defaultForm: {
      title: '',
      description: '',
      xp_reward: '10',
      target_count: '1',
      source_type: 'manual',
    },
    titlePlaceholder: 'e.g. Take thyroid medication',
    targetPlaceholder: '1',
  },
  weekly: {
    titleNoun: 'Weekly Quest',
    period: 'week',
    accentBar: 'bg-violet-500',
    accentBadge: 'bg-violet-900/60 text-violet-400',
    manualOptionLabel: 'Manual (adjust yourself)',
    autoLead: 'auto-increment',
    autoTail: 'No manual ±buttons will be shown.',
    emptyText: 'No weekly quests yet. Add one to get started.',
    deleteText:
      'Delete this weekly quest? All completions and XP events will be removed.',
    defaultForm: {
      title: '',
      description: '',
      xp_reward: '30',
      target_count: '3',
      source_type: 'manual',
    },
    titlePlaceholder: 'e.g. Workout',
    targetPlaceholder: '3',
  },
}
