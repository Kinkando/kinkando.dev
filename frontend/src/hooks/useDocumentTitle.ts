import { useEffect } from 'react'

export function useDocumentTitle(title: string) {
  useEffect(() => {
    document.title = `${title} | kinkando.dev`
    return () => {
      document.title = 'kinkando.dev'
    }
  }, [title])
}
