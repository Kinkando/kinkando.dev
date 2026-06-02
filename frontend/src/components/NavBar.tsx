import { useEffect, useRef, useState } from 'react'
import { Link, useLocation, useNavigate } from 'react-router-dom'
import { signOut } from 'firebase/auth'
import { auth } from '../lib/firebase'
import { useAuth } from '../auth/AuthContext'

const ICONS = {
  portfolio:
    'M17.982 18.725A7.488 7.488 0 0012 15.75a7.488 7.488 0 00-5.982 2.975m11.963 0a9 9 0 10-11.963 0m11.963 0A8.966 8.966 0 0112 21a8.966 8.966 0 01-5.982-2.275M15 9.75a3 3 0 11-6 0 3 3 0 016 0z',
  kanban:
    'M3.75 6A2.25 2.25 0 016 3.75h2.25A2.25 2.25 0 0110.5 6v2.25a2.25 2.25 0 01-2.25 2.25H6a2.25 2.25 0 01-2.25-2.25V6zM3.75 15.75A2.25 2.25 0 016 13.5h2.25a2.25 2.25 0 012.25 2.25V18a2.25 2.25 0 01-2.25 2.25H6A2.25 2.25 0 013.75 18v-2.25zM13.5 6a2.25 2.25 0 012.25-2.25H18A2.25 2.25 0 0120.25 6v2.25A2.25 2.25 0 0118 10.5h-2.25a2.25 2.25 0 01-2.25-2.25V6zM13.5 15.75a2.25 2.25 0 012.25-2.25H18a2.25 2.25 0 012.25 2.25V18A2.25 2.25 0 0118 20.25h-2.25A2.25 2.25 0 0113.5 18v-2.25z',
  archive:
    'M20.25 7.5l-.625 10.632a2.25 2.25 0 01-2.247 2.118H6.622a2.25 2.25 0 01-2.247-2.118L3.75 7.5M10 11.25h4M3.375 7.5h17.25c.621 0 1.125-.504 1.125-1.125v-1.5c0-.621-.504-1.125-1.125-1.125H3.375c-.621 0-1.125.504-1.125 1.125v1.5c0 .621.504 1.125 1.125 1.125z',
  finance:
    'M2.25 18.75a60.07 60.07 0 0115.797 2.101c.727.198 1.453-.342 1.453-1.096V18.75M3.75 4.5v.75A.75.75 0 013 6h-.75m0 0v-.375c0-.621.504-1.125 1.125-1.125H20.25M2.25 6v9m18-10.5v.75c0 .414.336.75.75.75h.75m-1.5-1.5h.375c.621 0 1.125.504 1.125 1.125v9.75c0 .621-.504 1.125-1.125 1.125h-.375m1.5-1.5H21a.75.75 0 00-.75.75v.75m0 0H3.75m0 0h-.375a1.125 1.125 0 01-1.125-1.125V15m1.5 1.5v-.75A.75.75 0 003 15h-.75M15 10.5a3 3 0 11-6 0 3 3 0 016 0zm3 0h.008v.008H18V10.5zm-12 0h.008v.008H6V10.5z',
  health:
    'M21 8.25c0-2.485-2.099-4.5-4.688-4.5-1.935 0-3.597 1.126-4.312 2.733-.715-1.607-2.377-2.733-4.313-2.733C5.1 3.75 3 5.765 3 8.25c0 7.22 9 12 9 12s9-4.78 9-12z',
  workout:
    'M5.25 5.653c0-.856.917-1.398 1.667-.986l11.54 6.347a1.125 1.125 0 010 1.972l-11.54 6.347a1.125 1.125 0 01-1.667-.986V5.653z',
  chat: 'M20.25 8.511c.884.284 1.5 1.128 1.5 2.097v4.286c0 1.136-.847 2.1-1.98 2.193-.34.027-.68.052-1.02.072v3.091l-3-3c-1.354 0-2.694-.055-4.02-.163a2.115 2.115 0 01-.825-.242m9.345-8.334a2.126 2.126 0 00-.476-.095 48.64 48.64 0 00-8.048 0c-1.131.094-1.976 1.057-1.976 2.192v4.286c0 .837.46 1.58 1.155 1.951m9.345-8.334V6.637c0-1.621-1.152-3.026-2.76-3.235A48.455 48.455 0 0011.25 3c-2.115 0-4.198.137-6.24.402-1.608.209-2.76 1.614-2.76 3.235v6.226c0 1.621 1.152 3.026 2.76 3.235.577.075 1.157.14 1.74.194V21l4.155-4.155',
  quest:
    'M11.48 3.499a.562.562 0 011.04 0l2.125 5.111a.563.563 0 00.475.345l5.518.442c.499.04.701.663.321.988l-4.204 3.602a.563.563 0 00-.182.557l1.285 5.385a.562.562 0 01-.84.61l-4.725-2.885a.563.563 0 00-.586 0L6.982 20.54a.562.562 0 01-.84-.61l1.285-5.386a.562.562 0 00-.182-.557l-4.204-3.602a.563.563 0 01.321-.988l5.518-.442a.563.563 0 00.475-.345L11.48 3.5z',
  news: 'M12 7.5h1.5m-1.5 3h1.5m-7.5 3h7.5m-7.5 3h7.5m3-9h3.375c.621 0 1.125.504 1.125 1.125V18a2.25 2.25 0 01-2.25 2.25M16.5 7.5V18a2.25 2.25 0 002.25 2.25M16.5 7.5V4.875c0-.621-.504-1.125-1.125-1.125H4.125C3.504 3.75 3 4.254 3 4.875V18a2.25 2.25 0 002.25 2.25h13.5M6 7.5h3v3H6v-3z',
  login:
    'M15.75 9V5.25A2.25 2.25 0 0013.5 3h-6a2.25 2.25 0 00-2.25 2.25v13.5A2.25 2.25 0 007.5 21h6a2.25 2.25 0 002.25-2.25V15M12 9l-3 3m0 0l3 3m-3-3h12.75',
  logout:
    'M15.75 9V5.25A2.25 2.25 0 0013.5 3h-6a2.25 2.25 0 00-2.25 2.25v13.5A2.25 2.25 0 007.5 21h6a2.25 2.25 0 002.25-2.25V15m3 0l3-3m0 0l-3-3m3 3H9',
} as const

