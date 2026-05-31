package gemini

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/generative-ai-go/genai"
	"github.com/google/uuid"
	"github.com/kinkando/personal-dashboard/internal/finance"
	financeSvc "github.com/kinkando/personal-dashboard/internal/finance/service"
	"github.com/kinkando/personal-dashboard/internal/kanban"
	kanbanRepo "github.com/kinkando/personal-dashboard/internal/kanban/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/api/option"
)

const defaultModel = "gemini-2.0-flash"

const systemInstruction = `You are a personal dashboard assistant for finance tracking and kanban task management.
Reply concisely in the same language the user writes in.
Always use tools to read or write data — never fabricate records or IDs.
When creating a finance record, call finance_list_categories first unless you already know the exact category name.
When creating a kanban card, call kanban_get_board first unless you already have the column ID.`

// Deps bundles dependencies for the Gemini client.
type Deps struct {
	APIKey      string
	Model       string
	FinSvc      *financeSvc.Service
	KanRepo     *kanbanRepo.Repository
	UserUUID    uuid.UUID
	FirebaseUID string
}

// Client wraps a Gemini generative model and dispatches tool calls to the dashboard services.
type Client struct {
	gc          *genai.Client
	model       *genai.GenerativeModel
	finSvc      *financeSvc.Service
	kanRepo     *kanbanRepo.Repository
	userUUID    uuid.UUID
	firebaseUID string
}

// New creates a Client. Returns an error if the Gemini API key is invalid or unreachable.
func New(ctx context.Context, d Deps) (*Client, error) {
	gc, err := genai.NewClient(ctx, option.WithAPIKey(d.APIKey))
	if err != nil {
		return nil, fmt.Errorf("gemini: new client: %w", err)
	}
	modelName := d.Model
	if modelName == "" {
		modelName = defaultModel
	}
	model := gc.GenerativeModel(modelName)
	model.Tools = []*genai.Tool{{FunctionDeclarations: AllTools()}}
	model.SystemInstruction = &genai.Content{
		Parts: []genai.Part{genai.Text(systemInstruction)},
	}
	return &Client{
		gc:          gc,
		model:       model,
		finSvc:      d.FinSvc,
		kanRepo:     d.KanRepo,
		userUUID:    d.UserUUID,
		firebaseUID: d.FirebaseUID,
	}, nil
}

// Close releases resources held by the underlying Gemini API client.
func (c *Client) Close() error { return c.gc.Close() }

// Chat sends userMsg to Gemini, executes any tool calls in a loop, and returns
// the final text reply.
func (c *Client) Chat(ctx context.Context, userMsg string) (string, error) {
	msg := fmt.Sprintf("[Today: %s]\n%s", time.Now().Format("2006-01-02"), userMsg)
	session := c.model.StartChat()
	parts := []genai.Part{genai.Text(msg)}

	for {
		resp, err := session.SendMessage(ctx, parts...)
		if err != nil {
			return "", fmt.Errorf("gemini: send message: %w", err)
		}
		if len(resp.Candidates) == 0 {
			return "", errors.New("gemini: no candidates in response")
		}

		var text string
		var calls []genai.FunctionCall
		for _, p := range resp.Candidates[0].Content.Parts {
			switch v := p.(type) {
			case genai.Text:
				text = string(v)
			case genai.FunctionCall:
				calls = append(calls, v)
			}
		}

		if len(calls) == 0 {
			return text, nil
		}

		parts = nil
		for _, fc := range calls {
			result, dispErr := c.dispatch(ctx, fc.Name, fc.Args)
			var response map[string]any
			if dispErr != nil {
				response = map[string]any{"error": dispErr.Error()}
			} else {
				response = map[string]any{"result": result}
			}
			parts = append(parts, genai.FunctionResponse{
				Name:     fc.Name,
				Response: response,
			})
		}
	}
}

