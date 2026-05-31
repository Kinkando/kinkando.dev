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

export interface MonthlySummary {
  month: string;
  income: number;
  expense: number;
  net: number;
  categories: CategorySummary[];
}

export interface CategorySummary {
  category: string;
  type: RecordType;
  total: number;
}