function NavIcon({ name }: { name: keyof typeof ICONS }) {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      className="h-4 w-4 shrink-0"
      fill="none"
      viewBox="0 0 24 24"
      stroke="currentColor"
      strokeWidth={1.5}
    >
      <path strokeLinecap="round" strokeLinejoin="round" d={ICONS[name]} />
    </svg>
  )
}

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
    `flex items-center gap-1.5 ${
      isActive(path)
        ? 'text-indigo-400'
        : 'text-gray-400 hover:text-gray-100 transition-colors'
    }`

  const navLinks = (
    <>
      <Link
        to="/portfolio"
        className={linkClass('/portfolio')}
        onClick={() => setOpen(false)}
      >
        <NavIcon name="portfolio" />
        Portfolio
      </Link>
      <Link
        to="/news"
        className={linkClass('/news')}
        onClick={() => setOpen(false)}
      >
        <NavIcon name="news" />
        News
      </Link>
      {user && (
        <>
          <Link
            to="/kanban"
            className={linkClass('/kanban')}
            onClick={() => setOpen(false)}
          >
            <NavIcon name="kanban" />
            Kanban
          </Link>
          <Link
            to="/kanban/archive"
            className={linkClass('/kanban/archive')}
            onClick={() => setOpen(false)}
          >
            <NavIcon name="archive" />
            Archive
          </Link>
          <Link
            to="/finance"
            className={linkClass('/finance')}
            onClick={() => setOpen(false)}
          >
            <NavIcon name="finance" />
            Finance
          </Link>
          <Link
            to="/health"
            className={linkClass('/health')}
            onClick={() => setOpen(false)}
          >
            <NavIcon name="health" />
            Health
          </Link>
          <Link
            to="/workout"
            className={linkClass('/workout')}
            onClick={() => setOpen(false)}
          >
            <NavIcon name="workout" />
            Workout
          </Link>
          <Link
            to="/quest"
            className={linkClass('/quest')}
            onClick={() => setOpen(false)}
          >
            <NavIcon name="quest" />
            Quest
          </Link>
          <Link
            to="/chat"
            className={linkClass('/chat')}
            onClick={() => setOpen(false)}
          >
            <NavIcon name="chat" />
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
          <NavIcon name="login" />
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
        <div className="hidden items-center gap-6 text-sm xl:flex">
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
                    className="flex w-full items-center gap-2 px-4 py-2 text-left text-sm text-gray-300 transition-colors hover:bg-gray-700 hover:text-gray-100"
                  >
                    <NavIcon name="logout" />
                    Logout
                  </button>
                </div>
              )}
            </div>
          )}
        </div>

        {/* Hamburger */}
        <button
          className="flex items-center justify-center rounded-md p-1.5 text-gray-400 hover:text-gray-100 xl:hidden"
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
        <div className="flex flex-col gap-4 border-t border-gray-800 px-6 py-4 text-sm xl:hidden">
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
                className="flex items-center gap-1.5 text-left text-gray-400 transition-colors hover:text-gray-100"
              >
                <NavIcon name="logout" />
                Logout
              </button>
            </>
          )}
        </div>
      )}
    </nav>
  )
}
