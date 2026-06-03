package mcpserver

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/kinkando/personal-dashboard/internal/finance"
	financeSvc "github.com/kinkando/personal-dashboard/internal/finance/service"
	"github.com/kinkando/personal-dashboard/internal/health"
	healthSvc "github.com/kinkando/personal-dashboard/internal/health/service"
	"github.com/kinkando/personal-dashboard/internal/kanban"
	kanbanRepo "github.com/kinkando/personal-dashboard/internal/kanban/repository"
	"github.com/kinkando/personal-dashboard/internal/medicine"
	medicineSvc "github.com/kinkando/personal-dashboard/internal/medicine/service"
	"github.com/kinkando/personal-dashboard/internal/tools"
	"github.com/kinkando/personal-dashboard/internal/workout"
	workoutSvc "github.com/kinkando/personal-dashboard/internal/workout/service"
	"github.com/kinkando/personal-dashboard/pkg/middleware"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
)

// Deps bundles the shared dependencies consumed by all MCP tool handlers.
type Deps struct {
	FinSvc      *financeSvc.Service
	KanRepo     *kanbanRepo.Repository
	WkSvc       *workoutSvc.Service
	HeaSvc      *healthSvc.Service
	MedSvc      *medicineSvc.Service
	UserUUID    uuid.UUID
	FirebaseUID string
	Logger      *zap.Logger
}

// New creates a new MCP server with all tools registered.
func New(d Deps) *mcp.Server {
	s := mcp.NewServer(&mcp.Implementation{Name: "kinkando-dashboard", Version: "0.1.0"}, nil)
	registerTools(s, d)
	return s
}

// ---- Finance output types ---------------------------------------------------

type recordDTO struct {
	ID         string  `json:"id"`
	UserID     string  `json:"user_id"`
	Type       string  `json:"type"`
	Amount     float64 `json:"amount"`
	CategoryID *string `json:"category_id"`
	Note       string  `json:"note"`
	Date       string  `json:"date"`
	CreatedAt  string  `json:"created_at"`
}

type categoryDTO struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Type  string `json:"type"`
	Icon  string `json:"icon"`
	Color string `json:"color"`
}

type listRecordsOut struct {
	Records []recordDTO `json:"records"`
}

type listCategoriesOut struct {
	Categories []categoryDTO `json:"categories"`
}

type createCategoryOut struct {
	Category categoryDTO `json:"category"`
}

type createRecordOut struct {
	Record recordDTO `json:"record"`
}

type deleteCategoryOut struct {
	Deleted bool `json:"deleted"`
}

type deleteRecordOut struct {
	Deleted bool `json:"deleted"`
}

type categorySummaryDTO struct {
	CategoryID string  `json:"category_id"`
	Category   string  `json:"category"`
	Type       string  `json:"type"`
	Total      float64 `json:"total"`
	Icon       string  `json:"icon"`
	Color      string  `json:"color"`
}

type monthlySummaryDTO struct {
	Month      string               `json:"month"`
	Income     float64              `json:"income"`
	Expense    float64              `json:"expense"`
	Net        float64              `json:"net"`
	Categories []categorySummaryDTO `json:"categories"`
}

type monthlySummaryOut struct {
	Summary *monthlySummaryDTO `json:"summary"`
}

// ---- Kanban output types ----------------------------------------------------

type boardDTO struct {
	ID        string `json:"id"`
	UserID    string `json:"user_id"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
}

type columnDTO struct {
	ID        string `json:"id"`
	BoardID   string `json:"board_id"`
	Name      string `json:"name"`
	Type      string `json:"type"`
	IsSystem  bool   `json:"is_system"`
	Order     int    `json:"order"`
	CreatedAt string `json:"created_at"`
}

type cardDTO struct {
	ID            string   `json:"id"`
	BoardID       string   `json:"board_id"`
	ColumnID      string   `json:"column_id"`
	Title         string   `json:"title"`
	Content       string   `json:"content"`
	Description   string   `json:"description"`
	Priority      string   `json:"priority"`
	DueDate       *string  `json:"due_date,omitempty"`
	Tags          []string `json:"tags"`
	Order         int      `json:"order"`
	CompletedAt   *string  `json:"completed_at,omitempty"`
	ArchivedAt    *string  `json:"archived_at,omitempty"`
	ArchiveReason string   `json:"archive_reason,omitempty"`
	CreatedAt     string   `json:"created_at"`
}

type boardStatsDTO struct {
	Total      int            `json:"total"`
	ByColumn   map[string]int `json:"by_column"`
	ByPriority map[string]int `json:"by_priority"`
	Overdue    int            `json:"overdue"`
	NoDueDate  int            `json:"no_due_date"`
}

type getBoardOut struct {
	Board   boardDTO    `json:"board"`
	Columns []columnDTO `json:"columns"`
	Cards   []cardDTO   `json:"cards"`
}

type createCardOut struct {
	Card cardDTO `json:"card"`
}

type updateCardOut struct {
	Card cardDTO `json:"card"`
}

type moveCardOut struct {
	Moved bool `json:"moved"`
}

type deleteCardOut struct {
	Deleted bool `json:"deleted"`
}

type boardStatsOut struct {
	Stats boardStatsDTO `json:"stats"`
}

type archiveCardOut struct {
	Card cardDTO `json:"card"`
}

type unarchiveCardOut struct {
	Card cardDTO `json:"card"`
}

type listArchivedCardsOut struct {
	Cards []cardDTO `json:"cards"`
}

// ---- DTO helpers ------------------------------------------------------------

func toRecordDTO(r *finance.Record) recordDTO {
	dto := recordDTO{
		ID:        r.ID.String(),
		UserID:    r.UserID.String(),
		Type:      string(r.Type),
		Amount:    r.Amount,
		Note:      r.Note,
		Date:      r.Date.Format(time.DateOnly),
		CreatedAt: r.CreatedAt.Format(time.RFC3339),
	}
	if r.CategoryID != nil {
		s := r.CategoryID.String()
		dto.CategoryID = &s
	}
	return dto
}

func toCategoryDTO(c *finance.Category) categoryDTO {
	return categoryDTO{
		ID:    c.ID.String(),
		Name:  c.Name,
		Type:  string(c.Type),
		Icon:  c.Icon,
		Color: c.Color,
	}
}

func toBoardDTO(b *kanban.Board) boardDTO {
	return boardDTO{
		ID:        b.ID.Hex(),
		UserID:    b.UserID,
		Name:      b.Name,
		CreatedAt: b.CreatedAt.Format(time.RFC3339),
	}
}

func toColumnDTO(c *kanban.Column) columnDTO {
	return columnDTO{
		ID:        c.ID.Hex(),
		BoardID:   c.BoardID.Hex(),
		Name:      c.Name,
		Type:      c.Type,
		IsSystem:  c.IsSystem,
		Order:     c.Order,
		CreatedAt: c.CreatedAt.Format(time.RFC3339),
	}
}

func toMonthlySummaryDTO(s *finance.MonthlySummary) *monthlySummaryDTO {
	if s == nil {
		return nil
	}
	cats := make([]categorySummaryDTO, 0, len(s.Categories))
	for _, c := range s.Categories {
		dto := categorySummaryDTO{
			Category: c.Category,
			Type:     string(c.Type),
			Total:    c.Total,
			Icon:     c.Icon,
			Color:    c.Color,
		}
		if c.CategoryID != nil {
			dto.CategoryID = c.CategoryID.String()
		}
		cats = append(cats, dto)
	}
	return &monthlySummaryDTO{
		Month:      s.Month,
		Income:     s.Income,
		Expense:    s.Expense,
		Net:        s.Net,
		Categories: cats,
	}
}

func toCardDTO(c *kanban.Card) cardDTO {
	priority := c.Priority
	if priority == "" {
		priority = kanban.PriorityNone
	}
	tags := c.Tags
	if tags == nil {
		tags = []string{}
	}
	dto := cardDTO{
		ID:            c.ID.Hex(),
		BoardID:       c.BoardID.Hex(),
		ColumnID:      c.ColumnID.Hex(),
		Title:         c.Title,
		Content:       c.Content,
		Description:   c.Description,
		Priority:      priority,
		Tags:          tags,
		Order:         c.Order,
		ArchiveReason: c.ArchiveReason,
		CreatedAt:     c.CreatedAt.Format(time.RFC3339),
	}
	if c.DueDate != nil {
		s := c.DueDate.Format(time.DateOnly)
		dto.DueDate = &s
	}
	if c.CompletedAt != nil {
		s := c.CompletedAt.Format(time.RFC3339)
		dto.CompletedAt = &s
	}
	if c.ArchivedAt != nil {
		s := c.ArchivedAt.Format(time.RFC3339)
		dto.ArchivedAt = &s
	}
	return dto
}

// ---- Food output types ------------------------------------------------------

type foodLogDTO struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	MealType   string   `json:"meal_type"`
	Calories   *int     `json:"calories"`
	ProteinG   *float64 `json:"protein_g"`
	CarbsG     *float64 `json:"carbs_g"`
	FatG       *float64 `json:"fat_g"`
	Notes      *string  `json:"notes"`
	ConsumedAt string   `json:"consumed_at"`
	CreatedAt  string   `json:"created_at"`
}

type listFoodLogsOut struct {
	Logs []foodLogDTO `json:"logs"`
}

type foodLogOut struct {
	Log foodLogDTO `json:"log"`
}

type deleteFoodOut struct {
	Deleted bool `json:"deleted"`
}

func toFoodLogDTO(f *health.FoodLog) foodLogDTO {
	return foodLogDTO{
		ID:         f.ID.String(),
		Name:       f.Name,
		MealType:   string(f.MealType),
		Calories:   f.Calories,
		ProteinG:   f.ProteinG,
		CarbsG:     f.CarbsG,
		FatG:       f.FatG,
		Notes:      f.Notes,
		ConsumedAt: f.ConsumedAt.Format(time.DateOnly),
		CreatedAt:  f.CreatedAt.Format(time.RFC3339),
	}
}

// ---- Sleep output types -----------------------------------------------------

type sleepLogDTO struct {
	ID              string  `json:"id"`
	StartedAt       string  `json:"started_at"`
	EndedAt         string  `json:"ended_at"`
	DurationMinutes int     `json:"duration_minutes"`
	Score           *int    `json:"score"`
	Notes           *string `json:"notes"`
	LoggedAt        string  `json:"logged_at"`
	CreatedAt       string  `json:"created_at"`
}

type listSleepLogsOut struct {
	Logs []sleepLogDTO `json:"logs"`
}

type sleepLogOut struct {
	Log sleepLogDTO `json:"log"`
}

type deleteSleepOut struct {
	Deleted bool `json:"deleted"`
}

// ---- Weight output types ----------------------------------------------------

type weightLogDTO struct {
	ID       string  `json:"id"`
	Weight   float64 `json:"weight"`
	LoggedAt string  `json:"logged_at"`
}

type listWeightLogsOut struct {
	Logs []weightLogDTO `json:"logs"`
}

type weightLogOut struct {
	Log weightLogDTO `json:"log"`
}

func toWeightLogDTO(w *health.WeightLog) weightLogDTO {
	return weightLogDTO{
		ID:       w.ID.String(),
		Weight:   w.Weight,
		LoggedAt: w.LoggedAt.Format(time.DateOnly),
	}
}

func toSleepLogDTO(s *health.SleepLog) sleepLogDTO {
	return sleepLogDTO{
		ID:              s.ID.String(),
		StartedAt:       s.StartedAt.Format(time.RFC3339),
		EndedAt:         s.EndedAt.Format(time.RFC3339),
		DurationMinutes: s.DurationMinutes,
		Score:           s.Score,
		Notes:           s.Notes,
		LoggedAt:        s.LoggedAt.Format(time.DateOnly),
		CreatedAt:       s.CreatedAt.Format(time.RFC3339),
	}
}

// ---- Workout output types ---------------------------------------------------

type workoutPresetDTO struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description,omitempty"`
	Exercises   int    `json:"exercise_count"`
}