func (c *Client) dispatch(ctx context.Context, name string, args map[string]any) (any, error) {
	switch name {
	case "finance_list_records":
		return c.finSvc.ListRecords(ctx, c.userUUID, strArg(args, "month"))

	case "finance_list_categories":
		return c.finSvc.ListCategories(ctx, c.userUUID)

	case "finance_create_category":
		return c.finSvc.CreateCategory(ctx, c.userUUID, finance.CreateCategoryInput{
			Name:  strArg(args, "name"),
			Type:  finance.RecordType(strArg(args, "type")),
			Icon:  strArg(args, "icon"),
			Color: strArg(args, "color"),
		})

	case "finance_delete_category":
		id, err := uuid.Parse(strArg(args, "id"))
		if err != nil {
			return nil, fmt.Errorf("invalid category id: %w", err)
		}
		return nil, c.finSvc.DeleteCategory(ctx, id, c.userUUID)

	case "finance_create_record":
		rt := finance.RecordType(strArg(args, "type"))
		cats, err := c.finSvc.ListCategories(ctx, c.userUUID)
		if err != nil {
			return nil, err
		}
		catName := strArg(args, "category")
		var catID uuid.UUID
		for _, cat := range cats {
			if strings.EqualFold(cat.Name, catName) && cat.Type == rt {
				catID = cat.ID
				break
			}
		}
		if catID == (uuid.UUID{}) {
			return nil, fmt.Errorf("no %s category %q", rt, catName)
		}
		date := strArg(args, "date")
		if date == "" {
			date = time.Now().Format("2006-01-02")
		}
		return c.finSvc.CreateRecord(ctx, c.userUUID, finance.CreateRecordInput{
			Type:       rt,
			Amount:     floatArg(args, "amount"),
			CategoryID: catID,
			Note:       strArg(args, "note"),
			Date:       date,
		})

	case "finance_delete_record":
		id, err := uuid.Parse(strArg(args, "id"))
		if err != nil {
			return nil, fmt.Errorf("invalid record id: %w", err)
		}
		return nil, c.finSvc.DeleteRecord(ctx, id, c.userUUID)

	case "finance_monthly_summary":
		return c.finSvc.MonthlySummary(ctx, c.userUUID, strArg(args, "month"))

	case "kanban_get_board":
		boards, err := c.kanRepo.ListBoards(ctx, c.firebaseUID)
		if err != nil || len(boards) == 0 {
			return nil, fmt.Errorf("list boards: %w", err)
		}
		board := boards[0]
		columns, err := c.kanRepo.GetColumns(ctx, board.ID)
		if err != nil {
			return nil, err
		}
		cards, err := c.kanRepo.GetCards(ctx, board.ID)
		if err != nil {
			return nil, err
		}
		return map[string]any{"board": board, "columns": columns, "cards": cards}, nil

	case "kanban_create_card":
		colID, err := primitive.ObjectIDFromHex(strArg(args, "column_id"))
		if err != nil {
			return nil, fmt.Errorf("invalid column_id: %w", err)
		}
		boards, err := c.kanRepo.ListBoards(ctx, c.firebaseUID)
		if err != nil || len(boards) == 0 {
			return nil, fmt.Errorf("list boards: %w", err)
		}
		board := boards[0]
		priority := strArg(args, "priority")
		if priority == "" {
			priority = kanban.PriorityNone
		}
		var dueDatePtr *string
		if dd := strArg(args, "due_date"); dd != "" {
			dueDatePtr = &dd
		}
		return c.kanRepo.CreateCard(ctx, board.ID, colID, kanban.CreateCardInput{
			BoardID:     board.ID.Hex(),
			ColumnID:    strArg(args, "column_id"),
			Title:       strArg(args, "title"),
			Content:     strArg(args, "content"),
			Description: strArg(args, "description"),
			Priority:    priority,
			DueDate:     dueDatePtr,
			Tags:        strSliceArg(args, "tags"),
		})

	case "kanban_update_card":
		cardID, err := primitive.ObjectIDFromHex(strArg(args, "card_id"))
		if err != nil {
			return nil, fmt.Errorf("invalid card_id: %w", err)
		}
		fields := make(map[string]bool)
		for _, f := range strSliceArg(args, "update_fields") {
			fields[f] = true
		}
		var input kanban.UpdateCardInput
		if fields["title"] {
			s := strArg(args, "title")
			input.Title = &s
		}
		if fields["description"] {
			s := strArg(args, "description")
			input.Description = &s
		}
		if fields["priority"] {
			s := strArg(args, "priority")
			input.Priority = &s
		}
		if fields["due_date"] {
			s := strArg(args, "due_date")
			input.DueDate = &s
		}
		if fields["tags"] {
			tags := strSliceArg(args, "tags")
			input.Tags = &tags
		}
		return c.kanRepo.UpdateCard(ctx, cardID, input)

	case "kanban_move_card":
		cardID, err := primitive.ObjectIDFromHex(strArg(args, "card_id"))
		if err != nil {
			return nil, fmt.Errorf("invalid card_id: %w", err)
		}
		err = c.kanRepo.MoveCard(ctx, cardID, kanban.MoveCardInput{
			ColumnID: strArg(args, "column_id"),
			Order:    int(floatArg(args, "order")),
		})
		return map[string]any{"moved": err == nil}, err

	case "kanban_delete_card":
		cardID, err := primitive.ObjectIDFromHex(strArg(args, "card_id"))
		if err != nil {
			return nil, fmt.Errorf("invalid card_id: %w", err)
		}
		return nil, c.kanRepo.DeleteCard(ctx, cardID)

	case "kanban_board_stats":
		boardID, err := primitive.ObjectIDFromHex(strArg(args, "board_id"))
		if err != nil {
			return nil, fmt.Errorf("invalid board_id: %w", err)
		}
		return c.kanRepo.GetBoardStats(ctx, boardID)

	default:
		return nil, fmt.Errorf("unknown tool: %s", name)
	}
}

func strArg(args map[string]any, key string) string {
	s, _ := args[key].(string)
	return s
}

func floatArg(args map[string]any, key string) float64 {
	f, _ := args[key].(float64)
	return f
}

func strSliceArg(args map[string]any, key string) []string {
	switch v := args[key].(type) {
	case []string:
		return v
	case []any:
		out := make([]string, 0, len(v))
		for _, item := range v {
			if s, ok := item.(string); ok {
				out = append(out, s)
			}
		}
		return out
	}
	return nil
}
