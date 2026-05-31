import { useEffect, useRef, useState } from 'react'
import { RotateCcw } from 'lucide-react'
import { streamChat } from '../lib/api/chat'
import type { ChatMessage, ChatUsage } from '../lib/api/types'
import MessageBubble from '../components/chat/MessageBubble'
import ChatInput from '../components/chat/ChatInput'

// Maximum number of history turns sent to the backend per request.
const MAX_HISTORY = 20

// Gemini 2.0 Flash pricing (USD per token)
const INPUT_PRICE_PER_TOKEN = 0.075 / 1_000_000
const OUTPUT_PRICE_PER_TOKEN = 0.3 / 1_000_000

type SessionUsage = {
  inputTokens: number
  outputTokens: number
}

function formatCost(usage: SessionUsage): string {
  const cost =
    usage.inputTokens * INPUT_PRICE_PER_TOKEN +
    usage.outputTokens * OUTPUT_PRICE_PER_TOKEN
  if (cost === 0) return '$0'
  if (cost < 0.000001) return '< $0.000001'
  return `$${cost.toFixed(6)}`
}

export default function ChatPage() {
  const [messages, setMessages] = useState<ChatMessage[]>([])
  const [streaming, setStreaming] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [sessionUsage, setSessionUsage] = useState<SessionUsage>({
    inputTokens: 0,
    outputTokens: 0,
  })
  const bottomRef = useRef<HTMLDivElement>(null)
  const abortRef = useRef<AbortController | null>(null)

  // Scroll to bottom whenever messages change.
  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [messages])

  function handleNewSession() {
    abortRef.current?.abort()
    setMessages([])
    setSessionUsage({ inputTokens: 0, outputTokens: 0 })
    setError(null)
    setStreaming(false)
  }

  async function handleSend(text: string) {
    setError(null)

    const userMessage: ChatMessage = { role: 'user', content: text }
    const assistantMessage: ChatMessage = { role: 'assistant', content: '' }

    setMessages((prev) => [...prev, userMessage, assistantMessage])
    setStreaming(true)

    // Build the history to send: all existing messages + the new user message.
    // Trim to the last MAX_HISTORY entries so the payload stays bounded.
    const history = [...messages, userMessage].slice(-MAX_HISTORY)

    const abort = new AbortController()
    abortRef.current = abort

    try {
      await streamChat(
        history,
        (token) => {
          setMessages((prev) => {
            const next = [...prev]
            const last = next[next.length - 1]
            if (last?.role === 'assistant') {
              next[next.length - 1] = {
                ...last,
                content: last.content + token,
              }
            }
            return next
          })
        },
        (usage: ChatUsage) => {
          setSessionUsage((prev) => ({
            inputTokens: prev.inputTokens + usage.inputTokens,
            outputTokens: prev.outputTokens + usage.outputTokens,
          }))
        },
        abort.signal,
      )
    } catch (err: unknown) {
      if ((err as { name?: string }).name === 'AbortError') return
      const msg =
        err instanceof Error
          ? err.message
          : 'Something went wrong. Please try again.'
      setError(msg)
      // Remove the empty assistant placeholder on error.
      setMessages((prev) => {
        const next = [...prev]
        if (
          next[next.length - 1]?.role === 'assistant' &&
          !next[next.length - 1].content
        ) {
          next.pop()
        }
        return next
      })
    } finally {
      setStreaming(false)
      abortRef.current = null
    }
  }

  const hasUsage = sessionUsage.inputTokens > 0 || sessionUsage.outputTokens > 0

  return (
    <main
      className="mx-auto flex max-w-3xl flex-col px-6 py-8"
      style={{ height: 'calc(100vh - 57px)' }}
    >
      {/* Header */}
      <div className="mb-6 flex items-center justify-between">
        <h1 className="text-xl font-semibold text-gray-100">AI Assistant</h1>
        <div className="flex items-center gap-3">
          {hasUsage && (
            <span className="text-xs text-gray-500">
              {sessionUsage.inputTokens.toLocaleString()} in ·{' '}
              {sessionUsage.outputTokens.toLocaleString()} out ·{' '}
              <span className="text-gray-400">{formatCost(sessionUsage)}</span>
            </span>
          )}
          <button
            onClick={handleNewSession}
            className="flex items-center gap-1.5 rounded-lg border border-gray-700 px-3 py-1.5 text-xs text-gray-400 transition hover:border-gray-500 hover:text-gray-200"
          >
            <RotateCcw size={12} />
            New session
          </button>
        </div>
      </div>

      {/* Message list */}
      <div className="flex-1 space-y-4 overflow-y-auto pr-1">
        {messages.length === 0 && (
          <p className="mt-16 text-center text-sm text-gray-500">
            Ask anything about your finances or tasks.
          </p>
        )}
        {messages.map((m, i) => {
          const isLastAssistant =
            i === messages.length - 1 && m.role === 'assistant'
          return (
            <MessageBubble
              key={i}
              message={m}
              isStreaming={isLastAssistant && streaming}
            />
          )
        })}
        <div ref={bottomRef} />
      </div>

      {/* Error banner */}
      {error && (
        <div className="mt-3 rounded-lg border border-red-800 bg-red-950 px-4 py-2 text-sm text-red-300">
          {error}
        </div>
      )}

      {/* Input */}
      <div className="mt-4">
        <ChatInput onSend={handleSend} disabled={streaming} />
      </div>
    </main>
  )
}