type scheduleEntryDTO struct {
	DayOfWeek  int    `json:"day_of_week"`
	DayName    string `json:"day_name"`
	PresetID   string `json:"preset_id"`
	PresetName string `json:"preset_name"`
	PresetType string `json:"preset_type"`
}

type sessionExerciseDTO struct {
	ID                    string   `json:"id"`
	Section               string   `json:"section"`
	OrderIndex            int      `json:"order_index"`
	Name                  string   `json:"name"`
	TargetSets            *int     `json:"target_sets,omitempty"`
	TargetReps            *int     `json:"target_reps,omitempty"`
	TargetDurationSeconds *int     `json:"target_duration_seconds,omitempty"`
	RestSeconds           *int     `json:"rest_seconds,omitempty"`
	ActualSets            *int     `json:"actual_sets,omitempty"`
	ActualReps            *int     `json:"actual_reps,omitempty"`
	ActualDurationSeconds *int     `json:"actual_duration_seconds,omitempty"`
	WeightKg              *float64 `json:"weight_kg,omitempty"`
	Completed             bool     `json:"completed"`
	Notes                 string   `json:"notes,omitempty"`
}

type workoutSessionDTO struct {
	ID              string               `json:"id"`
	Name            string               `json:"name"`
	Type            string               `json:"type"`
	PerformedAt     string               `json:"performed_at"`
	DurationMinutes *int                 `json:"duration_minutes,omitempty"`
	Notes           string               `json:"notes,omitempty"`
	Exercises       []sessionExerciseDTO `json:"exercises"`
}

type listPresetsOut struct {
	Presets []workoutPresetDTO `json:"presets"`
}

type getScheduleOut struct {
	Schedule []scheduleEntryDTO `json:"schedule"`
}

type listWorkoutSessionsOut struct {
	Sessions []workoutSessionDTO `json:"sessions"`
}

type startSessionOut struct {
	Session workoutSessionDTO `json:"session"`
}

type updateSessionOut struct {
	Session workoutSessionDTO `json:"session"`
}

type finishSessionOut struct {
	Session workoutSessionDTO `json:"session"`
}

type logExerciseOut struct {
	Exercise sessionExerciseDTO `json:"exercise"`
}

type addExerciseOut struct {
	Exercise sessionExerciseDTO `json:"exercise"`
}

type bulkLogExercisesOut struct {
	Exercises []sessionExerciseDTO `json:"exercises"`
}

type presetExerciseDTO struct {
	ID              string   `json:"id"`
	Section         string   `json:"section"`
	OrderIndex      int      `json:"order_index"`
	Name            string   `json:"name"`
	TargetMuscles   string   `json:"target_muscles,omitempty"`
	Sets            *int     `json:"sets,omitempty"`
	Reps            *int     `json:"reps,omitempty"`
	DurationSeconds *int     `json:"duration_seconds,omitempty"`
	RestSeconds     *int     `json:"rest_seconds,omitempty"`
	WeightKg        *float64 `json:"weight_kg,omitempty"`
	Equipment       string   `json:"equipment,omitempty"`
	Notes           string   `json:"notes,omitempty"`
}

type workoutPresetDetailDTO struct {
	ID          string              `json:"id"`
	Name        string              `json:"name"`
	Type        string              `json:"type"`
	Description string              `json:"description,omitempty"`
	Exercises   []presetExerciseDTO `json:"exercises"`
}

type getPresetOut struct {
	Preset workoutPresetDetailDTO `json:"preset"`
}

type createPresetOut struct {
	Preset workoutPresetDetailDTO `json:"preset"`
}

type updatePresetOut struct {
	Preset workoutPresetDetailDTO `json:"preset"`
}

type deletePresetOut struct {
	Deleted bool `json:"deleted"`
}

// ---- Workout DTO helpers ----------------------------------------------------

var dayNames = [7]string{"Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"}

func toPresetExerciseDTO(ex workout.PresetExercise) presetExerciseDTO {
	dto := presetExerciseDTO{
		ID:         ex.ID.String(),
		Section:    string(ex.Section),
		OrderIndex: ex.OrderIndex,
		Name:       ex.Name,
	}
	if ex.TargetMuscles != nil {
		dto.TargetMuscles = *ex.TargetMuscles
	}
	if ex.Sets != nil {
		dto.Sets = ex.Sets
	}
	if ex.Reps != nil {
		dto.Reps = ex.Reps
	}
	if ex.DurationSeconds != nil {
		dto.DurationSeconds = ex.DurationSeconds
	}
	if ex.RestSeconds != nil {
		dto.RestSeconds = ex.RestSeconds
	}
	if ex.WeightKg != nil {
		dto.WeightKg = ex.WeightKg
	}
	if ex.Equipment != nil {
		dto.Equipment = *ex.Equipment
	}
	if ex.Notes != nil {
		dto.Notes = *ex.Notes
	}
	return dto
}

func toWorkoutPresetDetailDTO(p *workout.Preset) workoutPresetDetailDTO {
	dto := workoutPresetDetailDTO{
		ID:   p.ID.String(),
		Name: p.Name,
		Type: string(p.Type),
	}
	if p.Description != nil {
		dto.Description = *p.Description
	}
	exs := make([]presetExerciseDTO, len(p.Exercises))
	for i, ex := range p.Exercises {
		exs[i] = toPresetExerciseDTO(ex)
	}
	dto.Exercises = exs
	return dto
}

func toWorkoutPresetDTO(p *workout.Preset) workoutPresetDTO {
	dto := workoutPresetDTO{
		ID:        p.ID.String(),
		Name:      p.Name,
		Type:      string(p.Type),
		Exercises: len(p.Exercises),
	}
	if p.Description != nil {
		dto.Description = *p.Description
	}
	return dto
}

func toSessionExerciseDTO(ex workout.SessionExercise) sessionExerciseDTO {
	dto := sessionExerciseDTO{
		ID:         ex.ID.String(),
		Section:    string(ex.Section),
		OrderIndex: ex.OrderIndex,
		Name:       ex.Name,
		Completed:  ex.Completed,
	}
	if ex.TargetSets != nil {
		dto.TargetSets = ex.TargetSets
	}
	if ex.TargetReps != nil {
		dto.TargetReps = ex.TargetReps
	}
	if ex.TargetDurationSeconds != nil {
		dto.TargetDurationSeconds = ex.TargetDurationSeconds
	}
	if ex.RestSeconds != nil {
		dto.RestSeconds = ex.RestSeconds
	}
	if ex.ActualSets != nil {
		dto.ActualSets = ex.ActualSets
	}
	if ex.ActualReps != nil {
		dto.ActualReps = ex.ActualReps
	}
	if ex.ActualDurationSeconds != nil {
		dto.ActualDurationSeconds = ex.ActualDurationSeconds
	}
	if ex.WeightKg != nil {
		dto.WeightKg = ex.WeightKg
	}
	if ex.Notes != nil {
		dto.Notes = *ex.Notes
	}
	return dto
}

func toWorkoutSessionDTO(s *workout.Session) workoutSessionDTO {
	dto := workoutSessionDTO{
		ID:          s.ID.String(),
		Name:        s.Name,
		Type:        string(s.Type),
		PerformedAt: s.PerformedAt.Format(time.DateOnly),
	}
	if s.DurationMinutes != nil {
		dto.DurationMinutes = s.DurationMinutes
	}
	if s.Notes != nil {
		dto.Notes = *s.Notes
	}
	exs := make([]sessionExerciseDTO, len(s.Exercises))
	for i, ex := range s.Exercises {
		exs[i] = toSessionExerciseDTO(ex)
	}
	dto.Exercises = exs
	return dto
}

// ---- Tool registration ------------------------------------------------------

