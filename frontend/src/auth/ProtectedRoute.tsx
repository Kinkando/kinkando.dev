import type { ReactNode } from 'react'
import { useLocation, Navigate } from 'react-router-dom'
import { useAuth } from './AuthContext'
import LoadingScreen from '../components/LoadingScreen'

export default function ProtectedRoute({ children }: { children: ReactNode }) {
  const { user, loading } = useAuth()
  const location = useLocation()

  if (loading) {
    return <LoadingScreen />
  }

  if (!user) {
    return (
      <Navigate
        to={`/login?redirect=${encodeURIComponent(location.pathname)}`}
        replace
      />
    )
  }

  return <>{children}</>
}
