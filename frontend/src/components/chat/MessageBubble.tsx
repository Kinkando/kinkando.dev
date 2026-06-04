import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
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
    <div
      className={`flex items-end gap-2 ${isUser ? 'justify-end' : 'justify-start'}`}
    >
      {!isUser && (
        <img
          src="/images/chibi.jpg"
          alt="AI"
          className="h-8 w-8 shrink-0 rounded-full object-cover"
        />
      )}
      <div
        className={`max-w-[80%] rounded-2xl px-4 py-3 text-sm leading-relaxed ${
          isUser
            ? 'bg-indigo-600 text-white'
            : 'border border-gray-700 bg-gray-800 text-gray-100'
        }`}
      >
        {isThinking ? (
          <ThinkingDots />
        ) : isUser ? (
          <span className="animate-fade-in whitespace-pre-wrap">
            {message.content}
          </span>
        ) : (
          <div className="animate-fade-in">
            <div className="prose prose-invert prose-sm max-w-none">
              <ReactMarkdown remarkPlugins={[remarkGfm]}>
                {message.content}
              </ReactMarkdown>
            </div>
            {isStreaming && (
              <span className="mt-1 block h-1.5 w-4 animate-pulse rounded-full bg-indigo-400" />
            )}
          </div>
        )}
      </div>
    </div>
  )
}
