import { useState } from 'react'
import { Link, useLocation, useNavigate } from 'react-router-dom'
import { signOut } from 'firebase/auth'
import { auth } from '../lib/firebase'
import { useAuth } from '../auth/AuthContext'

export default function NavBar() {
  const { user } = useAuth()
  const location = useLocation()
  const navigate = useNavigate()
  const [open, setOpen] = useState(false)

  function isActive(path: string) {
    return location.pathname === path
  }

  async function handleLogout() {
    await signOut(auth)
    navigate('/login')
  }

  const linkClass = (path: string) =>
    isActive(path)
      ? 'text-indigo-400'
      : 'text-gray-400 hover:text-gray-100 transition-colors'

  const navLinks = (
    <>
      <Link
        to="/portfolio"
        className={linkClass('/portfolio')}
        onClick={() => setOpen(false)}
      >
        Portfolio
      </Link>
      {user ? (
        <>
          <Link
            to="/kanban"
            className={linkClass('/kanban')}
            onClick={() => setOpen(false)}
          >
            Kanban
          </Link>
          <Link
            to="/finance"
            className={linkClass('/finance')}
            onClick={() => setOpen(false)}
          >
            Finance
          </Link>
          <Link
            to="/chat"
            className={linkClass('/chat')}
            onClick={() => setOpen(false)}
          >
            Chat
          </Link>
          <button
            onClick={() => {
              setOpen(false)
              handleLogout()
            }}
            className="text-left text-gray-400 transition-colors hover:text-gray-100"
          >
            Logout
          </button>
        </>
      ) : (
        <Link
          to="/login"
          className={linkClass('/login')}
          onClick={() => setOpen(false)}
        >
          Login
        </Link>
      )}
    </>
  )

  return (
    <nav className="border-b border-gray-800 bg-gray-900">
      <div className="flex items-center justify-between px-6 py-3">
        {/* Brand */}
        <Link
          to="/portfolio"
          className="text-lg font-bold tracking-tight text-indigo-400"
        >
          kinkando.dev
        </Link>

        {/* Desktop menu */}
        <div className="hidden items-center gap-6 text-sm sm:flex">
          {navLinks}
        </div>

        {/* Hamburger */}
        <button
          className="flex items-center justify-center rounded-md p-1.5 text-gray-400 hover:text-gray-100 sm:hidden"
          onClick={() => setOpen((o) => !o)}
          aria-label="Toggle menu"
        >
          {open ? (
            <svg
              xmlns="http://www.w3.org/2000/svg"
              className="h-5 w-5"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
              strokeWidth={2}
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                d="M6 18L18 6M6 6l12 12"
              />
            </svg>
          ) : (
            <svg
              xmlns="http://www.w3.org/2000/svg"
              className="h-5 w-5"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
              strokeWidth={2}
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                d="M4 6h16M4 12h16M4 18h16"
              />
            </svg>
          )}
        </button>
      </div>

      {/* Mobile drawer */}
      {open && (
        <div className="flex flex-col gap-4 border-t border-gray-800 px-6 py-4 text-sm sm:hidden">
          {navLinks}
        </div>
      )}
    </nav>
  )
}
