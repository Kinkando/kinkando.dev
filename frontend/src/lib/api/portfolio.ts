import { apiFetch } from './client'
import type { PortfolioProject, PortfolioSkill } from './types'

export function fetchProjects(): Promise<PortfolioProject[] | undefined> {
  return apiFetch<PortfolioProject[]>('/portfolio/projects')
}

export function fetchSkills(): Promise<PortfolioSkill[] | undefined> {
  return apiFetch<PortfolioSkill[]>('/portfolio/skills')
}
