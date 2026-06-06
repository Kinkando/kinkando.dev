import { useQuery } from '@tanstack/react-query'
import { fetchNews } from '../lib/api/news'
import { keys } from './keys'

export function useNews() {
  return useQuery({
    queryKey: keys.news,
    queryFn: fetchNews,
    // The server already caches feeds ~30m; keep the client copy fresh for 5m.
    staleTime: 5 * 60 * 1000,
  })
}
