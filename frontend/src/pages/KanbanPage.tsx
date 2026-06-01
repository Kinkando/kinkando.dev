import { useState, useEffect } from 'react'
import { useBoards, useBoard } from '../queries/useKanban'
import KanbanBoard from '../components/kanban/Board'
import BoardSwitcher from '../components/kanban/BoardSwitcher'
import { useDocumentTitle } from '../hooks/useDocumentTitle'

function Spinner() {
  return (
    <div className="flex items-center gap-3 text-gray-500">
      <div className="h-5 w-5 animate-spin rounded-full border-2 border-indigo-500 border-t-transparent" />
      Loading…
    </div>
  )
}

export default function KanbanPage() {
  useDocumentTitle('Kanban')
  const {
    data: boards,
    isLoading: boardsLoading,
    error: boardsError,
  } = useBoards()
  const [selectedBoardId, setSelectedBoardId] = useState<string>('')

  // Default to first board once loaded
  useEffect(() => {
    if (boards && boards.length > 0 && !selectedBoardId) {
      setSelectedBoardId(boards[0].id)
    }
  }, [boards, selectedBoardId])

  const effectiveId = selectedBoardId || boards?.[0]?.id || ''
  const { data: boardData, isLoading: boardLoading } = useBoard(effectiveId)

  return (
    <main className="px-6 py-12">
      <h1 className="mb-6 text-3xl font-bold text-gray-100">Kanban</h1>

      {boardsLoading ? (
        <Spinner />
      ) : boardsError ? (
        <p className="text-red-400">Failed to load boards.</p>
      ) : boards && boards.length > 0 ? (
        <>
          <BoardSwitcher
            boards={boards}
            selectedId={effectiveId}
            onSelect={setSelectedBoardId}
          />

          {boardLoading ? (
            <Spinner />
          ) : boardData ? (
            <KanbanBoard boardId={effectiveId} data={boardData} />
          ) : null}
        </>
      ) : null}
    </main>
  )
}
