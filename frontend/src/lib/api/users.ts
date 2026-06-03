import { apiFetch } from './client'
import type { LineLinkCode, User } from './types'

export function provisionUser(): Promise<User | undefined> {
  return apiFetch<User>('/users', { method: 'POST', auth: true })
}

export function getMe(): Promise<User | undefined> {
  return apiFetch<User>('/users/me', { auth: true })
}

export function createLineLinkCode(): Promise<LineLinkCode | undefined> {
  return apiFetch<LineLinkCode>('/users/line/link-code', {
    method: 'POST',
    auth: true,
  })
}

export function unlinkLine(): Promise<undefined> {
  return apiFetch('/users/line', {
    method: 'DELETE',
    auth: true,
  }) as Promise<undefined>
}
