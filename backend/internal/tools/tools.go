// Package tools is the single source of truth for all dashboard tool definitions.
// Each ToolDef carries the tool name, description, a zero-value input struct (for
// schema generation via reflection), and the list of required field names.
// Both the MCP server and the Gemini client import from here so names, descriptions,
// and input schemas are never declared twice.
package tools

// ToolDef describes a single callable tool.
type ToolDef struct {
	Name        string
	Description string
	// Input is a zero value of the tool's input struct. Nil means the tool takes
	// no parameters. Used by the Gemini schema generator via reflection.
	Input    any
	Required []string // field names that must be present in the model response
}

// All returns every tool definition in declaration order.
func All() []ToolDef {
	return []ToolDef{
		FinanceListRecords,
		FinanceListCategories,
		FinanceCreateCategory,
		FinanceDeleteCategory,
		FinanceCreateRecord,
		FinanceDeleteRecord,
		FinanceMonthlySummary,
		KanbanGetBoard,
		KanbanCreateCard,
		KanbanUpdateCard,
		KanbanMoveCard,
		KanbanDeleteCard,
		KanbanBoardStats,
		KanbanArchiveCard,
		KanbanUnarchiveCard,
		KanbanListArchivedCards,
		WorkoutListSessions,
		WorkoutListPresets,
		WorkoutGetPreset,
		WorkoutGetSchedule,
		WorkoutCreatePreset,
		WorkoutUpdatePreset,
		WorkoutDeletePreset,
		WorkoutStartSession,
		WorkoutUpdateSession,
		WorkoutLogExercise,
		WorkoutAddExercise,
		WorkoutBulkLogExercises,
		WorkoutFinishSession,
		FoodListLogs,
		FoodLogMeal,
		FoodUpdateMeal,
		FoodDeleteMeal,
		SleepListLogs,
		SleepLogNight,
		SleepUpdateNight,
		SleepDeleteNight,
		WeightListLogs,
		WeightCreateLog,
		WeightUpdateLog,
		MedicineList,
		MedicineTake,
		MedicineAdjustStock,
		MedicineListIntakes,
		MedicineListStockAdjustments,
		QuestGetDashboard,
		QuestListDaily,
		QuestListWeekly,
		QuestGetXPSummary,
		QuestListHistory,
	}
}

// ---- Finance ----------------------------------------------------------------

var FinanceListRecords = ToolDef{
	Name:        "finance_list_records",
	Description: "List all finance records (income/expense) for a given month.",
	Input:       ListRecordsIn{},
	Required:    []string{"month"},
}

var FinanceListCategories = ToolDef{
	Name:        "finance_list_categories",
	Description: "List all finance categories. Call before finance_create_record to get valid category names.",
}

var FinanceCreateCategory = ToolDef{
	Name:        "finance_create_category",
	Description: "Create a new finance category. Icon must be a valid lucide-react icon name.",
	Input:       CreateCategoryIn{},
	Required:    []string{"name", "type", "icon", "color"},
}

var FinanceDeleteCategory = ToolDef{
	Name:        "finance_delete_category",
	Description: "Delete a finance category by its UUID.",
	Input:       DeleteCategoryIn{},
	Required:    []string{"id"},
}

var FinanceCreateRecord = ToolDef{
	Name:        "finance_create_record",
	Description: "Create a new income or expense record. Call finance_list_categories first to get valid category names.",
	Input:       CreateRecordIn{},
	Required:    []string{"type", "amount", "category", "date"},
}

var FinanceDeleteRecord = ToolDef{
	Name:        "finance_delete_record",
	Description: "Delete a finance record by its UUID.",
	Input:       DeleteRecordIn{},
	Required:    []string{"id"},
}

var FinanceMonthlySummary = ToolDef{
	Name:        "finance_monthly_summary",
	Description: "Get a monthly summary of income, expenses, net balance, and per-category breakdown.",
	Input:       MonthlySummaryIn{},
	Required:    []string{"month"},
}

// ---- Kanban -----------------------------------------------------------------

var KanbanGetBoard = ToolDef{
	Name:        "kanban_get_board",
	Description: "Retrieve the full kanban board: metadata, columns (with type: todo/in_progress/done/custom and is_system flag), and all active (non-archived) cards. Use the returned IDs for subsequent calls.",
}

var KanbanCreateCard = ToolDef{
	Name:        "kanban_create_card",
	Description: "Create a new card in a column. Call kanban_get_board first to get column IDs.",
	Input:       CreateCardIn{},
	Required:    []string{"column_id", "title"},
}

