import type { ChatMessage } from '../../lib/api/types'

type Props = {
  message: ChatMessage
  isStreaming?: boolean
}

function ThinkingDots() {
  return (
    <span className="flex items-center gap-1.5 py-0.5">
      {[0, 1, 2].map((i) => (
        <span
          key={i}
          className="inline-block h-2 w-2 animate-bounce rounded-full bg-gray-500"
          style={{ animationDelay: `${i * 0.15}s`, animationDuration: '0.8s' }}
        />
      ))}
    </span>
  )
}

export default function MessageBubble({ message, isStreaming }: Props) {
  const isUser = message.role === 'user'
  const isThinking = !isUser && isStreaming && !message.content

  return (
    <div className={`flex ${isUser ? 'justify-end' : 'justify-start'}`}>
      <div
        className={`max-w-[80%] rounded-2xl px-4 py-3 text-sm leading-relaxed whitespace-pre-wrap ${
          isUser
            ? 'bg-indigo-600 text-white'
            : 'border border-gray-700 bg-gray-800 text-gray-100'
        }`}
      >
        {isThinking ? (
          <ThinkingDots />
        ) : (
          <>
            <span className="animate-fade-in">{message.content}</span>
            {isStreaming && (
              <span className="ml-1 inline-block h-4 w-2 animate-pulse rounded-sm bg-indigo-400 align-middle" />
            )}
          </>
        )}
      </div>
    </div>
  )
}
