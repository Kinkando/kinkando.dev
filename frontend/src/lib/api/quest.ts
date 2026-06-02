import { apiFetch } from './client'
import type {
  Quest,
  CreateQuestInput,
  UpdateQuestInput,
  QuestOverview,
  XPEvent,
} from './types'

export function fetchQuestOverview(): Promise<QuestOverview | undefined> {
  return apiFetch<QuestOverview>('/quest/overview', { auth: true })
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

export function completeDaily(id: string): Promise<undefined> {
  return apiFetch<undefined>(`/quest/quests/${id}/complete`, {
    method: 'POST',
    auth: true,
  })
}

export function uncompleteDaily(id: string): Promise<undefined> {
  return apiFetch<undefined>(`/quest/quests/${id}/complete`, {
    method: 'DELETE',
    auth: true,
  })
}

export function incrementWeekly(id: string): Promise<undefined> {
  return apiFetch<undefined>(`/quest/quests/${id}/increment`, {
    method: 'POST',
    auth: true,
  })
}

export function decrementWeekly(id: string): Promise<undefined> {
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