func registerTools(s *mcp.Server, d Deps) {
	// Finance
	mcp.AddTool(s, &mcp.Tool{
		Name:        tools.FinanceListRecords.Name,
		Description: tools.FinanceListRecords.Description,
	}, middleware.MCPRequestLogger(d.Logger, tools.FinanceListRecords.Name, func(ctx context.Context, _ *mcp.CallToolRequest, in tools.ListRecordsIn) (*mcp.CallToolResult, listRecordsOut, error) {
		if in.Month == "" {
			return nil, listRecordsOut{}, fmt.Errorf("month is required (YYYY-MM)")
		}
		records, err := d.FinSvc.ListRecords(ctx, d.UserUUID, in.Month)
		if err != nil {
			return nil, listRecordsOut{}, fmt.Errorf("list records: %w", err)
		}
		if records == nil {
			records = []*finance.Record{}
		}
		dtos := make([]recordDTO, len(records))
		for i, rec := range records {
			dtos[i] = toRecordDTO(rec)
		}
		return nil, listRecordsOut{Records: dtos}, nil
	}))

	mcp.AddTool(s, &mcp.Tool{
		Name:        tools.FinanceListCategories.Name,
		Description: tools.FinanceListCategories.Description,
	}, middleware.MCPRequestLogger(d.Logger, tools.FinanceListCategories.Name, func(ctx context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, listCategoriesOut, error) {
		cats, err := d.FinSvc.ListCategories(ctx, d.UserUUID)
		if err != nil {
			return nil, listCategoriesOut{}, fmt.Errorf("list categories: %w", err)
		}
		if cats == nil {
			cats = []*finance.Category{}
		}
		dtos := make([]categoryDTO, len(cats))
		for i, c := range cats {
			dtos[i] = toCategoryDTO(c)
		}
		return nil, listCategoriesOut{Categories: dtos}, nil
	}))

	mcp.AddTool(s, &mcp.Tool{
		Name:        tools.FinanceCreateCategory.Name,
		Description: tools.FinanceCreateCategory.Description,
	}, middleware.MCPRequestLogger(d.Logger, tools.FinanceCreateCategory.Name, func(ctx context.Context, _ *mcp.CallToolRequest, in tools.CreateCategoryIn) (*mcp.CallToolResult, createCategoryOut, error) {
		rt := finance.RecordType(in.Type)
		if rt != finance.RecordTypeIncome && rt != finance.RecordTypeExpense {
			return nil, createCategoryOut{}, fmt.Errorf("type must be %q or %q, got %q", finance.RecordTypeIncome, finance.RecordTypeExpense, in.Type)
		}
		if in.Name == "" {
			return nil, createCategoryOut{}, fmt.Errorf("name is required")
		}
		cat, err := d.FinSvc.CreateCategory(ctx, d.UserUUID, finance.CreateCategoryInput{
			Name:  in.Name,
			Type:  rt,
			Icon:  in.Icon,
			Color: in.Color,
		})
		if err != nil {
			return nil, createCategoryOut{}, fmt.Errorf("create category: %w", err)
		}
		return nil, createCategoryOut{Category: toCategoryDTO(cat)}, nil
	}))

	mcp.AddTool(s, &mcp.Tool{
		Name:        tools.FinanceDeleteCategory.Name,
		Description: tools.FinanceDeleteCategory.Description,
	}, middleware.MCPRequestLogger(d.Logger, tools.FinanceDeleteCategory.Name, func(ctx context.Context, _ *mcp.CallToolRequest, in tools.DeleteCategoryIn) (*mcp.CallToolResult, deleteCategoryOut, error) {
		id, err := uuid.Parse(in.ID)
		if err != nil {
			return nil, deleteCategoryOut{}, fmt.Errorf("invalid category id %q: %w", in.ID, err)
		}
		if err := d.FinSvc.DeleteCategory(ctx, id, d.UserUUID); err != nil {
			return nil, deleteCategoryOut{}, fmt.Errorf("delete category: %w", err)
		}
		return nil, deleteCategoryOut{Deleted: true}, nil
	}))

	mcp.AddTool(s, &mcp.Tool{
		Name:        tools.FinanceCreateRecord.Name,
		Description: tools.FinanceCreateRecord.Description,
	}, middleware.MCPRequestLogger(d.Logger, tools.FinanceCreateRecord.Name, func(ctx context.Context, _ *mcp.CallToolRequest, in tools.CreateRecordIn) (*mcp.CallToolResult, createRecordOut, error) {
		rt := finance.RecordType(in.Type)
		if rt != finance.RecordTypeIncome && rt != finance.RecordTypeExpense {
			return nil, createRecordOut{}, fmt.Errorf("type must be %q or %q, got %q", finance.RecordTypeIncome, finance.RecordTypeExpense, in.Type)
		}
		if in.Amount <= 0 {
			return nil, createRecordOut{}, fmt.Errorf("amount must be positive, got %v", in.Amount)
		}
		cats, err := d.FinSvc.ListCategories(ctx, d.UserUUID)
		if err != nil {
			return nil, createRecordOut{}, fmt.Errorf("list categories: %w", err)
		}
		var catID uuid.UUID
		found := false
		for _, c := range cats {
			if strings.EqualFold(c.Name, in.Category) && c.Type == rt {
				catID = c.ID
				found = true
				break
			}
		}
		if !found {
			return nil, createRecordOut{}, fmt.Errorf("no %s category named %q — call finance_list_categories to see available names", in.Type, in.Category)
		}
		rec, err := d.FinSvc.CreateRecord(ctx, d.UserUUID, finance.CreateRecordInput{
			Type:       rt,
			Amount:     in.Amount,
			CategoryID: catID,
			Note:       in.Note,
			Date:       in.Date,
		})
		if err != nil {
			return nil, createRecordOut{}, fmt.Errorf("create record: %w", err)
		}
		return nil, createRecordOut{Record: toRecordDTO(rec)}, nil
	}))

	mcp.AddTool(s, &mcp.Tool{
		Name:        tools.FinanceDeleteRecord.Name,
		Description: tools.FinanceDeleteRecord.Description,
	}, middleware.MCPRequestLogger(d.Logger, tools.FinanceDeleteRecord.Name, func(ctx context.Context, _ *mcp.CallToolRequest, in tools.DeleteRecordIn) (*mcp.CallToolResult, deleteRecordOut, error) {
		id, err := uuid.Parse(in.ID)
		if err != nil {
			return nil, deleteRecordOut{}, fmt.Errorf("invalid record id %q: %w", in.ID, err)
		}
		if err := d.FinSvc.DeleteRecord(ctx, id, d.UserUUID); err != nil {
			return nil, deleteRecordOut{}, fmt.Errorf("delete record: %w", err)
		}
		return nil, deleteRecordOut{Deleted: true}, nil
	}))

	mcp.AddTool(s, &mcp.Tool{
		Name:        tools.FinanceMonthlySummary.Name,
		Description: tools.FinanceMonthlySummary.Description,
	}, middleware.MCPRequestLogger(d.Logger, tools.FinanceMonthlySummary.Name, func(ctx context.Context, _ *mcp.CallToolRequest, in tools.MonthlySummaryIn) (*mcp.CallToolResult, monthlySummaryOut, error) {
		if in.Month == "" {
			return nil, monthlySummaryOut{}, fmt.Errorf("month is required (YYYY-MM)")
		}
		summary, err := d.FinSvc.MonthlySummary(ctx, d.UserUUID, in.Month)
		if err != nil {
			return nil, monthlySummaryOut{}, fmt.Errorf("monthly summary: %w", err)
		}
		return nil, monthlySummaryOut{Summary: toMonthlySummaryDTO(summary)}, nil
	}))

	// Kanban
	mcp.AddTool(s, &mcp.Tool{
		Name:        tools.KanbanGetBoard.Name,
		Description: tools.KanbanGetBoard.Description,
	}, middleware.MCPRequestLogger(d.Logger, tools.KanbanGetBoard.Name, func(ctx context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, getBoardOut, error) {
		boards, err := d.KanRepo.ListBoards(ctx, d.FirebaseUID)
		if err != nil {
			return nil, getBoardOut{}, fmt.Errorf("list boards: %w", err)
		}
		board := boards[0]
		columns, err := d.KanRepo.GetColumns(ctx, board.ID)
		if err != nil {
			return nil, getBoardOut{}, fmt.Errorf("get columns: %w", err)
		}
		cards, err := d.KanRepo.GetCards(ctx, board.ID)
		if err != nil {
			return nil, getBoardOut{}, fmt.Errorf("get cards: %w", err)
		}
		if columns == nil {
			columns = []*kanban.Column{}
		}
		if cards == nil {
			cards = []*kanban.Card{}
		}
		columnDTOs := make([]columnDTO, len(columns))
		for i, col := range columns {
			columnDTOs[i] = toColumnDTO(col)
		}
		cardDTOs := make([]cardDTO, len(cards))
		for i, card := range cards {
			cardDTOs[i] = toCardDTO(card)
		}
		return nil, getBoardOut{
			Board:   toBoardDTO(board),
			Columns: columnDTOs,
			Cards:   cardDTOs,
		}, nil
	}))

	mcp.AddTool(s, &mcp.Tool{
		Name:        tools.KanbanCreateCard.Name,
		Description: tools.KanbanCreateCard.Description,
	}, middleware.MCPRequestLogger(d.Logger, tools.KanbanCreateCard.Name, func(ctx context.Context, _ *mcp.CallToolRequest, in tools.CreateCardIn) (*mcp.CallToolResult, createCardOut, error) {
		if in.Title == "" {
			return nil, createCardOut{}, fmt.Errorf("title is required")
		}
		colID, err := primitive.ObjectIDFromHex(in.ColumnID)
		if err != nil {
			return nil, createCardOut{}, fmt.Errorf("invalid column_id %q: %w", in.ColumnID, err)
		}
		boards, err := d.KanRepo.ListBoards(ctx, d.FirebaseUID)
		if err != nil {
			return nil, createCardOut{}, fmt.Errorf("list boards: %w", err)
		}
		board := boards[0]

		var dueDateStr *string
		if in.DueDate != "" {
			dueDateStr = &in.DueDate
		}

		priority := in.Priority
		if priority == "" {
			priority = kanban.PriorityNone
		}

		card, err := d.KanRepo.CreateCard(ctx, board.ID, colID, kanban.CreateCardInput{
			BoardID:     board.ID.Hex(),
			ColumnID:    in.ColumnID,
			Title:       in.Title,
			Content:     in.Content,
			Description: in.Description,
			Priority:    priority,
			DueDate:     dueDateStr,
			Tags:        in.Tags,
		})
		if err != nil {
			return nil, createCardOut{}, fmt.Errorf("create card: %w", err)
		}
		return nil, createCardOut{Card: toCardDTO(card)}, nil
	}))

	mcp.AddTool(s, &mcp.Tool{
		Name:        tools.KanbanUpdateCard.Name,
		Description: tools.KanbanUpdateCard.Description,
	}, middleware.MCPRequestLogger(d.Logger, tools.KanbanUpdateCard.Name, func(ctx context.Context, _ *mcp.CallToolRequest, in tools.UpdateCardIn) (*mcp.CallToolResult, updateCardOut, error) {
		cardID, err := primitive.ObjectIDFromHex(in.CardID)
		if err != nil {
			return nil, updateCardOut{}, fmt.Errorf("invalid card_id %q: %w", in.CardID, err)
		}
		fields := make(map[string]bool, len(in.UpdateFields))
		for _, f := range in.UpdateFields {
			fields[f] = true
		}
		var input kanban.UpdateCardInput
		if fields["title"] {
			input.Title = &in.Title
		}
		if fields["description"] {
			input.Description = &in.Description
		}
		if fields["priority"] {
			if !kanban.ValidPriority(in.Priority) {
				return nil, updateCardOut{}, fmt.Errorf("invalid priority %q", in.Priority)
			}
			input.Priority = &in.Priority
		}
		if fields["due_date"] {
			input.DueDate = &in.DueDate
		}
		if fields["tags"] {
			tags := in.Tags
			if tags == nil {
				tags = []string{}
			}
			input.Tags = &tags
		}
		card, err := d.KanRepo.UpdateCard(ctx, cardID, input)
		if err != nil {
			return nil, updateCardOut{}, fmt.Errorf("update card: %w", err)
		}
		return nil, updateCardOut{Card: toCardDTO(card)}, nil
	}))

	mcp.AddTool(s, &mcp.Tool{
		Name:        tools.KanbanMoveCard.Name,
		Description: tools.KanbanMoveCard.Description,
	}, middleware.MCPRequestLogger(d.Logger, tools.KanbanMoveCard.Name, func(ctx context.Context, _ *mcp.CallToolRequest, in tools.MoveCardIn) (*mcp.CallToolResult, moveCardOut, error) {
		cardID, err := primitive.ObjectIDFromHex(in.CardID)
		if err != nil {
			return nil, moveCardOut{}, fmt.Errorf("invalid card_id %q: %w", in.CardID, err)
		}
		if err := d.KanRepo.MoveCard(ctx, cardID, kanban.MoveCardInput{
			ColumnID: in.ColumnID,
			Order:    in.Order,
		}); err != nil {
			return nil, moveCardOut{}, fmt.Errorf("move card: %w", err)
		}
		return nil, moveCardOut{Moved: true}, nil
	}))

	mcp.AddTool(s, &mcp.Tool{
		Name:        tools.KanbanDeleteCard.Name,
		Description: tools.KanbanDeleteCard.Description,
	}, middleware.MCPRequestLogger(d.Logger, tools.KanbanDeleteCard.Name, func(ctx context.Context, _ *mcp.CallToolRequest, in tools.DeleteCardIn) (*mcp.CallToolResult, deleteCardOut, error) {
		cardID, err := primitive.ObjectIDFromHex(in.CardID)
		if err != nil {
			return nil, deleteCardOut{}, fmt.Errorf("invalid card_id %q: %w", in.CardID, err)
		}
		if err := d.KanRepo.DeleteCard(ctx, cardID); err != nil {
			return nil, deleteCardOut{}, fmt.Errorf("delete card: %w", err)
		}
		return nil, deleteCardOut{Deleted: true}, nil
	}))

	mcp.AddTool(s, &mcp.Tool{
		Name:        tools.KanbanBoardStats.Name,
		Description: tools.KanbanBoardStats.Description,
	}, middleware.MCPRequestLogger(d.Logger, tools.KanbanBoardStats.Name, func(ctx context.Context, _ *mcp.CallToolRequest, in tools.BoardStatsIn) (*mcp.CallToolResult, boardStatsOut, error) {
		boardID, err := primitive.ObjectIDFromHex(in.BoardID)
		if err != nil {
			return nil, boardStatsOut{}, fmt.Errorf("invalid board_id %q: %w", in.BoardID, err)
		}
		stats, err := d.KanRepo.GetBoardStats(ctx, boardID)
		if err != nil {
			return nil, boardStatsOut{}, fmt.Errorf("get stats: %w", err)
		}
		dto := boardStatsDTO{
			Total:      stats.Total,
			ByColumn:   stats.ByColumn,
			ByPriority: stats.ByPriority,
			Overdue:    stats.Overdue,
			NoDueDate:  stats.NoDueDate,
		}
		return nil, boardStatsOut{Stats: dto}, nil
	}))

	mcp.AddTool(s, &mcp.Tool{
		Name:        tools.KanbanArchiveCard.Name,
		Description: tools.KanbanArchiveCard.Description,
	}, middleware.MCPRequestLogger(d.Logger, tools.KanbanArchiveCard.Name, func(ctx context.Context, _ *mcp.CallToolRequest, in tools.ArchiveCardIn) (*mcp.CallToolResult, archiveCardOut, error) {
		cardID, err := primitive.ObjectIDFromHex(in.CardID)
		if err != nil {
			return nil, archiveCardOut{}, fmt.Errorf("invalid card_id %q: %w", in.CardID, err)
		}
		if in.Reason != "" && !kanban.ValidUserArchiveReason(in.Reason) {
			return nil, archiveCardOut{}, fmt.Errorf("invalid reason %q: must be cancelled, duplicate, or stale", in.Reason)
		}
		card, err := d.KanRepo.ArchiveCard(ctx, cardID, in.Reason)
		if err != nil {
			return nil, archiveCardOut{}, fmt.Errorf("archive card: %w", err)
		}
		return nil, archiveCardOut{Card: toCardDTO(card)}, nil
	}))

	mcp.AddTool(s, &mcp.Tool{
		Name:        tools.KanbanUnarchiveCard.Name,
		Description: tools.KanbanUnarchiveCard.Description,
	}, middleware.MCPRequestLogger(d.Logger, tools.KanbanUnarchiveCard.Name, func(ctx context.Context, _ *mcp.CallToolRequest, in tools.UnarchiveCardIn) (*mcp.CallToolResult, unarchiveCardOut, error) {
		cardID, err := primitive.ObjectIDFromHex(in.CardID)
		if err != nil {
			return nil, unarchiveCardOut{}, fmt.Errorf("invalid card_id %q: %w", in.CardID, err)
		}
		card, err := d.KanRepo.UnarchiveCard(ctx, cardID)
		if err != nil {
			return nil, unarchiveCardOut{}, fmt.Errorf("unarchive card: %w", err)
		}
		return nil, unarchiveCardOut{Card: toCardDTO(card)}, nil
	}))

	mcp.AddTool(s, &mcp.Tool{
		Name:        tools.KanbanListArchivedCards.Name,
		Description: tools.KanbanListArchivedCards.Description,
	}, middleware.MCPRequestLogger(d.Logger, tools.KanbanListArchivedCards.Name, func(ctx context.Context, _ *mcp.CallToolRequest, in tools.ListArchivedCardsIn) (*mcp.CallToolResult, listArchivedCardsOut, error) {
		boardID, err := primitive.ObjectIDFromHex(in.BoardID)
		if err != nil {
			return nil, listArchivedCardsOut{}, fmt.Errorf("invalid board_id %q: %w", in.BoardID, err)
		}
		cards, err := d.KanRepo.ListArchivedCards(ctx, boardID, kanban.ListArchivedFilter{
			Reason: in.Reason,
			Month:  in.Month,
			Year:   in.Year,
		})
		if err != nil {
			return nil, listArchivedCardsOut{}, fmt.Errorf("list archived cards: %w", err)
		}
		if cards == nil {
			cards = []*kanban.Card{}
		}
		dtos := make([]cardDTO, len(cards))
		for i, card := range cards {
			dtos[i] = toCardDTO(card)
		}
		return nil, listArchivedCardsOut{Cards: dtos}, nil
	}))

	// Workout
	mcp.AddTool(s, &mcp.Tool{
		Name:        tools.WorkoutListSessions.Name,
		Description: tools.WorkoutListSessions.Description,
	}, middleware.MCPRequestLogger(d.Logger, tools.WorkoutListSessions.Name, func(ctx context.Context, _ *mcp.CallToolRequest, in tools.WorkoutListSessionsIn) (*mcp.CallToolResult, listWorkoutSessionsOut, error) {
		now := time.Now().UTC().Truncate(24 * time.Hour)
		from := now.AddDate(0, 0, -30)
		to := now
		if in.From != "" {
			t, err := time.Parse(time.DateOnly, in.From)
			if err != nil {
				return nil, listWorkoutSessionsOut{}, fmt.Errorf("invalid from date: %w", err)
			}
			from = t
		}
		if in.To != "" {
			t, err := time.Parse(time.DateOnly, in.To)
			if err != nil {
				return nil, listWorkoutSessionsOut{}, fmt.Errorf("invalid to date: %w", err)
			}
			to = t
		}
		sessions, err := d.WkSvc.ListSessions(ctx, d.UserUUID, from, to)
		if err != nil {
			return nil, listWorkoutSessionsOut{}, fmt.Errorf("list sessions: %w", err)
		}
		dtos := make([]workoutSessionDTO, len(sessions))
		for i, s := range sessions {
			dtos[i] = toWorkoutSessionDTO(s)
		}
		return nil, listWorkoutSessionsOut{Sessions: dtos}, nil
	}))

	mcp.AddTool(s, &mcp.Tool{
		Name:        tools.WorkoutListPresets.Name,
		Description: tools.WorkoutListPresets.Description,
	}, middleware.MCPRequestLogger(d.Logger, tools.WorkoutListPresets.Name, func(ctx context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, listPresetsOut, error) {
		presets, err := d.WkSvc.ListPresets(ctx, d.UserUUID)
		if err != nil {
			return nil, listPresetsOut{}, fmt.Errorf("list presets: %w", err)
		}
		dtos := make([]workoutPresetDTO, len(presets))
		for i, p := range presets {
			dtos[i] = toWorkoutPresetDTO(p)
		}
		return nil, listPresetsOut{Presets: dtos}, nil
	}))

	mcp.AddTool(s, &mcp.Tool{
		Name:        tools.WorkoutGetPreset.Name,
		Description: tools.WorkoutGetPreset.Description,
	}, middleware.MCPRequestLogger(d.Logger, tools.WorkoutGetPreset.Name, func(ctx context.Context, _ *mcp.CallToolRequest, in tools.WorkoutGetPresetIn) (*mcp.CallToolResult, getPresetOut, error) {
		presetID, err := uuid.Parse(in.PresetID)
		if err != nil {
			return nil, getPresetOut{}, fmt.Errorf("invalid preset_id %q: %w", in.PresetID, err)
		}
		preset, err := d.WkSvc.GetPreset(ctx, presetID, d.UserUUID)
		if err != nil {
			return nil, getPresetOut{}, fmt.Errorf("get preset: %w", err)
		}
		return nil, getPresetOut{Preset: toWorkoutPresetDetailDTO(preset)}, nil
	}))

	mcp.AddTool(s, &mcp.Tool{
		Name:        tools.WorkoutGetSchedule.Name,
		Description: tools.WorkoutGetSchedule.Description,
	}, middleware.MCPRequestLogger(d.Logger, tools.WorkoutGetSchedule.Name, func(ctx context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, getScheduleOut, error) {
		entries, err := d.WkSvc.GetSchedule(ctx, d.UserUUID)
		if err != nil {
			return nil, getScheduleOut{}, fmt.Errorf("get schedule: %w", err)
		}
		dtos := make([]scheduleEntryDTO, len(entries))
		for i, e := range entries {
			dtos[i] = scheduleEntryDTO{
				DayOfWeek:  e.DayOfWeek,
				DayName:    dayNames[e.DayOfWeek],
				PresetID:   e.PresetID.String(),
				PresetName: e.PresetName,
				PresetType: string(e.PresetType),
			}
		}
		return nil, getScheduleOut{Schedule: dtos}, nil
	}))

	mcp.AddTool(s, &mcp.Tool{
		Name:        tools.WorkoutStartSession.Name,
		Description: tools.WorkoutStartSession.Description,
	}, middleware.MCPRequestLogger(d.Logger, tools.WorkoutStartSession.Name, func(ctx context.Context, _ *mcp.CallToolRequest, in tools.WorkoutStartSessionIn) (*mcp.CallToolResult, startSessionOut, error) {
		if in.PresetName == "" && in.Type == "" {
			return nil, startSessionOut{}, fmt.Errorf("preset_name or type is required")
		}
		var nameOverride *string
		if in.Name != "" {
			nameOverride = &in.Name
		}
		input := workout.CreateSessionInput{
			Date: in.Date,
			Name: nameOverride,
		}
		if in.PresetName != "" {
			presets, err := d.WkSvc.ListPresets(ctx, d.UserUUID)
			if err != nil {
				return nil, startSessionOut{}, fmt.Errorf("list presets: %w", err)
			}
			var presetID *uuid.UUID
			for _, p := range presets {
				if strings.EqualFold(p.Name, in.PresetName) {
					id := p.ID
					presetID = &id
					break
				}
			}
			if presetID == nil {
				return nil, startSessionOut{}, fmt.Errorf("no preset named %q — call workout_list_presets to see available names", in.PresetName)
			}
			input.PresetID = presetID
		} else {
			t := workout.Type(in.Type)
			input.Type = &t
		}
		session, err := d.WkSvc.CreateSession(ctx, d.UserUUID, input)
		if err != nil {
			return nil, startSessionOut{}, fmt.Errorf("start session: %w", err)
		}
		return nil, startSessionOut{Session: toWorkoutSessionDTO(session)}, nil
	}))

	mcp.AddTool(s, &mcp.Tool{
		Name:        tools.WorkoutUpdateSession.Name,
		Description: tools.WorkoutUpdateSession.Description,
	}, middleware.MCPRequestLogger(d.Logger, tools.WorkoutUpdateSession.Name, func(ctx context.Context, _ *mcp.CallToolRequest, in tools.WorkoutUpdateSessionIn) (*mcp.CallToolResult, updateSessionOut, error) {
		if strings.TrimSpace(in.Name) == "" {
			return nil, updateSessionOut{}, fmt.Errorf("name is required")
		}
		sessionID, err := uuid.Parse(in.SessionID)
		if err != nil {
			return nil, updateSessionOut{}, fmt.Errorf("invalid session_id %q: %w", in.SessionID, err)
		}
		var durationMinutes *int
		if in.DurationMinutes > 0 {
			durationMinutes = &in.DurationMinutes
		}
		var notes *string
		if in.Notes != "" {
			notes = &in.Notes
		}
		session, err := d.WkSvc.UpdateSession(ctx, sessionID, d.UserUUID, workout.UpdateSessionInput{
			Name:            in.Name,
			DurationMinutes: durationMinutes,
			Notes:           notes,
		})
		if err != nil {
			return nil, updateSessionOut{}, fmt.Errorf("update session: %w", err)
		}
		return nil, updateSessionOut{Session: toWorkoutSessionDTO(session)}, nil
	}))

	mcp.AddTool(s, &mcp.Tool{
		Name:        tools.WorkoutLogExercise.Name,
		Description: tools.WorkoutLogExercise.Description,
	}, middleware.MCPRequestLogger(d.Logger, tools.WorkoutLogExercise.Name, func(ctx context.Context, _ *mcp.CallToolRequest, in tools.WorkoutLogExerciseIn) (*mcp.CallToolResult, logExerciseOut, error) {
		sessionID, err := uuid.Parse(in.SessionID)
		if err != nil {
			return nil, logExerciseOut{}, fmt.Errorf("invalid session_id %q: %w", in.SessionID, err)
		}
		exerciseID, err := uuid.Parse(in.ExerciseID)
		if err != nil {
			return nil, logExerciseOut{}, fmt.Errorf("invalid exercise_id %q: %w", in.ExerciseID, err)
		}
		input := workout.UpdateSessionExerciseInput{
			Completed: in.Completed,
		}
		if in.ActualSets > 0 {
			input.ActualSets = &in.ActualSets
		}
		if in.ActualReps > 0 {
			input.ActualReps = &in.ActualReps
		}
		if in.ActualDurationSeconds > 0 {
			input.ActualDurationSeconds = &in.ActualDurationSeconds
		}
		if in.WeightKg > 0 {
			w := in.WeightKg
			input.WeightKg = &w
		}
		if in.Notes != "" {
			input.Notes = &in.Notes
		}
		ex, err := d.WkSvc.UpdateSessionExercise(ctx, exerciseID, sessionID, d.UserUUID, input)
		if err != nil {
			return nil, logExerciseOut{}, fmt.Errorf("log exercise: %w", err)
		}
		return nil, logExerciseOut{Exercise: toSessionExerciseDTO(*ex)}, nil
	}))

	mcp.AddTool(s, &mcp.Tool{
		Name:        tools.WorkoutAddExercise.Name,
		Description: tools.WorkoutAddExercise.Description,
	}, middleware.MCPRequestLogger(d.Logger, tools.WorkoutAddExercise.Name, func(ctx context.Context, _ *mcp.CallToolRequest, in tools.WorkoutAddExerciseIn) (*mcp.CallToolResult, addExerciseOut, error) {
		if strings.TrimSpace(in.Name) == "" {
			return nil, addExerciseOut{}, fmt.Errorf("name is required")
		}
		sessionID, err := uuid.Parse(in.SessionID)
		if err != nil {
			return nil, addExerciseOut{}, fmt.Errorf("invalid session_id %q: %w", in.SessionID, err)
		}
		section := workout.Section(in.Section)
		if section == "" {
			section = workout.SectionMain
		}
		input := workout.AddSessionExerciseInput{
			Section: section,
			Name:    in.Name,
		}
		if in.TargetSets > 0 {
			input.TargetSets = &in.TargetSets
		}
		if in.TargetReps > 0 {
			input.TargetReps = &in.TargetReps
		}
		if in.TargetDurationSeconds > 0 {
			input.TargetDurationSeconds = &in.TargetDurationSeconds
		}
		if in.RestSeconds > 0 {
			input.RestSeconds = &in.RestSeconds
		}
		ex, err := d.WkSvc.AddSessionExercise(ctx, sessionID, d.UserUUID, input)
		if err != nil {
			return nil, addExerciseOut{}, fmt.Errorf("add exercise: %w", err)
		}
		return nil, addExerciseOut{Exercise: toSessionExerciseDTO(*ex)}, nil
	}))

	mcp.AddTool(s, &mcp.Tool{
		Name:        tools.WorkoutBulkLogExercises.Name,
		Description: tools.WorkoutBulkLogExercises.Description,
	}, middleware.MCPRequestLogger(d.Logger, tools.WorkoutBulkLogExercises.Name, func(ctx context.Context, _ *mcp.CallToolRequest, in tools.WorkoutBulkLogExercisesIn) (*mcp.CallToolResult, bulkLogExercisesOut, error) {
		if len(in.Items) == 0 {
			return nil, bulkLogExercisesOut{}, fmt.Errorf("items must not be empty")
		}
		sessionID, err := uuid.Parse(in.SessionID)
		if err != nil {
			return nil, bulkLogExercisesOut{}, fmt.Errorf("invalid session_id %q: %w", in.SessionID, err)
		}
		items := make([]workout.BulkUpdateSessionExerciseItem, 0, len(in.Items))
		for _, item := range in.Items {
			exID, err := uuid.Parse(item.ExerciseID)
			if err != nil {
				return nil, bulkLogExercisesOut{}, fmt.Errorf("invalid exercise_id %q: %w", item.ExerciseID, err)
			}
			inp := workout.BulkUpdateSessionExerciseItem{
				ID:        exID,
				Completed: item.Completed,
			}
			if item.ActualSets > 0 {
				s := item.ActualSets
				inp.ActualSets = &s
			}
			if item.ActualReps > 0 {
				rp := item.ActualReps
				inp.ActualReps = &rp
			}
			if item.ActualDurationSeconds > 0 {
				d := item.ActualDurationSeconds
				inp.ActualDurationSeconds = &d
			}
			if item.WeightKg > 0 {
				w := item.WeightKg
				inp.WeightKg = &w
			}
			if item.Notes != "" {
				inp.Notes = &item.Notes
			}
			items = append(items, inp)
		}
		exercises, err := d.WkSvc.BulkUpdateSessionExercises(ctx, sessionID, d.UserUUID, items)
		if err != nil {
			return nil, bulkLogExercisesOut{}, fmt.Errorf("bulk log exercises: %w", err)
		}
		dtos := make([]sessionExerciseDTO, len(exercises))
		for i, ex := range exercises {
			dtos[i] = toSessionExerciseDTO(ex)
		}
		return nil, bulkLogExercisesOut{Exercises: dtos}, nil
	}))

	mcp.AddTool(s, &mcp.Tool{
		Name:        tools.WorkoutFinishSession.Name,
		Description: tools.WorkoutFinishSession.Description,
	}, middleware.MCPRequestLogger(d.Logger, tools.WorkoutFinishSession.Name, func(ctx context.Context, _ *mcp.CallToolRequest, in tools.WorkoutFinishSessionIn) (*mcp.CallToolResult, finishSessionOut, error) {
		sessionID, err := uuid.Parse(in.SessionID)
		if err != nil {
			return nil, finishSessionOut{}, fmt.Errorf("invalid session_id %q: %w", in.SessionID, err)
		}
		session, err := d.WkSvc.FinishSession(ctx, sessionID, d.UserUUID)
		if err != nil {
			return nil, finishSessionOut{}, fmt.Errorf("finish session: %w", err)
		}
		return nil, finishSessionOut{Session: toWorkoutSessionDTO(session)}, nil
	}))

	mcp.AddTool(s, &mcp.Tool{
		Name:        tools.WorkoutCreatePreset.Name,
		Description: tools.WorkoutCreatePreset.Description,
	}, middleware.MCPRequestLogger(d.Logger, tools.WorkoutCreatePreset.Name, func(ctx context.Context, _ *mcp.CallToolRequest, in tools.WorkoutCreatePresetIn) (*mcp.CallToolResult, createPresetOut, error) {
		if strings.TrimSpace(in.Name) == "" {
			return nil, createPresetOut{}, fmt.Errorf("name is required")
		}
		t := workout.Type(in.Type)
		switch t {
		case workout.TypeWeightTraining, workout.TypeBodyWeight, workout.TypeRunning,
			workout.TypeWalking, workout.TypeCardio, workout.TypeMobility, workout.TypeCustom:
		default:
			return nil, createPresetOut{}, fmt.Errorf("invalid type %q: must be weight_training, body_weight, running, walking, cardio, mobility, or custom", in.Type)
		}
		exInputs := make([]workout.PresetExerciseInput, len(in.Exercises))
		for i, ex := range in.Exercises {
			section := workout.Section(ex.Section)
			if section == "" {
				section = workout.SectionMain
			}
			exIn := workout.PresetExerciseInput{
				Section: section,
				Name:    ex.Name,
			}
			if ex.TargetMuscles != "" {
				exIn.TargetMuscles = &ex.TargetMuscles
			}
			if ex.Instructions != "" {
				exIn.Instructions = &ex.Instructions
			}
			if ex.Sets > 0 {
				exIn.Sets = &ex.Sets
			}
			if ex.Reps > 0 {
				exIn.Reps = &ex.Reps
			}
			if ex.DurationSeconds > 0 {
				exIn.DurationSeconds = &ex.DurationSeconds
			}
			if ex.RestSeconds > 0 {
				exIn.RestSeconds = &ex.RestSeconds
			}
			if ex.WeightKg > 0 {
				exIn.WeightKg = &ex.WeightKg
			}
			if ex.Equipment != "" {
				exIn.Equipment = &ex.Equipment
			}
			if ex.Notes != "" {
				exIn.Notes = &ex.Notes
			}
			exInputs[i] = exIn
		}
		input := workout.CreatePresetInput{
			Name:      in.Name,
			Type:      t,
			Exercises: exInputs,
		}
		if in.Description != "" {
			input.Description = &in.Description
		}
		preset, err := d.WkSvc.CreatePreset(ctx, d.UserUUID, input)
		if err != nil {
			return nil, createPresetOut{}, fmt.Errorf("create preset: %w", err)
		}
		return nil, createPresetOut{Preset: toWorkoutPresetDetailDTO(preset)}, nil
	}))

	mcp.AddTool(s, &mcp.Tool{
		Name:        tools.WorkoutUpdatePreset.Name,
		Description: tools.WorkoutUpdatePreset.Description,
	}, middleware.MCPRequestLogger(d.Logger, tools.WorkoutUpdatePreset.Name, func(ctx context.Context, _ *mcp.CallToolRequest, in tools.WorkoutUpdatePresetIn) (*mcp.CallToolResult, updatePresetOut, error) {
		presetID, err := uuid.Parse(in.PresetID)
		if err != nil {
			return nil, updatePresetOut{}, fmt.Errorf("invalid preset_id %q: %w", in.PresetID, err)
		}
		if strings.TrimSpace(in.Name) == "" {
			return nil, updatePresetOut{}, fmt.Errorf("name is required")
		}
		t := workout.Type(in.Type)
		switch t {
		case workout.TypeWeightTraining, workout.TypeBodyWeight, workout.TypeRunning,
			workout.TypeWalking, workout.TypeCardio, workout.TypeMobility, workout.TypeCustom:
		default:
			return nil, updatePresetOut{}, fmt.Errorf("invalid type %q", in.Type)
		}
		exInputs := make([]workout.PresetExerciseInput, len(in.Exercises))
		for i, ex := range in.Exercises {
			section := workout.Section(ex.Section)
			if section == "" {
				section = workout.SectionMain
			}
			exIn := workout.PresetExerciseInput{Section: section, Name: ex.Name}
			if ex.TargetMuscles != "" {
				exIn.TargetMuscles = &ex.TargetMuscles
			}
			if ex.Instructions != "" {
				exIn.Instructions = &ex.Instructions
			}
			if ex.Sets > 0 {
				exIn.Sets = &ex.Sets
			}
			if ex.Reps > 0 {
				exIn.Reps = &ex.Reps
			}
			if ex.DurationSeconds > 0 {
				exIn.DurationSeconds = &ex.DurationSeconds
			}
			if ex.RestSeconds > 0 {
				exIn.RestSeconds = &ex.RestSeconds
			}
			if ex.WeightKg > 0 {
				exIn.WeightKg = &ex.WeightKg
			}
			if ex.Equipment != "" {
				exIn.Equipment = &ex.Equipment
			}
			if ex.Notes != "" {
				exIn.Notes = &ex.Notes
			}
			exInputs[i] = exIn
		}
		input := workout.UpdatePresetInput{Name: in.Name, Type: t, Exercises: exInputs}
		if in.Description != "" {
			input.Description = &in.Description
		}
		preset, err := d.WkSvc.UpdatePreset(ctx, presetID, d.UserUUID, input)
		if err != nil {
			return nil, updatePresetOut{}, fmt.Errorf("update preset: %w", err)
		}
		return nil, updatePresetOut{Preset: toWorkoutPresetDetailDTO(preset)}, nil
	}))

	mcp.AddTool(s, &mcp.Tool{
		Name:        tools.WorkoutDeletePreset.Name,
		Description: tools.WorkoutDeletePreset.Description,
	}, middleware.MCPRequestLogger(d.Logger, tools.WorkoutDeletePreset.Name, func(ctx context.Context, _ *mcp.CallToolRequest, in tools.WorkoutDeletePresetIn) (*mcp.CallToolResult, deletePresetOut, error) {
		presetID, err := uuid.Parse(in.PresetID)
		if err != nil {
			return nil, deletePresetOut{}, fmt.Errorf("invalid preset_id %q: %w", in.PresetID, err)
		}
		if err := d.WkSvc.DeletePreset(ctx, presetID, d.UserUUID); err != nil {
			return nil, deletePresetOut{}, fmt.Errorf("delete preset: %w", err)
		}
		return nil, deletePresetOut{Deleted: true}, nil
	}))

	// Food
	mcp.AddTool(s, &mcp.Tool{
		Name:        tools.FoodListLogs.Name,
		Description: tools.FoodListLogs.Description,
	}, middleware.MCPRequestLogger(d.Logger, tools.FoodListLogs.Name, func(ctx context.Context, _ *mcp.CallToolRequest, in tools.FoodListLogsIn) (*mcp.CallToolResult, listFoodLogsOut, error) {
		logs, err := d.HeaSvc.ListFoodLogs(ctx, d.UserUUID)
		if err != nil {
			return nil, listFoodLogsOut{}, fmt.Errorf("list food logs: %w", err)
		}
		// filter by date range if provided
		from := time.Time{}
		to := time.Now().UTC().Add(24 * time.Hour)
		if in.From != "" {
			t, err := time.Parse(time.DateOnly, in.From)
			if err != nil {
				return nil, listFoodLogsOut{}, fmt.Errorf("invalid from date: %w", err)
			}
			from = t
		}
		if in.To != "" {
			t, err := time.Parse(time.DateOnly, in.To)
			if err != nil {
				return nil, listFoodLogsOut{}, fmt.Errorf("invalid to date: %w", err)
			}
			to = t.Add(24 * time.Hour) // inclusive
		}
		dtos := make([]foodLogDTO, 0, len(logs))
		for _, l := range logs {
			if (!from.IsZero() && l.ConsumedAt.Before(from)) || l.ConsumedAt.After(to) {
				continue
			}
			dtos = append(dtos, toFoodLogDTO(l))
		}
		return nil, listFoodLogsOut{Logs: dtos}, nil
	}))

	mcp.AddTool(s, &mcp.Tool{
		Name:        tools.FoodLogMeal.Name,
		Description: tools.FoodLogMeal.Description,
	}, middleware.MCPRequestLogger(d.Logger, tools.FoodLogMeal.Name, func(ctx context.Context, _ *mcp.CallToolRequest, in tools.FoodLogMealIn) (*mcp.CallToolResult, foodLogOut, error) {
		inp := health.CreateFoodInput{
			Name:       in.Name,
			MealType:   health.MealType(in.MealType),
			Notes:      nilStr(in.Notes),
			ConsumedAt: in.ConsumedAt,
		}
		if in.Calories > 0 {
			c := in.Calories
			inp.Calories = &c
		}
		if in.ProteinG > 0 {
			inp.ProteinG = &in.ProteinG
		}
		if in.CarbsG > 0 {
			inp.CarbsG = &in.CarbsG
		}
		if in.FatG > 0 {
			inp.FatG = &in.FatG
		}
		log, err := d.HeaSvc.CreateFoodLog(ctx, d.UserUUID, inp)
		if err != nil {
			return nil, foodLogOut{}, fmt.Errorf("create food log: %w", err)
		}
		return nil, foodLogOut{Log: toFoodLogDTO(log)}, nil
	}))

	mcp.AddTool(s, &mcp.Tool{
		Name:        tools.FoodUpdateMeal.Name,
		Description: tools.FoodUpdateMeal.Description,
	}, middleware.MCPRequestLogger(d.Logger, tools.FoodUpdateMeal.Name, func(ctx context.Context, _ *mcp.CallToolRequest, in tools.FoodUpdateMealIn) (*mcp.CallToolResult, foodLogOut, error) {
		logID, err := uuid.Parse(in.LogID)
		if err != nil {
			return nil, foodLogOut{}, fmt.Errorf("invalid log_id %q: %w", in.LogID, err)
		}
		inp := health.UpdateFoodInput{
			Name:       in.Name,
			MealType:   health.MealType(in.MealType),
			Notes:      nilStr(in.Notes),
			ConsumedAt: in.ConsumedAt,
		}
		if in.Calories > 0 {
			c := in.Calories
			inp.Calories = &c
		}
		if in.ProteinG > 0 {
			inp.ProteinG = &in.ProteinG
		}
		if in.CarbsG > 0 {
			inp.CarbsG = &in.CarbsG
		}
		if in.FatG > 0 {
			inp.FatG = &in.FatG
		}
		log, err := d.HeaSvc.UpdateFoodLog(ctx, logID, d.UserUUID, inp)
		if err != nil {
			return nil, foodLogOut{}, fmt.Errorf("update food log: %w", err)
		}
		return nil, foodLogOut{Log: toFoodLogDTO(log)}, nil
	}))

	mcp.AddTool(s, &mcp.Tool{
		Name:        tools.FoodDeleteMeal.Name,
		Description: tools.FoodDeleteMeal.Description,
	}, middleware.MCPRequestLogger(d.Logger, tools.FoodDeleteMeal.Name, func(ctx context.Context, _ *mcp.CallToolRequest, in tools.FoodDeleteMealIn) (*mcp.CallToolResult, deleteFoodOut, error) {
		logID, err := uuid.Parse(in.LogID)
		if err != nil {
			return nil, deleteFoodOut{}, fmt.Errorf("invalid log_id %q: %w", in.LogID, err)
		}
		if err := d.HeaSvc.DeleteFoodLog(ctx, logID, d.UserUUID); err != nil {
			return nil, deleteFoodOut{}, fmt.Errorf("delete food log: %w", err)
		}
		return nil, deleteFoodOut{Deleted: true}, nil
	}))

	// Weight
	mcp.AddTool(s, &mcp.Tool{
		Name:        tools.HealthListWeightLogs.Name,
		Description: tools.HealthListWeightLogs.Description,
	}, middleware.MCPRequestLogger(d.Logger, tools.HealthListWeightLogs.Name, func(ctx context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, listWeightLogsOut, error) {
		logs, err := d.HeaSvc.ListWeightLogs(ctx, d.UserUUID)
		if err != nil {
			return nil, listWeightLogsOut{}, fmt.Errorf("list weight logs: %w", err)
		}
		dtos := make([]weightLogDTO, len(logs))
		for i, l := range logs {
			dtos[i] = toWeightLogDTO(l)
		}
		return nil, listWeightLogsOut{Logs: dtos}, nil
	}))

	mcp.AddTool(s, &mcp.Tool{
		Name:        tools.HealthLogWeight.Name,
		Description: tools.HealthLogWeight.Description,
	}, middleware.MCPRequestLogger(d.Logger, tools.HealthLogWeight.Name, func(ctx context.Context, _ *mcp.CallToolRequest, in tools.HealthLogWeightIn) (*mcp.CallToolResult, weightLogOut, error) {
		if in.Weight <= 0 {
			return nil, weightLogOut{}, fmt.Errorf("weight must be positive, got %v", in.Weight)
		}
		log, err := d.HeaSvc.CreateWeightLog(ctx, d.UserUUID, health.CreateWeightInput{
			Weight:   in.Weight,
			LoggedAt: in.LoggedAt,
		})
		if err != nil {
			return nil, weightLogOut{}, fmt.Errorf("log weight: %w", err)
		}
		return nil, weightLogOut{Log: toWeightLogDTO(log)}, nil
	}))

	mcp.AddTool(s, &mcp.Tool{
		Name:        tools.HealthUpdateWeight.Name,
		Description: tools.HealthUpdateWeight.Description,
	}, middleware.MCPRequestLogger(d.Logger, tools.HealthUpdateWeight.Name, func(ctx context.Context, _ *mcp.CallToolRequest, in tools.HealthUpdateWeightIn) (*mcp.CallToolResult, weightLogOut, error) {
		logID, err := uuid.Parse(in.LogID)
		if err != nil {
			return nil, weightLogOut{}, fmt.Errorf("invalid log_id %q: %w", in.LogID, err)
		}
		if in.Weight <= 0 {
			return nil, weightLogOut{}, fmt.Errorf("weight must be positive, got %v", in.Weight)
		}
		log, err := d.HeaSvc.UpdateWeightLog(ctx, logID, d.UserUUID, health.UpdateWeightInput{
			Weight:   in.Weight,
			LoggedAt: in.LoggedAt,
		})
		if err != nil {
			return nil, weightLogOut{}, fmt.Errorf("update weight log: %w", err)
		}
		return nil, weightLogOut{Log: toWeightLogDTO(log)}, nil
	}))

	// Medicine
	registerMedicineTools(s, d)

	// Sleep
	mcp.AddTool(s, &mcp.Tool{
		Name:        tools.SleepListLogs.Name,
		Description: tools.SleepListLogs.Description,
	}, middleware.MCPRequestLogger(d.Logger, tools.SleepListLogs.Name, func(ctx context.Context, _ *mcp.CallToolRequest, in tools.SleepListLogsIn) (*mcp.CallToolResult, listSleepLogsOut, error) {
		logs, err := d.HeaSvc.ListSleepLogs(ctx, d.UserUUID)
		if err != nil {
			return nil, listSleepLogsOut{}, fmt.Errorf("list sleep logs: %w", err)
		}
		from := time.Time{}
		to := time.Now().UTC().Add(24 * time.Hour)
		if in.From != "" {
			t, err := time.Parse(time.DateOnly, in.From)
			if err != nil {
				return nil, listSleepLogsOut{}, fmt.Errorf("invalid from date: %w", err)
			}
			from = t
		}
		if in.To != "" {
			t, err := time.Parse(time.DateOnly, in.To)
			if err != nil {
				return nil, listSleepLogsOut{}, fmt.Errorf("invalid to date: %w", err)
			}
			to = t.Add(24 * time.Hour) // inclusive
		}
		dtos := make([]sleepLogDTO, 0, len(logs))
		for _, l := range logs {
			if (!from.IsZero() && l.LoggedAt.Before(from)) || l.LoggedAt.After(to) {
				continue
			}
			dtos = append(dtos, toSleepLogDTO(l))
		}
		return nil, listSleepLogsOut{Logs: dtos}, nil
	}))

	mcp.AddTool(s, &mcp.Tool{
		Name:        tools.SleepLogNight.Name,
		Description: tools.SleepLogNight.Description,
	}, middleware.MCPRequestLogger(d.Logger, tools.SleepLogNight.Name, func(ctx context.Context, _ *mcp.CallToolRequest, in tools.SleepLogNightIn) (*mcp.CallToolResult, sleepLogOut, error) {
		inp := health.CreateSleepInput{
			StartedAt: in.StartedAt,
			EndedAt:   in.EndedAt,
			Notes:     nilStr(in.Notes),
			LoggedAt:  in.LoggedAt,
		}
		if in.Score > 0 {
			s := in.Score
			inp.Score = &s
		}
		log, err := d.HeaSvc.CreateSleepLog(ctx, d.UserUUID, inp)
		if err != nil {
			return nil, sleepLogOut{}, fmt.Errorf("create sleep log: %w", err)
		}
		return nil, sleepLogOut{Log: toSleepLogDTO(log)}, nil
	}))

	mcp.AddTool(s, &mcp.Tool{
		Name:        tools.SleepUpdateNight.Name,
		Description: tools.SleepUpdateNight.Description,
	}, middleware.MCPRequestLogger(d.Logger, tools.SleepUpdateNight.Name, func(ctx context.Context, _ *mcp.CallToolRequest, in tools.SleepUpdateNightIn) (*mcp.CallToolResult, sleepLogOut, error) {
		logID, err := uuid.Parse(in.LogID)
		if err != nil {
			return nil, sleepLogOut{}, fmt.Errorf("invalid log_id %q: %w", in.LogID, err)
		}
		inp := health.UpdateSleepInput{
			StartedAt: in.StartedAt,
			EndedAt:   in.EndedAt,
			Notes:     nilStr(in.Notes),
			LoggedAt:  in.LoggedAt,
		}
		if in.Score > 0 {
			s := in.Score
			inp.Score = &s
		}
		log, err := d.HeaSvc.UpdateSleepLog(ctx, logID, d.UserUUID, inp)
		if err != nil {
			return nil, sleepLogOut{}, fmt.Errorf("update sleep log: %w", err)
		}
		return nil, sleepLogOut{Log: toSleepLogDTO(log)}, nil
	}))

	mcp.AddTool(s, &mcp.Tool{
		Name:        tools.SleepDeleteNight.Name,
		Description: tools.SleepDeleteNight.Description,
	}, middleware.MCPRequestLogger(d.Logger, tools.SleepDeleteNight.Name, func(ctx context.Context, _ *mcp.CallToolRequest, in tools.SleepDeleteNightIn) (*mcp.CallToolResult, deleteSleepOut, error) {
		logID, err := uuid.Parse(in.LogID)
		if err != nil {
			return nil, deleteSleepOut{}, fmt.Errorf("invalid log_id %q: %w", in.LogID, err)
		}
		if err := d.HeaSvc.DeleteSleepLog(ctx, logID, d.UserUUID); err != nil {
			return nil, deleteSleepOut{}, fmt.Errorf("delete sleep log: %w", err)
		}
		return nil, deleteSleepOut{Deleted: true}, nil
	}))
}

