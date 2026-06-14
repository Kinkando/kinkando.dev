package handler

import (
	"context"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/kinkando/personal-dashboard/internal/auth"
	"github.com/kinkando/personal-dashboard/internal/kanban"
	"github.com/kinkando/personal-dashboard/pkg/respond"
	"github.com/kinkando/personal-dashboard/pkg/validate"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Repository interface {
	ListBoards(ctx context.Context, userID string) ([]*kanban.Board, error)
	GetBoardByID(ctx context.Context, boardID primitive.ObjectID, userID string) (*kanban.Board, error)
	CreateBoard(ctx context.Context, userID, name string) (*kanban.Board, error)
	UpdateBoard(ctx context.Context, boardID primitive.ObjectID, name string) error
	DeleteBoard(ctx context.Context, boardID primitive.ObjectID) error
	GetColumns(ctx context.Context, boardID primitive.ObjectID) ([]*kanban.Column, error)
	GetColumn(ctx context.Context, columnID primitive.ObjectID) (*kanban.Column, error)
	CreateColumn(ctx context.Context, boardID primitive.ObjectID, name string) (*kanban.Column, error)
	UpdateColumn(ctx context.Context, columnID primitive.ObjectID, name string) error
	ReorderColumns(ctx context.Context, boardID primitive.ObjectID, columnIDs []string) error
	DeleteColumn(ctx context.Context, columnID primitive.ObjectID, action, targetColumnID string) error
	GetCards(ctx context.Context, boardID primitive.ObjectID) ([]*kanban.Card, error)
	GetCard(ctx context.Context, cardID primitive.ObjectID) (*kanban.Card, error)
	CreateCard(ctx context.Context, boardID, columnID primitive.ObjectID, in kanban.CreateCardInput) (*kanban.Card, error)
	UpdateCard(ctx context.Context, cardID primitive.ObjectID, in kanban.UpdateCardInput) (*kanban.Card, error)
	MoveCard(ctx context.Context, cardID primitive.ObjectID, in kanban.MoveCardInput) error
	DeleteCard(ctx context.Context, cardID primitive.ObjectID) error
	AddAttachment(ctx context.Context, cardID primitive.ObjectID, in kanban.AddAttachmentInput) (*kanban.Attachment, error)
	RemoveAttachment(ctx context.Context, cardID, attachmentID primitive.ObjectID) (*kanban.Attachment, error)
	ArchiveCard(ctx context.Context, cardID primitive.ObjectID, reason kanban.ArchiveReason) (*kanban.Card, error)
	UnarchiveCard(ctx context.Context, cardID primitive.ObjectID) (*kanban.Card, error)
	ListArchivedCards(ctx context.Context, boardID primitive.ObjectID, filter kanban.ListArchivedFilter) ([]*kanban.Card, error)
	GetBoardStats(ctx context.Context, boardID primitive.ObjectID) (*kanban.BoardStats, error)
}

type Handler struct {
	repo Repository
}

func New(repo Repository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) Register(router fiber.Router) {
	// Board routes
	router.Get("/boards", h.listBoards)
	router.Post("/boards", h.createBoard)
	router.Get("/boards/:id/stats", h.getBoardStats)
	router.Get("/boards/:id/archive", h.listArchive)
	router.Patch("/boards/:id/columns/reorder", h.reorderColumns)
	router.Get("/boards/:id", h.getBoard)
	router.Patch("/boards/:id", h.updateBoard)
	router.Delete("/boards/:id", h.deleteBoard)
	// Column routes
	router.Post("/columns", h.createColumn)
	router.Patch("/columns/:id", h.updateColumn)
	router.Delete("/columns/:id", h.deleteColumn)
	// Card routes
	router.Post("/cards", h.createCard)
	router.Patch("/cards/:id/move", h.moveCard)
	router.Patch("/cards/:id/archive", h.archiveCard)
	router.Patch("/cards/:id/unarchive", h.unarchiveCard)
	router.Post("/cards/:id/attachments", h.addAttachment)
	router.Delete("/cards/:id/attachments/:attachmentId", h.removeAttachment)
	router.Patch("/cards/:id", h.updateCard)
	router.Delete("/cards/:id", h.deleteCard)
}

// ---- Board handlers --------------------------------------------------------

func (h *Handler) listBoards(c *fiber.Ctx) error {
	userID := auth.GetUserID(c)
	boards, err := h.repo.ListBoards(c.Context(), userID)
	if err != nil {
		return respond.Internal(c, err)
	}
	if boards == nil {
		boards = []*kanban.Board{}
	}
	return respond.Data(c, boards)
}

func (h *Handler) createBoard(c *fiber.Ctx) error {
	userID := auth.GetUserID(c)
	var in kanban.CreateBoardInput
	if err := c.BodyParser(&in); err != nil {
		return respond.BadRequest(c, "invalid request body")
	}
	if err := validate.Struct(in); err != nil {
		return err
	}
	board, err := h.repo.CreateBoard(c.Context(), userID, in.Name)
	if err != nil {
		return respond.Internal(c, err)
	}
	return respond.Created(c, board)
}

func (h *Handler) getBoard(c *fiber.Ctx) error {
	userID := auth.GetUserID(c)
	boardID, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return respond.BadRequest(c, "invalid board id")
	}
	board, err := h.repo.GetBoardByID(c.Context(), boardID, userID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return respond.NotFound(c, "board not found")
		}
		return respond.Internal(c, err)
	}
	columns, err := h.repo.GetColumns(c.Context(), board.ID)
	if err != nil {
		return respond.Internal(c, err)
	}
	cards, err := h.repo.GetCards(c.Context(), board.ID)
	if err != nil {
		return respond.Internal(c, err)
	}
	if columns == nil {
		columns = []*kanban.Column{}
	}
	if cards == nil {
		cards = []*kanban.Card{}
	}
	// Normalise zero values on old documents.
	for _, card := range cards {
		if card.Tags == nil {
			card.Tags = []string{}
		}
		if card.Attachments == nil {
			card.Attachments = []kanban.Attachment{}
		}
		if card.Priority == "" {
			card.Priority = kanban.PriorityNone
		}
	}
	// Default column type if somehow empty after backfill.
	for _, col := range columns {
		if col.Type == "" {
			col.Type = kanban.ColumnTypeCustom
		}
	}
	return respond.Data(c, fiber.Map{
		"board":   board,
		"columns": columns,
		"cards":   cards,
	})
}

