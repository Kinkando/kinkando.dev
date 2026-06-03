import { Navigate } from 'react-router-dom'
import AuthForm from '../components/AuthForm'
import { useAuth } from '../auth/AuthContext'
import { useDocumentTitle } from '../hooks/useDocumentTitle'
import LoadingScreen from '../components/LoadingScreen'

export default function LoginPage() {
  useDocumentTitle('Login')
  const { user, loading } = useAuth()

  if (loading) {
    return <LoadingScreen />
  }

  if (user) return <Navigate to="/news" replace />

  return <AuthForm mode="login" />
}