// ---- Medicine output types --------------------------------------------------

type medicineDTO struct {
	ID                string  `json:"id"`
	Name              string  `json:"name"`
	GenericName       *string `json:"generic_name,omitempty"`
	StockQuantity     float64 `json:"stock_quantity"`
	StockUnit         string  `json:"stock_unit"`
	DosageAmount      float64 `json:"dosage_amount"`
	DosageUnit        *string `json:"dosage_unit,omitempty"`
	FrequencyType     string  `json:"frequency_type"`
	FrequencyValue    *int    `json:"frequency_value,omitempty"`
	Timing            *string `json:"timing,omitempty"`
	LowStockThreshold float64 `json:"low_stock_threshold"`
	ArchivedAt        *string `json:"archived_at,omitempty"`
}

type medicineIntakeDTO struct {
	ID            string  `json:"id"`
	MedicineName  string  `json:"medicine_name"`
	TakenAt       string  `json:"taken_at"`
	QuantityTaken float64 `json:"quantity_taken"`
	StockBefore   float64 `json:"stock_before"`
	StockAfter    float64 `json:"stock_after"`
	Status        string  `json:"status"`
	Note          *string `json:"note,omitempty"`
	CreatedAt     string  `json:"created_at"`
}

type medicineStockAdjustmentDTO struct {
	ID           string  `json:"id"`
	MedicineName string  `json:"medicine_name,omitempty"`
	Type         string  `json:"type"`
	Quantity     float64 `json:"quantity"`
	StockBefore  float64 `json:"stock_before"`
	StockAfter   float64 `json:"stock_after"`
	Reason       *string `json:"reason,omitempty"`
	CreatedAt    string  `json:"created_at"`
}

