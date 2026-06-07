import { useEffect, useRef, useState, type KeyboardEvent } from 'react'
import { Mic, SendHorizonal, Square, X } from 'lucide-react'
import { transcribeAudio } from '../../lib/api/chat'

type Props = {
  onSend: (text: string) => void
  disabled: boolean
  onError?: (msg: string) => void
}

const MAX_HEIGHT = 160 // px — matches roughly max-h-40

export default function ChatInput({ onSend, disabled, onError }: Props) {
  const [value, setValue] = useState('')
  const [recording, setRecording] = useState(false)
  const [transcribing, setTranscribing] = useState(false)
  const textareaRef = useRef<HTMLTextAreaElement>(null)
  const recorderRef = useRef<MediaRecorder | null>(null)
  const streamRef = useRef<MediaStream | null>(null)
  const chunksRef = useRef<Blob[]>([])
  const transcribeAbortRef = useRef<AbortController | null>(null)
  const cancelRecordingRef = useRef(false)

  // Auto-expand height to fit content, capped at MAX_HEIGHT.
  useEffect(() => {
    const el = textareaRef.current
    if (!el) return
    el.style.height = 'auto'
    el.style.height = `${Math.min(el.scrollHeight, MAX_HEIGHT)}px`
  }, [value])

  // Stop recording and cancel any in-flight transcription on unmount.
  useEffect(() => {
    return () => {
      cancelRecordingRef.current = true
      recorderRef.current?.stop()
      streamRef.current?.getTracks().forEach((t) => t.stop())
      transcribeAbortRef.current?.abort()
    }
  }, [])

  function handleSubmit(e: React.SubmitEvent) {
    e.preventDefault()
    const text = value.trim()
    if (!text || disabled) return
    onSend(text)
    setValue('')
  }

  function handleKeyDown(e: KeyboardEvent<HTMLTextAreaElement>) {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      handleSubmit(e as unknown as React.SubmitEvent)
    }
  }

  async function handleMicClick() {
    if (transcribing) {
      transcribeAbortRef.current?.abort()
      return
    }

    if (recording) {
      // Stop recording — onstop will handle transcription.
      recorderRef.current?.stop()
      return
    }

    try {
      const stream = await navigator.mediaDevices.getUserMedia({ audio: true })
      streamRef.current = stream
      chunksRef.current = []

      const recorder = new MediaRecorder(stream)
      recorderRef.current = recorder

      recorder.ondataavailable = (e) => {
        if (e.data.size > 0) chunksRef.current.push(e.data)
      }

      recorder.onstop = async () => {
        // Stop all mic tracks so the browser releases the indicator.
        stream.getTracks().forEach((t) => t.stop())
        streamRef.current = null
        setRecording(false)

        // User cancelled — discard audio without transcribing.
        if (cancelRecordingRef.current) {
          cancelRecordingRef.current = false
          chunksRef.current = []
          return
        }

        const blob = new Blob(chunksRef.current, { type: recorder.mimeType })
        chunksRef.current = []

        const abort = new AbortController()
        transcribeAbortRef.current = abort
        setTranscribing(true)
        try {
          const text = await transcribeAudio(blob, abort.signal)
          if (text) {
            setValue((prev) => (prev ? `${prev} ${text}` : text))
          }
        } catch (err: unknown) {
          if ((err as { name?: string }).name !== 'AbortError') {
            onError?.('Transcription failed. Please try again.')
          }
        } finally {
          transcribeAbortRef.current = null
          setTranscribing(false)
        }
      }

      recorder.start()
      setRecording(true)
    } catch {
      onError?.('Microphone access denied or unavailable.')
    }
  }

  function handleCancelRecording() {
    cancelRecordingRef.current = true
    recorderRef.current?.stop()
  }

  const micBusy = recording || transcribing
  const sendDisabled = disabled || micBusy || !value.trim()

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
      />
      <div className="flex items-center justify-between px-2 pb-2">
        {/* Mic button */}
        <div className="flex items-center gap-2 pl-1">
          <button
            type="button"
            onClick={handleMicClick}
            className={`flex h-8 w-8 cursor-pointer items-center justify-center rounded-lg transition ${
              recording
                ? 'bg-red-600 text-white hover:bg-red-500'
                : transcribing
                  ? 'text-gray-400 hover:bg-gray-700 hover:text-gray-200'
                  : 'text-gray-400 hover:bg-gray-700 hover:text-gray-200'
            } disabled:opacity-40`}
            aria-label={
              transcribing
                ? 'Cancel transcription'
                : recording
                  ? 'Stop recording'
                  : 'Start voice input'
            }
          >
            {transcribing ? (
              <X size={16} />
            ) : recording ? (
              <Square size={16} />
            ) : (
              <Mic size={16} />
            )}
          </button>
          {recording && (
            <span className="flex items-center gap-1.5 text-xs text-red-400">
              <span className="inline-block h-2 w-2 animate-pulse rounded-full bg-red-500" />
              Recording…
              <button
                type="button"
                onClick={handleCancelRecording}
                className="flex h-5 w-5 cursor-pointer items-center justify-center rounded hover:bg-red-950 disabled:opacity-40"
                aria-label="Cancel recording"
              >
                <X size={12} />
              </button>
            </span>
          )}
          {transcribing && (
            <span className="text-xs text-gray-500">Transcribing…</span>
          )}
        </div>

        {/* Send button */}
        <button
          type="submit"
          disabled={sendDisabled}
          className="flex h-9 w-9 cursor-pointer items-center justify-center rounded-lg bg-indigo-600 text-white transition hover:bg-indigo-500 disabled:opacity-40"
          aria-label="Send message"
        >
          <SendHorizonal size={16} />
        </button>
      </div>
    </form>
  )
}
