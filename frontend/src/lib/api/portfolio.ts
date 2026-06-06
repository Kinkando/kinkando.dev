import { apiFetch } from './client'
import type {
  PortfolioEducation,
  PortfolioExperience,
  PortfolioProfile,
  PortfolioProject,
  PortfolioSkill,
} from './types'

export function fetchProfile(): Promise<PortfolioProfile | undefined> {
  return apiFetch<PortfolioProfile>('/portfolio/profile')
}

export function fetchExperience(): Promise<PortfolioExperience[] | undefined> {
  return apiFetch<PortfolioExperience[]>('/portfolio/experience')
}

export function fetchEducation(): Promise<PortfolioEducation[] | undefined> {
  return apiFetch<PortfolioEducation[]>('/portfolio/education')
}

export function fetchProjects(): Promise<PortfolioProject[] | undefined> {
  return apiFetch<PortfolioProject[]>('/portfolio/projects')
}

export function fetchSkills(): Promise<PortfolioSkill[] | undefined> {
  return apiFetch<PortfolioSkill[]>('/portfolio/skills')
}
