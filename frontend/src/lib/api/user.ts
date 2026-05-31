import { apiFetch } from '@/lib/api';
import type { User } from '@/types/user';

export function ensureUser(): Promise<User> {
  return apiFetch('/api/v1/users', { method: 'POST' });
}
