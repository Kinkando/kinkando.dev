import { useEffect, useRef } from 'react'
import { useLocation } from 'react-router-dom'
import NProgress from 'nprogress'

NProgress.configure({ showSpinner: false, speed: 300, minimum: 0.08 })

export default function TopLoader() {
  const location = useLocation()
  const started = useRef(false)

  // Start progress on any internal link/button navigation intent
  useEffect(() => {
    const handleClick = (e: MouseEvent) => {
      const anchor = (e.target as Element).closest('a')
      if (!anchor) return
      const href = anchor.getAttribute('href')
      if (
        !href ||
        href.startsWith('http') ||
        href.startsWith('mailto:') ||
        href.startsWith('#')
      )
        return
      NProgress.start()
      started.current = true
    }

    // Back/forward buttons
    const handlePopState = () => {
      NProgress.start()
      started.current = true
    }

    document.addEventListener('click', handleClick)
    window.addEventListener('popstate', handlePopState)
    return () => {
      document.removeEventListener('click', handleClick)
      window.removeEventListener('popstate', handlePopState)
    }
  }, [])

  // Complete progress when route actually changes
  useEffect(() => {
    if (started.current) {
      NProgress.done()
      started.current = false
    }
  }, [location.key])

  return null
}
