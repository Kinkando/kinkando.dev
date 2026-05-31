import type { ChatMessage } from '../../lib/api/types'

type Props = {
  message: ChatMessage
  isStreaming?: boolean
}

export default function MessageBubble({ message, isStreaming }: Props) {
  const isUser = message.role === 'user'

  return (
    <div className={`flex ${isUser ? 'justify-end' : 'justify-start'}`}>
      <div
        className={`max-w-[80%] rounded-2xl px-4 py-3 text-sm leading-relaxed whitespace-pre-wrap ${
          isUser
            ? 'bg-indigo-600 text-white'
            : 'border border-gray-700 bg-gray-800 text-gray-100'
        }`}
      >
        {message.content}
        {isStreaming && (
          <span className="ml-1 inline-block h-4 w-2 animate-pulse rounded-sm bg-indigo-400 align-middle" />
        )}
      </div>
    </div>
  )
}
