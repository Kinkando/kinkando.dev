import { apiFetch } from './client'
import type {
  Quest,
  CreateQuestInput,
  UpdateQuestInput,
  QuestOverview,
  QuestStreaks,
  AchievementsSummary,
  XPEvent,
} from './types'

export function fetchQuestOverview(): Promise<QuestOverview | undefined> {
  return apiFetch<QuestOverview>('/quest/overview', { auth: true })
}

export function fetchQuestStreaks(): Promise<QuestStreaks | undefined> {
  return apiFetch<QuestStreaks>('/quest/streaks', { auth: true })
}

export function fetchAchievements(): Promise<AchievementsSummary | undefined> {
  return apiFetch<AchievementsSummary>('/achievements', { auth: true })
}

export function fetchQuests(type: string): Promise<Quest[] | undefined> {
  return apiFetch<Quest[]>('/quest/quests', { auth: true, query: { type } })
}

export function createQuest(
  input: CreateQuestInput,
): Promise<Quest | undefined> {
  return apiFetch<Quest>('/quest/quests', {
    method: 'POST',
    body: input,
    auth: true,
  })
}

export function updateQuest(
  id: string,
  input: UpdateQuestInput,
): Promise<Quest | undefined> {
  return apiFetch<Quest>(`/quest/quests/${id}`, {
    method: 'PATCH',
    body: input,
    auth: true,
  })
}

export function deleteQuest(id: string): Promise<undefined> {
  return apiFetch<undefined>(`/quest/quests/${id}`, {
    method: 'DELETE',
    auth: true,
  })
}

export function activateQuest(id: string): Promise<Quest | undefined> {
  return apiFetch<Quest>(`/quest/quests/${id}/activate`, {
    method: 'POST',
    auth: true,
  })
}

export function deactivateQuest(id: string): Promise<Quest | undefined> {
  return apiFetch<Quest>(`/quest/quests/${id}/deactivate`, {
    method: 'POST',
    auth: true,
  })
}

export function incrementQuest(id: string): Promise<undefined> {
  return apiFetch<undefined>(`/quest/quests/${id}/increment`, {
    method: 'POST',
    auth: true,
  })
}

export function decrementQuest(id: string): Promise<undefined> {
  return apiFetch<undefined>(`/quest/quests/${id}/decrement`, {
    method: 'POST',
    auth: true,
  })
}

export function fetchQuestHistory(limit = 50): Promise<XPEvent[] | undefined> {
  return apiFetch<XPEvent[]>('/quest/history', {
    auth: true,
    query: { limit: String(limit) },
  })
}
