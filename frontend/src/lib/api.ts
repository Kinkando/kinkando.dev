import { getFirebaseAuth } from '@/lib/firebase';
import type { ApiError, ApiResponse } from '@/types/api';

const BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

export class ApiRequestError extends Error {
  constructor(
    public status: number,
    message: string
  ) {
    super(message);
    this.name = 'ApiRequestError';
  }
}

export async function apiFetch<T>(path: string, options: RequestInit = {}): Promise<T> {
  const headers = new Headers(options.headers);

  // Attach Firebase token if user is signed in
  try {
    const auth = await getFirebaseAuth();
    const user = auth.currentUser;
    if (user) {
      const token = await user.getIdToken();
      headers.set('Authorization', `Bearer ${token}`);
    }
  } catch {
    // Firebase not available (SSR) — skip token attachment
  }

  // Set JSON content type for mutation requests
  if (options.body && !headers.has('Content-Type')) {
    headers.set('Content-Type', 'application/json');
  }

  const res = await fetch(`${BASE_URL}${path}`, {
    ...options,
    headers
  });

  // 204 No Content
  if (res.status === 204) {
    return undefined as T;
  }

  const json = await res.json();

  if (!res.ok) {
    const errorBody = json as ApiError;
    throw new ApiRequestError(res.status, errorBody.error || 'Unknown error');
  }

  return (json as ApiResponse<T>).data;
}
