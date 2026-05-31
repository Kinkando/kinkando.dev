import { getIdToken } from '../firebase'
import env from '../../config/env'

export class ApiError extends Error {
  constructor(
    public status: number,
    message: string,
  ) {
    super(message)
    this.name = 'ApiError'
  }
}

const BASE = `${env.apiUrl}/api/v1`

type Options = {
  method?: string
  body?: unknown
  auth?: boolean
  query?: Record<string, string>
}

export async function apiFetch<T>(
  path: string,
  opts: Options = {},
): Promise<T | undefined> {
  const { method = 'GET', body, auth = false, query } = opts
  const headers: HeadersInit = { 'Content-Type': 'application/json' }

  if (auth) {
    const token = await getIdToken()
    if (token) headers['Authorization'] = `Bearer ${token}`
  }

  let url = `${BASE}${path}`
  if (query) {
    url += `?${new URLSearchParams(query).toString()}`
  }

  const res = await fetch(url, {
    method,
    headers,
    body: body != null ? JSON.stringify(body) : undefined,
  })

  if (res.status === 204) return undefined

  const json = await res.json()
  if (!res.ok) {
    throw new ApiError(
      res.status,
      (json as { error?: string }).error ?? `HTTP ${res.status}`,
    )
  }

  return (json as { data: T }).data
}
