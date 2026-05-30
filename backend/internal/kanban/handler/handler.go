package handler

import (
	"context"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/kinkando/personal-dashboard/internal/auth"
	"github.com/kinkando/personal-dashboard/internal/kanban"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Repository interface {
	GetBoard(ctx context.Context, userID string) (*kanban.Board, error)
	GetColumns(ctx context.Context, boardID primitive.ObjectID) ([]*kanban.Column, error)
	GetCards(ctx context.Context, boardID primitive.ObjectID) ([]*kanban.Card, error)
	CreateCard(ctx context.Context, boardID, columnID primitive.ObjectID, in kanban.CreateCardInput) (*kanban.Card, error)
	MoveCard(ctx context.Context, cardID primitive.ObjectID, in kanban.MoveCardInput) error
	DeleteCard(ctx context.Context, cardID primitive.ObjectID) error
}

type Handler struct {
	repo Repository
}

func New(repo Repository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) Register(router fiber.Router) {
	router.Get("/board", h.getBoard)
	router.Post("/cards", h.createCard)
	router.Patch("/cards/:id/move", h.moveCard)
	router.Delete("/cards/:id", h.deleteCard)
}

func (h *Handler) getBoard(c *fiber.Ctx) error {
	userID := auth.GetUserID(c)
	board, err := h.repo.GetBoard(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	columns, err := h.repo.GetColumns(c.Context(), board.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	cards, err := h.repo.GetCards(c.Context(), board.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	if columns == nil {
		columns = []*kanban.Column{}
	}
	if cards == nil {
		cards = []*kanban.Card{}
	}
	return c.JSON(fiber.Map{"data": fiber.Map{
		"board":   board,
		"columns": columns,
		"cards":   cards,
	}})
}

func (h *Handler) createCard(c *fiber.Ctx) error {
	userID := auth.GetUserID(c)
	var in kanban.CreateCardInput
	if err := c.BodyParser(&in); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	if in.Title == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "title is required"})
	}
	colID, err := primitive.ObjectIDFromHex(in.ColumnID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid column_id"})
	}
	board, err := h.repo.GetBoard(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	card, err := h.repo.CreateCard(c.Context(), board.ID, colID, in)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"data": card})
}

func (h *Handler) moveCard(c *fiber.Ctx) error {
	cardID, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid card id"})
	}
	var in kanban.MoveCardInput
	if err := c.BodyParser(&in); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	if err := h.repo.MoveCard(c.Context(), cardID, in); err != nil {
		if strings.Contains(err.Error(), "invalid column_id") {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) deleteCard(c *fiber.Ctx) error {
	cardID, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid card id"})
	}
	if err := h.repo.DeleteCard(c.Context(), cardID); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "card not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.SendStatus(fiber.StatusNoContent)
}