var KanbanUpdateCard = ToolDef{
	Name:        "kanban_update_card",
	Description: "Update specific fields on a card. Only fields listed in update_fields are changed. Valid names: title, description, priority, due_date, tags.",
	Input:       UpdateCardIn{},
	Required:    []string{"card_id", "update_fields"},
}

var KanbanMoveCard = ToolDef{
	Name:        "kanban_move_card",
	Description: "Move a card to a different column and/or position.",
	Input:       MoveCardIn{},
	Required:    []string{"card_id", "column_id", "order"},
}

var KanbanDeleteCard = ToolDef{
	Name:        "kanban_delete_card",
	Description: "Delete a kanban card by its ID.",
	Input:       DeleteCardIn{},
	Required:    []string{"card_id"},
}

var KanbanBoardStats = ToolDef{
	Name:        "kanban_board_stats",
	Description: "Get board statistics: total cards, cards per column, cards per priority, overdue count, and cards without a due date.",
	Input:       BoardStatsIn{},
	Required:    []string{"board_id"},
}

// ---- Input structs ----------------------------------------------------------
// json tags: wire field names; jsonschema tags: Gemini/MCP field descriptions.

type ListRecordsIn struct {
	Month string `json:"month" jsonschema:"Month in YYYY-MM format (e.g. 2026-05)"`
}

type CreateCategoryIn struct {
	Name  string `json:"name"  jsonschema:"Category name"`
	Type  string `json:"type"  jsonschema:"Category type: income or expense"`
	Icon  string `json:"icon"  jsonschema:"Lucide icon name (e.g. ShoppingCart, Utensils, Home, Car, Briefcase, Gift, HeartPulse, Plane, Wallet, PiggyBank, Coffee, Gamepad2, Zap, TrendingUp, CircleDollarSign)"`
	Color string `json:"color" jsonschema:"Hex color code (e.g. #6366f1)"`
}

type DeleteCategoryIn struct {
	ID string `json:"id" jsonschema:"UUID of the finance category to delete"`
}

type CreateRecordIn struct {
	Type     string  `json:"type"     jsonschema:"Record type: income or expense"`
	Amount   float64 `json:"amount"   jsonschema:"Amount as a positive number"`
	Category string  `json:"category" jsonschema:"Category name — must match an existing category of the given type"`
	Note     string  `json:"note"     jsonschema:"Optional note or description"`
	Date     string  `json:"date"     jsonschema:"Date in YYYY-MM-DD format; use today's date if the user did not specify"`
}

type DeleteRecordIn struct {
	ID string `json:"id" jsonschema:"UUID of the finance record to delete"`
}

type MonthlySummaryIn struct {
	Month string `json:"month" jsonschema:"Month in YYYY-MM format (e.g. 2026-05)"`
}

type CreateCardIn struct {
	ColumnID    string   `json:"column_id"    jsonschema:"Target column ID (hex ObjectID) — get it from kanban_get_board"`
	Title       string   `json:"title"        jsonschema:"Card title (required)"`
	Content     string   `json:"content"      jsonschema:"Card body text (optional)"`
	Description string   `json:"description"  jsonschema:"Detailed description (optional)"`
	Priority    string   `json:"priority"     jsonschema:"Priority: none, low, medium, high, or urgent (default: none)"`
	DueDate     string   `json:"due_date"     jsonschema:"Due date in YYYY-MM-DD format (optional)"`
	Tags        []string `json:"tags"         jsonschema:"List of freeform tags (optional)"`
}

type UpdateCardIn struct {
	CardID       string   `json:"card_id"       jsonschema:"Card ID (hex ObjectID) to update"`
	Title        string   `json:"title"         jsonschema:"New title (omit to keep current)"`
	Description  string   `json:"description"   jsonschema:"New description (omit to keep current)"`
	Priority     string   `json:"priority"      jsonschema:"New priority: none, low, medium, high, urgent (omit to keep current)"`
	DueDate      string   `json:"due_date"      jsonschema:"New due date YYYY-MM-DD, empty string to clear, omit to keep current"`
	Tags         []string `json:"tags"          jsonschema:"New tags list (omit to keep current, send empty array to clear)"`
	UpdateFields []string `json:"update_fields" jsonschema:"Field names to apply: title, description, priority, due_date, tags"`
}

