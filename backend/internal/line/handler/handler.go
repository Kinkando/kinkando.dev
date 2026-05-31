package handler

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/kinkando/personal-dashboard/internal/finance"
	financeSvc "github.com/kinkando/personal-dashboard/internal/finance/service"
	"github.com/kinkando/personal-dashboard/internal/kanban"
	kanbanRepo "github.com/kinkando/personal-dashboard/internal/kanban/repository"
	linepkg "github.com/kinkando/personal-dashboard/internal/line"
	"go.uber.org/zap"
)

const usageText = "Unknown command. Use:\n" +
	"  fin <income|expense> <amount> <category> [note]\n" +
	"  card [column] <title>"

// Deps bundles dependencies for the LINE webhook handler.
type Deps struct {
	ChannelID     string
	ChannelSecret string
	Client        *linepkg.Client
	FinSvc        *financeSvc.Service
	KanRepo       *kanbanRepo.Repository
	UserUUID      uuid.UUID // finance user
	FirebaseUID   string    // kanban user
	Logger        *zap.Logger
}

// Handler handles LINE Messaging API webhook events.
type Handler struct {
	channelID     string
	channelSecret string
	client        *linepkg.Client
	finSvc        *financeSvc.Service
	kanRepo       *kanbanRepo.Repository
	userUUID      uuid.UUID
	firebaseUID   string
	logger        *zap.Logger
}

// New creates a Handler from the provided dependencies.
func New(d Deps) *Handler {
	return &Handler{
		channelID:     d.ChannelID,
		channelSecret: d.ChannelSecret,
		client:        d.Client,
		finSvc:        d.FinSvc,
		kanRepo:       d.KanRepo,
		userUUID:      d.UserUUID,
		firebaseUID:   d.FirebaseUID,
		logger:        d.Logger,
	}
}

// Register mounts routes onto the given router (no auth middleware — webhook is
// self-authenticated via X-Line-Signature HMAC verification).
func (h *Handler) Register(router fiber.Router) {
	router.Post("/webhook", h.webhook)
}

func (h *Handler) webhook(c *fiber.Ctx) error {
	body := c.Body()

	sig := c.Get("X-Line-Signature")
	if !h.validSignature(body, sig) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid signature"})
	}

	var payload linepkg.WebhookBody
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	// Always acknowledge 200 before processing — LINE retries on non-2xx.
	// Process events asynchronously so the ack is immediate.
	ctx := c.Context()
	for _, ev := range payload.Events {
		if ev.Type != "message" || ev.Message == nil || ev.Message.Type != "text" || ev.ReplyToken == "" {
			continue
		}
		replyText := h.handleText(ctx, ev.Message.Text)
		if err := h.client.Reply(ctx, ev.ReplyToken, []linepkg.ReplyMessage{linepkg.TextMessage(replyText)}); err != nil {
			h.logger.Error("LINE reply failed", zap.String("replyToken", ev.ReplyToken), zap.Error(err))
		}
	}

	return c.SendStatus(fiber.StatusOK)
}

// handleText routes the message text to the correct feature handler and returns
// a human-readable reply string.
func (h *Handler) handleText(ctx context.Context, text string) string {
	text = strings.TrimSpace(text)
	idx := strings.IndexByte(text, ' ')

	var cmd, rest string
	if idx == -1 {
		cmd = strings.ToLower(text)
		rest = ""
	} else {
		cmd = strings.ToLower(text[:idx])
		rest = strings.TrimSpace(text[idx+1:])
	}

	switch cmd {
	case "fin":
		return h.createFinance(ctx, rest)
	case "card":
		return h.createCard(ctx, rest)
	default:
		return usageText
	}
}

