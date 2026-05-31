// Portfolio
export interface Project {
  name: string;
  description: string;
  url: string;
  tags: string[];
}

export interface SkillGroup {
  category: string;
  items: string[];
}

// Users
export interface AppUser {
  id: string;
  firebase_uid: string;
  email: string;
  created_at: string;
}

// Finance
export type RecordType = 'income' | 'expense';

export interface FinanceRecord {
  id: string;
  user_id: string;
  type: RecordType;
  amount: number;
  category: string;
  note: string;
  date: string;
  created_at: string;
}

export interface CreateRecordInput {
  type: RecordType;
  amount: number;
  category: string;
  note: string;
  date: string; // YYYY-MM-DD
}

export interface CategorySummary {
  category: string;
  type: RecordType;
  total: number;
}

export interface FinanceSummary {
  month: string;
  income: number;
  expense: number;
  net: number;
  categories: CategorySummary[];
}

// Kanban
export interface KanbanBoard {
  id: string;
  user_id: string;
  created_at: string;
}

export interface KanbanColumn {
  id: string;
  board_id: string;
  name: string;
  order: number;
  created_at: string;
}

export interface KanbanCard {
  id: string;
  board_id: string;
  column_id: string;
  title: string;
  content: string;
  order: number;
  created_at: string;
}

export interface BoardResponse {
  board: KanbanBoard;
  columns: KanbanColumn[];
  cards: KanbanCard[];
}

export interface CreateCardInput {
  column_id: string;
  title: string;
  content: string;
}

export interface MoveCardInput {
  column_id: string;
  order: number;
}
