package handler

import (
	"context"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/kinkando/personal-dashboard/internal/auth"
	"github.com/kinkando/personal-dashboard/internal/finance"
	"github.com/kinkando/personal-dashboard/pkg/validate"
)

type Service interface {
	CreateCategory(ctx context.Context, userID uuid.UUID, in finance.CreateCategoryInput) (*finance.Category, error)
	ListCategories(ctx context.Context, userID uuid.UUID) ([]*finance.Category, error)
	UpdateCategory(ctx context.Context, id uuid.UUID, userID uuid.UUID, in finance.UpdateCategoryInput) (*finance.Category, error)
	DeleteCategory(ctx context.Context, id uuid.UUID, userID uuid.UUID) error

	CreateRecord(ctx context.Context, userID uuid.UUID, in finance.CreateRecordInput) (*finance.Record, error)
	ListRecords(ctx context.Context, userID uuid.UUID, month string) ([]*finance.Record, error)
	MonthlySummary(ctx context.Context, userID uuid.UUID, month string) (*finance.MonthlySummary, error)
	DeleteRecord(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
}

// UserResolver resolves a Firebase UID to the internal UUID stored in the users table.
type UserResolver interface {
	GetIDByFirebaseUID(ctx context.Context, firebaseUID string) (uuid.UUID, error)
}

type Handler struct {
	svc   Service
	users UserResolver
}

func New(svc Service, users UserResolver) *Handler {
	return &Handler{svc: svc, users: users}
}

func (h *Handler) Register(router fiber.Router) {
	router.Get("/categories", h.listCategories)
	router.Post("/categories", h.createCategory)
	router.Patch("/categories/:id", h.updateCategory)
	router.Delete("/categories/:id", h.deleteCategory)

	router.Get("/records", h.listRecords)
	router.Post("/records", h.createRecord)
	router.Delete("/records/:id", h.deleteRecord)
	router.Get("/summary", h.summary)
}

// ── Category handlers ─────────────────────────────────────────────────────────

func (h *Handler) listCategories(c *fiber.Ctx) error {
	userID, err := h.resolveUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid user"})
	}
	cats, err := h.svc.ListCategories(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	if cats == nil {
		cats = []*finance.Category{}
	}
	return c.JSON(fiber.Map{"data": cats})
}

func (h *Handler) createCategory(c *fiber.Ctx) error {
	userID, err := h.resolveUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid user"})
	}
	var in finance.CreateCategoryInput
	if err := c.BodyParser(&in); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	if err := validate.Struct(in); err != nil {
		return err
	}
	cat, err := h.svc.CreateCategory(c.Context(), userID, in)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"data": cat})
}

func (h *Handler) updateCategory(c *fiber.Ctx) error {
	userID, err := h.resolveUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid user"})
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid category id"})
	}
	var in finance.UpdateCategoryInput
	if err := c.BodyParser(&in); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	if err := validate.Struct(in); err != nil {
		return err
	}
	cat, err := h.svc.UpdateCategory(c.Context(), id, userID, in)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "category not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"data": cat})
}

func (h *Handler) deleteCategory(c *fiber.Ctx) error {
	userID, err := h.resolveUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid user"})
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid category id"})
	}
	if err := h.svc.DeleteCategory(c.Context(), id, userID); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "category not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// ── Record handlers ───────────────────────────────────────────────────────────

func (h *Handler) listRecords(c *fiber.Ctx) error {
	userID, err := h.resolveUserID(c)
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
	userID, err := h.resolveUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid user"})
	}
	var in finance.CreateRecordInput
	if err := c.BodyParser(&in); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	if err := validate.Struct(in); err != nil {
		return err
	}
	rec, err := h.svc.CreateRecord(c.Context(), userID, in)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
		}
		if strings.Contains(err.Error(), "does not match") {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"data": rec})
}

func (h *Handler) deleteRecord(c *fiber.Ctx) error {
	userID, err := h.resolveUserID(c)
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
	userID, err := h.resolveUserID(c)
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

// resolveUserID looks up the internal UUID for the Firebase UID in the request context.
func (h *Handler) resolveUserID(c *fiber.Ctx) (uuid.UUID, error) {
	firebaseUID := auth.GetUserID(c)
	if firebaseUID == "" {
		return uuid.UUID{}, fiber.ErrUnauthorized
	}
	return h.users.GetIDByFirebaseUID(c.Context(), firebaseUID)
}
