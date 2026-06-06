import { useDocumentTitle } from '../hooks/useDocumentTitle'
import SleepTab from '../components/health/SleepTab'

export default function SleepPage() {
  useDocumentTitle('Sleep')
  return (
    <main className="mx-auto max-w-5xl px-6 py-12">
      <h1 className="mb-6 text-2xl font-bold text-gray-100">Sleep</h1>
      <SleepTab />
    </main>
  )
}
