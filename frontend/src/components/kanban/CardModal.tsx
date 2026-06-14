import { createPortal } from 'react-dom'
import { useEffect, useRef, useState } from 'react'
import type { FormEvent } from 'react'
import type { Attachment, Card, Column, Priority } from '../../lib/api/types'
import { PRIORITY_META } from '../../lib/kanban'
import {
  useCreateCard,
  useUpdateCard,
  useArchiveCard,
  useMoveCard,
  useUploadAttachment,
  useRemoveAttachment,
} from '../../queries/useKanban'

type Props = {
  boardId: string
  columns: Column[]
  cards?: Card[]
  columnId?: string
  initial?: Card
  onClose: () => void
}

const inputClass =
  'w-full rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100 placeholder-gray-500 focus:border-indigo-500 focus:outline-none'

export default function CardModal({
  boardId,
  columns,
  cards,
  columnId,
  initial,
  onClose,
}: Props) {
  const isEdit = !!initial
  const createCard = useCreateCard(boardId)
  const updateCard = useUpdateCard(boardId)
  const archiveCard = useArchiveCard(boardId)
  const moveCard = useMoveCard(boardId)
  const uploadAttachment = useUploadAttachment(boardId)
  const removeAttachment = useRemoveAttachment(boardId)
  const fileInputRef = useRef<HTMLInputElement>(null)
  const [uploadError, setUploadError] = useState<string | null>(null)

  // Use the latest card snapshot from the board query so attachments refresh
  // after upload/delete; fall back to the initial prop for the first render.
  const liveCard = initial
    ? ((cards ?? []).find((c) => c.id === initial.id) ?? initial)
    : null
  const attachments: Attachment[] = liveCard?.attachments ?? []

  const [title, setTitle] = useState(initial?.title ?? '')
  const [description, setDescription] = useState(initial?.description ?? '')
  const [priority, setPriority] = useState<Priority>(
    initial?.priority ?? 'none',
  )
  const [dueDate, setDueDate] = useState(initial?.due_date?.slice(0, 10) ?? '')
  const [tags, setTags] = useState<string[]>(initial?.tags ?? [])
  const [tagInput, setTagInput] = useState('')
  const [showArchiveReason, setShowArchiveReason] = useState(false)
  const [selectedColumnId, setSelectedColumnId] = useState(
    initial?.column_id ?? columnId ?? '',
  )

  const backdropRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    function onKey(e: KeyboardEvent) {
      if (e.key === 'Escape') onClose()
    }
    window.addEventListener('keydown', onKey)
    return () => window.removeEventListener('keydown', onKey)
  }, [onClose])

  async function handleFiles(files: FileList | null) {
    if (!files || !initial) return
    setUploadError(null)
    for (const file of Array.from(files)) {
      try {
        await uploadAttachment.mutateAsync({ cardId: initial.id, file })
      } catch (err) {
        setUploadError(err instanceof Error ? err.message : 'upload failed')
        break
      }
    }
    if (fileInputRef.current) fileInputRef.current.value = ''
  }

  function handleRemoveAttachment(attachmentId: string) {
    if (!initial) return
    removeAttachment.mutate({ cardId: initial.id, attachmentId })
  }

  function formatBytes(n: number) {
    if (n < 1024) return `${n} B`
    if (n < 1024 * 1024) return `${(n / 1024).toFixed(1)} KB`
    return `${(n / (1024 * 1024)).toFixed(1)} MB`
  }

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
      if (selectedColumnId && selectedColumnId !== initial!.column_id) {
        const targetCards = (cards ?? []).filter(
          (c) => c.column_id === selectedColumnId,
        )
        await moveCard.mutateAsync({
          id: initial!.id,
          input: { column_id: selectedColumnId, order: targetCards.length },
        })
      }
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

  // Determine column type for the card being edited.
  const editCardColumnType = isEdit
    ? (columns.find((c) => c.id === initial?.column_id)?.type ?? 'custom')
    : null

  function handleArchiveClick() {
    if (!isEdit) return
    if (editCardColumnType === 'done') {
      archiveCard.mutate({ id: initial!.id, input: {} }, { onSuccess: onClose })
    } else {
      setShowArchiveReason(true)
    }
  }

  function handleArchiveWithReason(
    reason: 'cancelled' | 'duplicate' | 'stale',
  ) {
    if (!isEdit) return
    archiveCard.mutate(
      { id: initial!.id, input: { reason } },
      { onSuccess: onClose },
    )
    setShowArchiveReason(false)
  }

  const isPending =
    createCard.isPending ||
    updateCard.isPending ||
    archiveCard.isPending ||
    moveCard.isPending

  return createPortal(
    <div
      ref={backdropRef}
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4"
      onMouseDown={(e) => {
        if (e.target === backdropRef.current) onClose()
      }}
    >
      <div className="max-h-[90vh] w-full max-w-md overflow-y-auto rounded-xl border border-gray-700 bg-gray-900 p-6 shadow-2xl">
        <div className="mb-5 flex items-center justify-between">
          <h2 className="text-lg font-semibold text-gray-100">
            {isEdit ? 'Edit Card' : 'Add Card'}
          </h2>
          {isEdit && !showArchiveReason && (
            <button
              type="button"
              onClick={handleArchiveClick}
              disabled={isPending}
              className="cursor-pointer text-xs text-gray-500 hover:text-amber-400 disabled:opacity-50"
              title={
                editCardColumnType === 'done'
                  ? 'Archive as completed'
                  : 'Archive card'
              }
            >
              ⊙ Archive
            </button>
          )}
        </div>

        {/* Archive reason picker (inline) */}
        {showArchiveReason && (
          <div className="mb-5 rounded-lg border border-amber-600/40 bg-gray-800 p-4">
            <p className="mb-3 text-sm font-medium text-gray-200">
              Why is this card being archived?
            </p>
            <div className="flex flex-col gap-2">
              {(
                [
                  { value: 'cancelled', label: 'Cancelled' },
                  { value: 'duplicate', label: 'Duplicate' },
                  { value: 'stale', label: 'Stale' },
                ] as const
              ).map((r) => (
                <button
                  key={r.value}
                  type="button"
                  onClick={() => handleArchiveWithReason(r.value)}
                  className="cursor-pointer rounded-lg border border-gray-700 py-2 text-sm text-gray-300 hover:border-amber-500 hover:text-amber-300"
                >
                  {r.label}
                </button>
              ))}
            </div>
            <button
              type="button"
              onClick={() => setShowArchiveReason(false)}
              className="mt-2 w-full cursor-pointer text-xs text-gray-500 hover:text-gray-400"
            >
              Cancel
            </button>
          </div>
        )}

        {!showArchiveReason && (
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

            {/* Column (edit mode only) */}
            {isEdit && columns.length > 1 && (
              <div>
                <label className="mb-1 block text-xs font-medium text-gray-400">
                  Column
                </label>
                <select
                  value={selectedColumnId}
                  onChange={(e) => setSelectedColumnId(e.target.value)}
                  className={inputClass}
                >
                  {columns.map((col) => (
                    <option key={col.id} value={col.id}>
                      {col.name}
                    </option>
                  ))}
                </select>
              </div>
            )}

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
                    className={`cursor-pointer rounded-lg px-3 py-1 text-xs font-medium transition-all ${
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
                        className="cursor-pointer leading-none text-gray-400 hover:text-gray-100"
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

            {/* Attachments — edit mode only; needs a card id to upload against. */}
            {isEdit && (
              <div>
                <label className="mb-1 block text-xs font-medium text-gray-400">
                  Attachments
                </label>
                {attachments.length > 0 && (
                  <ul className="mb-2 flex flex-col gap-1.5">
                    {attachments.map((att) => (
                      <li
                        key={att.id}
                        className="flex items-center gap-2 rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm"
                      >
                        <a
                          href={att.url}
                          target="_blank"
                          rel="noopener noreferrer"
                          className="min-w-0 flex-1 truncate text-indigo-400 hover:text-indigo-300"
                          title={att.name}
                        >
                          {att.name}
                        </a>
                        <span className="flex-shrink-0 text-xs text-gray-500">
                          {formatBytes(att.size)}
                        </span>
                        <button
                          type="button"
                          onClick={() => handleRemoveAttachment(att.id)}
                          disabled={removeAttachment.isPending}
                          className="flex-shrink-0 cursor-pointer text-xs text-gray-500 hover:text-red-400 disabled:opacity-50"
                          title="Remove attachment"
                        >
                          ✕
                        </button>
                      </li>
                    ))}
                  </ul>
                )}
                <input
                  ref={fileInputRef}
                  type="file"
                  multiple
                  onChange={(e) => handleFiles(e.target.files)}
                  className="hidden"
                />
                <button
                  type="button"
                  onClick={() => fileInputRef.current?.click()}
                  disabled={uploadAttachment.isPending}
                  className="cursor-pointer rounded-lg border border-dashed border-gray-700 px-3 py-2 text-xs text-gray-400 hover:border-indigo-500 hover:text-indigo-400 disabled:opacity-50"
                >
                  {uploadAttachment.isPending ? 'Uploading…' : '＋ Upload file'}
                </button>
                {uploadError && (
                  <p className="mt-1 text-xs text-red-400">{uploadError}</p>
                )}
              </div>
            )}

            {/* Actions */}
            <div className="flex justify-end gap-2 pt-1">
              <button
                type="submit"
                disabled={isPending || !title.trim()}
                className="cursor-pointer rounded-lg bg-indigo-600 px-4 py-2 text-sm font-medium text-white hover:bg-indigo-500 disabled:opacity-50"
              >
                {isEdit ? 'Save Changes' : 'Add Card'}
              </button>
              <button
                type="button"
                onClick={onClose}
                className="cursor-pointer rounded-lg bg-gray-800 px-4 py-2 text-sm font-medium text-gray-400 hover:bg-gray-700"
              >
                Cancel
              </button>
            </div>
          </form>
        )}
      </div>
    </div>,
    document.body,
  )
}
