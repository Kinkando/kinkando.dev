import { Link } from 'react-router-dom'
import { Home } from 'lucide-react'
import { useDocumentTitle } from '../hooks/useDocumentTitle'

export default function NotFoundPage() {
  useDocumentTitle('Not Found')

  return (
    <main className="mx-auto flex min-h-[calc(100vh-8rem)] max-w-2xl flex-col items-center justify-center px-6 py-16 text-center">
      <p className="text-7xl font-bold text-gray-700 sm:text-8xl">404</p>
      <h1 className="mt-6 text-2xl font-bold text-gray-100 sm:text-3xl">
        Page not found
      </h1>
      <p className="mt-3 max-w-md text-gray-400">
        The page you're looking for doesn't exist or has been moved.
      </p>
      <Link
        to="/"
        className="mt-8 inline-flex cursor-pointer items-center gap-2 rounded-lg bg-emerald-500 px-5 py-2.5 text-sm font-semibold text-gray-950 transition hover:bg-emerald-400"
      >
        <Home className="h-4 w-4" />
        Back to home
      </Link>
    </main>
  )
}
