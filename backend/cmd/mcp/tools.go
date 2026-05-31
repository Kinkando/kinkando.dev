package main

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/kinkando/personal-dashboard/internal/finance"
	financeSvc "github.com/kinkando/personal-dashboard/internal/finance/service"
	"github.com/kinkando/personal-dashboard/internal/kanban"
	kanbanRepo "github.com/kinkando/personal-dashboard/internal/kanban/repository"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// deps bundles the shared dependencies consumed by all MCP tool handlers.
type deps struct {
	finSvc      *financeSvc.Service
	kanRepo     *kanbanRepo.Repository
	userUUID    uuid.UUID
	firebaseUID string
}

// ---- Finance types --------------------------------------------------------

type listRecordsIn struct {
	Month string `json:"month" jsonschema:"description=Month to query in YYYY-MM format (e.g. 2026-05)"`
}
type listRecordsOut struct {
	Records []*finance.Record `json:"records"`
}

type createRecordIn struct {
	Type     string  `json:"type"     jsonschema:"description=Record type: income or expense"`
	Amount   float64 `json:"amount"   jsonschema:"description=Amount (must be a positive number)"`
	Category string  `json:"category" jsonschema:"description=Category name"`
	Note     string  `json:"note"     jsonschema:"description=Optional note"`
	Date     string  `json:"date"     jsonschema:"description=Date in YYYY-MM-DD format"`
}
type createRecordOut struct {
	Record *finance.Record `json:"record"`
}

type deleteRecordIn struct {
	ID string `json:"id" jsonschema:"description=UUID of the finance record to delete"`
}
type deleteRecordOut struct {
	Deleted bool `json:"deleted"`
}

type monthlySummaryIn struct {
	Month string `json:"month" jsonschema:"description=Month in YYYY-MM format (e.g. 2026-05)"`
}
type monthlySummaryOut struct {
	Summary *finance.MonthlySummary `json:"summary"`
}

// ---- Kanban types ---------------------------------------------------------

type getBoardOut struct {
	Board   *kanban.Board    `json:"board"`
	Columns []*kanban.Column `json:"columns"`
	Cards   []*kanban.Card   `json:"cards"`
}

type createCardIn struct {
	ColumnID string `json:"column_id" jsonschema:"description=Target column ID (hex ObjectID) — get it from kanban_get_board"`
	Title    string `json:"title"     jsonschema:"description=Card title (required)"`
	Content  string `json:"content"   jsonschema:"description=Card body text (optional)"`
}
type createCardOut struct {
	Card *kanban.Card `json:"card"`
}

type moveCardIn struct {
	CardID   string `json:"card_id"   jsonschema:"description=Card ID (hex ObjectID) to move"`
	ColumnID string `json:"column_id" jsonschema:"description=Target column ID (hex ObjectID)"`
	Order    int    `json:"order"     jsonschema:"description=New 0-indexed position within the target column"`
}
type moveCardOut struct {
	Moved bool `json:"moved"`
}

type deleteCardIn struct {
	CardID string `json:"card_id" jsonschema:"description=Card ID (hex ObjectID) to delete"`
}
type deleteCardOut struct {
	Deleted bool `json:"deleted"`
}

// ---- Registration ---------------------------------------------------------

// registerTools adds all kanban and finance MCP tools to the server, binding
// each handler to the supplied dependencies via closure.
func registerTools(s *mcp.Server, d deps) {
	// Finance
	mcp.AddTool(s, &mcp.Tool{
		Name:        "finance_list_records",
		Description: "List all finance records (income/expense) for a given month.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in listRecordsIn) (*mcp.CallToolResult, listRecordsOut, error) {
		if in.Month == "" {
			return nil, listRecordsOut{}, fmt.Errorf("month is required (YYYY-MM)")
		}
		records, err := d.finSvc.ListRecords(ctx, d.userUUID, in.Month)
		if err != nil {
			return nil, listRecordsOut{}, fmt.Errorf("list records: %w", err)
		}
		if records == nil {
			records = []*finance.Record{}
		}
		return nil, listRecordsOut{Records: records}, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "finance_create_record",
		Description: "Create a new income or expense record.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in createRecordIn) (*mcp.CallToolResult, createRecordOut, error) {
		rt := finance.RecordType(in.Type)
		if rt != finance.RecordTypeIncome && rt != finance.RecordTypeExpense {
			return nil, createRecordOut{}, fmt.Errorf("type must be %q or %q, got %q", finance.RecordTypeIncome, finance.RecordTypeExpense, in.Type)
		}
		if in.Amount <= 0 {
			return nil, createRecordOut{}, fmt.Errorf("amount must be positive, got %v", in.Amount)
		}
		rec, err := d.finSvc.CreateRecord(ctx, d.userUUID, finance.CreateRecordInput{
			Type:     rt,
			Amount:   in.Amount,
			Category: in.Category,
			Note:     in.Note,
			Date:     in.Date,
		})
		if err != nil {
			return nil, createRecordOut{}, fmt.Errorf("create record: %w", err)
		}
		return nil, createRecordOut{Record: rec}, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "finance_delete_record",
		Description: "Delete a finance record by its UUID.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in deleteRecordIn) (*mcp.CallToolResult, deleteRecordOut, error) {
		id, err := uuid.Parse(in.ID)
		if err != nil {
			return nil, deleteRecordOut{}, fmt.Errorf("invalid record id %q: %w", in.ID, err)
		}
		if err := d.finSvc.DeleteRecord(ctx, id, d.userUUID); err != nil {
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
		summary, err := d.finSvc.MonthlySummary(ctx, d.userUUID, in.Month)
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
		board, err := d.kanRepo.GetBoard(ctx, d.firebaseUID)
		if err != nil {
			return nil, getBoardOut{}, fmt.Errorf("get board: %w", err)
		}
		columns, err := d.kanRepo.GetColumns(ctx, board.ID)
		if err != nil {
			return nil, getBoardOut{}, fmt.Errorf("get columns: %w", err)
		}
		cards, err := d.kanRepo.GetCards(ctx, board.ID)
		if err != nil {
			return nil, getBoardOut{}, fmt.Errorf("get cards: %w", err)
		}
		if columns == nil {
			columns = []*kanban.Column{}
		}
		if cards == nil {
			cards = []*kanban.Card{}
		}
		return nil, getBoardOut{Board: board, Columns: columns, Cards: cards}, nil
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
		board, err := d.kanRepo.GetBoard(ctx, d.firebaseUID)
		if err != nil {
			return nil, createCardOut{}, fmt.Errorf("get board: %w", err)
		}
		card, err := d.kanRepo.CreateCard(ctx, board.ID, colID, kanban.CreateCardInput{
			ColumnID: in.ColumnID,
			Title:    in.Title,
			Content:  in.Content,
		})
		if err != nil {
			return nil, createCardOut{}, fmt.Errorf("create card: %w", err)
		}
		return nil, createCardOut{Card: card}, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "kanban_move_card",
		Description: "Move a card to a different column and/or position.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in moveCardIn) (*mcp.CallToolResult, moveCardOut, error) {
		cardID, err := primitive.ObjectIDFromHex(in.CardID)
		if err != nil {
			return nil, moveCardOut{}, fmt.Errorf("invalid card_id %q: %w", in.CardID, err)
		}
		if err := d.kanRepo.MoveCard(ctx, cardID, kanban.MoveCardInput{
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
		if err := d.kanRepo.DeleteCard(ctx, cardID); err != nil {
			return nil, deleteCardOut{}, fmt.Errorf("delete card: %w", err)
		}
		return nil, deleteCardOut{Deleted: true}, nil
	})
}