type listMedicinesOut struct {
	Medicines []medicineDTO `json:"medicines"`
}

type medicineTakeOut struct {
	Intake   medicineIntakeDTO `json:"intake"`
	Medicine medicineDTO       `json:"medicine"`
}

type medicineAdjustStockOut struct {
	Adjustment medicineStockAdjustmentDTO `json:"adjustment"`
	Medicine   medicineDTO                `json:"medicine"`
}

type listMedicineIntakesOut struct {
	Intakes []medicineIntakeDTO `json:"intakes"`
}

type listMedicineStockAdjustmentsOut struct {
	Adjustments []medicineStockAdjustmentDTO `json:"adjustments"`
}

func toMedicineDTO(m *medicine.Medicine) medicineDTO {
	dto := medicineDTO{
		ID:                m.ID.String(),
		Name:              m.Name,
		GenericName:       m.GenericName,
		StockQuantity:     m.StockQuantity,
		StockUnit:         m.StockUnit,
		DosageAmount:      m.DosageAmount,
		DosageUnit:        m.DosageUnit,
		FrequencyType:     string(m.FrequencyType),
		FrequencyValue:    m.FrequencyValue,
		LowStockThreshold: m.LowStockThreshold,
	}
	if m.Timing != nil {
		t := string(*m.Timing)
		dto.Timing = &t
	}
	if m.ArchivedAt != nil {
		s := m.ArchivedAt.Format(time.RFC3339)
		dto.ArchivedAt = &s
	}
	return dto
}

