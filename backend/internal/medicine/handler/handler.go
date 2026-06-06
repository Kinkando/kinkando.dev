package handler

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/kinkando/personal-dashboard/internal/auth"
	"github.com/kinkando/personal-dashboard/internal/medicine"
	"github.com/kinkando/personal-dashboard/pkg/validate"
)

type Service interface {
	ListMedicines(ctx context.Context, userID uuid.UUID, includeArchived bool, sourceType *medicine.SourceType) ([]*medicine.Medicine, error)
	CreateMedicine(ctx context.Context, userID uuid.UUID, in medicine.CreateMedicineInput) (*medicine.Medicine, error)
	UpdateMedicine(ctx context.Context, id uuid.UUID, userID uuid.UUID, in medicine.UpdateMedicineInput) (*medicine.Medicine, error)
	SetArchived(ctx context.Context, id uuid.UUID, userID uuid.UUID, archived bool) (*medicine.Medicine, error)

	Take(ctx context.Context, userID uuid.UUID, medicineID uuid.UUID, in medicine.TakeMedicineInput) (*medicine.MedicineIntake, *medicine.Medicine, error)

	AdjustStock(ctx context.Context, userID uuid.UUID, medicineID uuid.UUID, in medicine.AdjustStockInput) (*medicine.MedicineStockAdjustment, *medicine.Medicine, error)

	ListIntakes(ctx context.Context, userID uuid.UUID, opts medicine.ListIntakeOpts) ([]*medicine.MedicineIntake, error)

	ListStockAdjustments(ctx context.Context, userID uuid.UUID, opts medicine.ListAdjustmentOpts) ([]*medicine.MedicineStockAdjustment, error)
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
	// Collection routes — must be registered before /:id routes so
	// literal paths like /intakes are not matched by the /:id wildcard.
	router.Get("/", h.listMedicines)
	router.Post("/", h.createMedicine)
	router.Get("/intakes", h.listAllIntakes)
	router.Get("/stock-adjustments", h.listAllStockAdjustments)

	// Per-medicine routes
	router.Patch("/:id", h.updateMedicine)
	router.Post("/:id/archive", h.archiveMedicine)
	router.Post("/:id/unarchive", h.unarchiveMedicine)
	router.Post("/:id/take", h.takeMedicine)
	router.Post("/:id/stock", h.adjustStock)
	router.Get("/:id/intakes", h.listMedicineIntakes)
	router.Get("/:id/stock-adjustments", h.listMedicineStockAdjustments)
}

// ── Medicines ─────────────────────────────────────────────────────────────────

func (h *Handler) listMedicines(c *fiber.Ctx) error {
	userID, err := h.resolveUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid user"})
	}
	includeArchived := strings.ToLower(c.Query("include_archived")) == "true"
	sourceType, err := parseSourceType(c)
	if err != nil {
		return err
	}
	meds, err := h.svc.ListMedicines(c.Context(), userID, includeArchived, sourceType)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	if meds == nil {
		meds = []*medicine.Medicine{}
	}
	return c.JSON(fiber.Map{"data": meds})
}

func (h *Handler) createMedicine(c *fiber.Ctx) error {
	userID, err := h.resolveUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid user"})
	}
	var in medicine.CreateMedicineInput
	if err := c.BodyParser(&in); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	if err := validate.Struct(in); err != nil {
		return err
	}
	med, err := h.svc.CreateMedicine(c.Context(), userID, in)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"data": med})
}

func (h *Handler) updateMedicine(c *fiber.Ctx) error {
	userID, err := h.resolveUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid user"})
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid medicine id"})
	}
	var in medicine.UpdateMedicineInput
	if err := c.BodyParser(&in); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	if err := validate.Struct(in); err != nil {
		return err
	}
	med, err := h.svc.UpdateMedicine(c.Context(), id, userID, in)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "medicine not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"data": med})
}

func (h *Handler) archiveMedicine(c *fiber.Ctx) error {
	return h.setArchived(c, true)
}

func (h *Handler) unarchiveMedicine(c *fiber.Ctx) error {
	return h.setArchived(c, false)
}

func (h *Handler) setArchived(c *fiber.Ctx, archived bool) error {
	userID, err := h.resolveUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid user"})
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid medicine id"})
	}
	med, err := h.svc.SetArchived(c.Context(), id, userID, archived)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "medicine not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"data": med})
}

// ── Take ──────────────────────────────────────────────────────────────────────

func (h *Handler) takeMedicine(c *fiber.Ctx) error {
	userID, err := h.resolveUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid user"})
	}
	medicineID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid medicine id"})
	}
	var in medicine.TakeMedicineInput
	if err := c.BodyParser(&in); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	if err := validate.Struct(in); err != nil {
		return err
	}
	intake, med, err := h.svc.Take(c.Context(), userID, medicineID, in)
	if err != nil {
		if errors.Is(err, medicine.ErrInsufficientStock) {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "insufficient stock; set allow_negative to true to override"})
		}
		if strings.Contains(err.Error(), "not found") {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "medicine not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"data": fiber.Map{
		"intake":   intake,
		"medicine": med,
	}})
}

