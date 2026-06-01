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

export type ColumnType = 'todo' | 'in_progress' | 'done' | 'custom'

export type ArchiveReason = 'completed' | 'cancelled' | 'duplicate' | 'stale'

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
  type: ColumnType
  is_system: boolean
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
  completed_at?: string
  archived_at?: string
  archive_reason?: ArchiveReason
  created_at: string
}

export type CreateBoardInput = {
  name: string
}

export type UpdateBoardInput = {
  name: string
}

export type CreateColumnInput = {
  board_id: string
  name: string
}

export type UpdateColumnInput = {
  name: string
}

export type ReorderColumnsInput = {
  column_ids: string[]
}

export type DeleteColumnInput = {
  action: 'move' | 'archive'
  target_column_id?: string
}

export type ArchiveCardInput = {
  reason?: 'cancelled' | 'duplicate' | 'stale'
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

export type ArchiveFilter = {
  reason?: 'completed' | 'general'
  month?: number
  year?: number
}

export type ChatMessage = {
  role: 'user' | 'assistant'
  content: string
}

export type ChatUsage = {
  inputTokens: number
  outputTokens: number
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
