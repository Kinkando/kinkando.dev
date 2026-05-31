import {
  useEffect,
  useRef,
  useState,
  type FormEvent,
  type KeyboardEvent,
} from 'react'
import { SendHorizonal } from 'lucide-react'

type Props = {
  onSend: (text: string) => void
  disabled: boolean
}

const MAX_HEIGHT = 160 // px — matches roughly max-h-40

export default function ChatInput({ onSend, disabled }: Props) {
  const [value, setValue] = useState('')
  const textareaRef = useRef<HTMLTextAreaElement>(null)

  // Auto-expand height to fit content, capped at MAX_HEIGHT.
  useEffect(() => {
    const el = textareaRef.current
    if (!el) return
    el.style.height = 'auto'
    el.style.height = `${Math.min(el.scrollHeight, MAX_HEIGHT)}px`
  }, [value])

  // Restore focus when the AI finishes responding.
  useEffect(() => {
    if (!disabled) textareaRef.current?.focus()
  }, [disabled])

  function handleSubmit(e: FormEvent) {
    e.preventDefault()
    const text = value.trim()
    if (!text || disabled) return
    onSend(text)
    setValue('')
  }

  function handleKeyDown(e: KeyboardEvent<HTMLTextAreaElement>) {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      handleSubmit(e as unknown as FormEvent)
    }
  }

  return (
    <form
      onSubmit={handleSubmit}
      className="flex flex-col rounded-xl border border-gray-700 bg-gray-800 transition-colors focus-within:border-indigo-500"
    >
      <textarea
        ref={textareaRef}
        className="block w-full resize-none bg-transparent px-4 pt-3 text-sm text-gray-100 placeholder-gray-500 focus:outline-none disabled:opacity-50"
        style={{ minHeight: 48, maxHeight: MAX_HEIGHT, overflowY: 'auto' }}
        placeholder="Message the assistant… (Enter to send, Shift+Enter for new line)"
        rows={1}
        value={value}
        onChange={(e) => setValue(e.target.value)}
        onKeyDown={handleKeyDown}
        disabled={disabled}
      />
      <div className="flex justify-end px-2 pb-2">
        <button
          type="submit"
          disabled={disabled || !value.trim()}
          className="flex h-9 w-9 items-center justify-center rounded-lg bg-indigo-600 text-white transition hover:bg-indigo-500 disabled:opacity-40"
          aria-label="Send message"
        >
          <SendHorizonal size={16} />
        </button>
      </div>
    </form>
  )
}
