import type { ArchiveFilter } from '../lib/api/types'

export const keys = {
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
  healthWeight: ['health', 'weight'] as const,
  healthFood: ['health', 'food'] as const,
  healthSleep: ['health', 'sleep'] as const,
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
}