// ── Stock adjustment ──────────────────────────────────────────────────────────

func (h *Handler) adjustStock(c *fiber.Ctx) error {
	userID, err := h.resolveUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid user"})
	}
	medicineID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid medicine id"})
	}
	var in medicine.AdjustStockInput
	if err := c.BodyParser(&in); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	if err := validate.Struct(in); err != nil {
		return err
	}
	adj, med, err := h.svc.AdjustStock(c.Context(), userID, medicineID, in)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "medicine not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"data": fiber.Map{
		"adjustment": adj,
		"medicine":   med,
	}})
}

// ── Intakes ───────────────────────────────────────────────────────────────────

func (h *Handler) listAllIntakes(c *fiber.Ctx) error {
	return h.listIntakesHandler(c, nil)
}

func (h *Handler) listMedicineIntakes(c *fiber.Ctx) error {
	medicineID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid medicine id"})
	}
	return h.listIntakesHandler(c, &medicineID)
}

func (h *Handler) listIntakesHandler(c *fiber.Ctx, medicineID *uuid.UUID) error {
	userID, err := h.resolveUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid user"})
	}

	opts := medicine.ListIntakeOpts{MedicineID: medicineID}
	sourceType, err := parseSourceType(c)
	if err != nil {
		return err
	}
	opts.SourceType = sourceType
	if dateStr := c.Query("date"); dateStr != "" {
		t, err := time.Parse(time.DateOnly, dateStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid date format; use YYYY-MM-DD"})
		}
		opts.Date = &t
	}
	if limitStr := c.Query("limit"); limitStr != "" {
		n, err := strconv.Atoi(limitStr)
		if err != nil || n <= 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "limit must be a positive integer"})
		}
		opts.Limit = n
	}

	intakes, err := h.svc.ListIntakes(c.Context(), userID, opts)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	if intakes == nil {
		intakes = []*medicine.MedicineIntake{}
	}
	return c.JSON(fiber.Map{"data": intakes})
}

// ── Stock adjustments ─────────────────────────────────────────────────────────

func (h *Handler) listAllStockAdjustments(c *fiber.Ctx) error {
	return h.listStockAdjustmentsHandler(c, nil)
}

func (h *Handler) listMedicineStockAdjustments(c *fiber.Ctx) error {
	medicineID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid medicine id"})
	}
	return h.listStockAdjustmentsHandler(c, &medicineID)
}

func (h *Handler) listStockAdjustmentsHandler(c *fiber.Ctx, medicineID *uuid.UUID) error {
	userID, err := h.resolveUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid user"})
	}

	opts := medicine.ListAdjustmentOpts{MedicineID: medicineID}
	sourceType, err := parseSourceType(c)
	if err != nil {
		return err
	}
	opts.SourceType = sourceType
	if dateStr := c.Query("date"); dateStr != "" {
		t, err := time.Parse(time.DateOnly, dateStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid date format; use YYYY-MM-DD"})
		}
		opts.Date = &t
	}
	if limitStr := c.Query("limit"); limitStr != "" {
		n, err := strconv.Atoi(limitStr)
		if err != nil || n <= 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "limit must be a positive integer"})
		}
		opts.Limit = n
	}

	adjs, err := h.svc.ListStockAdjustments(c.Context(), userID, opts)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	if adjs == nil {
		adjs = []*medicine.MedicineStockAdjustment{}
	}
	return c.JSON(fiber.Map{"data": adjs})
}

// ── Helpers ───────────────────────────────────────────────────────────────────

// listQuery binds optional list filters from the query string so they can be
// validated declaratively via `validate` struct tags (see docs/backend/rules.md),
// rather than with hand-rolled checks.
type listQuery struct {
	SourceType string `query:"source_type" json:"source_type" validate:"omitempty,oneof=medication supplement"`
}

// parseSourceType binds and validates the optional ?source_type= query param.
// Returns nil when absent.
func parseSourceType(c *fiber.Ctx) (*medicine.SourceType, error) {
	var q listQuery
	if err := c.QueryParser(&q); err != nil {
		return nil, fiber.NewError(fiber.StatusBadRequest, "invalid query parameters")
	}
	if err := validate.Struct(q); err != nil {
		return nil, err // *fiber.Error 400 with a descriptive field message
	}
	if q.SourceType == "" {
		return nil, nil
	}
	st := medicine.SourceType(q.SourceType)
	return &st, nil
}

func (h *Handler) resolveUserID(c *fiber.Ctx) (uuid.UUID, error) {
	firebaseUID := auth.GetUserID(c)
	if firebaseUID == "" {
		return uuid.UUID{}, fiber.ErrUnauthorized
	}
	return h.users.GetIDByFirebaseUID(c.Context(), firebaseUID)
}