type MoveCardIn struct {
	CardID   string `json:"card_id"   jsonschema:"Card ID (hex ObjectID) to move"`
	ColumnID string `json:"column_id" jsonschema:"Target column ID (hex ObjectID)"`
	Order    int    `json:"order"     jsonschema:"New 0-indexed position within the target column"`
}

type DeleteCardIn struct {
	CardID string `json:"card_id" jsonschema:"Card ID (hex ObjectID) to delete"`
}

type BoardStatsIn struct {
	BoardID string `json:"board_id" jsonschema:"Board ID (hex ObjectID) — get it from kanban_get_board"`
}

var KanbanArchiveCard = ToolDef{
	Name:        "kanban_archive_card",
	Description: "Archive a kanban card. When the card is in a Done column the reason is automatically set to 'completed'; for other columns supply a reason: cancelled, duplicate, or stale.",
	Input:       ArchiveCardIn{},
	Required:    []string{"card_id"},
}

var KanbanUnarchiveCard = ToolDef{
	Name:        "kanban_unarchive_card",
	Description: "Restore an archived kanban card back to its original column.",
	Input:       UnarchiveCardIn{},
	Required:    []string{"card_id"},
}

var KanbanListArchivedCards = ToolDef{
	Name:        "kanban_list_archived_cards",
	Description: "List archived kanban cards for a board. Use reason='completed' for the completed archive, reason='general' for cancelled/duplicate/stale cards, or omit for all archived. Optionally filter by year and month.",
	Input:       ListArchivedCardsIn{},
	Required:    []string{"board_id"},
}

type ArchiveCardIn struct {
	CardID string `json:"card_id" jsonschema:"Card ID (hex ObjectID) to archive"`
	Reason string `json:"reason"  jsonschema:"Archive reason: cancelled, duplicate, or stale. Omit when archiving from a Done column — the server assigns 'completed' automatically."`
}

type UnarchiveCardIn struct {
	CardID string `json:"card_id" jsonschema:"Card ID (hex ObjectID) to restore from the archive"`
}

type ListArchivedCardsIn struct {
	BoardID string `json:"board_id" jsonschema:"Board ID (hex ObjectID) — get from kanban_get_board"`
	Reason  string `json:"reason"   jsonschema:"Filter: 'completed' for done-archive, 'general' for all non-completed, or empty for all archived cards"`
	Month   int    `json:"month"    jsonschema:"Filter by month (1–12); 0 or omit for all months"`
	Year    int    `json:"year"     jsonschema:"Filter by year (e.g. 2026); 0 or omit for all years"`
}

// ---- Workout ----------------------------------------------------------------

var WorkoutListSessions = ToolDef{
	Name:        "workout_list_sessions",
	Description: "List workout sessions in a date range. Returns sessions with their exercises and logged actuals. Defaults to the last 30 days if dates are omitted.",
	Input:       WorkoutListSessionsIn{},
}

var WorkoutListPresets = ToolDef{
	Name:        "workout_list_presets",
	Description: "List all saved workout preset templates. Returns preset name, type, description, and exercise count. Call before workout_start_session to get a valid preset name.",
}

var WorkoutGetPreset = ToolDef{
	Name:        "workout_get_preset",
	Description: "Get a single workout preset by ID with its full exercise list (sets, reps, duration, rest, target muscles, equipment, notes). Call workout_list_presets first to get the preset ID.",
	Input:       WorkoutGetPresetIn{},
	Required:    []string{"preset_id"},
}

var WorkoutGetSchedule = ToolDef{
	Name:        "workout_get_schedule",
	Description: "Get the weekly workout schedule: which preset template is assigned to each day of the week (0=Sun … 6=Sat). Days without an entry are rest days.",
}

var WorkoutStartSession = ToolDef{
	Name:        "workout_start_session",
	Description: "Start a new workout session. Provide preset_name to copy a saved template (call workout_list_presets first), or provide type for a quick start with no exercises. Returns the new session with its exercise list.",
	Input:       WorkoutStartSessionIn{},
}

var WorkoutUpdateSession = ToolDef{
	Name:        "workout_update_session",
	Description: "Update a session's name, duration, and/or notes. Provide the current session name if you only want to change duration or notes.",
	Input:       WorkoutUpdateSessionIn{},
	Required:    []string{"session_id", "name"},
}

var WorkoutLogExercise = ToolDef{
	Name:        "workout_log_exercise",
	Description: "Log actual performance for a session exercise: sets, reps, duration, weight, and completion status. Pass 0 for any numeric field you don't want to update.",
	Input:       WorkoutLogExerciseIn{},
	Required:    []string{"session_id", "exercise_id"},
}

