package mcpserver

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/kinkando/personal-dashboard/internal/finance"
	financeSvc "github.com/kinkando/personal-dashboard/internal/finance/service"
	"github.com/kinkando/personal-dashboard/internal/kanban"
	kanbanRepo "github.com/kinkando/personal-dashboard/internal/kanban/repository"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Deps bundles the shared dependencies consumed by all MCP tool handlers.
type Deps struct {
	FinSvc      *financeSvc.Service
	KanRepo     *kanbanRepo.Repository
	UserUUID    uuid.UUID
	FirebaseUID string
}

// New creates a new MCP server with all tools registered.
func New(d Deps) *mcp.Server {
	s := mcp.NewServer(&mcp.Implementation{Name: "kinkando-dashboard", Version: "0.1.0"}, nil)
	registerTools(s, d)
	return s
}

// ---- Finance types --------------------------------------------------------

type recordDTO struct {
	ID           string  `json:"id"`
	UserID       string  `json:"user_id"`
	Type         string  `json:"type"`
	Amount       float64 `json:"amount"`
	CategoryID   *string `json:"category_id"`
	CategoryName string  `json:"category_name"`
	Note         string  `json:"note"`
	Date         string  `json:"date"`
	CreatedAt    string  `json:"created_at"`
}

type categoryDTO struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Type  string `json:"type"`
	Icon  string `json:"icon"`
	Color string `json:"color"`
}

type listRecordsIn struct {
	Month string `json:"month" jsonschema:"Month to query in YYYY-MM format (e.g. 2026-05)"`
}
type listRecordsOut struct {
	Records []recordDTO `json:"records"`
}

type listCategoriesOut struct {
	Categories []categoryDTO `json:"categories"`
}

type createCategoryIn struct {
	Name  string `json:"name"  jsonschema:"Category name (required)"`
	Type  string `json:"type"  jsonschema:"Category type: income or expense"`
	Icon  string `json:"icon"  jsonschema:"Lucide icon name (e.g. ShoppingCart, Utensils, Home, Car, Briefcase, Gift, HeartPulse, Plane, Wallet, PiggyBank, Coffee, Gamepad2, Zap, TrendingUp, CircleDollarSign)"`
	Color string `json:"color" jsonschema:"Hex color code (e.g. #6366f1)"`
}
type createCategoryOut struct {
	Category categoryDTO `json:"category"`
}

type createRecordIn struct {
	Type     string  `json:"type"     jsonschema:"Record type: income or expense"`
	Amount   float64 `json:"amount"   jsonschema:"Amount (must be a positive number)"`
	Category string  `json:"category" jsonschema:"Category name — must match an existing category of the given type (call finance_list_categories first)"`
	Note     string  `json:"note"     jsonschema:"Optional note"`
	Date     string  `json:"date"     jsonschema:"Date in YYYY-MM-DD format"`
}
type createRecordOut struct {
	Record recordDTO `json:"record"`
}

type deleteCategoryIn struct {
	ID string `json:"id" jsonschema:"UUID of the finance category to delete"`
}
type deleteCategoryOut struct {
	Deleted bool `json:"deleted"`
}

type deleteRecordIn struct {
	ID string `json:"id" jsonschema:"UUID of the finance record to delete"`
}
type deleteRecordOut struct {
	Deleted bool `json:"deleted"`
}

type monthlySummaryIn struct {
	Month string `json:"month" jsonschema:"Month in YYYY-MM format (e.g. 2026-05)"`
}
type monthlySummaryOut struct {
	Summary *finance.MonthlySummary `json:"summary"`
}

// ---- Kanban types ---------------------------------------------------------

type boardDTO struct {
	ID        string `json:"id"`
	UserID    string `json:"user_id"`
	CreatedAt string `json:"created_at"`
}

type columnDTO struct {
	ID        string `json:"id"`
	BoardID   string `json:"board_id"`
	Name      string `json:"name"`
	Order     int    `json:"order"`
	CreatedAt string `json:"created_at"`
}

