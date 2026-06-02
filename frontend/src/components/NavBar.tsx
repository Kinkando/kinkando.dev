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
  finance:
    'M2.25 18.75a60.07 60.07 0 0115.797 2.101c.727.198 1.453-.342 1.453-1.096V18.75M3.75 4.5v.75A.75.75 0 013 6h-.75m0 0v-.375c0-.621.504-1.125 1.125-1.125H20.25M2.25 6v9m18-10.5v.75c0 .414.336.75.75.75h.75m-1.5-1.5h.375c.621 0 1.125.504 1.125 1.125v9.75c0 .621-.504 1.125-1.125 1.125h-.375m1.5-1.5H21a.75.75 0 00-.75.75v.75m0 0H3.75m0 0h-.375a1.125 1.125 0 01-1.125-1.125V15m1.5 1.5v-.75A.75.75 0 003 15h-.75M15 10.5a3 3 0 11-6 0 3 3 0 016 0zm3 0h.008v.008H18V10.5zm-12 0h.008v.008H6V10.5z',
  health:
    'M21 8.25c0-2.485-2.099-4.5-4.688-4.5-1.935 0-3.597 1.126-4.312 2.733-.715-1.607-2.377-2.733-4.313-2.733C5.1 3.75 3 5.765 3 8.25c0 7.22 9 12 9 12s9-4.78 9-12z',
  chat: 'M20.25 8.511c.884.284 1.5 1.128 1.5 2.097v4.286c0 1.136-.847 2.1-1.98 2.193-.34.027-.68.052-1.02.072v3.091l-3-3c-1.354 0-2.694-.055-4.02-.163a2.115 2.115 0 01-.825-.242m9.345-8.334a2.126 2.126 0 00-.476-.095 48.64 48.64 0 00-8.048 0c-1.131.094-1.976 1.057-1.976 2.192v4.286c0 .837.46 1.58 1.155 1.951m9.345-8.334V6.637c0-1.621-1.152-3.026-2.76-3.235A48.455 48.455 0 0011.25 3c-2.115 0-4.198.137-6.24.402-1.608.209-2.76 1.614-2.76 3.235v6.226c0 1.621 1.152 3.026 2.76 3.235.577.075 1.157.14 1.74.194V21l4.155-4.155',
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

type SubItem = { label: string; path: string }
type NavGroup = {
  label: string
  icon: keyof typeof ICONS
  path: string
  protected?: boolean
  subItems?: SubItem[]
}

