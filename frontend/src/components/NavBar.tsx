import { Link, useLocation, useNavigate } from 'react-router-dom'
import { signOut } from 'firebase/auth'
import { auth } from '../lib/firebase'
import { useAuth } from '../auth/AuthContext'

export default function NavBar() {
  const { user } = useAuth()
  const location = useLocation()
  const navigate = useNavigate()

  function isActive(path: string) {
    return location.pathname === path
  }

  async function handleLogout() {
    await signOut(auth)
    navigate('/login')
  }

  const linkClass = (path: string) =>
    isActive(path) ? 'text-indigo-400' : 'text-gray-400 hover:text-gray-100'

  return (
    <nav className="flex items-center gap-6 border-b border-gray-800 bg-gray-900 px-6 py-3">
      <Link
        to="/portfolio"
        className="text-lg font-bold tracking-tight text-indigo-400"
      >
        kinkando.dev
      </Link>
      <div className="flex items-center gap-4 text-sm">
        <Link to="/portfolio" className={linkClass('/portfolio')}>
          Portfolio
        </Link>
        {user ? (
          <>
            <Link to="/kanban" className={linkClass('/kanban')}>
              Kanban
            </Link>
            <Link to="/finance" className={linkClass('/finance')}>
              Finance
            </Link>
            <button
              onClick={handleLogout}
              className="text-gray-400 hover:text-gray-100"
            >
              Logout
            </button>
          </>
        ) : (
          <>
            <Link to="/login" className={linkClass('/login')}>
              Login
            </Link>
            <Link to="/register" className={linkClass('/register')}>
              Register
            </Link>
          </>
        )}
      </div>
    </nav>
  )
}