var WorkoutAddExercise = ToolDef{
	Name:        "workout_add_exercise",
	Description: "Add a new exercise to an existing session. Useful for quick-start sessions that begin with no exercises.",
	Input:       WorkoutAddExerciseIn{},
	Required:    []string{"session_id", "name"},
}

var WorkoutCreatePreset = ToolDef{
	Name:        "workout_create_preset",
	Description: "Create a new workout preset template with a list of exercises. Each exercise belongs to a section (warmup, main, or cooldown) and may have sets, reps, duration, rest, weight, and notes.",
	Input:       WorkoutCreatePresetIn{},
	Required:    []string{"name", "type"},
}

var WorkoutUpdatePreset = ToolDef{
	Name:        "workout_update_preset",
	Description: "Replace all fields of an existing workout preset (name, type, description, and full exercise list). Call workout_list_presets first to get the preset ID.",
	Input:       WorkoutUpdatePresetIn{},
	Required:    []string{"preset_id", "name", "type"},
}

var WorkoutDeletePreset = ToolDef{
	Name:        "workout_delete_preset",
	Description: "Delete a workout preset by its UUID. Call workout_list_presets first to get the preset ID.",
	Input:       WorkoutDeletePresetIn{},
	Required:    []string{"preset_id"},
}

var WorkoutBulkLogExercises = ToolDef{
	Name:        "workout_bulk_log_exercises",
	Description: "Log actual performance for multiple session exercises in one call. Use instead of repeated workout_log_exercise calls when saving all exercises at once.",
	Input:       WorkoutBulkLogExercisesIn{},
	Required:    []string{"session_id", "items"},
}

var WorkoutFinishSession = ToolDef{
	Name:        "workout_finish_session",
	Description: "Mark a workout session as completed. Sets the completion timestamp and finalizes the session — no further edits are allowed after finishing. Returns the finished session.",
	Input:       WorkoutFinishSessionIn{},
	Required:    []string{"session_id"},
}

type WorkoutListSessionsIn struct {
	From string `json:"from" jsonschema:"Start date YYYY-MM-DD (default: 30 days ago)"`
	To   string `json:"to"   jsonschema:"End date YYYY-MM-DD (default: today)"`
}

type WorkoutStartSessionIn struct {
	PresetName string `json:"preset_name" jsonschema:"Preset template name to start from (case-insensitive match). Call workout_list_presets to get available names. If omitted, type is required."`
	Type       string `json:"type"        jsonschema:"Workout type for a quick start (no exercises): weight_training, body_weight, running, walking, cardio, mobility, or custom. Required when preset_name is not provided."`
	Date       string `json:"date"        jsonschema:"Session date YYYY-MM-DD (default: today)"`
	Name       string `json:"name"        jsonschema:"Optional session name override"`
}

type WorkoutUpdateSessionIn struct {
	SessionID       string `json:"session_id"       jsonschema:"UUID of the session to update"`
	Name            string `json:"name"             jsonschema:"Session name — use the existing name if you only want to change duration or notes"`
	DurationMinutes int    `json:"duration_minutes" jsonschema:"Workout duration in minutes (0 to clear)"`
	Notes           string `json:"notes"            jsonschema:"Session notes (empty string to clear)"`
}

type WorkoutLogExerciseIn struct {
	SessionID             string  `json:"session_id"              jsonschema:"UUID of the workout session (from workout_list_sessions or workout_start_session)"`
	ExerciseID            string  `json:"exercise_id"             jsonschema:"UUID of the session exercise to log"`
	ActualSets            int     `json:"actual_sets"             jsonschema:"Sets completed (0 = not logging this field)"`
	ActualReps            int     `json:"actual_reps"             jsonschema:"Reps completed (0 = not logging this field)"`
	ActualDurationSeconds int     `json:"actual_duration_seconds" jsonschema:"Duration in seconds (0 = not logging this field)"`
	WeightKg              float64 `json:"weight_kg"               jsonschema:"Weight used in kg (0 = not logging this field)"`
	Completed             bool    `json:"completed"               jsonschema:"Mark exercise as completed"`
	Notes                 string  `json:"notes"                   jsonschema:"Optional notes for this exercise"`
}

