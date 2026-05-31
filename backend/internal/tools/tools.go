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
	Description: "Retrieve the full kanban board: metadata, all columns, and all cards. Use the returned IDs for subsequent calls.",
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
