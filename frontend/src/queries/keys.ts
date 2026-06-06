import type { ArchiveFilter } from '../lib/api/types'

export const keys = {
  userMe: ['user', 'me'] as const,
  portfolioProjects: ['portfolio', 'projects'] as const,
  portfolioSkills: ['portfolio', 'skills'] as const,
  financeRecords: (month: string) => ['finance', 'records', month] as const,
  financeSummary: (month: string) => ['finance', 'summary', month] as const,
  financeCategories: ['finance', 'categories'] as const,
  kanbanBoards: ['kanban', 'boards'] as const,
  kanbanBoard: (id: string) => ['kanban', 'board', id] as const,
  kanbanStats: (id: string) => ['kanban', 'stats', id] as const,
  kanbanArchive: (id: string, filter?: ArchiveFilter) =>
    ['kanban', 'archive', id, filter] as const,
  healthProfile: ['health', 'profile'] as const,
  healthWeight: (range: string) => ['health', 'weight', range] as const,
  healthFood: ['health', 'food'] as const,
  healthSleep: (range: string) => ['health', 'sleep', range] as const,
  workoutPresets: ['workout', 'presets'] as const,
  workoutPreset: (id: string) => ['workout', 'preset', id] as const,
  workoutSchedule: ['workout', 'schedule'] as const,
  workoutSessions: (range: string) => ['workout', 'sessions', range] as const,
  workoutSession: (id: string) => ['workout', 'session', id] as const,
  medicines: (includeArchived: boolean) =>
    ['medicine', 'list', includeArchived] as const,
  medicineIntakes: (date?: string) => ['medicine', 'intakes', date] as const,
  medicineAdjustments: (date?: string) =>
    ['medicine', 'adjustments', date] as const,
  questOverview: ['quest', 'overview'] as const,
  questStreaks: ['quest', 'streaks'] as const,
  questList: (type: string) => ['quest', 'list', type] as const,
  questHistory: ['quest', 'history'] as const,
  notificationSettings: ['notifications', 'settings'] as const,
  notificationDevice: ['notifications', 'device'] as const,
}
