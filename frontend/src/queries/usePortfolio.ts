import { useQuery } from '@tanstack/react-query'
import { fetchProjects, fetchSkills } from '../lib/api/portfolio'
import { keys } from './keys'

export function useProjects() {
  return useQuery({
    queryKey: keys.portfolioProjects,
    queryFn: fetchProjects,
  })
}

export function useSkills() {
  return useQuery({
    queryKey: keys.portfolioSkills,
    queryFn: fetchSkills,
  })
}