func (h *Handler) getBoardStats(c *fiber.Ctx) error {
	userID := auth.GetUserID(c)
	boardID, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return respond.BadRequest(c, "invalid board id")
	}
	if _, err := h.repo.GetBoardByID(c.Context(), boardID, userID); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return respond.NotFound(c, "board not found")
		}
		return respond.Internal(c, err)
	}
	stats, err := h.repo.GetBoardStats(c.Context(), boardID)
	if err != nil {
		return respond.Internal(c, err)
	}
	return respond.Data(c, stats)
}

func (h *Handler) updateBoard(c *fiber.Ctx) error {
	userID := auth.GetUserID(c)
	boardID, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return respond.BadRequest(c, "invalid board id")
	}
	if _, err := h.repo.GetBoardByID(c.Context(), boardID, userID); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return respond.NotFound(c, "board not found")
		}
		return respond.Internal(c, err)
	}
	var in kanban.UpdateBoardInput
	if err := c.BodyParser(&in); err != nil {
		return respond.BadRequest(c, "invalid request body")
	}
	if err := validate.Struct(in); err != nil {
		return err
	}
	if err := h.repo.UpdateBoard(c.Context(), boardID, in.Name); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return respond.NotFound(c, "board not found")
		}
		return respond.Internal(c, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) deleteBoard(c *fiber.Ctx) error {
	userID := auth.GetUserID(c)
	boardID, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return respond.BadRequest(c, "invalid board id")
	}
	if _, err := h.repo.GetBoardByID(c.Context(), boardID, userID); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return respond.NotFound(c, "board not found")
		}
		return respond.Internal(c, err)
	}
	if err := h.repo.DeleteBoard(c.Context(), boardID); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return respond.NotFound(c, "board not found")
		}
		return respond.Internal(c, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// ---- Column handlers -------------------------------------------------------

func (h *Handler) createColumn(c *fiber.Ctx) error {
	userID := auth.GetUserID(c)
	var in kanban.CreateColumnInput
	if err := c.BodyParser(&in); err != nil {
		return respond.BadRequest(c, "invalid request body")
	}
	if err := validate.Struct(in); err != nil {
		return err
	}
	boardID, err := primitive.ObjectIDFromHex(in.BoardID)
	if err != nil {
		return respond.BadRequest(c, "invalid board_id")
	}
	if _, err := h.repo.GetBoardByID(c.Context(), boardID, userID); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return respond.NotFound(c, "board not found")
		}
		return respond.Internal(c, err)
	}
	col, err := h.repo.CreateColumn(c.Context(), boardID, in.Name)
	if err != nil {
		return respond.Internal(c, err)
	}
	return respond.Created(c, col)
}

