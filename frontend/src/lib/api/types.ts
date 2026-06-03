export type User = {
  id: string
  firebase_uid: string
  email: string
  line_id: string | null
  created_at: string
}

export type LineLinkCode = {
  code: string
  expires_at: string
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

export type ChatResponseMode = 'text' | 'voice'

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

// ---- Health

export type Gender = 'male' | 'female' | 'other'

export type HealthGoal = 'lose_weight' | 'maintain' | 'gain_muscle'

export type HealthProfile = {
  id: string
  user_id: string
  height: number | null // cm
  age: number | null
  gender: Gender | null
  goal: HealthGoal | null
  created_at: string
  updated_at: string
}

export type UpsertProfileInput = {
  height: number | null
  age: number | null
  gender: Gender | null
  goal: HealthGoal | null
}

export type WeightLog = {
  id: string
  weight: number // kg
  logged_at: string // ISO date
}

export type CreateWeightInput = {
  weight: number
  logged_at?: string // YYYY-MM-DD, optional
}

// ---- Food

export type MealType = 'breakfast' | 'lunch' | 'dinner' | 'snack'

export type FoodLog = {
  id: string
  user_id: string
  name: string
  meal_type: MealType
  calories: number | null
  protein_g: number | null
  carbs_g: number | null
  fat_g: number | null
  notes: string | null
  consumed_at: string // ISO date
  created_at: string
}

export type CreateFoodInput = {
  name: string
  meal_type: MealType
  calories: number | null
  protein_g: number | null
  carbs_g: number | null
  fat_g: number | null
  notes: string | null
  consumed_at: string // YYYY-MM-DD
}

export type UpdateFoodInput = CreateFoodInput

// ---- Sleep

export type SleepLog = {
  id: string
  user_id: string
  started_at: string // RFC3339
  ended_at: string // RFC3339
  duration_minutes: number
  score: number | null // 0–100 (Samsung Health)
  notes: string | null
  logged_at: string // ISO date
  created_at: string
}

export type CreateSleepInput = {
  started_at: string // RFC3339
  ended_at: string // RFC3339
  score: number | null
  notes: string | null
  logged_at?: string // YYYY-MM-DD, optional
}

export type UpdateSleepInput = CreateSleepInput

// ---- Workout

export type WorkoutType =
  | 'weight_training'
  | 'body_weight'
  | 'running'
  | 'walking'
  | 'cardio'
  | 'mobility'
  | 'custom'

export type ExerciseSection = 'warmup' | 'main' | 'cooldown'

export type WorkoutPresetExercise = {
  id: string
  preset_id: string
  section: ExerciseSection
  order_index: number
  name: string
  target_muscles: string | null
  instructions: string | null
  sets: number | null
  reps: number | null
  duration_seconds: number | null
  rest_seconds: number | null
  weight_kg: number | null
  equipment: string | null
  notes: string | null
}

export type WorkoutPreset = {
  id: string
  user_id: string
  name: string
  type: WorkoutType
  description: string | null
  exercises: WorkoutPresetExercise[]
  created_at: string
  updated_at: string
}

export type WorkoutPresetExerciseInput = {
  section: ExerciseSection
  name: string
  target_muscles: string | null
  instructions: string | null
  sets: number | null
  reps: number | null
  duration_seconds: number | null
  rest_seconds: number | null
  weight_kg: number | null
  equipment: string | null
  notes: string | null
}

export type CreatePresetInput = {
  name: string
  type: WorkoutType
  description: string | null
  exercises: WorkoutPresetExerciseInput[]
}

export type UpdatePresetInput = CreatePresetInput

export type WorkoutScheduleEntry = {
  id: string
  user_id: string
  day_of_week: number // 0=Sun … 6=Sat
  preset_id: string
  preset_name: string
  preset_type: WorkoutType
  created_at: string
}

export type SetScheduleInput = {
  entries: { day_of_week: number; preset_id: string }[]
}

export type WorkoutSessionExercise = {
  id: string
  session_id: string
  section: ExerciseSection
  order_index: number
  name: string
  target_muscles: string | null
  instructions: string | null
  target_sets: number | null
  target_reps: number | null
  target_duration_seconds: number | null
  rest_seconds: number | null
  actual_sets: number | null
  actual_reps: number | null
  actual_duration_seconds: number | null
  weight_kg: number | null
  completed: boolean
  notes: string | null
}

export type WorkoutSession = {
  id: string
  user_id: string
  preset_id: string | null
  name: string
  type: WorkoutType
  performed_at: string // ISO date
  duration_minutes: number | null
  notes: string | null
  completed_at: string | null // RFC3339, null if not yet finished
  exercises: WorkoutSessionExercise[]
  created_at: string
  updated_at: string
}

export type UpdateSessionExerciseInput = {
  actual_sets: number | null
  actual_reps: number | null
  actual_duration_seconds: number | null
  weight_kg: number | null
  completed: boolean
  notes: string | null
}

export type BulkUpdateSessionExerciseItem = {
  id: string
  actual_sets: number | null
  actual_reps: number | null
  actual_duration_seconds: number | null
  weight_kg: number | null
  completed: boolean
  notes: string | null
}

export type BulkUpdateSessionExercisesInput = {
  items: BulkUpdateSessionExerciseItem[]
}

export type AddSessionExerciseInput = {
  section: ExerciseSection
  name: string
  target_muscles: string | null
  instructions: string | null
  target_sets: number | null
  target_reps: number | null
  target_duration_seconds: number | null
  rest_seconds: number | null
}

export type UpdateSessionInput = {
  name: string
  duration_minutes: number | null
  notes: string | null
}

// ---- Medicine

export type MedicineSourceType = 'medication' | 'supplement'

export type FrequencyType = 'daily' | 'weekly' | 'as_needed' | 'custom'

export type MedicineTiming =
  | 'before_meal'
  | 'after_meal'
  | 'before_breakfast'
  | 'after_breakfast'
  | 'before_lunch'
  | 'after_lunch'
  | 'before_dinner'
  | 'after_dinner'
  | 'before_bed'
  | 'anytime'

export type IntakeStatus = 'taken' | 'skipped' | 'missed'

export type AdjustmentType = 'add' | 'remove' | 'correction'

export type Medicine = {
  id: string
  user_id: string
  name: string
  source_type: MedicineSourceType
  generic_name: string | null
  description: string | null
  stock_quantity: number
  stock_unit: string
  dosage_amount: number
  dosage_unit: string | null
  frequency_type: FrequencyType
  frequency_value: number | null
  timing: MedicineTiming | null
  start_date: string | null // ISO date
  end_date: string | null // ISO date
  low_stock_threshold: number
  note: string | null
  created_at: string
  updated_at: string
  archived_at: string | null
}

export type MedicineIntake = {
  id: string
  medicine_id: string
  user_id: string
  medicine_name: string
  taken_at: string
  quantity_taken: number
  stock_before: number
  stock_after: number
  status: IntakeStatus
  note: string | null
  created_at: string
}

export type MedicineStockAdjustment = {
  id: string
  medicine_id: string
  user_id: string
  type: AdjustmentType
  quantity: number
  stock_before: number
  stock_after: number
  reason: string | null
  created_at: string
}

export type CreateMedicineInput = {
  name: string
  source_type: MedicineSourceType
  generic_name: string | null
  description: string | null
  stock_quantity: number
  stock_unit: string
  dosage_amount: number
  dosage_unit: string | null
  frequency_type: FrequencyType
  frequency_value: number | null
  timing: MedicineTiming | null
  start_date?: string // YYYY-MM-DD, optional
  end_date?: string // YYYY-MM-DD, optional
  low_stock_threshold: number
  note: string | null
}

export type UpdateMedicineInput = CreateMedicineInput

export type TakeMedicineInput = {
  quantity_taken: number
  status?: IntakeStatus
  note?: string | null
  taken_at?: string // RFC3339, optional
  allow_negative?: boolean
}

export type AdjustStockInput = {
  type: AdjustmentType
  quantity: number
  reason?: string | null
}

export type TakeResponse = {
  intake: MedicineIntake
  medicine: Medicine
}

export type AdjustStockResponse = {
  adjustment: MedicineStockAdjustment
  medicine: Medicine
}

// ---- Quest

export type QuestType = 'daily' | 'weekly'

export type SourceType = 'manual' | 'medicine' | 'workout' | 'supplement'

export type Quest = {
  id: string
  user_id: string
  type: QuestType
  source_type: SourceType
  title: string
  description: string
  xp_reward: number
  target_count: number
  is_active: boolean
  created_at: string
  updated_at: string
}

export type DailyQuestStatus = Quest & {
  current_count: number
  completed: boolean
}

export type WeeklyQuestStatus = Quest & {
  current_count: number
  completed: boolean
}

export type XPSummary = {
  total_xp: number
  level: number
  xp_into_level: number
  xp_for_level: number
  xp_to_next: number
}

export type QuestOverview = {
  date: string
  week_start: string
  xp: XPSummary
  daily: DailyQuestStatus[]
  weekly: WeeklyQuestStatus[]
  daily_done: number
  daily_total: number
  weekly_done: number
  weekly_total: number
}

export type XPEvent = {
  id: string
  quest_id: string | null
  quest_title: string
  source: string
  period_start: string
  xp: number
  created_at: string
}

export type CreateQuestInput = {
  type: QuestType
  source_type?: SourceType
  title: string
  description: string
  xp_reward: number
  target_count: number
}

export type UpdateQuestInput = {
  source_type?: SourceType
  title: string
  description: string
  xp_reward: number
  target_count: number
  is_active: boolean
}
