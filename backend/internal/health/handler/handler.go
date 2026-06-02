package handler

import (
	"context"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/kinkando/personal-dashboard/internal/auth"
	"github.com/kinkando/personal-dashboard/internal/health"
	"github.com/kinkando/personal-dashboard/pkg/validate"
)

type Service interface {
	GetProfile(ctx context.Context, userID uuid.UUID) (*health.Profile, error)
	UpsertProfile(ctx context.Context, userID uuid.UUID, in health.UpsertProfileInput) (*health.Profile, error)

	ListWeightLogs(ctx context.Context, userID uuid.UUID) ([]*health.WeightLog, error)
	CreateWeightLog(ctx context.Context, userID uuid.UUID, in health.CreateWeightInput) (*health.WeightLog, error)
	DeleteWeightLog(ctx context.Context, id uuid.UUID, userID uuid.UUID) error

	ListFoodLogs(ctx context.Context, userID uuid.UUID) ([]*health.FoodLog, error)
	CreateFoodLog(ctx context.Context, userID uuid.UUID, in health.CreateFoodInput) (*health.FoodLog, error)
	UpdateFoodLog(ctx context.Context, id uuid.UUID, userID uuid.UUID, in health.UpdateFoodInput) (*health.FoodLog, error)
	DeleteFoodLog(ctx context.Context, id uuid.UUID, userID uuid.UUID) error

	ListSleepLogs(ctx context.Context, userID uuid.UUID) ([]*health.SleepLog, error)
	CreateSleepLog(ctx context.Context, userID uuid.UUID, in health.CreateSleepInput) (*health.SleepLog, error)
	UpdateSleepLog(ctx context.Context, id uuid.UUID, userID uuid.UUID, in health.UpdateSleepInput) (*health.SleepLog, error)
	DeleteSleepLog(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
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
	router.Get("/profile", h.getProfile)
	router.Put("/profile", h.upsertProfile)

	router.Get("/weight", h.listWeightLogs)
	router.Post("/weight", h.createWeightLog)
	router.Delete("/weight/:id", h.deleteWeightLog)

	router.Get("/food", h.listFoodLogs)
	router.Post("/food", h.createFoodLog)
	router.Patch("/food/:id", h.updateFoodLog)
	router.Delete("/food/:id", h.deleteFoodLog)

	router.Get("/sleep", h.listSleepLogs)
	router.Post("/sleep", h.createSleepLog)
	router.Patch("/sleep/:id", h.updateSleepLog)
	router.Delete("/sleep/:id", h.deleteSleepLog)
}

// ── Profile handlers ──────────────────────────────────────────────────────────

func (h *Handler) getProfile(c *fiber.Ctx) error {
	userID, err := h.resolveUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid user"})
	}
	profile, err := h.svc.GetProfile(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	// profile may be nil (no row yet) — return {"data": null} so the frontend
	// distinguishes "not set up yet" from an actual error.
	return c.JSON(fiber.Map{"data": profile})
}

func (h *Handler) upsertProfile(c *fiber.Ctx) error {
	userID, err := h.resolveUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid user"})
	}
	var in health.UpsertProfileInput
	if err := c.BodyParser(&in); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	if err := validate.Struct(in); err != nil {
		return err
	}
	profile, err := h.svc.UpsertProfile(c.Context(), userID, in)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"data": profile})
}

// ── Weight log handlers ───────────────────────────────────────────────────────

func (h *Handler) listWeightLogs(c *fiber.Ctx) error {
	userID, err := h.resolveUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid user"})
	}
	logs, err := h.svc.ListWeightLogs(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	if logs == nil {
		logs = []*health.WeightLog{}
	}
	return c.JSON(fiber.Map{"data": logs})
}

func (h *Handler) createWeightLog(c *fiber.Ctx) error {
	userID, err := h.resolveUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid user"})
	}
	var in health.CreateWeightInput
	if err := c.BodyParser(&in); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	if err := validate.Struct(in); err != nil {
		return err
	}
	log, err := h.svc.CreateWeightLog(c.Context(), userID, in)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"data": log})
}

func (h *Handler) deleteWeightLog(c *fiber.Ctx) error {
	userID, err := h.resolveUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid user"})
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid weight log id"})
	}
	if err := h.svc.DeleteWeightLog(c.Context(), id, userID); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "weight log not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// ── Food log handlers ─────────────────────────────────────────────────────────

func (h *Handler) listFoodLogs(c *fiber.Ctx) error {
	userID, err := h.resolveUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid user"})
	}
	logs, err := h.svc.ListFoodLogs(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	if logs == nil {
		logs = []*health.FoodLog{}
	}
	return c.JSON(fiber.Map{"data": logs})
}

func (h *Handler) createFoodLog(c *fiber.Ctx) error {
	userID, err := h.resolveUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid user"})
	}
	var in health.CreateFoodInput
	if err := c.BodyParser(&in); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	if err := validate.Struct(in); err != nil {
		return err
	}
	log, err := h.svc.CreateFoodLog(c.Context(), userID, in)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"data": log})
}

func (h *Handler) updateFoodLog(c *fiber.Ctx) error {
	userID, err := h.resolveUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid user"})
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid food log id"})
	}
	var in health.UpdateFoodInput
	if err := c.BodyParser(&in); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	if err := validate.Struct(in); err != nil {
		return err
	}
	log, err := h.svc.UpdateFoodLog(c.Context(), id, userID, in)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "food log not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"data": log})
}

func (h *Handler) deleteFoodLog(c *fiber.Ctx) error {
	userID, err := h.resolveUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid user"})
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid food log id"})
	}
	if err := h.svc.DeleteFoodLog(c.Context(), id, userID); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "food log not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// ── Sleep log handlers ────────────────────────────────────────────────────────

func (h *Handler) listSleepLogs(c *fiber.Ctx) error {
	userID, err := h.resolveUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid user"})
	}
	logs, err := h.svc.ListSleepLogs(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	if logs == nil {
		logs = []*health.SleepLog{}
	}
	return c.JSON(fiber.Map{"data": logs})
}

func (h *Handler) createSleepLog(c *fiber.Ctx) error {
	userID, err := h.resolveUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid user"})
	}
	var in health.CreateSleepInput
	if err := c.BodyParser(&in); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	if err := validate.Struct(in); err != nil {
		return err
	}
	log, err := h.svc.CreateSleepLog(c.Context(), userID, in)
	if err != nil {
		if strings.Contains(err.Error(), "ended_at must be after started_at") {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"data": log})
}

func (h *Handler) updateSleepLog(c *fiber.Ctx) error {
	userID, err := h.resolveUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid user"})
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid sleep log id"})
	}
	var in health.UpdateSleepInput
	if err := c.BodyParser(&in); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	if err := validate.Struct(in); err != nil {
		return err
	}
	log, err := h.svc.UpdateSleepLog(c.Context(), id, userID, in)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "sleep log not found"})
		}
		if strings.Contains(err.Error(), "ended_at must be after started_at") {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"data": log})
}

func (h *Handler) deleteSleepLog(c *fiber.Ctx) error {
	userID, err := h.resolveUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid user"})
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid sleep log id"})
	}
	if err := h.svc.DeleteSleepLog(c.Context(), id, userID); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "sleep log not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// resolveUserID looks up the internal UUID for the Firebase UID in the request context.
func (h *Handler) resolveUserID(c *fiber.Ctx) (uuid.UUID, error) {
	firebaseUID := auth.GetUserID(c)
	if firebaseUID == "" {
		return uuid.UUID{}, fiber.ErrUnauthorized
	}
	return h.users.GetIDByFirebaseUID(c.Context(), firebaseUID)
}
