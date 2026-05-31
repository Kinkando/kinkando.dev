export const keys = {
  portfolioProjects: ['portfolio', 'projects'] as const,
  portfolioSkills: ['portfolio', 'skills'] as const,
  financeRecords: (month: string) => ['finance', 'records', month] as const,
  financeSummary: (month: string) => ['finance', 'summary', month] as const,
  financeCategories: ['finance', 'categories'] as const,
  kanbanBoards: ['kanban', 'boards'] as const,
  kanbanBoard: (id: string) => ['kanban', 'board', id] as const,
  kanbanStats: (id: string) => ['kanban', 'stats', id] as const,
}
