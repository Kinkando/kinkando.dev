import { apiFetch } from './client'
import type { User } from './types'

export function provisionUser(): Promise<User | undefined> {
  return apiFetch<User>('/users', { method: 'POST', auth: true })
}
