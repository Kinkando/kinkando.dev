import AuthForm from '../components/AuthForm'
import { useDocumentTitle } from '../hooks/useDocumentTitle'

export default function RegisterPage() {
  useDocumentTitle('Register')
  return <AuthForm mode="register" />
}
