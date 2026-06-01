package handler

import (
	"context"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/kinkando/personal-dashboard/internal/auth"
	"github.com/kinkando/personal-dashboard/internal/health"
)

type Service interface {
	GetProfile(ctx context.Context, userID uuid.UUID) (*health.Profile, error)
	UpsertProfile(ctx context.Context, userID uuid.UUID, in health.UpsertProfileInput) (*health.Profile, error)

	ListWeightLogs(ctx context.Context, userID uuid.UUID) ([]*health.WeightLog, error)
	CreateWeightLog(ctx context.Context, userID uuid.UUID, in health.CreateWeightInput) (*health.WeightLog, error)
	DeleteWeightLog(ctx context.Context, id uuid.UUID, userID uuid.UUID) error

	ListExercises(ctx context.Context, userID uuid.UUID) ([]*health.Exercise, error)
	CreateExercise(ctx context.Context, userID uuid.UUID, in health.CreateExerciseInput) (*health.Exercise, error)
	UpdateExercise(ctx context.Context, id uuid.UUID, userID uuid.UUID, in health.UpdateExerciseInput) (*health.Exercise, error)
	DeleteExercise(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
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

	router.Get("/exercises", h.listExercises)
	router.Post("/exercises", h.createExercise)
	router.Patch("/exercises/:id", h.updateExercise)
	router.Delete("/exercises/:id", h.deleteExercise)
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
	if in.Height != nil && *in.Height <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "height must be positive"})
	}
	if in.Age != nil && *in.Age <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "age must be positive"})
	}
	if in.Gender != nil {
		switch *in.Gender {
		case health.GenderMale, health.GenderFemale, health.GenderOther:
		default:
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "gender must be male, female, or other"})
		}
	}
	if in.Goal != nil {
		switch *in.Goal {
		case health.GoalLoseWeight, health.GoalMaintain, health.GoalGainMuscle:
		default:
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "goal must be lose_weight, maintain, or gain_muscle"})
		}
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
	if in.Weight <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "weight must be positive"})
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

// ── Exercise handlers ─────────────────────────────────────────────────────────

func (h *Handler) listExercises(c *fiber.Ctx) error {
	userID, err := h.resolveUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid user"})
	}
	exercises, err := h.svc.ListExercises(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	if exercises == nil {
		exercises = []*health.Exercise{}
	}
	return c.JSON(fiber.Map{"data": exercises})
}

func (h *Handler) createExercise(c *fiber.Ctx) error {
	userID, err := h.resolveUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid user"})
	}
	var in health.CreateExerciseInput
	if err := c.BodyParser(&in); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	if strings.TrimSpace(in.Name) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "name is required"})
	}
	switch in.Type {
	case health.ExerciseTypeCardio, health.ExerciseTypeStrength, health.ExerciseTypeFlexibility:
	default:
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "type must be cardio, strength, or flexibility"})
	}
	ex, err := h.svc.CreateExercise(c.Context(), userID, in)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"data": ex})
}

func (h *Handler) updateExercise(c *fiber.Ctx) error {
	userID, err := h.resolveUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid user"})
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid exercise id"})
	}
	var in health.UpdateExerciseInput
	if err := c.BodyParser(&in); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	if strings.TrimSpace(in.Name) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "name is required"})
	}
	switch in.Type {
	case health.ExerciseTypeCardio, health.ExerciseTypeStrength, health.ExerciseTypeFlexibility:
	default:
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "type must be cardio, strength, or flexibility"})
	}
	ex, err := h.svc.UpdateExercise(c.Context(), id, userID, in)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "exercise not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"data": ex})
}

func (h *Handler) deleteExercise(c *fiber.Ctx) error {
	userID, err := h.resolveUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid user"})
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid exercise id"})
	}
	if err := h.svc.DeleteExercise(c.Context(), id, userID); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "exercise not found"})
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