func (h *Handler) updateColumn(c *fiber.Ctx) error {
	userID := auth.GetUserID(c)
	columnID, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return respond.BadRequest(c, "invalid column id")
	}
	col, err := h.repo.GetColumn(c.Context(), columnID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return respond.NotFound(c, "column not found")
		}
		return respond.Internal(c, err)
	}
	if _, err := h.repo.GetBoardByID(c.Context(), col.BoardID, userID); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return respond.NotFound(c, "column not found")
		}
		return respond.Internal(c, err)
	}
	var in kanban.UpdateColumnInput
	if err := c.BodyParser(&in); err != nil {
		return respond.BadRequest(c, "invalid request body")
	}
	if err := validate.Struct(in); err != nil {
		return err
	}
	if err := h.repo.UpdateColumn(c.Context(), columnID, in.Name); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return respond.NotFound(c, "column not found")
		}
		return respond.Internal(c, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) reorderColumns(c *fiber.Ctx) error {
	userID := auth.GetUserID(c)
	boardID, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return respond.BadRequest(c, "invalid board id")
	}
	if _, err := h.repo.GetBoardByID(c.Context(), boardID, userID); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return respond.NotFound(c, "board not found")
		}
		return respond.Internal(c, err)
	}
	var in kanban.ReorderColumnsInput
	if err := c.BodyParser(&in); err != nil {
		return respond.BadRequest(c, "invalid request body")
	}
	if err := validate.Struct(in); err != nil {
		return err
	}
	if err := h.repo.ReorderColumns(c.Context(), boardID, in.ColumnIDs); err != nil {
		return respond.Internal(c, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) deleteColumn(c *fiber.Ctx) error {
	userID := auth.GetUserID(c)
	columnID, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return respond.BadRequest(c, "invalid column id")
	}
	col, err := h.repo.GetColumn(c.Context(), columnID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return respond.NotFound(c, "column not found")
		}
		return respond.Internal(c, err)
	}
	if _, err := h.repo.GetBoardByID(c.Context(), col.BoardID, userID); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return respond.NotFound(c, "column not found")
		}
		return respond.Internal(c, err)
	}
	if col.IsSystem {
		return respond.Conflict(c, "system columns cannot be deleted")
	}
	var in kanban.DeleteColumnInput
	if err := c.BodyParser(&in); err != nil {
		return respond.BadRequest(c, "invalid request body")
	}
	if err := validate.Struct(in); err != nil {
		return err
	}
	if err := h.repo.DeleteColumn(c.Context(), columnID, in.Action, in.TargetColumnID); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return respond.NotFound(c, err.Error())
		}
		if strings.Contains(err.Error(), "system column") {
			return respond.Conflict(c, err.Error())
		}
		if strings.Contains(err.Error(), "invalid") {
			return respond.BadRequest(c, err.Error())
		}
		return respond.Internal(c, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// ---- Card handlers ---------------------------------------------------------

func (h *Handler) createCard(c *fiber.Ctx) error {
	userID := auth.GetUserID(c)
	var in kanban.CreateCardInput
	if err := c.BodyParser(&in); err != nil {
		return respond.BadRequest(c, "invalid request body")
	}
	if err := validate.Struct(in); err != nil {
		return err
	}
	boardID, err := primitive.ObjectIDFromHex(in.BoardID)
	if err != nil {
		return respond.BadRequest(c, "invalid board_id")
	}
	if _, err := h.repo.GetBoardByID(c.Context(), boardID, userID); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return respond.NotFound(c, "board not found")
		}
		return respond.Internal(c, err)
	}
	colID, err := primitive.ObjectIDFromHex(in.ColumnID)
	if err != nil {
		return respond.BadRequest(c, "invalid column_id")
	}
	card, err := h.repo.CreateCard(c.Context(), boardID, colID, in)
	if err != nil {
		return respond.Internal(c, err)
	}
	if card.Tags == nil {
		card.Tags = []string{}
	}
	if card.Attachments == nil {
		card.Attachments = []kanban.Attachment{}
	}
	return respond.Created(c, card)
}

