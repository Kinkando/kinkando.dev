import AuthForm from '../components/AuthForm'
import { useDocumentTitle } from '../hooks/useDocumentTitle'

export default function LoginPage() {
  useDocumentTitle('Login')
  return <AuthForm mode="login" />
}
