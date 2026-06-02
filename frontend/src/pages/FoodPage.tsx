import { useDocumentTitle } from '../hooks/useDocumentTitle'
import { useFoodLogs } from '../queries/useHealth'
import FoodTab from '../components/health/FoodTab'

export default function FoodPage() {
  useDocumentTitle('Food')
  const foodQuery = useFoodLogs()
  return (
    <main className="mx-auto max-w-5xl px-6 py-12">
      <h1 className="mb-6 text-2xl font-bold text-gray-100">Food</h1>
      <FoodTab foodLogs={foodQuery.data} />
    </main>
  )
}
