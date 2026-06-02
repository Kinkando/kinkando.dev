import { useState } from 'react'
import type {
  WorkoutSession,
  WorkoutSessionExercise,
  ExerciseSection,
} from '../../lib/api/types'
import { useSession } from '../../queries/useWorkout'
import { WORKOUT_TYPE_LABELS } from '../../lib/workout'

const SECTION_LABELS: Record<ExerciseSection, string> = {
  warmup: '🔥 Warm-up',
  main: '🏋️ Main',
  cooldown: '❄️ Cool-down',
}

type Props = {
  sessions: WorkoutSession[] | undefined
}

function formatDate(iso: string) {
  return new Date(iso).toLocaleDateString(undefined, {
    weekday: 'short',
    month: 'short',
    day: 'numeric',
    year: 'numeric',
  })
}

function ExerciseSummaryRow({ ex }: { ex: WorkoutSessionExercise }) {
  const done = ex.completed
  return (
    <div
      className={`flex items-baseline justify-between gap-2 rounded px-2 py-1 text-xs ${
        done ? 'text-gray-500' : 'text-gray-400'
      }`}
    >
      <span className={done ? 'line-through' : ''}>{ex.name}</span>
      <span className="shrink-0 text-gray-600">
        {[
          ex.actual_sets != null && ex.actual_reps != null
            ? `${ex.actual_sets}×${ex.actual_reps}`
            : ex.actual_sets != null
              ? `${ex.actual_sets} sets`
              : null,
          ex.actual_duration_seconds != null
            ? `${ex.actual_duration_seconds}s`
            : null,
          ex.weight_kg != null ? `${ex.weight_kg}kg` : null,
        ]
          .filter(Boolean)
          .join(' · ')}
      </span>
    </div>
  )
}

function SessionDetail({ sessionId }: { sessionId: string }) {
  const { data: session, isLoading } = useSession(sessionId)

  if (isLoading) {
    return <p className="py-4 text-center text-xs text-gray-500">Loading…</p>
  }

  if (!session) {
    return (
      <p className="py-4 text-center text-xs text-gray-500">
        Session not found.
      </p>
    )
  }

  const sections: ExerciseSection[] = ['warmup', 'main', 'cooldown']
  const bySection: Partial<Record<ExerciseSection, WorkoutSessionExercise[]>> =
    {}
  for (const ex of session.exercises) {
    ;(bySection[ex.section] ??= []).push(ex)
  }

  const completedCount = session.exercises.filter((e) => e.completed).length

  return (
    <div className="mt-3 space-y-3 border-t border-gray-800 pt-3">
      <div className="flex flex-wrap gap-3 text-xs text-gray-500">
        {session.duration_minutes != null && (
          <span>⏱ {session.duration_minutes} min</span>
        )}
        <span>
          ✓ {completedCount}/{session.exercises.length} completed
        </span>
        {session.notes && <span>📝 {session.notes}</span>}
      </div>

      {sections.map((section) => {
        const exs = bySection[section]
        if (!exs?.length) return null
        return (
          <div key={section}>
            <p className="mb-1 text-xs font-semibold tracking-wide text-gray-600 uppercase">
              {SECTION_LABELS[section]}
            </p>
            <div className="space-y-0.5">
              {exs.map((ex) => (
                <ExerciseSummaryRow key={ex.id} ex={ex} />
              ))}
            </div>
          </div>
        )
      })}

      {session.exercises.length === 0 && (
        <p className="text-xs text-gray-600">No exercises logged.</p>
      )}
    </div>
  )
}

export default function HistoryTab({ sessions }: Props) {
  const [expandedId, setExpandedId] = useState<string | null>(null)

  function toggleExpand(id: string) {
    setExpandedId((prev) => (prev === id ? null : id))
  }

  if (!sessions || sessions.length === 0) {
    return (
      <div className="rounded-xl border border-dashed border-gray-700 py-12 text-center">
        <p className="text-sm text-gray-500">No sessions logged yet.</p>
        <p className="mt-1 text-xs text-gray-600">
          Start a workout from the Today tab to see your history here.
        </p>
      </div>
    )
  }

  return (
    <div className="space-y-2">
      {sessions.map((session) => {
        const isExpanded = expandedId === session.id
        const completedCount = session.exercises.filter(
          (e) => e.completed,
        ).length

        return (
          <div
            key={session.id}
            className="rounded-xl border border-gray-800 bg-gray-900"
          >
            <button
              className="flex w-full cursor-pointer items-center gap-3 p-4 text-left"
              onClick={() => toggleExpand(session.id)}
            >
              <div className="min-w-0 flex-1">
                <div className="flex flex-wrap items-center gap-2">
                  <span className="rounded bg-indigo-900/50 px-2 py-0.5 text-xs text-indigo-400">
                    {WORKOUT_TYPE_LABELS[session.type]}
                  </span>
                  {session.completed_at && (
                    <span className="rounded bg-green-900/50 px-2 py-0.5 text-xs text-green-400">
                      Finished
                    </span>
                  )}
                  <span className="truncate text-sm font-medium text-gray-100">
                    {session.name}
                  </span>
                </div>
                <div className="mt-0.5 flex flex-wrap gap-2 text-xs text-gray-500">
                  <span>{formatDate(session.performed_at)}</span>
                  {session.duration_minutes != null && (
                    <span>⏱ {session.duration_minutes} min</span>
                  )}
                  {session.exercises.length > 0 && (
                    <span>
                      {completedCount}/{session.exercises.length} done
                    </span>
                  )}
                </div>
              </div>
              <svg
                xmlns="http://www.w3.org/2000/svg"
                className={`h-4 w-4 shrink-0 text-gray-600 transition-transform ${isExpanded ? 'rotate-180' : ''}`}
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
                strokeWidth={2}
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  d="M19 9l-7 7-7-7"
                />
              </svg>
            </button>

            {isExpanded && (
              <div className="px-4 pb-4">
                <SessionDetail sessionId={session.id} />
              </div>
            )}
          </div>
        )
      })}
    </div>
  )
}