type cardDTO struct {
	ID        string `json:"id"`
	BoardID   string `json:"board_id"`
	ColumnID  string `json:"column_id"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	Order     int    `json:"order"`
	CreatedAt string `json:"created_at"`
}

type getBoardOut struct {
	Board   boardDTO    `json:"board"`
	Columns []columnDTO `json:"columns"`
	Cards   []cardDTO   `json:"cards"`
}

type createCardIn struct {
	ColumnID string `json:"column_id" jsonschema:"Target column ID (hex ObjectID) — get it from kanban_get_board"`
	Title    string `json:"title"     jsonschema:"Card title (required)"`
	Content  string `json:"content"   jsonschema:"Card body text (optional)"`
}
type createCardOut struct {
	Card cardDTO `json:"card"`
}

type moveCardIn struct {
	CardID   string `json:"card_id"   jsonschema:"Card ID (hex ObjectID) to move"`
	ColumnID string `json:"column_id" jsonschema:"Target column ID (hex ObjectID)"`
	Order    int    `json:"order"     jsonschema:"New 0-indexed position within the target column"`
}
type moveCardOut struct {
	Moved bool `json:"moved"`
}

type deleteCardIn struct {
	CardID string `json:"card_id" jsonschema:"Card ID (hex ObjectID) to delete"`
}
type deleteCardOut struct {
	Deleted bool `json:"deleted"`
}

// ---- Registration ---------------------------------------------------------

func toRecordDTO(r *finance.Record) recordDTO {
	dto := recordDTO{
		ID:           r.ID.String(),
		UserID:       r.UserID.String(),
		Type:         string(r.Type),
		Amount:       r.Amount,
		CategoryName: r.CategoryName,
		Note:         r.Note,
		Date:         r.Date.Format("2006-01-02"),
		CreatedAt:    r.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
	if r.CategoryID != nil {
		s := r.CategoryID.String()
		dto.CategoryID = &s
	}
	if r.Category != nil {
		dto.CategoryName = r.Category.Name
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
		CreatedAt: b.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func toColumnDTO(c *kanban.Column) columnDTO {
	return columnDTO{
		ID:        c.ID.Hex(),
		BoardID:   c.BoardID.Hex(),
		Name:      c.Name,
		Order:     c.Order,
		CreatedAt: c.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func toCardDTO(c *kanban.Card) cardDTO {
	return cardDTO{
		ID:        c.ID.Hex(),
		BoardID:   c.BoardID.Hex(),
		ColumnID:  c.ColumnID.Hex(),
		Title:     c.Title,
		Content:   c.Content,
		Order:     c.Order,
		CreatedAt: c.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func registerTools(s *mcp.Server, d Deps) {
	// Finance
	mcp.AddTool(s, &mcp.Tool{
		Name:        "finance_list_records",
		Description: "List all finance records (income/expense) for a given month.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in listRecordsIn) (*mcp.CallToolResult, listRecordsOut, error) {
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
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "finance_list_categories",
		Description: "List all finance categories (income/expense) for the user. Call this before finance_create_record to get valid category names.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, listCategoriesOut, error) {
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
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "finance_create_category",
		Description: "Create a new finance category. Icon must be a valid lucide-react icon name from the registry (e.g. ShoppingCart, Utensils, Home, Car, Briefcase, Gift, HeartPulse, Plane, Wallet, PiggyBank, Coffee, Gamepad2, Zap, TrendingUp, CircleDollarSign).",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in createCategoryIn) (*mcp.CallToolResult, createCategoryOut, error) {
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
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "finance_delete_category",
		Description: "Delete a finance category by its UUID.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in deleteCategoryIn) (*mcp.CallToolResult, deleteCategoryOut, error) {
		id, err := uuid.Parse(in.ID)
		if err != nil {
			return nil, deleteCategoryOut{}, fmt.Errorf("invalid category id %q: %w", in.ID, err)
		}
		if err := d.FinSvc.DeleteCategory(ctx, id, d.UserUUID); err != nil {
			return nil, deleteCategoryOut{}, fmt.Errorf("delete category: %w", err)
		}
		return nil, deleteCategoryOut{Deleted: true}, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "finance_create_record",
		Description: "Create a new income or expense record. The 'category' field must match an existing category name of the given type — call finance_list_categories first to see available names.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in createRecordIn) (*mcp.CallToolResult, createRecordOut, error) {
		rt := finance.RecordType(in.Type)
		if rt != finance.RecordTypeIncome && rt != finance.RecordTypeExpense {
			return nil, createRecordOut{}, fmt.Errorf("type must be %q or %q, got %q", finance.RecordTypeIncome, finance.RecordTypeExpense, in.Type)
		}
		if in.Amount <= 0 {
			return nil, createRecordOut{}, fmt.Errorf("amount must be positive, got %v", in.Amount)
		}
		// Resolve category name → ID
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
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "finance_delete_record",
		Description: "Delete a finance record by its UUID.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in deleteRecordIn) (*mcp.CallToolResult, deleteRecordOut, error) {
		id, err := uuid.Parse(in.ID)
		if err != nil {
			return nil, deleteRecordOut{}, fmt.Errorf("invalid record id %q: %w", in.ID, err)
		}
		if err := d.FinSvc.DeleteRecord(ctx, id, d.UserUUID); err != nil {
			return nil, deleteRecordOut{}, fmt.Errorf("delete record: %w", err)
		}
		return nil, deleteRecordOut{Deleted: true}, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "finance_monthly_summary",
		Description: "Get a monthly summary of income, expenses, net, and per-category breakdown.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in monthlySummaryIn) (*mcp.CallToolResult, monthlySummaryOut, error) {
		if in.Month == "" {
			return nil, monthlySummaryOut{}, fmt.Errorf("month is required (YYYY-MM)")
		}
		summary, err := d.FinSvc.MonthlySummary(ctx, d.UserUUID, in.Month)
		if err != nil {
			return nil, monthlySummaryOut{}, fmt.Errorf("monthly summary: %w", err)
		}
		return nil, monthlySummaryOut{Summary: summary}, nil
	})

	// Kanban
	mcp.AddTool(s, &mcp.Tool{
		Name:        "kanban_get_board",
		Description: "Retrieve the full kanban board: board metadata, all columns, and all cards. Auto-creates the board on first call.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, getBoardOut, error) {
		board, err := d.KanRepo.GetBoard(ctx, d.FirebaseUID)
		if err != nil {
			return nil, getBoardOut{}, fmt.Errorf("get board: %w", err)
		}
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
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "kanban_create_card",
		Description: "Create a new card in a column. Use kanban_get_board to get column IDs first.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in createCardIn) (*mcp.CallToolResult, createCardOut, error) {
		if in.Title == "" {
			return nil, createCardOut{}, fmt.Errorf("title is required")
		}
		colID, err := primitive.ObjectIDFromHex(in.ColumnID)
		if err != nil {
			return nil, createCardOut{}, fmt.Errorf("invalid column_id %q: %w", in.ColumnID, err)
		}
		board, err := d.KanRepo.GetBoard(ctx, d.FirebaseUID)
		if err != nil {
			return nil, createCardOut{}, fmt.Errorf("get board: %w", err)
		}
		card, err := d.KanRepo.CreateCard(ctx, board.ID, colID, kanban.CreateCardInput{
			ColumnID: in.ColumnID,
			Title:    in.Title,
			Content:  in.Content,
		})
		if err != nil {
			return nil, createCardOut{}, fmt.Errorf("create card: %w", err)
		}
		return nil, createCardOut{Card: toCardDTO(card)}, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "kanban_move_card",
		Description: "Move a card to a different column and/or position.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in moveCardIn) (*mcp.CallToolResult, moveCardOut, error) {
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
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "kanban_delete_card",
		Description: "Delete a kanban card by its ID.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in deleteCardIn) (*mcp.CallToolResult, deleteCardOut, error) {
		cardID, err := primitive.ObjectIDFromHex(in.CardID)
		if err != nil {
			return nil, deleteCardOut{}, fmt.Errorf("invalid card_id %q: %w", in.CardID, err)
		}
		if err := d.KanRepo.DeleteCard(ctx, cardID); err != nil {
			return nil, deleteCardOut{}, fmt.Errorf("delete card: %w", err)
		}
		return nil, deleteCardOut{Deleted: true}, nil
	})
}
