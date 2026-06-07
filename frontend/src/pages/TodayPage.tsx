import { useDocumentTitle } from '../hooks/useDocumentTitle'
import { todayDate } from '../lib/date'
import { useQuestOverview } from '../queries/useQuest'
import { useSchedule, useSessions } from '../queries/useWorkout'
import { useMedicines, useMedicineIntakes } from '../queries/useMedicine'
import { useFoodLogs, useWeightLogs, useSleepLogs } from '../queries/useHealth'
import QuestSummaryCard from '../components/today/QuestSummaryCard'
import WorkoutTodayCard from '../components/today/WorkoutTodayCard'
import MedicineDueCard from '../components/today/MedicineDueCard'
import CaloriesCard from '../components/today/CaloriesCard'
import WeightCard from '../components/today/WeightCard'
import SleepCard from '../components/today/SleepCard'

function greeting(): string {
  const hour = new Date().getHours()
  if (hour < 12) return 'Good morning'
  if (hour < 17) return 'Good afternoon'
  return 'Good evening'
}

/** Day-of-week (0=Sun … 6=Sat) derived from the Bangkok date string to avoid
 *  local-timezone ambiguity on midnight boundaries. */
function dayOfWeekFromDate(yyyyMmDd: string): number {
  return new Date(yyyyMmDd + 'T12:00:00').getDay()
}

export default function TodayPage() {
  useDocumentTitle('Today')
  const today = todayDate()
  const dayOfWeek = dayOfWeekFromDate(today)

  const questQuery = useQuestOverview()
  const scheduleQuery = useSchedule()
  const sessionsQuery = useSessions({ from: today, to: today })
  const medicinesQuery = useMedicines(undefined, false)
  const intakesQuery = useMedicineIntakes(undefined, today)
  const foodQuery = useFoodLogs()
  const weightQuery = useWeightLogs()
  const sleepQuery = useSleepLogs()

  const dateLabel = new Date(today + 'T12:00:00').toLocaleDateString(
    undefined,
    {
      weekday: 'long',
      month: 'long',
      day: 'numeric',
    },
  )

  return (
    <main className="mx-auto max-w-5xl px-4 py-8 sm:px-6 sm:py-12">
      {/* Header */}
      <div className="mb-8">
        <h1 className="text-2xl font-black tracking-tight text-gray-100">
          {greeting()}
        </h1>
        <p className="mt-0.5 text-sm text-gray-500">{dateLabel}</p>
      </div>

      {/* Card grid */}
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
        {/* Quest overview spans both columns */}
        <div className="sm:col-span-2">
          <QuestSummaryCard overview={questQuery.data} />
        </div>

        <WorkoutTodayCard
          schedule={scheduleQuery.data}
          sessions={sessionsQuery.data}
          dayOfWeek={dayOfWeek}
        />

        <MedicineDueCard
          medicines={medicinesQuery.data}
          intakes={intakesQuery.data}
        />

        <CaloriesCard foodLogs={foodQuery.data} today={today} />

        <WeightCard weightLogs={weightQuery.data} today={today} />

        {/* Sleep spans both columns */}
        <div className="sm:col-span-2">
          <SleepCard sleepLogs={sleepQuery.data} today={today} />
        </div>
      </div>
    </main>
  )
}
