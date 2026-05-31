import { getIdToken } from '../firebase'
import env from '../../config/env'
import { ApiError } from './client'
import type { ChatMessage } from './types'

/**
 * Streams an AI chat response from the backend.
 *
 * Sends the full message history (including the new user turn) as JSON and
 * reads the SSE response, calling onToken for each text chunk. Resolves when
 * the stream ends normally, rejects on error.
 */
export async function streamChat(
  messages: ChatMessage[],
  onToken: (token: string) => void,
  signal?: AbortSignal,
): Promise<void> {
  const token = await getIdToken()

  const res = await fetch(`${env.apiUrl}/api/v1/ai-chat`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
    },
    body: JSON.stringify({ messages }),
    signal,
  })

  if (!res.ok) {
    const json = await res.json().catch(() => ({}))
    throw new ApiError(
      res.status,
      (json as { error?: string }).error ?? `HTTP ${res.status}`,
    )
  }

  const reader = res.body!.getReader()
  const decoder = new TextDecoder()
  let buf = ''

  while (true) {
    const { done, value } = await reader.read()
    if (done) break

    buf += decoder.decode(value, { stream: true })

    // Split on double-newline SSE frame boundaries.
    const frames = buf.split('\n\n')
    buf = frames.pop() ?? '' // last element may be incomplete

    for (const frame of frames) {
      if (!frame.trim()) continue

      let eventType = 'message'
      let dataLine = ''

      for (const line of frame.split('\n')) {
        if (line.startsWith('event: ')) eventType = line.slice(7).trim()
        if (line.startsWith('data: ')) dataLine = line.slice(6).trim()
      }

      if (eventType === 'done') return
      if (eventType === 'error') {
        const json = JSON.parse(dataLine) as { error?: string }
        throw new ApiError(500, json.error ?? 'Stream error')
      }

      // Default event: a token chunk.
      if (dataLine) {
        const json = JSON.parse(dataLine) as { token?: string }
        if (json.token) onToken(json.token)
      }
    }
  }
}