// createFinance parses "fin <income|expense> <amount> <category> [note...]" and
// creates a finance record for today.
//
// Example: fin expense 120 Food lunch at work
func (h *Handler) createFinance(ctx context.Context, rest string) string {
	if rest == "" {
		return "Usage: fin <income|expense> <amount> <category> [note]"
	}

	parts := strings.Fields(rest)
	if len(parts) < 3 {
		return "Usage: fin <income|expense> <amount> <category> [note]"
	}

	recType := finance.RecordType(strings.ToLower(parts[0]))
	if recType != finance.RecordTypeIncome && recType != finance.RecordTypeExpense {
		return fmt.Sprintf("Type must be \"income\" or \"expense\", got %q.", parts[0])
	}

	amount, err := strconv.ParseFloat(parts[1], 64)
	if err != nil || amount <= 0 {
		return fmt.Sprintf("Invalid amount %q — must be a positive number.", parts[1])
	}

	categoryName := parts[2]
	note := ""
	if len(parts) > 3 {
		note = strings.Join(parts[3:], " ")
	}

	// Resolve category name → ID (mirrors MCP server logic).
	cats, err := h.finSvc.ListCategories(ctx, h.userUUID)
	if err != nil {
		h.logger.Error("LINE finance: list categories", zap.Error(err))
		return "⚠️ Could not load categories. Try again."
	}
	var catID uuid.UUID
	found := false
	for _, c := range cats {
		if strings.EqualFold(c.Name, categoryName) && c.Type == recType {
			catID = c.ID
			found = true
			break
		}
	}
	if !found {
		return fmt.Sprintf("No %s category named %q. Check your categories in the dashboard.", recType, categoryName)
	}

	rec, err := h.finSvc.CreateRecord(ctx, h.userUUID, finance.CreateRecordInput{
		Type:       recType,
		Amount:     amount,
		CategoryID: catID,
		Note:       note,
		Date:       time.Now().Format("2006-01-02"),
	})
	if err != nil {
		h.logger.Error("LINE finance: create record", zap.Error(err))
		return "⚠️ Could not create record. Try again."
	}

	noteStr := ""
	if rec.Note != "" {
		noteStr = " — " + rec.Note
	}
	return fmt.Sprintf("✅ %s %.2f (%s)%s recorded for %s.", rec.Type, rec.Amount, categoryName, noteStr, rec.Date.Format("2006-01-02"))
}

// createCard parses "card [column] <title>" and creates a kanban card.
// If the text starts with a known column name (case-insensitive), that column is
// used and the remainder becomes the title. Otherwise the first column is used.
//
// Examples:
//
//	card To Do Buy milk          → column "To Do", title "Buy milk"
//	card Buy milk                → column (first), title "Buy milk"
func (h *Handler) createCard(ctx context.Context, rest string) string {
	if rest == "" {
		return "Usage: card [column] <title>"
	}

	boards, err := h.kanRepo.ListBoards(ctx, h.firebaseUID)
	if err != nil || len(boards) == 0 {
		h.logger.Error("LINE kanban: list boards", zap.Error(err))
		return "⚠️ Could not load kanban board. Try again."
	}
	board := boards[0]

	columns, err := h.kanRepo.GetColumns(ctx, board.ID)
	if err != nil || len(columns) == 0 {
		h.logger.Error("LINE kanban: get columns", zap.Error(err))
		return "⚠️ Could not load columns. Try again."
	}

	// Find the longest column-name prefix match (greedy) so "In Progress" beats "In".
	var targetCol *kanban.Column
	var title string

	lowerRest := strings.ToLower(rest)
	bestLen := 0
	for _, col := range columns {
		lowerName := strings.ToLower(col.Name)
		// Match "<name> <rest-of-title>" or exact "<name>" (no title yet).
		if strings.HasPrefix(lowerRest, lowerName+" ") && len(col.Name) > bestLen {
			bestLen = len(col.Name)
			targetCol = col
			title = strings.TrimSpace(rest[len(col.Name)+1:])
		} else if strings.EqualFold(rest, col.Name) {
			// User typed only the column name with no title.
			return "Usage: card [column] <title>"
		}
	}

	if targetCol == nil {
		// No column matched — use first column and entire rest as title.
		targetCol = columns[0]
		title = rest
	}

	if title == "" {
		return "Usage: card [column] <title>"
	}

	card, err := h.kanRepo.CreateCard(ctx, board.ID, targetCol.ID, kanban.CreateCardInput{
		BoardID:  board.ID.Hex(),
		ColumnID: targetCol.ID.Hex(),
		Title:    title,
		Priority: kanban.PriorityNone,
		Tags:     []string{},
	})
	if err != nil {
		h.logger.Error("LINE kanban: create card", zap.Error(err))
		return "⚠️ Could not create card. Try again."
	}

	return fmt.Sprintf("✅ Card %q added to \"%s\".", card.Title, targetCol.Name)
}

func (h *Handler) validSignature(body []byte, sig string) bool {
	mac := hmac.New(sha256.New, []byte(h.channelSecret))
	mac.Write(body) //nolint:errcheck
	expected := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(sig))
}
