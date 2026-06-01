package mcpserver

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/kinkando/personal-dashboard/internal/finance"
	financeSvc "github.com/kinkando/personal-dashboard/internal/finance/service"
	"github.com/kinkando/personal-dashboard/internal/kanban"
	kanbanRepo "github.com/kinkando/personal-dashboard/internal/kanban/repository"
	"github.com/kinkando/personal-dashboard/internal/tools"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
)

// Deps bundles the shared dependencies consumed by all MCP tool handlers.
type Deps struct {
	FinSvc      *financeSvc.Service
	KanRepo     *kanbanRepo.Repository
	UserUUID    uuid.UUID
	FirebaseUID string
	Logger      *zap.Logger
}

// withLog wraps an MCP tool handler with structured logging (input, output, latency, errors).
func withLog[In, Out any](logger *zap.Logger, name string, fn func(context.Context, *mcp.CallToolRequest, In) (*mcp.CallToolResult, Out, error)) func(context.Context, *mcp.CallToolRequest, In) (*mcp.CallToolResult, Out, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, in In) (*mcp.CallToolResult, Out, error) {
		start := time.Now()
		inJSON, _ := json.Marshal(in)
		result, out, err := fn(ctx, req, in)
		latency := time.Since(start)
		args := []zap.Field{
			zap.String("tool", name),
			zap.Duration("latency", latency),
		}
		if len(inJSON) > 0 {
			args = append(args, zap.Any("input", in))
		}
		if err != nil {
			args = append(args, zap.Error(err))
			logger.Error("mcp", args...)
		} else {
			outJSON, _ := json.Marshal(out)
			if len(outJSON) > 0 {
				args = append(args, zap.Any("output", out))
			}
			logger.Info("mcp", args...)
		}
		return result, out, err
	}
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

type monthlySummaryOut struct {
	Summary *finance.MonthlySummary `json:"summary"`
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
		Date:      r.Date.Format("2006-01-02"),
		CreatedAt: r.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
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
		CreatedAt: b.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
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
		CreatedAt: c.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
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
		CreatedAt:     c.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
	if c.DueDate != nil {
		s := c.DueDate.Format("2006-01-02")
		dto.DueDate = &s
	}
	if c.CompletedAt != nil {
		s := c.CompletedAt.Format("2006-01-02T15:04:05Z07:00")
		dto.CompletedAt = &s
	}
	if c.ArchivedAt != nil {
		s := c.ArchivedAt.Format("2006-01-02T15:04:05Z07:00")
		dto.ArchivedAt = &s
	}
	return dto
}

// ---- Tool registration ------------------------------------------------------

func registerTools(s *mcp.Server, d Deps) {
	// Finance
	mcp.AddTool(s, &mcp.Tool{
		Name:        tools.FinanceListRecords.Name,
		Description: tools.FinanceListRecords.Description,
	}, withLog(d.Logger, tools.FinanceListRecords.Name, func(ctx context.Context, _ *mcp.CallToolRequest, in tools.ListRecordsIn) (*mcp.CallToolResult, listRecordsOut, error) {
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
	}, withLog(d.Logger, tools.FinanceListCategories.Name, func(ctx context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, listCategoriesOut, error) {
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
	}, withLog(d.Logger, tools.FinanceCreateCategory.Name, func(ctx context.Context, _ *mcp.CallToolRequest, in tools.CreateCategoryIn) (*mcp.CallToolResult, createCategoryOut, error) {
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
	}, withLog(d.Logger, tools.FinanceDeleteCategory.Name, func(ctx context.Context, _ *mcp.CallToolRequest, in tools.DeleteCategoryIn) (*mcp.CallToolResult, deleteCategoryOut, error) {
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
	}, withLog(d.Logger, tools.FinanceCreateRecord.Name, func(ctx context.Context, _ *mcp.CallToolRequest, in tools.CreateRecordIn) (*mcp.CallToolResult, createRecordOut, error) {
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
	}, withLog(d.Logger, tools.FinanceDeleteRecord.Name, func(ctx context.Context, _ *mcp.CallToolRequest, in tools.DeleteRecordIn) (*mcp.CallToolResult, deleteRecordOut, error) {
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
	}, withLog(d.Logger, tools.FinanceMonthlySummary.Name, func(ctx context.Context, _ *mcp.CallToolRequest, in tools.MonthlySummaryIn) (*mcp.CallToolResult, monthlySummaryOut, error) {
		if in.Month == "" {
			return nil, monthlySummaryOut{}, fmt.Errorf("month is required (YYYY-MM)")
		}
		summary, err := d.FinSvc.MonthlySummary(ctx, d.UserUUID, in.Month)
		if err != nil {
			return nil, monthlySummaryOut{}, fmt.Errorf("monthly summary: %w", err)
		}
		return nil, monthlySummaryOut{Summary: summary}, nil
	}))

	// Kanban
	mcp.AddTool(s, &mcp.Tool{
		Name:        tools.KanbanGetBoard.Name,
		Description: tools.KanbanGetBoard.Description,
	}, withLog(d.Logger, tools.KanbanGetBoard.Name, func(ctx context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, getBoardOut, error) {
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
	}, withLog(d.Logger, tools.KanbanCreateCard.Name, func(ctx context.Context, _ *mcp.CallToolRequest, in tools.CreateCardIn) (*mcp.CallToolResult, createCardOut, error) {
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
	}, withLog(d.Logger, tools.KanbanUpdateCard.Name, func(ctx context.Context, _ *mcp.CallToolRequest, in tools.UpdateCardIn) (*mcp.CallToolResult, updateCardOut, error) {
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
	}, withLog(d.Logger, tools.KanbanMoveCard.Name, func(ctx context.Context, _ *mcp.CallToolRequest, in tools.MoveCardIn) (*mcp.CallToolResult, moveCardOut, error) {
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
	}, withLog(d.Logger, tools.KanbanDeleteCard.Name, func(ctx context.Context, _ *mcp.CallToolRequest, in tools.DeleteCardIn) (*mcp.CallToolResult, deleteCardOut, error) {
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
	}, withLog(d.Logger, tools.KanbanBoardStats.Name, func(ctx context.Context, _ *mcp.CallToolRequest, in tools.BoardStatsIn) (*mcp.CallToolResult, boardStatsOut, error) {
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
	}, withLog(d.Logger, tools.KanbanArchiveCard.Name, func(ctx context.Context, _ *mcp.CallToolRequest, in tools.ArchiveCardIn) (*mcp.CallToolResult, archiveCardOut, error) {
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
	}, withLog(d.Logger, tools.KanbanUnarchiveCard.Name, func(ctx context.Context, _ *mcp.CallToolRequest, in tools.UnarchiveCardIn) (*mcp.CallToolResult, unarchiveCardOut, error) {
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
	}, withLog(d.Logger, tools.KanbanListArchivedCards.Name, func(ctx context.Context, _ *mcp.CallToolRequest, in tools.ListArchivedCardsIn) (*mcp.CallToolResult, listArchivedCardsOut, error) {
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
}