func (h *Handler) updateCard(c *fiber.Ctx) error {
	userID := auth.GetUserID(c)
	cardID, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return respond.BadRequest(c, "invalid card id")
	}
	card, err := h.repo.GetCard(c.Context(), cardID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return respond.NotFound(c, "card not found")
		}
		return respond.Internal(c, err)
	}
	if _, err := h.repo.GetBoardByID(c.Context(), card.BoardID, userID); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return respond.NotFound(c, "card not found")
		}
		return respond.Internal(c, err)
	}
	var in kanban.UpdateCardInput
	if err := c.BodyParser(&in); err != nil {
		return respond.BadRequest(c, "invalid request body")
	}
	if err := validate.Struct(in); err != nil {
		return err
	}
	updated, err := h.repo.UpdateCard(c.Context(), cardID, in)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return respond.NotFound(c, "card not found")
		}
		return respond.Internal(c, err)
	}
	if updated.Tags == nil {
		updated.Tags = []string{}
	}
	if updated.Attachments == nil {
		updated.Attachments = []kanban.Attachment{}
	}
	if updated.Priority == "" {
		updated.Priority = kanban.PriorityNone
	}
	return respond.Data(c, updated)
}

func (h *Handler) moveCard(c *fiber.Ctx) error {
	userID := auth.GetUserID(c)
	cardID, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return respond.BadRequest(c, "invalid card id")
	}
	card, err := h.repo.GetCard(c.Context(), cardID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return respond.NotFound(c, "card not found")
		}
		return respond.Internal(c, err)
	}
	if _, err := h.repo.GetBoardByID(c.Context(), card.BoardID, userID); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return respond.NotFound(c, "card not found")
		}
		return respond.Internal(c, err)
	}
	var in kanban.MoveCardInput
	if err := c.BodyParser(&in); err != nil {
		return respond.BadRequest(c, "invalid request body")
	}
	if err := validate.Struct(in); err != nil {
		return err
	}
	if err := h.repo.MoveCard(c.Context(), cardID, in); err != nil {
		if strings.Contains(err.Error(), "invalid column_id") {
			return respond.BadRequest(c, err.Error())
		}
		return respond.Internal(c, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) deleteCard(c *fiber.Ctx) error {
	userID := auth.GetUserID(c)
	cardID, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return respond.BadRequest(c, "invalid card id")
	}
	card, err := h.repo.GetCard(c.Context(), cardID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return respond.NotFound(c, "card not found")
		}
		return respond.Internal(c, err)
	}
	if _, err := h.repo.GetBoardByID(c.Context(), card.BoardID, userID); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return respond.NotFound(c, "card not found")
		}
		return respond.Internal(c, err)
	}
	if err := h.repo.DeleteCard(c.Context(), cardID); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return respond.NotFound(c, "card not found")
		}
		return respond.Internal(c, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) archiveCard(c *fiber.Ctx) error {
	userID := auth.GetUserID(c)
	cardID, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return respond.BadRequest(c, "invalid card id")
	}
	card, err := h.repo.GetCard(c.Context(), cardID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return respond.NotFound(c, "card not found")
		}
		return respond.Internal(c, err)
	}
	if _, err := h.repo.GetBoardByID(c.Context(), card.BoardID, userID); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return respond.NotFound(c, "card not found")
		}
		return respond.Internal(c, err)
	}
	var in kanban.ArchiveCardInput
	if err := c.BodyParser(&in); err != nil {
		return respond.BadRequest(c, "invalid request body")
	}
	// Reject client-supplied "completed" before general validation — it is
	// system-assigned based on column type and warrants a specific message.
	if in.Reason == kanban.ArchiveReasonCompleted {
		return respond.BadRequest(c, "reason 'completed' is reserved; omit it and the server will assign it when archiving from a Done column")
	}
	if err := validate.Struct(in); err != nil {
		return err
	}
	archived, err := h.repo.ArchiveCard(c.Context(), cardID, in.Reason)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return respond.NotFound(c, "card not found")
		}
		return respond.Internal(c, err)
	}
	if archived.Tags == nil {
		archived.Tags = []string{}
	}
	if archived.Attachments == nil {
		archived.Attachments = []kanban.Attachment{}
	}
	return respond.Data(c, archived)
}

