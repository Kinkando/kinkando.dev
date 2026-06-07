package handler

import (
	"context"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/kinkando/personal-dashboard/internal/auth"
	"github.com/kinkando/personal-dashboard/internal/finance"
	"github.com/kinkando/personal-dashboard/pkg/respond"
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

type Handler struct {
	svc   Service
	users auth.UserResolver
}

func New(svc Service, users auth.UserResolver) *Handler {
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
	userID, err := auth.ResolveUserID(c, h.users)
	if err != nil {
		return respond.Unauthorized(c, "invalid user")
	}
	cats, err := h.svc.ListCategories(c.Context(), userID)
	if err != nil {
		return respond.Internal(c, err)
	}
	if cats == nil {
		cats = []*finance.Category{}
	}
	return respond.Data(c, cats)
}

func (h *Handler) createCategory(c *fiber.Ctx) error {
	userID, err := auth.ResolveUserID(c, h.users)
	if err != nil {
		return respond.Unauthorized(c, "invalid user")
	}
	var in finance.CreateCategoryInput
	if err := c.BodyParser(&in); err != nil {
		return respond.BadRequest(c, "invalid request body")
	}
	if err := validate.Struct(in); err != nil {
		return err
	}
	cat, err := h.svc.CreateCategory(c.Context(), userID, in)
	if err != nil {
		return respond.Internal(c, err)
	}
	return respond.Created(c, cat)
}

func (h *Handler) updateCategory(c *fiber.Ctx) error {
	userID, err := auth.ResolveUserID(c, h.users)
	if err != nil {
		return respond.Unauthorized(c, "invalid user")
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return respond.BadRequest(c, "invalid category id")
	}
	var in finance.UpdateCategoryInput
	if err := c.BodyParser(&in); err != nil {
		return respond.BadRequest(c, "invalid request body")
	}
	if err := validate.Struct(in); err != nil {
		return err
	}
	cat, err := h.svc.UpdateCategory(c.Context(), id, userID, in)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return respond.NotFound(c, "category not found")
		}
		return respond.Internal(c, err)
	}
	return respond.Data(c, cat)
}

func (h *Handler) deleteCategory(c *fiber.Ctx) error {
	userID, err := auth.ResolveUserID(c, h.users)
	if err != nil {
		return respond.Unauthorized(c, "invalid user")
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return respond.BadRequest(c, "invalid category id")
	}
	if err := h.svc.DeleteCategory(c.Context(), id, userID); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return respond.NotFound(c, "category not found")
		}
		return respond.Internal(c, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// ── Record handlers ───────────────────────────────────────────────────────────

func (h *Handler) listRecords(c *fiber.Ctx) error {
	userID, err := auth.ResolveUserID(c, h.users)
	if err != nil {
		return respond.Unauthorized(c, "invalid user")
	}
	month := c.Query("month")
	if month == "" {
		return respond.BadRequest(c, "month query param required (YYYY-MM)")
	}
	records, err := h.svc.ListRecords(c.Context(), userID, month)
	if err != nil {
		return respond.Internal(c, err)
	}
	if records == nil {
		records = []*finance.Record{}
	}
	return respond.Data(c, records)
}

func (h *Handler) createRecord(c *fiber.Ctx) error {
	userID, err := auth.ResolveUserID(c, h.users)
	if err != nil {
		return respond.Unauthorized(c, "invalid user")
	}
	var in finance.CreateRecordInput
	if err := c.BodyParser(&in); err != nil {
		return respond.BadRequest(c, "invalid request body")
	}
	if err := validate.Struct(in); err != nil {
		return err
	}
	rec, err := h.svc.CreateRecord(c.Context(), userID, in)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return respond.NotFound(c, err.Error())
		}
		if strings.Contains(err.Error(), "does not match") {
			return respond.BadRequest(c, err.Error())
		}
		return respond.Internal(c, err)
	}
	return respond.Created(c, rec)
}

func (h *Handler) deleteRecord(c *fiber.Ctx) error {
	userID, err := auth.ResolveUserID(c, h.users)
	if err != nil {
		return respond.Unauthorized(c, "invalid user")
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return respond.BadRequest(c, "invalid record id")
	}
	if err := h.svc.DeleteRecord(c.Context(), id, userID); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return respond.NotFound(c, "record not found")
		}
		return respond.Internal(c, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) summary(c *fiber.Ctx) error {
	userID, err := auth.ResolveUserID(c, h.users)
	if err != nil {
		return respond.Unauthorized(c, "invalid user")
	}
	month := c.Query("month")
	if month == "" {
		return respond.BadRequest(c, "month query param required (YYYY-MM)")
	}
	s, err := h.svc.MonthlySummary(c.Context(), userID, month)
	if err != nil {
		return respond.Internal(c, err)
	}
	return respond.Data(c, s)
}