type WorkoutAddExerciseIn struct {
	SessionID             string `json:"session_id"              jsonschema:"UUID of the session to add the exercise to"`
	Name                  string `json:"name"                    jsonschema:"Exercise name (required)"`
	Section               string `json:"section"                 jsonschema:"Exercise section: warmup, main, or cooldown (default: main)"`
	TargetSets            int    `json:"target_sets"             jsonschema:"Target number of sets (0 = no target)"`
	TargetReps            int    `json:"target_reps"             jsonschema:"Target reps per set (0 = no target)"`
	TargetDurationSeconds int    `json:"target_duration_seconds" jsonschema:"Target duration in seconds (0 = no target)"`
	RestSeconds           int    `json:"rest_seconds"            jsonschema:"Rest between sets in seconds (0 = no target)"`
}

type WorkoutGetPresetIn struct {
	PresetID string `json:"preset_id" jsonschema:"UUID of the preset to retrieve — get from workout_list_presets"`
}

type WorkoutPresetExerciseIn struct {
	Section         string  `json:"section"          jsonschema:"Exercise section: warmup, main, or cooldown (default: main)"`
	Name            string  `json:"name"             jsonschema:"Exercise name (required)"`
	TargetMuscles   string  `json:"target_muscles"   jsonschema:"Target muscle groups (optional)"`
	Instructions    string  `json:"instructions"     jsonschema:"Exercise instructions (optional)"`
	Sets            int     `json:"sets"             jsonschema:"Number of sets (0 = no target)"`
	Reps            int     `json:"reps"             jsonschema:"Reps per set (0 = no target)"`
	DurationSeconds int     `json:"duration_seconds" jsonschema:"Duration in seconds (0 = no target)"`
	RestSeconds     int     `json:"rest_seconds"     jsonschema:"Rest between sets in seconds (0 = no target)"`
	WeightKg        float64 `json:"weight_kg"        jsonschema:"Weight in kg (0 = no target)"`
	Equipment       string  `json:"equipment"        jsonschema:"Equipment needed (optional)"`
	Notes           string  `json:"notes"            jsonschema:"Additional notes (optional)"`
}

type WorkoutCreatePresetIn struct {
	Name        string                    `json:"name"        jsonschema:"Preset name (required)"`
	Type        string                    `json:"type"        jsonschema:"Workout type: weight_training, body_weight, running, walking, cardio, mobility, or custom"`
	Description string                    `json:"description" jsonschema:"Optional description"`
	Exercises   []WorkoutPresetExerciseIn `json:"exercises"   jsonschema:"Ordered list of exercises in the preset"`
}

type WorkoutUpdatePresetIn struct {
	PresetID    string                    `json:"preset_id"   jsonschema:"UUID of the preset to update — get from workout_list_presets"`
	Name        string                    `json:"name"        jsonschema:"Preset name (required)"`
	Type        string                    `json:"type"        jsonschema:"Workout type: weight_training, body_weight, running, walking, cardio, mobility, or custom"`
	Description string                    `json:"description" jsonschema:"Optional description"`
	Exercises   []WorkoutPresetExerciseIn `json:"exercises"   jsonschema:"Full exercise list to replace existing exercises"`
}

type WorkoutDeletePresetIn struct {
	PresetID string `json:"preset_id" jsonschema:"UUID of the preset to delete — get from workout_list_presets"`
}

type WorkoutBulkLogExerciseItem struct {
	ExerciseID            string  `json:"exercise_id"             jsonschema:"UUID of the session exercise to log"`
	ActualSets            int     `json:"actual_sets"             jsonschema:"Sets completed (0 = not logging this field)"`
	ActualReps            int     `json:"actual_reps"             jsonschema:"Reps completed (0 = not logging this field)"`
	ActualDurationSeconds int     `json:"actual_duration_seconds" jsonschema:"Duration in seconds (0 = not logging this field)"`
	WeightKg              float64 `json:"weight_kg"               jsonschema:"Weight used in kg (0 = not logging this field)"`
	Completed             bool    `json:"completed"               jsonschema:"Mark exercise as completed"`
	Notes                 string  `json:"notes"                   jsonschema:"Optional notes for this exercise"`
}

type WorkoutBulkLogExercisesIn struct {
	SessionID string                       `json:"session_id" jsonschema:"UUID of the workout session"`
	Items     []WorkoutBulkLogExerciseItem `json:"items"      jsonschema:"List of exercises to update; must include all exercises you want to save"`
}

type WorkoutFinishSessionIn struct {
	SessionID string `json:"session_id" jsonschema:"UUID of the workout session to finish"`
}