func (h *Handler) unarchiveCard(c *fiber.Ctx) error {
	userID := auth.GetUserID(c)
	cardID, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return respond.BadRequest(c, "invalid card id")
	}
	card, err := h.repo.GetCard(c.Context(), cardID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return respond.NotFound(c, "card not found")
		}
		return respond.Internal(c, err)
	}
	if _, err := h.repo.GetBoardByID(c.Context(), card.BoardID, userID); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return respond.NotFound(c, "card not found")
		}
		return respond.Internal(c, err)
	}
	restored, err := h.repo.UnarchiveCard(c.Context(), cardID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return respond.NotFound(c, "card not found")
		}
		if strings.Contains(err.Error(), "not archived") {
			return respond.BadRequest(c, err.Error())
		}
		return respond.Internal(c, err)
	}
	if restored.Tags == nil {
		restored.Tags = []string{}
	}
	if restored.Attachments == nil {
		restored.Attachments = []kanban.Attachment{}
	}
	return respond.Data(c, restored)
}

func (h *Handler) listArchive(c *fiber.Ctx) error {
	userID := auth.GetUserID(c)
	boardID, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return respond.BadRequest(c, "invalid board id")
	}
	if _, err := h.repo.GetBoardByID(c.Context(), boardID, userID); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return respond.NotFound(c, "board not found")
		}
		return respond.Internal(c, err)
	}

	filter := kanban.ListArchivedFilter{
		Reason: c.Query("reason"), // "completed" | "general" | ""
	}
	if monthStr := c.Query("month"); monthStr != "" {
		if m, err := strconv.Atoi(monthStr); err == nil && m >= 1 && m <= 12 {
			filter.Month = m
		}
	}
	if yearStr := c.Query("year"); yearStr != "" {
		if y, err := strconv.Atoi(yearStr); err == nil && y > 0 {
			filter.Year = y
		}
	}

	cards, err := h.repo.ListArchivedCards(c.Context(), boardID, filter)
	if err != nil {
		return respond.Internal(c, err)
	}
	if cards == nil {
		cards = []*kanban.Card{}
	}
	for _, card := range cards {
		if card.Tags == nil {
			card.Tags = []string{}
		}
		if card.Attachments == nil {
			card.Attachments = []kanban.Attachment{}
		}
		if card.Priority == "" {
			card.Priority = kanban.PriorityNone
		}
	}
	return respond.Data(c, cards)
}

// ---- Attachment handlers ---------------------------------------------------

// addAttachment registers a file the client has already uploaded to Firebase Storage.
// The bytes never pass through the backend; only metadata is persisted.
func (h *Handler) addAttachment(c *fiber.Ctx) error {
	userID := auth.GetUserID(c)
	cardID, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return respond.BadRequest(c, "invalid card id")
	}
	card, err := h.repo.GetCard(c.Context(), cardID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return respond.NotFound(c, "card not found")
		}
		return respond.Internal(c, err)
	}
	if _, err := h.repo.GetBoardByID(c.Context(), card.BoardID, userID); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return respond.NotFound(c, "card not found")
		}
		return respond.Internal(c, err)
	}
	var in kanban.AddAttachmentInput
	if err := c.BodyParser(&in); err != nil {
		return respond.BadRequest(c, "invalid request body")
	}
	if err := validate.Struct(in); err != nil {
		return err
	}
	att, err := h.repo.AddAttachment(c.Context(), cardID, in)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return respond.NotFound(c, "card not found")
		}
		return respond.Internal(c, err)
	}
	return respond.Created(c, att)
}

// removeAttachment removes a single attachment from a card. The client is
// responsible for deleting the underlying object from Firebase Storage; the
// removed attachment is returned so the client knows the storage_path to delete.
func (h *Handler) removeAttachment(c *fiber.Ctx) error {
	userID := auth.GetUserID(c)
	cardID, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return respond.BadRequest(c, "invalid card id")
	}
	attachmentID, err := primitive.ObjectIDFromHex(c.Params("attachmentId"))
	if err != nil {
		return respond.BadRequest(c, "invalid attachment id")
	}
	card, err := h.repo.GetCard(c.Context(), cardID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return respond.NotFound(c, "card not found")
		}
		return respond.Internal(c, err)
	}
	if _, err := h.repo.GetBoardByID(c.Context(), card.BoardID, userID); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return respond.NotFound(c, "card not found")
		}
		return respond.Internal(c, err)
	}
	removed, err := h.repo.RemoveAttachment(c.Context(), cardID, attachmentID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return respond.NotFound(c, err.Error())
		}
		return respond.Internal(c, err)
	}
	return respond.Data(c, removed)
}
