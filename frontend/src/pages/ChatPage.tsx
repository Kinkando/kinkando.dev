import { useEffect, useRef, useState } from 'react'
import { RotateCcw } from 'lucide-react'
import { streamChat } from '../lib/api/chat'
import type { ChatMessage, ChatUsage } from '../lib/api/types'
import MessageBubble from '../components/chat/MessageBubble'
import ChatInput from '../components/chat/ChatInput'
import { useDocumentTitle } from '../hooks/useDocumentTitle'

// Maximum number of history turns sent to the backend per request.
const MAX_HISTORY = 20

// Gemini 2.0 Flash pricing (USD per token)
const INPUT_PRICE_PER_TOKEN = 0.075 / 1_000_000
const OUTPUT_PRICE_PER_TOKEN = 0.3 / 1_000_000

// Approximate exchange rate — update as needed
const THB_PER_USD = 33

type SessionUsage = {
  inputTokens: number
  outputTokens: number
}

function calcCostUSD(usage: SessionUsage): number {
  return (
    usage.inputTokens * INPUT_PRICE_PER_TOKEN +
    usage.outputTokens * OUTPUT_PRICE_PER_TOKEN
  )
}

function formatUSD(cost: number): string {
  if (cost === 0) return '$0'
  if (cost < 0.000001) return '< $0.000001'
  return `$${cost.toFixed(6)}`
}

function formatTHB(cost: number): string {
  const baht = cost * THB_PER_USD
  if (baht === 0) return '฿0'
  if (baht < 0.00001) return '< ฿0.00001'
  return `฿${baht.toFixed(5)}`
}

const SUGGESTIONS = [
  {
    label: 'Finance summary',
    description: 'Summarize my spending or income this month',
    prompt: 'Give me a summary of my finances this month.',
  },
  {
    label: 'Task overview',
    description: 'What tasks do I have open or in progress?',
    prompt: 'Give me an overview of my current kanban tasks.',
  },
  {
    label: 'Budget advice',
    description: 'Tips on managing my budget based on my records',
    prompt: 'Based on my finance records, give me some budget advice.',
  },
  {
    label: 'Productivity tips',
    description: 'Help me prioritize or plan my work',
    prompt: 'Help me prioritize my tasks and plan my day.',
  },
]

export default function ChatPage() {
  useDocumentTitle('Chat')
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
      <div className="mb-6 flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
        <h1 className="text-xl font-semibold text-gray-100">AI Assistant</h1>
        <div className="flex flex-wrap items-center gap-3">
          {hasUsage && (
            <span className="text-xs text-gray-500">
              {sessionUsage.inputTokens.toLocaleString()} in ·{' '}
              {sessionUsage.outputTokens.toLocaleString()} out ·{' '}
              <span className="text-gray-400">
                {formatUSD(calcCostUSD(sessionUsage))}
              </span>
              <span className="mx-1 text-gray-600">/</span>
              <span className="text-gray-400">
                {formatTHB(calcCostUSD(sessionUsage))}
              </span>
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
          <div className="mt-8 flex flex-col items-center gap-6">
            <div className="flex flex-col items-center gap-3 text-center">
              <img
                src="/chibi.jpg"
                alt="AI"
                className="h-16 w-16 rounded-full object-cover ring-2 ring-indigo-500/40"
              />
              <div>
                <p className="text-base font-medium text-gray-200">
                  Hello! I'm your AI Assistant
                </p>
                <p className="mt-1 text-sm text-gray-500">
                  Here's what I can help you with:
                </p>
              </div>
            </div>
            <div className="grid w-full grid-cols-1 gap-2 sm:grid-cols-2">
              {SUGGESTIONS.map((s) => (
                <button
                  key={s.label}
                  onClick={() => handleSend(s.prompt)}
                  className="flex flex-col gap-1 rounded-xl border border-gray-700 bg-gray-800/50 px-4 py-3 text-left transition hover:border-indigo-500/60 hover:bg-gray-800"
                >
                  <span className="text-sm font-medium text-gray-200">
                    {s.label}
                  </span>
                  <span className="text-xs text-gray-500">{s.description}</span>
                </button>
              ))}
            </div>
          </div>
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