func toMedicineIntakeDTO(i *medicine.MedicineIntake) medicineIntakeDTO {
	return medicineIntakeDTO{
		ID:            i.ID.String(),
		MedicineName:  i.MedicineName,
		TakenAt:       i.TakenAt.Format(time.RFC3339),
		QuantityTaken: i.QuantityTaken,
		StockBefore:   i.StockBefore,
		StockAfter:    i.StockAfter,
		Status:        string(i.Status),
		Note:          i.Note,
		CreatedAt:     i.CreatedAt.Format(time.RFC3339),
	}
}

func toMedicineStockAdjustmentDTO(a *medicine.MedicineStockAdjustment) medicineStockAdjustmentDTO {
	return medicineStockAdjustmentDTO{
		ID:          a.ID.String(),
		Type:        string(a.Type),
		Quantity:    a.Quantity,
		StockBefore: a.StockBefore,
		StockAfter:  a.StockAfter,
		Reason:      a.Reason,
		CreatedAt:   a.CreatedAt.Format(time.RFC3339),
	}
}

// registerMedicineTools appends Medicine tool handlers to the server.
func registerMedicineTools(s *mcp.Server, d Deps) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        tools.MedicineList.Name,
		Description: tools.MedicineList.Description,
	}, middleware.MCPRequestLogger(d.Logger, tools.MedicineList.Name, func(ctx context.Context, _ *mcp.CallToolRequest, in tools.MedicineListIn) (*mcp.CallToolResult, listMedicinesOut, error) {
		meds, err := d.MedSvc.ListMedicines(ctx, d.UserUUID, in.IncludeArchived)
		if err != nil {
			return nil, listMedicinesOut{}, fmt.Errorf("list medicines: %w", err)
		}
		dtos := make([]medicineDTO, len(meds))
		for i, m := range meds {
			dtos[i] = toMedicineDTO(m)
		}
		return nil, listMedicinesOut{Medicines: dtos}, nil
	}))

	mcp.AddTool(s, &mcp.Tool{
		Name:        tools.MedicineTake.Name,
		Description: tools.MedicineTake.Description,
	}, middleware.MCPRequestLogger(d.Logger, tools.MedicineTake.Name, func(ctx context.Context, _ *mcp.CallToolRequest, in tools.MedicineTakeIn) (*mcp.CallToolResult, medicineTakeOut, error) {
		if strings.TrimSpace(in.MedicineName) == "" {
			return nil, medicineTakeOut{}, fmt.Errorf("medicine_name is required")
		}
		meds, err := d.MedSvc.ListMedicines(ctx, d.UserUUID, false)
		if err != nil {
			return nil, medicineTakeOut{}, fmt.Errorf("list medicines: %w", err)
		}
		var target *medicine.Medicine
		for _, m := range meds {
			if strings.EqualFold(m.Name, in.MedicineName) {
				target = m
				break
			}
		}
		if target == nil {
			return nil, medicineTakeOut{}, fmt.Errorf("no active medicine named %q — call medicine_list to see available names", in.MedicineName)
		}
		qty := in.QuantityTaken
		if qty <= 0 {
			qty = target.DosageAmount
		}
		inp := medicine.TakeMedicineInput{
			QuantityTaken: qty,
			AllowNegative: in.AllowNegative,
		}
		if in.Note != "" {
			inp.Note = &in.Note
		}
		intake, med, err := d.MedSvc.Take(ctx, d.UserUUID, target.ID, inp)
		if err != nil {
			return nil, medicineTakeOut{}, fmt.Errorf("take medicine: %w", err)
		}
		return nil, medicineTakeOut{Intake: toMedicineIntakeDTO(intake), Medicine: toMedicineDTO(med)}, nil
	}))

	mcp.AddTool(s, &mcp.Tool{
		Name:        tools.MedicineAdjustStock.Name,
		Description: tools.MedicineAdjustStock.Description,
	}, middleware.MCPRequestLogger(d.Logger, tools.MedicineAdjustStock.Name, func(ctx context.Context, _ *mcp.CallToolRequest, in tools.MedicineAdjustStockIn) (*mcp.CallToolResult, medicineAdjustStockOut, error) {
		if strings.TrimSpace(in.MedicineName) == "" {
			return nil, medicineAdjustStockOut{}, fmt.Errorf("medicine_name is required")
		}
		adjType := medicine.AdjustmentType(in.Type)
		switch adjType {
		case medicine.AdjustmentTypeAdd, medicine.AdjustmentTypeRemove, medicine.AdjustmentTypeCorrection:
		default:
			return nil, medicineAdjustStockOut{}, fmt.Errorf("type must be add, remove, or correction, got %q", in.Type)
		}
		if in.Quantity <= 0 {
			return nil, medicineAdjustStockOut{}, fmt.Errorf("quantity must be greater than 0")
		}
		meds, err := d.MedSvc.ListMedicines(ctx, d.UserUUID, false)
		if err != nil {
			return nil, medicineAdjustStockOut{}, fmt.Errorf("list medicines: %w", err)
		}
		var target *medicine.Medicine
		for _, m := range meds {
			if strings.EqualFold(m.Name, in.MedicineName) {
				target = m
				break
			}
		}
		if target == nil {
			return nil, medicineAdjustStockOut{}, fmt.Errorf("no active medicine named %q — call medicine_list to see available names", in.MedicineName)
		}
		inp := medicine.AdjustStockInput{
			Type:     adjType,
			Quantity: in.Quantity,
		}
		if in.Reason != "" {
			inp.Reason = &in.Reason
		}
		adj, med, err := d.MedSvc.AdjustStock(ctx, d.UserUUID, target.ID, inp)
		if err != nil {
			return nil, medicineAdjustStockOut{}, fmt.Errorf("adjust stock: %w", err)
		}
		return nil, medicineAdjustStockOut{Adjustment: toMedicineStockAdjustmentDTO(adj), Medicine: toMedicineDTO(med)}, nil
	}))

	mcp.AddTool(s, &mcp.Tool{
		Name:        tools.MedicineListIntakes.Name,
		Description: tools.MedicineListIntakes.Description,
	}, middleware.MCPRequestLogger(d.Logger, tools.MedicineListIntakes.Name, func(ctx context.Context, _ *mcp.CallToolRequest, in tools.MedicineListIntakesIn) (*mcp.CallToolResult, listMedicineIntakesOut, error) {
		opts := medicine.ListIntakeOpts{Limit: in.Limit}
		if in.Date != "" {
			t, err := time.Parse(time.DateOnly, in.Date)
			if err != nil {
				return nil, listMedicineIntakesOut{}, fmt.Errorf("invalid date %q (use YYYY-MM-DD): %w", in.Date, err)
			}
			opts.Date = &t
		}
		intakes, err := d.MedSvc.ListIntakes(ctx, d.UserUUID, opts)
		if err != nil {
			return nil, listMedicineIntakesOut{}, fmt.Errorf("list intakes: %w", err)
		}
		dtos := make([]medicineIntakeDTO, len(intakes))
		for i, it := range intakes {
			dtos[i] = toMedicineIntakeDTO(it)
		}
		return nil, listMedicineIntakesOut{Intakes: dtos}, nil
	}))

	mcp.AddTool(s, &mcp.Tool{
		Name:        tools.MedicineListStockAdjustments.Name,
		Description: tools.MedicineListStockAdjustments.Description,
	}, middleware.MCPRequestLogger(d.Logger, tools.MedicineListStockAdjustments.Name, func(ctx context.Context, _ *mcp.CallToolRequest, in tools.MedicineListStockAdjustmentsIn) (*mcp.CallToolResult, listMedicineStockAdjustmentsOut, error) {
		opts := medicine.ListAdjustmentOpts{Limit: in.Limit}
		if in.Date != "" {
			t, err := time.Parse(time.DateOnly, in.Date)
			if err != nil {
				return nil, listMedicineStockAdjustmentsOut{}, fmt.Errorf("invalid date %q (use YYYY-MM-DD): %w", in.Date, err)
			}
			opts.Date = &t
		}
		adjs, err := d.MedSvc.ListStockAdjustments(ctx, d.UserUUID, opts)
		if err != nil {
			return nil, listMedicineStockAdjustmentsOut{}, fmt.Errorf("list stock adjustments: %w", err)
		}
		dtos := make([]medicineStockAdjustmentDTO, len(adjs))
		for i, a := range adjs {
			dtos[i] = toMedicineStockAdjustmentDTO(a)
		}
		return nil, listMedicineStockAdjustmentsOut{Adjustments: dtos}, nil
	}))
}

// nilStr returns nil when s is empty, otherwise a pointer to s.
func nilStr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
