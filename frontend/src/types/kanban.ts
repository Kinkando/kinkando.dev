export interface Board {
  id: string;
  user_id: string;
  created_at: string;
}

export interface Column {
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

export interface CreateCardInput {
  column_id: string;
  title: string;
  content: string;
}

export interface MoveCardInput {
  column_id: string;
  order: number;
}

export interface BoardData {
  board: Board;
  columns: Column[];
  cards: KanbanCard[];
}
