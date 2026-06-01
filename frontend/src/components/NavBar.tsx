import { useEffect, useRef, useState } from 'react'
import { Link, useLocation, useNavigate } from 'react-router-dom'
import { signOut } from 'firebase/auth'
import { auth } from '../lib/firebase'
import { useAuth } from '../auth/AuthContext'

export default function NavBar() {
  const { user } = useAuth()
  const location = useLocation()
  const navigate = useNavigate()
  const [open, setOpen] = useState(false)
  const [avatarOpen, setAvatarOpen] = useState(false)
  const avatarRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    function handleClickOutside(e: MouseEvent) {
      if (avatarRef.current && !avatarRef.current.contains(e.target as Node)) {
        setAvatarOpen(false)
      }
    }
    document.addEventListener('mousedown', handleClickOutside)
    return () => document.removeEventListener('mousedown', handleClickOutside)
  }, [])

  function isActive(path: string) {
    return location.pathname === path
  }

  async function handleLogout() {
    await signOut(auth)
    navigate('/login')
  }

  function getInitials() {
    if (user?.displayName) return user.displayName.charAt(0).toUpperCase()
    if (user?.email) return user.email.charAt(0).toUpperCase()
    return '?'
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
      {user && (
        <>
          <Link
            to="/kanban"
            className={linkClass('/kanban')}
            onClick={() => setOpen(false)}
          >
            Kanban
          </Link>
          <Link
            to="/kanban/archive"
            className={linkClass('/kanban/archive')}
            onClick={() => setOpen(false)}
          >
            Archive
          </Link>
          <Link
            to="/finance"
            className={linkClass('/finance')}
            onClick={() => setOpen(false)}
          >
            Finance
          </Link>
          <Link
            to="/health"
            className={linkClass('/health')}
            onClick={() => setOpen(false)}
          >
            Health
          </Link>
          <Link
            to="/chat"
            className={linkClass('/chat')}
            onClick={() => setOpen(false)}
          >
            Chat
          </Link>
        </>
      )}
      {!user && (
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
          className="flex items-center gap-2 text-lg font-bold tracking-tight text-indigo-400"
        >
          <img src="/logo.png" alt="" className="h-7 w-auto" />
          kinkando.dev
        </Link>

        {/* Desktop menu */}
        <div className="hidden items-center gap-6 text-sm lg:flex">
          {navLinks}
          {user && (
            <div ref={avatarRef} className="relative">
              <button
                onClick={() => setAvatarOpen((o) => !o)}
                className="flex items-center gap-2 rounded-md px-1 py-0.5 text-gray-300 transition-colors hover:text-gray-100"
              >
                {user.photoURL ? (
                  <img
                    src={user.photoURL}
                    alt=""
                    className="h-7 w-7 rounded-full object-cover"
                  />
                ) : (
                  <span className="flex h-7 w-7 items-center justify-center rounded-full bg-indigo-600 text-xs font-semibold text-white">
                    {getInitials()}
                  </span>
                )}
                <span className="max-w-[120px] truncate text-sm">
                  {user.displayName ?? user.email}
                </span>
                <svg
                  xmlns="http://www.w3.org/2000/svg"
                  className={`h-3.5 w-3.5 transition-transform ${avatarOpen ? 'rotate-180' : ''}`}
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                  strokeWidth={2}
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    d="M19 9l-7 7-7-7"
                  />
                </svg>
              </button>
              {avatarOpen && (
                <div className="absolute top-full right-0 mt-1 w-44 rounded-md border border-gray-700 bg-gray-800 py-1 shadow-lg">
                  <button
                    onClick={() => {
                      setAvatarOpen(false)
                      handleLogout()
                    }}
                    className="w-full px-4 py-2 text-left text-sm text-gray-300 transition-colors hover:bg-gray-700 hover:text-gray-100"
                  >
                    Logout
                  </button>
                </div>
              )}
            </div>
          )}
        </div>

        {/* Hamburger */}
        <button
          className="flex items-center justify-center rounded-md p-1.5 text-gray-400 hover:text-gray-100 lg:hidden"
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
        <div className="flex flex-col gap-4 border-t border-gray-800 px-6 py-4 text-sm lg:hidden">
          {navLinks}
          {user && (
            <>
              <div className="flex items-center gap-2 border-t border-gray-800 pt-3">
                {user.photoURL ? (
                  <img
                    src={user.photoURL}
                    alt=""
                    className="h-7 w-7 rounded-full object-cover"
                  />
                ) : (
                  <span className="flex h-7 w-7 items-center justify-center rounded-full bg-indigo-600 text-xs font-semibold text-white">
                    {getInitials()}
                  </span>
                )}
                <span className="truncate text-gray-300">
                  {user.displayName ?? user.email}
                </span>
              </div>
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
          )}
        </div>
      )}
    </nav>
  )
}
