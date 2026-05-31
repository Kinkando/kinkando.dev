import { useBoard } from '../queries/useKanban'
import KanbanBoard from '../components/kanban/Board'

export default function KanbanPage() {
  const { data, isLoading, error } = useBoard()

  return (
    <main className="px-6 py-12">
      <h1 className="mb-8 text-3xl font-bold text-gray-100">Kanban</h1>
      {isLoading ? (
        <div className="flex items-center gap-3 text-gray-500">
          <div className="h-5 w-5 animate-spin rounded-full border-2 border-indigo-500 border-t-transparent" />
          Loading board…
        </div>
      ) : error ? (
        <p className="text-red-400">Failed to load board.</p>
      ) : data ? (
        <KanbanBoard data={data} />
      ) : null}
    </main>
  )
}
