export type User = {
  id: string
  firebase_uid: string
  email: string
  created_at: string
}

export type RecordType = 'income' | 'expense'

export type FinanceRecord = {
  id: string
  user_id: string
  type: RecordType
  amount: number
  category: string
  note: string
  date: string
  created_at: string
}

export type CreateRecordInput = {
  type: RecordType
  amount: number
  category: string
  note: string
  date: string
}

export type CategorySummary = {
  category: string
  type: RecordType
  total: number
}

export type MonthlySummary = {
  month: string
  income: number
  expense: number
  net: number
  categories: CategorySummary[]
}

export type Board = {
  id: string
  user_id: string
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
  order: number
  created_at: string
}

export type CreateCardInput = {
  column_id: string
  title: string
  content: string
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
