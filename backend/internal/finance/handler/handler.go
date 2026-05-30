package handler

import (
	"context"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/kinkando/personal-dashboard/internal/auth"
	"github.com/kinkando/personal-dashboard/internal/finance"
)

type Service interface {
	CreateRecord(ctx context.Context, userID uuid.UUID, in finance.CreateRecordInput) (*finance.Record, error)
	ListRecords(ctx context.Context, userID uuid.UUID, month string) ([]*finance.Record, error)
	MonthlySummary(ctx context.Context, userID uuid.UUID, month string) (*finance.MonthlySummary, error)
	DeleteRecord(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
}

type Handler struct {
	svc Service
}

func New(svc Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Register(router fiber.Router) {
	router.Get("/records", h.listRecords)
	router.Post("/records", h.createRecord)
	router.Delete("/records/:id", h.deleteRecord)
	router.Get("/summary", h.summary)
}

func (h *Handler) listRecords(c *fiber.Ctx) error {
	userID, err := parseUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid user"})
	}
	month := c.Query("month")
	if month == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "month query param required (YYYY-MM)"})
	}
	records, err := h.svc.ListRecords(c.Context(), userID, month)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	if records == nil {
		records = []*finance.Record{}
	}
	return c.JSON(fiber.Map{"data": records})
}

func (h *Handler) createRecord(c *fiber.Ctx) error {
	userID, err := parseUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid user"})
	}
	var in finance.CreateRecordInput
	if err := c.BodyParser(&in); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	if in.Type != finance.RecordTypeIncome && in.Type != finance.RecordTypeExpense {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "type must be income or expense"})
	}
	if in.Amount <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "amount must be positive"})
	}
	rec, err := h.svc.CreateRecord(c.Context(), userID, in)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"data": rec})
}

func (h *Handler) deleteRecord(c *fiber.Ctx) error {
	userID, err := parseUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid user"})
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid record id"})
	}
	if err := h.svc.DeleteRecord(c.Context(), id, userID); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "record not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) summary(c *fiber.Ctx) error {
	userID, err := parseUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid user"})
	}
	month := c.Query("month")
	if month == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "month query param required (YYYY-MM)"})
	}
	s, err := h.svc.MonthlySummary(c.Context(), userID, month)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"data": s})
}

func parseUserID(c *fiber.Ctx) (uuid.UUID, error) {
	return uuid.Parse(auth.GetUserID(c))
}