// ---- Food -------------------------------------------------------------------

var FoodListLogs = ToolDef{
	Name:        "food_list_logs",
	Description: "List food/nutrition logs in a date range (consumed_at). Defaults to the last 30 days if dates are omitted.",
	Input:       FoodListLogsIn{},
}

var FoodLogMeal = ToolDef{
	Name:        "food_log_meal",
	Description: "Log a meal or snack with nutritional details (calories, protein, carbs, fat). Returns the created food log entry.",
	Input:       FoodLogMealIn{},
	Required:    []string{"name", "meal_type"},
}

var FoodUpdateMeal = ToolDef{
	Name:        "food_update_meal",
	Description: "Update an existing food log entry. Call food_list_logs first to get the log ID.",
	Input:       FoodUpdateMealIn{},
	Required:    []string{"log_id", "name", "meal_type"},
}

var FoodDeleteMeal = ToolDef{
	Name:        "food_delete_meal",
	Description: "Delete a food log entry by its UUID. Call food_list_logs first to get the log ID.",
	Input:       FoodDeleteMealIn{},
	Required:    []string{"log_id"},
}

type FoodListLogsIn struct {
	From string `json:"from" jsonschema:"Start date YYYY-MM-DD (default: 30 days ago)"`
	To   string `json:"to"   jsonschema:"End date YYYY-MM-DD (default: today)"`
}

type FoodLogMealIn struct {
	Name       string  `json:"name"        jsonschema:"Food or meal name (required)"`
	MealType   string  `json:"meal_type"   jsonschema:"Meal type: breakfast, lunch, dinner, or snack (required)"`
	Calories   int     `json:"calories"    jsonschema:"Calories (0 = not logging)"`
	ProteinG   float64 `json:"protein_g"   jsonschema:"Protein in grams (0 = not logging)"`
	CarbsG     float64 `json:"carbs_g"     jsonschema:"Carbohydrates in grams (0 = not logging)"`
	FatG       float64 `json:"fat_g"       jsonschema:"Fat in grams (0 = not logging)"`
	Notes      string  `json:"notes"       jsonschema:"Optional notes"`
	ConsumedAt string  `json:"consumed_at" jsonschema:"Date YYYY-MM-DD (default: today)"`
}

type FoodUpdateMealIn struct {
	LogID      string  `json:"log_id"      jsonschema:"UUID of the food log to update — get from food_list_logs"`
	Name       string  `json:"name"        jsonschema:"Food or meal name (required)"`
	MealType   string  `json:"meal_type"   jsonschema:"Meal type: breakfast, lunch, dinner, or snack (required)"`
	Calories   int     `json:"calories"    jsonschema:"Calories (0 = clear)"`
	ProteinG   float64 `json:"protein_g"   jsonschema:"Protein in grams (0 = clear)"`
	CarbsG     float64 `json:"carbs_g"     jsonschema:"Carbohydrates in grams (0 = clear)"`
	FatG       float64 `json:"fat_g"       jsonschema:"Fat in grams (0 = clear)"`
	Notes      string  `json:"notes"       jsonschema:"Notes (empty string to clear)"`
	ConsumedAt string  `json:"consumed_at" jsonschema:"Date YYYY-MM-DD (default: today)"`
}

type FoodDeleteMealIn struct {
	LogID string `json:"log_id" jsonschema:"UUID of the food log to delete — get from food_list_logs"`
}

// ---- Sleep ------------------------------------------------------------------

var SleepListLogs = ToolDef{
	Name:        "sleep_list_logs",
	Description: "List sleep logs in a date range (logged_at). Defaults to the last 30 days if dates are omitted. Each entry includes duration, score (0–100), and bedtime/wake timestamps.",
	Input:       SleepListLogsIn{},
}

var SleepLogNight = ToolDef{
	Name:        "sleep_log_night",
	Description: "Log a night of sleep with bedtime, wake time, Samsung Health sleep score (0–100), and optional notes. One entry per day is enforced — logging the same date again updates the existing entry.",
	Input:       SleepLogNightIn{},
	Required:    []string{"started_at", "ended_at"},
}

var SleepUpdateNight = ToolDef{
	Name:        "sleep_update_night",
	Description: "Update an existing sleep log entry. Call sleep_list_logs first to get the log ID.",
	Input:       SleepUpdateNightIn{},
	Required:    []string{"log_id", "started_at", "ended_at"},
}

