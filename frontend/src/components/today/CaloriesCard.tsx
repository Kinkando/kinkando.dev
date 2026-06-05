import { useNavigate } from 'react-router-dom'
import type { FoodLog } from '../../lib/api/types'

type Props = {
  foodLogs: FoodLog[] | undefined
  today: string // YYYY-MM-DD (Bangkok)
}

export default function CaloriesCard({ foodLogs, today }: Props) {
  const navigate = useNavigate()

  const todayLogs =
    foodLogs?.filter((f) => f.consumed_at.slice(0, 10) === today) ?? []
  const kcal = todayLogs.reduce((sum, f) => sum + (f.calories ?? 0), 0)
  const protein = todayLogs.reduce((sum, f) => sum + (f.protein_g ?? 0), 0)
  const carbs = todayLogs.reduce((sum, f) => sum + (f.carbs_g ?? 0), 0)
  const fat = todayLogs.reduce((sum, f) => sum + (f.fat_g ?? 0), 0)

  return (
    <div
      className="cursor-pointer rounded-xl border border-gray-800 bg-gray-900 p-4 transition-colors hover:border-gray-700"
      onClick={() => navigate('/health/food')}
    >
      <p className="mb-3 text-xs font-semibold tracking-widest text-orange-600 uppercase">
        Calories Today
      </p>

      {todayLogs.length > 0 ? (
        <>
          <p className="mb-3 text-2xl font-black text-gray-100">
            {Math.round(kcal)}
            <span className="ml-1 text-sm font-normal text-gray-500">kcal</span>
          </p>
          <div className="grid grid-cols-3 gap-1.5 text-center text-xs">
            <div className="rounded bg-gray-800/60 py-1.5">
              <p className="text-gray-500">Protein</p>
              <p className="font-semibold text-blue-400">
                {Math.round(protein)}g
              </p>
            </div>
            <div className="rounded bg-gray-800/60 py-1.5">
              <p className="text-gray-500">Carbs</p>
              <p className="font-semibold text-yellow-400">
                {Math.round(carbs)}g
              </p>
            </div>
            <div className="rounded bg-gray-800/60 py-1.5">
              <p className="text-gray-500">Fat</p>
              <p className="font-semibold text-pink-400">{Math.round(fat)}g</p>
            </div>
          </div>
        </>
      ) : (
        <p className="text-sm text-gray-500">No food logged today.</p>
      )}
    </div>
  )
}
