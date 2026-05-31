import clientConfig from '$config/client';
import { getToken } from '$lib/firebase/token';

export class ApiError extends Error {
  constructor(
    message: string,
    public readonly status: number
  ) {
    super(message);
    this.name = 'ApiError';
  }
}

type QueryParams = Record<string, string | number | undefined>;

function buildQuery(query?: QueryParams): string {
  if (!query) return '';
  const params = new URLSearchParams();
  for (const [k, v] of Object.entries(query)) {
    if (v !== undefined) params.set(k, String(v));
  }
  const qs = params.toString();
  return qs ? '?' + qs : '';
}

const BASE = `${clientConfig.apiUrl}/api/v1`;

export async function apiFetch<T>(
  path: string,
  {
    method = 'GET',
    body,
    auth = false,
    query
  }: {
    method?: string;
    body?: unknown;
    auth?: boolean;
    query?: QueryParams;
  } = {}
): Promise<T> {
  const headers: Record<string, string> = {
    'Content-Type': 'application/json'
  };

  if (auth) {
    const token = await getToken();
    if (!token) throw new ApiError('Not authenticated', 401);
    headers['Authorization'] = `Bearer ${token}`;
  }

  const res = await fetch(`${BASE}${path}${buildQuery(query)}`, {
    method,
    headers,
    body: body !== undefined ? JSON.stringify(body) : undefined
  });

  // 204 No Content (deletes, move)
  if (res.status === 204) return undefined as T;

  const json = await res.json().catch(() => null);

  if (!res.ok) {
    throw new ApiError(json?.error ?? res.statusText, res.status);
  }

  return json?.data as T;
}