var SleepDeleteNight = ToolDef{
	Name:        "sleep_delete_night",
	Description: "Delete a sleep log entry by its UUID. Call sleep_list_logs first to get the log ID.",
	Input:       SleepDeleteNightIn{},
	Required:    []string{"log_id"},
}

type SleepListLogsIn struct {
	From string `json:"from" jsonschema:"Start date YYYY-MM-DD (default: 30 days ago)"`
	To   string `json:"to"   jsonschema:"End date YYYY-MM-DD (default: today)"`
}

type SleepLogNightIn struct {
	StartedAt string `json:"started_at" jsonschema:"Bedtime in RFC3339 format, e.g. 2026-06-01T22:30:00+07:00 (required)"`
	EndedAt   string `json:"ended_at"   jsonschema:"Wake time in RFC3339 format, e.g. 2026-06-02T06:45:00+07:00 (required)"`
	Score     int    `json:"score"      jsonschema:"Samsung Health sleep score 0–100 (0 = not logging)"`
	Notes     string `json:"notes"      jsonschema:"Optional notes"`
	LoggedAt  string `json:"logged_at"  jsonschema:"Night-of date YYYY-MM-DD (defaults to started_at date); one log per day is enforced — same date updates the existing entry"`
}

type SleepUpdateNightIn struct {
	LogID     string `json:"log_id"     jsonschema:"UUID of the sleep log to update — get from sleep_list_logs"`
	StartedAt string `json:"started_at" jsonschema:"Bedtime in RFC3339 format (required)"`
	EndedAt   string `json:"ended_at"   jsonschema:"Wake time in RFC3339 format (required)"`
	Score     int    `json:"score"      jsonschema:"Samsung Health sleep score 0–100 (0 = clear)"`
	Notes     string `json:"notes"      jsonschema:"Notes (empty string to clear)"`
	LoggedAt  string `json:"logged_at"  jsonschema:"Night-of date YYYY-MM-DD (defaults to started_at date)"`
}

type SleepDeleteNightIn struct {
	LogID string `json:"log_id" jsonschema:"UUID of the sleep log to delete — get from sleep_list_logs"`
}

// ---- Weight -----------------------------------------------------------------

var WeightListLogs = ToolDef{
	Name:        "weight_list_logs",
	Description: "List all body weight logs sorted by date ascending. Returns id, weight (kg), and logged_at date. Call before weight_update_log to get the log ID.",
}

var WeightCreateLog = ToolDef{
	Name:        "weight_create_log",
	Description: "Log a body weight measurement in kg. One entry per day is enforced — logging the same date twice will error.",
	Input:       WeightCreateLogIn{},
	Required:    []string{"weight"},
}

var WeightUpdateLog = ToolDef{
	Name:        "weight_update_log",
	Description: "Update an existing weight log entry. Call weight_list_logs first to get the log ID.",
	Input:       WeightUpdateLogIn{},
	Required:    []string{"log_id", "weight"},
}

type WeightCreateLogIn struct {
	Weight   float64 `json:"weight"    jsonschema:"Body weight in kg (required, must be > 0)"`
	LoggedAt string  `json:"logged_at" jsonschema:"Date YYYY-MM-DD (default: today); one log per day is enforced"`
}

type WeightUpdateLogIn struct {
	LogID    string  `json:"log_id"    jsonschema:"UUID of the weight log to update — get from weight_list_logs"`
	Weight   float64 `json:"weight"    jsonschema:"New body weight in kg (required, must be > 0)"`
	LoggedAt string  `json:"logged_at" jsonschema:"New date YYYY-MM-DD (default: today)"`
}

// ---- Medicine ---------------------------------------------------------------

var MedicineList = ToolDef{
	Name:        "medicine_list",
	Description: "List medicines with their current stock, dosage, frequency, and timing. Call before medicine_take or medicine_adjust_stock to get valid medicine names.",
	Input:       MedicineListIn{},
}

var MedicineTake = ToolDef{
	Name:        "medicine_take",
	Description: "Record taking a medicine. Finds the medicine by name, decrements stock by quantity_taken, and creates an intake log. Returns the intake log and updated medicine. Set allow_negative to true to proceed even if stock would go below zero.",
	Input:       MedicineTakeIn{},
	Required:    []string{"medicine_name", "quantity_taken"},
}

