import { apiFetch } from './client';
import type { Project, SkillGroup } from './types';

export function getProjects(): Promise<Project[]> {
  return apiFetch<Project[]>('/portfolio/projects');
}

export function getSkills(): Promise<SkillGroup[]> {
  return apiFetch<SkillGroup[]>('/portfolio/skills');
}
