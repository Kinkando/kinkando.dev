import { useState } from 'react'
import type { FormEvent } from 'react'
import type { Board } from '../../lib/api/types'
import {
  useCreateBoard,
  useUpdateBoard,
  useDeleteBoard,
} from '../../queries/useKanban'

type Props = {
  boards: Board[]
  selectedId: string
  onSelect: (id: string) => void
}

export default function BoardSwitcher({ boards, selectedId, onSelect }: Props) {
  const createBoard = useCreateBoard()
  const updateBoard = useUpdateBoard()
  const deleteBoard = useDeleteBoard()

  const [creating, setCreating] = useState(false)
  const [newName, setNewName] = useState('')
  const [editingId, setEditingId] = useState<string | null>(null)
  const [editName, setEditName] = useState('')

  async function handleCreate(e: FormEvent) {
    e.preventDefault()
    if (!newName.trim()) return
    const board = await createBoard.mutateAsync({ name: newName.trim() })
    setNewName('')
    setCreating(false)
    if (board) onSelect(board.id)
  }

  async function handleRename(e: FormEvent, boardId: string) {
    e.preventDefault()
    if (!editName.trim()) return
    await updateBoard.mutateAsync({
      id: boardId,
      input: { name: editName.trim() },
    })
    setEditingId(null)
  }

  async function handleDelete(boardId: string, boardName: string) {
    if (
      !confirm(
        `Delete board "${boardName}" and all its cards? This cannot be undone.`,
      )
    )
      return
    await deleteBoard.mutateAsync(boardId)
    const remaining = boards.filter((b) => b.id !== boardId)
    if (remaining.length > 0) onSelect(remaining[0].id)
  }

  return (
    <div className="mb-6 flex flex-wrap items-center gap-2">
      {boards.map((board) => (
        <div key={board.id} className="group relative">
          {editingId === board.id ? (
            <form
              onSubmit={(e) => handleRename(e, board.id)}
              className="flex gap-1"
            >
              <input
                autoFocus
                value={editName}
                onChange={(e) => setEditName(e.target.value)}
                onKeyDown={(e) => e.key === 'Escape' && setEditingId(null)}
                className="rounded-lg border border-indigo-500 bg-gray-800 px-3 py-1.5 text-sm text-gray-100 focus:outline-none"
              />
              <button
                type="submit"
                className="rounded-lg bg-indigo-600 px-2.5 py-1.5 text-xs text-white hover:bg-indigo-500"
              >
                ✓
              </button>
              <button
                type="button"
                onClick={() => setEditingId(null)}
                className="rounded-lg bg-gray-700 px-2.5 py-1.5 text-xs text-gray-400 hover:bg-gray-600"
              >
                ✕
              </button>
            </form>
          ) : (
            <>
              <button
                onClick={() => onSelect(board.id)}
                className={`rounded-lg border px-4 py-1.5 text-sm font-medium transition-colors ${
                  board.id === selectedId
                    ? 'border-indigo-500 bg-indigo-600 text-white'
                    : 'border-gray-700 bg-gray-900 text-gray-300 hover:border-gray-500 hover:text-gray-100'
                }`}
              >
                {board.name}
              </button>
              {board.id === selectedId && (
                <div className="absolute -top-2 -right-1 hidden gap-0.5 group-hover:flex">
                  <button
                    onClick={() => {
                      setEditingId(board.id)
                      setEditName(board.name)
                    }}
                    title="Rename"
                    className="rounded bg-gray-700 px-1 py-0.5 text-xs text-gray-400 hover:text-gray-100"
                  >
                    ✎
                  </button>
                  {boards.length > 1 && (
                    <button
                      onClick={() => handleDelete(board.id, board.name)}
                      title="Delete"
                      className="rounded bg-gray-700 px-1 py-0.5 text-xs text-gray-400 hover:text-red-400"
                    >
                      ✕
                    </button>
                  )}
                </div>
              )}
            </>
          )}
        </div>
      ))}

      {creating ? (
        <form onSubmit={handleCreate} className="flex gap-1">
          <input
            autoFocus
            value={newName}
            onChange={(e) => setNewName(e.target.value)}
            onKeyDown={(e) => e.key === 'Escape' && setCreating(false)}
            placeholder="Board name"
            className="rounded-lg border border-gray-700 bg-gray-800 px-3 py-1.5 text-sm text-gray-100 placeholder-gray-500 focus:border-indigo-500 focus:outline-none"
          />
          <button
            type="submit"
            disabled={createBoard.isPending}
            className="rounded-lg bg-indigo-600 px-3 py-1.5 text-sm text-white hover:bg-indigo-500 disabled:opacity-50"
          >
            Add
          </button>
          <button
            type="button"
            onClick={() => setCreating(false)}
            className="rounded-lg bg-gray-700 px-3 py-1.5 text-sm text-gray-400 hover:bg-gray-600"
          >
            Cancel
          </button>
        </form>
      ) : (
        <button
          onClick={() => setCreating(true)}
          className="rounded-lg border border-dashed border-gray-700 px-4 py-1.5 text-sm text-gray-500 hover:border-gray-500 hover:text-gray-300"
        >
          + New Board
        </button>
      )}
    </div>
  )
}