var MedicineAdjustStock = ToolDef{
	Name:        "medicine_adjust_stock",
	Description: "Adjust a medicine's stock: add (restock), remove (discard), or correction (set exact amount). Finds medicine by name. Returns the adjustment log and updated medicine.",
	Input:       MedicineAdjustStockIn{},
	Required:    []string{"medicine_name", "type", "quantity"},
}

var MedicineListIntakes = ToolDef{
	Name:        "medicine_list_intakes",
	Description: "List recent medicine intake logs, optionally filtered to a specific date. Returns up to 50 entries sorted newest first.",
	Input:       MedicineListIntakesIn{},
}

var MedicineListStockAdjustments = ToolDef{
	Name:        "medicine_list_stock_adjustments",
	Description: "List recent medicine stock adjustment logs (restocks, removals, corrections), optionally filtered to a specific date. Returns up to 50 entries sorted newest first.",
	Input:       MedicineListStockAdjustmentsIn{},
}

type MedicineListIn struct {
	IncludeArchived bool   `json:"include_archived" jsonschema:"Include archived medicines (default: false)"`
	SourceType      string `json:"source_type"      jsonschema:"Filter by type: medication or supplement (optional; omit for all)" validate:"omitempty,oneof=medication supplement"`
}

type MedicineTakeIn struct {
	MedicineName  string  `json:"medicine_name"  jsonschema:"Medicine name (case-insensitive) — call medicine_list first to confirm the name"`
	QuantityTaken float64 `json:"quantity_taken" jsonschema:"Amount taken (supports decimals like 0.5); defaults to the medicine's dosage_amount if omitted"`
	Note          string  `json:"note"           jsonschema:"Optional note (e.g. 'taken with food')"`
	AllowNegative bool    `json:"allow_negative" jsonschema:"Set to true to record the intake even if stock would go below zero"`
}

type MedicineAdjustStockIn struct {
	MedicineName string  `json:"medicine_name" jsonschema:"Medicine name (case-insensitive) — call medicine_list first to confirm the name"`
	Type         string  `json:"type"          jsonschema:"Adjustment type: add (restock), remove (discard/waste), or correction (set exact amount)"`
	Quantity     float64 `json:"quantity"      jsonschema:"Quantity to add/remove, or the new exact stock for correction (must be > 0)"`
	Reason       string  `json:"reason"        jsonschema:"Optional reason (e.g. 'bought new pack', 'expired tablets')"`
}

type MedicineListIntakesIn struct {
	Date       string `json:"date"        jsonschema:"Filter to a specific date YYYY-MM-DD (optional; omit for recent 50)"`
	SourceType string `json:"source_type" jsonschema:"Filter by type: medication or supplement (optional; omit for all)" validate:"omitempty,oneof=medication supplement"`
	Limit      int    `json:"limit"       jsonschema:"Max number of results (default: 50)"`
}

type MedicineListStockAdjustmentsIn struct {
	Date       string `json:"date"        jsonschema:"Filter to a specific date YYYY-MM-DD (optional; omit for recent 50)"`
	SourceType string `json:"source_type" jsonschema:"Filter by type: medication or supplement (optional; omit for all)" validate:"omitempty,oneof=medication supplement"`
	Limit      int    `json:"limit"       jsonschema:"Max number of results (default: 50)"`
}

// ---- Quest ------------------------------------------------------------------

var QuestGetDashboard = ToolDef{
	Name:        "quest_get_dashboard",
	Description: "Get the full quest dashboard for today: date, week start, XP summary (total XP, level, progress), daily and weekly quest statuses with current progress and completion, and done/total counts.",
}

var QuestListDaily = ToolDef{
	Name:        "quest_list_daily",
	Description: "List today's daily quests with current progress (current_count / target_count) and completion status.",
}

var QuestListWeekly = ToolDef{
	Name:        "quest_list_weekly",
	Description: "List this week's weekly quests with current progress (current_count / target_count) and completion status.",
}

var QuestGetXPSummary = ToolDef{
	Name:        "quest_get_xp_summary",
	Description: "Get the user's current XP summary: total XP earned, current level, XP earned into this level, XP required for the level, and XP remaining to next level.",
}

var QuestListHistory = ToolDef{
	Name:        "quest_list_history",
	Description: "List XP event history (quests completed, XP earned) sorted newest first. Pass limit to cap results; omit or set to 0 for all events.",
	Input:       QuestListHistoryIn{},
}

type QuestListHistoryIn struct {
	Limit int `json:"limit" jsonschema:"Max number of recent XP events to return; omit or 0 for all"`
}
