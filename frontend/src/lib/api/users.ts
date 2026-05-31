import { apiFetch } from './client';
import type { AppUser } from './types';

export function ensureUser(): Promise<AppUser> {
  return apiFetch<AppUser>('/users', { method: 'POST', auth: true });
}
