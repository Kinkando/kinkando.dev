import { apiFetch } from '@/lib/api';
import type { Project, SkillCategory } from '@/types/portfolio';

export function getProjects(): Promise<Project[]> {
  return apiFetch('/api/v1/portfolio/projects');
}

export function getSkills(): Promise<SkillCategory[]> {
  return apiFetch('/api/v1/portfolio/skills');
}
