export type User = {
  id: string
  firebase_uid: string
  email: string
  created_at: string
}

export type RecordType = 'income' | 'expense'

export type CategoryRef = {
  id: string
  name: string
  icon: string
  color: string
}

export type FinanceRecord = {
  id: string
  user_id: string
  type: RecordType
  amount: number
  category_id: string | null
  category: CategoryRef | null
  category_name: string
  note: string
  date: string
  created_at: string
}

export type CreateRecordInput = {
  type: RecordType
  amount: number
  category_id: string
  note: string
  date: string
}

export type CategorySummary = {
  category_id: string | null
  category: string
  type: RecordType
  total: number
  icon: string
  color: string
}

export type MonthlySummary = {
  month: string
  income: number
  expense: number
  net: number
  categories: CategorySummary[]
}

export type Category = {
  id: string
  user_id: string
  name: string
  type: RecordType
  icon: string
  color: string
  created_at: string
}

export type CreateCategoryInput = {
  name: string
  type: RecordType
  icon: string
  color: string
}

export type UpdateCategoryInput = {
  name: string
  icon: string
  color: string
}

export type Priority = 'none' | 'low' | 'medium' | 'high' | 'urgent'

export type Board = {
  id: string
  user_id: string
  name: string
  created_at: string
}

export type Column = {
  id: string
  board_id: string
  name: string
  order: number
  created_at: string
}

export type Card = {
  id: string
  board_id: string
  column_id: string
  title: string
  content: string
  description: string
  priority: Priority
  due_date?: string
  tags: string[]
  order: number
  created_at: string
}

export type CreateBoardInput = {
  name: string
}

export type UpdateBoardInput = {
  name: string
}

export type CreateCardInput = {
  board_id: string
  column_id: string
  title: string
  content: string
  description?: string
  priority?: Priority
  due_date?: string
  tags?: string[]
}

export type UpdateCardInput = {
  title?: string
  description?: string
  priority?: Priority
  due_date?: string
  tags?: string[]
}

export type MoveCardInput = {
  column_id: string
  order: number
}

export type KanbanBoard = {
  board: Board
  columns: Column[]
  cards: Card[]
}

export type BoardStats = {
  total: number
  by_column: Record<string, number>
  by_priority: Record<string, number>
  overdue: number
  no_due_date: number
}

export type PortfolioProject = {
  name: string
  description: string
  url: string
  tags: string[]
}

export type PortfolioSkill = {
  category: string
  items: string[]
}