const GROUPS: NavGroup[] = [
  { label: 'Portfolio', icon: 'portfolio', path: '/portfolio' },
  { label: 'News', icon: 'news', path: '/news' },
  {
    label: 'Kanban',
    icon: 'kanban',
    path: '/kanban',
    protected: true,
    subItems: [
      { label: 'Board', path: '/kanban' },
      { label: 'Archive', path: '/kanban/archive' },
    ],
  },
  { label: 'Finance', icon: 'finance', path: '/finance', protected: true },
  {
    label: 'Health',
    icon: 'health',
    path: '/health',
    protected: true,
    subItems: [
      { label: 'Dashboard', path: '/health' },
      { label: 'Workout', path: '/health/workout' },
      { label: 'Medicine', path: '/health/medicine' },
      { label: 'Food', path: '/health/food' },
      { label: 'Sleep', path: '/health/sleep' },
    ],
  },
  { label: 'Chat', icon: 'chat', path: '/chat', protected: true },
]

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

  async function handleLogout() {
    await signOut(auth)
    navigate('/login')
  }

  function getInitials() {
    if (user?.displayName) return user.displayName.charAt(0).toUpperCase()
    if (user?.email) return user.email.charAt(0).toUpperCase()
    return '?'
  }

  const pathname = location.pathname

  function isGroupActive(g: NavGroup) {
    return pathname === g.path || pathname.startsWith(g.path + '/')
  }

  const visibleGroups = GROUPS.filter((g) => !g.protected || user)
  const activeGroup = visibleGroups.find(isGroupActive)

  const groupLinkClass = (g: NavGroup) =>
    `flex items-center gap-1.5 ${
      isGroupActive(g)
        ? 'text-indigo-400'
        : 'text-gray-400 hover:text-gray-100 transition-colors'
    }`

  const subLinkClass = (path: string) =>
    `${
      pathname === path
        ? 'text-indigo-400'
        : 'text-gray-400 hover:text-gray-100 transition-colors'
    }`

  return (
    <nav className="border-b border-gray-800 bg-gray-900">
      {/* Main row */}
      <div className="flex items-center justify-between px-6 py-3">
        {/* Brand */}
        <Link
          to="/portfolio"
          className="flex items-center gap-2 text-lg font-bold tracking-tight text-indigo-400"
        >
          <img src="/logo.png" alt="" className="h-7 w-auto" />
          kinkando.dev
        </Link>

        {/* Desktop main links */}
        <div className="hidden items-center gap-6 text-sm xl:flex">
          {visibleGroups.map((g) => (
            <Link key={g.path} to={g.path} className={groupLinkClass(g)}>
              <NavIcon name={g.icon} />
              {g.label}
            </Link>
          ))}
          {!user && (
            <Link
              to="/login"
              className={`flex items-center gap-1.5 ${
                pathname === '/login'
                  ? 'text-indigo-400'
                  : 'text-gray-400 transition-colors hover:text-gray-100'
              }`}
            >
              <NavIcon name="login" />
              Login
            </Link>
          )}
          {user && (
            <div ref={avatarRef} className="relative">
              <button
                onClick={() => setAvatarOpen((o) => !o)}
                className="flex cursor-pointer items-center gap-2 rounded-md px-1 py-0.5 text-gray-300 transition-colors hover:text-gray-100"
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
                    className="flex w-full cursor-pointer items-center gap-2 px-4 py-2 text-left text-sm text-gray-300 transition-colors hover:bg-gray-700 hover:text-gray-100"
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
          className="flex cursor-pointer items-center justify-center rounded-md p-1.5 text-gray-400 hover:text-gray-100 xl:hidden"
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

      {/* Sub row — desktop only, shown when the active group has sub-items */}
      {activeGroup?.subItems && (
        <div className="hidden border-t border-gray-800 px-6 py-2 xl:flex">
          <div className="flex items-center gap-6 text-sm">
            {activeGroup.subItems.map((sub) => (
              <Link
                key={sub.path}
                to={sub.path}
                className={subLinkClass(sub.path)}
              >
                {sub.label}
              </Link>
            ))}
          </div>
        </div>
      )}

      {/* Mobile drawer */}
      {open && (
        <div className="flex flex-col gap-4 border-t border-gray-800 px-6 py-4 text-sm xl:hidden">
          {visibleGroups.map((g) => (
            <div key={g.path}>
              <Link
                to={g.path}
                className={groupLinkClass(g)}
                onClick={() => setOpen(false)}
              >
                <NavIcon name={g.icon} />
                {g.label}
              </Link>
              {g.subItems && (
                <div className="mt-2 flex flex-col gap-2 pl-6">
                  {g.subItems.map((sub) => (
                    <Link
                      key={sub.path}
                      to={sub.path}
                      className={subLinkClass(sub.path)}
                      onClick={() => setOpen(false)}
                    >
                      {sub.label}
                    </Link>
                  ))}
                </div>
              )}
            </div>
          ))}
          {!user && (
            <Link
              to="/login"
              className={`flex items-center gap-1.5 ${
                pathname === '/login'
                  ? 'text-indigo-400'
                  : 'text-gray-400 transition-colors hover:text-gray-100'
              }`}
              onClick={() => setOpen(false)}
            >
              <NavIcon name="login" />
              Login
            </Link>
          )}
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
                className="flex cursor-pointer items-center gap-1.5 text-left text-gray-400 transition-colors hover:text-gray-100"
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
