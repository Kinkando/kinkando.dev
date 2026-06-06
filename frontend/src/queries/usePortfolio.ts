import { useQuery } from '@tanstack/react-query'
import {
  fetchEducation,
  fetchExperience,
  fetchProfile,
  fetchProjects,
  fetchSkills,
} from '../lib/api/portfolio'
import { keys } from './keys'

export function useProfile() {
  return useQuery({
    queryKey: keys.portfolioProfile,
    queryFn: fetchProfile,
  })
}

export function useExperience() {
  return useQuery({
    queryKey: keys.portfolioExperience,
    queryFn: fetchExperience,
  })
}

export function useEducation() {
  return useQuery({
    queryKey: keys.portfolioEducation,
    queryFn: fetchEducation,
  })
}

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
