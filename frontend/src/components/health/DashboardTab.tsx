import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
} from 'recharts'
import type {
  HealthProfile,
  WeightLog,
  HealthExercise,
} from '../../lib/api/types'
import { computeBmi, bmiCategory, bmiColor } from './bmi'

type Props = {
  profile: HealthProfile | null | undefined
  weightLogs: WeightLog[] | undefined
  exercises: HealthExercise[] | undefined
  onGoToSettings: () => void
}

const GOAL_LABELS: Record<string, string> = {
  lose_weight: 'Lose Weight',
  maintain: 'Maintain',
  gain_muscle: 'Gain Muscle',
}

const TYPE_LABELS: Record<string, string> = {
  cardio: 'Cardio',
  strength: 'Strength',
  flexibility: 'Flexibility',
}

function formatDate(iso: string) {
  return new Date(iso).toLocaleDateString(undefined, {
    month: 'short',
    day: 'numeric',
  })
}

export default function DashboardTab({
  profile,
  weightLogs,
  exercises,
  onGoToSettings,
}: Props) {
  const latestWeight =
    weightLogs && weightLogs.length > 0
      ? weightLogs[weightLogs.length - 1].weight
      : null

  const bmi =
    latestWeight != null && profile?.height
      ? computeBmi(latestWeight, profile.height)
      : null
  const bmiCat = bmi != null ? bmiCategory(bmi) : null

  const recentExercises = exercises?.slice(0, 6) ?? []

  const sparkData =
    weightLogs
      ?.slice(-20)
      .map((w) => ({ date: formatDate(w.logged_at), weight: w.weight })) ?? []

  if (!profile && !latestWeight) {
    return (
      <div className="flex flex-col items-center gap-4 py-16 text-center">
        <p className="text-gray-400">
          Set up your profile to see your health dashboard.
        </p>
        <button
          onClick={onGoToSettings}
          className="rounded-lg bg-indigo-600 px-4 py-2 text-sm font-medium text-white hover:bg-indigo-500"
        >
          Go to Settings
        </button>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      {/* Stat cards */}
      <div className="grid grid-cols-2 gap-4 sm:grid-cols-4">
        <StatCard
          label="Current Weight"
          value={latestWeight != null ? `${latestWeight} kg` : '—'}
        />
        <StatCard
          label="BMI"
          value={
            bmi != null ? (
              <span>
                <span>{bmi.toFixed(1)}</span>
                {bmiCat && (
                  <span className={`ml-1.5 text-xs ${bmiColor(bmiCat)}`}>
                    {bmiCat}
                  </span>
                )}
              </span>
            ) : (
              '—'
            )
          }
        />
        <StatCard
          label="Goal"
          value={
            profile?.goal ? (GOAL_LABELS[profile.goal] ?? profile.goal) : '—'
          }
        />
        <StatCard
          label="Age"
          value={profile?.age != null ? `${profile.age} yrs` : '—'}
        />
      </div>

      {/* Weight sparkline */}
      {sparkData.length > 0 && (
        <div className="rounded-xl border border-gray-800 bg-gray-900 p-4">
          <h3 className="mb-3 text-sm font-medium text-gray-400">
            Weight Trend
          </h3>
          <ResponsiveContainer width="100%" height={160}>
            <LineChart data={sparkData}>
              <CartesianGrid strokeDasharray="3 3" stroke="#374151" />
              <XAxis
                dataKey="date"
                tick={{ fill: '#9ca3af', fontSize: 11 }}
                tickLine={false}
                axisLine={false}
              />
              <YAxis
                domain={['auto', 'auto']}
                tick={{ fill: '#9ca3af', fontSize: 11 }}
                tickLine={false}
                axisLine={false}
                width={36}
              />
              <Tooltip
                contentStyle={{
                  background: '#1f2937',
                  border: '1px solid #374151',
                  borderRadius: '8px',
                  color: '#f3f4f6',
                  fontSize: 12,
                }}
                labelStyle={{ color: '#9ca3af' }}
              />
              <Line
                type="monotone"
                dataKey="weight"
                stroke="#6366f1"
                strokeWidth={2}
                dot={false}
              />
            </LineChart>
          </ResponsiveContainer>
        </div>
      )}

      {/* Recent exercises */}
      <div className="rounded-xl border border-gray-800 bg-gray-900 p-4">
        <h3 className="mb-3 text-sm font-medium text-gray-400">
          Recent Exercises
        </h3>
        {recentExercises.length === 0 ? (
          <p className="text-sm text-gray-500">No exercises logged yet.</p>
        ) : (
          <ul className="divide-y divide-gray-800">
            {recentExercises.map((ex) => (
              <li key={ex.id} className="flex items-center gap-3 py-2.5">
                <span className="rounded bg-gray-800 px-2 py-0.5 text-xs text-indigo-400">
                  {TYPE_LABELS[ex.type] ?? ex.type}
                </span>
                <span className="flex-1 text-sm text-gray-100">{ex.name}</span>
                <span className="text-xs text-gray-500">
                  {ex.duration_minutes != null
                    ? `${ex.duration_minutes} min`
                    : ''}
                </span>
                <span className="text-xs text-gray-600">
                  {formatDate(ex.performed_at)}
                </span>
              </li>
            ))}
          </ul>
        )}
      </div>
    </div>
  )
}

function StatCard({ label, value }: { label: string; value: React.ReactNode }) {
  return (
    <div className="rounded-xl border border-gray-800 bg-gray-900 p-4">
      <p className="mb-1 text-xs text-gray-500">{label}</p>
      <p className="text-lg font-semibold text-gray-100">{value}</p>
    </div>
  )
}
