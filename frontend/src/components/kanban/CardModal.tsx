import { createPortal } from 'react-dom'
import { useEffect, useRef, useState } from 'react'
import type { FormEvent } from 'react'
import type { Card, Priority } from '../../lib/api/types'
import { PRIORITY_META } from '../../lib/kanban'
import { useCreateCard, useUpdateCard } from '../../queries/useKanban'

type Props = {
  boardId: string
  columnId?: string
  initial?: Card
  onClose: () => void
}

const inputClass =
  'w-full rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100 placeholder-gray-500 focus:border-indigo-500 focus:outline-none'

export default function CardModal({
  boardId,
  columnId,
  initial,
  onClose,
}: Props) {
  const isEdit = !!initial
  const createCard = useCreateCard(boardId)
  const updateCard = useUpdateCard(boardId)

  const [title, setTitle] = useState(initial?.title ?? '')
  const [description, setDescription] = useState(initial?.description ?? '')
  const [priority, setPriority] = useState<Priority>(
    initial?.priority ?? 'none',
  )
  const [dueDate, setDueDate] = useState(initial?.due_date?.slice(0, 10) ?? '')
  const [tags, setTags] = useState<string[]>(initial?.tags ?? [])
  const [tagInput, setTagInput] = useState('')

  const backdropRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    function onKey(e: KeyboardEvent) {
      if (e.key === 'Escape') onClose()
    }
    window.addEventListener('keydown', onKey)
    return () => window.removeEventListener('keydown', onKey)
  }, [onClose])

  function addTag() {
    const t = tagInput.trim().replace(/,$/, '')
    if (t && !tags.includes(t)) setTags([...tags, t])
    setTagInput('')
  }

  function removeTag(tag: string) {
    setTags(tags.filter((t) => t !== tag))
  }

  async function handleSubmit(e: FormEvent) {
    e.preventDefault()
    if (!title.trim()) return
    if (isEdit) {
      await updateCard.mutateAsync({
        id: initial!.id,
        input: {
          title: title.trim(),
          description,
          priority,
          due_date: dueDate,
          tags,
        },
      })
    } else {
      await createCard.mutateAsync({
        board_id: boardId,
        column_id: columnId!,
        title: title.trim(),
        content: '',
        description,
        priority,
        due_date: dueDate || undefined,
        tags,
      })
    }
    onClose()
  }

  const isPending = createCard.isPending || updateCard.isPending

  return createPortal(
    <div
      ref={backdropRef}
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4"
      onMouseDown={(e) => {
        if (e.target === backdropRef.current) onClose()
      }}
    >
      <div className="w-full max-w-md rounded-xl border border-gray-700 bg-gray-900 p-6 shadow-2xl">
        <h2 className="mb-5 text-lg font-semibold text-gray-100">
          {isEdit ? 'Edit Card' : 'Add Card'}
        </h2>
        <form onSubmit={handleSubmit} className="flex flex-col gap-4">
          {/* Title */}
          <div>
            <label className="mb-1 block text-xs font-medium text-gray-400">
              Title *
            </label>
            <input
              autoFocus
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              placeholder="Card title"
              className={inputClass}
            />
          </div>

          {/* Description */}
          <div>
            <label className="mb-1 block text-xs font-medium text-gray-400">
              Description
            </label>
            <textarea
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="Add a description…"
              rows={3}
              className={inputClass + ' resize-none'}
            />
          </div>

          {/* Priority */}
          <div>
            <label className="mb-2 block text-xs font-medium text-gray-400">
              Priority
            </label>
            <div className="flex flex-wrap gap-2">
              {(
                Object.entries(PRIORITY_META) as [
                  Priority,
                  { label: string; color: string },
                ][]
              ).map(([p, meta]) => (
                <button
                  key={p}
                  type="button"
                  onClick={() => setPriority(p)}
                  className={`rounded-lg px-3 py-1 text-xs font-medium transition-all ${
                    priority === p
                      ? 'ring-2 ring-offset-1 ring-offset-gray-900'
                      : 'opacity-50 hover:opacity-80'
                  }`}
                  style={
                    {
                      backgroundColor: meta.color + '26',
                      color: meta.color,
                      '--tw-ring-color': meta.color,
                    } as React.CSSProperties
                  }
                >
                  {meta.label}
                </button>
              ))}
            </div>
          </div>

          {/* Due date */}
          <div>
            <label className="mb-1 block text-xs font-medium text-gray-400">
              Due Date
            </label>
            <input
              type="date"
              value={dueDate}
              onChange={(e) => setDueDate(e.target.value)}
              className={inputClass}
            />
          </div>

          {/* Tags */}
          <div>
            <label className="mb-1 block text-xs font-medium text-gray-400">
              Tags
            </label>
            {tags.length > 0 && (
              <div className="mb-2 flex flex-wrap gap-1.5">
                {tags.map((tag) => (
                  <span
                    key={tag}
                    className="flex items-center gap-1 rounded-full bg-gray-700 px-2.5 py-0.5 text-xs text-gray-200"
                  >
                    {tag}
                    <button
                      type="button"
                      onClick={() => removeTag(tag)}
                      className="leading-none text-gray-400 hover:text-gray-100"
                    >
                      ×
                    </button>
                  </span>
                ))}
              </div>
            )}
            <input
              value={tagInput}
              onChange={(e) => setTagInput(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === 'Enter' || e.key === ',') {
                  e.preventDefault()
                  addTag()
                }
              }}
              onBlur={addTag}
              placeholder="Type a tag and press Enter"
              className={inputClass}
            />
          </div>

          {/* Actions */}
          <div className="flex gap-2 pt-1">
            <button
              type="submit"
              disabled={isPending || !title.trim()}
              className="flex-1 rounded-lg bg-indigo-600 py-2 text-sm font-medium text-white hover:bg-indigo-500 disabled:opacity-50"
            >
              {isEdit ? 'Save Changes' : 'Add Card'}
            </button>
            <button
              type="button"
              onClick={onClose}
              className="flex-1 rounded-lg bg-gray-800 py-2 text-sm font-medium text-gray-400 hover:bg-gray-700"
            >
              Cancel
            </button>
          </div>
        </form>
      </div>
    </div>,
    document.body,
  )
}
