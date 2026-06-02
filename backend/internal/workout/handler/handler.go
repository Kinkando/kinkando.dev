package handler

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/kinkando/personal-dashboard/internal/auth"
	"github.com/kinkando/personal-dashboard/internal/workout"
)

type Service interface {
	ListPresets(ctx context.Context, userID uuid.UUID) ([]*workout.Preset, error)
	GetPreset(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*workout.Preset, error)
	CreatePreset(ctx context.Context, userID uuid.UUID, in workout.CreatePresetInput) (*workout.Preset, error)
	UpdatePreset(ctx context.Context, id uuid.UUID, userID uuid.UUID, in workout.UpdatePresetInput) (*workout.Preset, error)
	DeletePreset(ctx context.Context, id uuid.UUID, userID uuid.UUID) error

	GetSchedule(ctx context.Context, userID uuid.UUID) ([]*workout.ScheduleEntry, error)
	SetSchedule(ctx context.Context, userID uuid.UUID, entries []workout.ScheduleEntryInput) ([]*workout.ScheduleEntry, error)

	ListSessions(ctx context.Context, userID uuid.UUID, from, to time.Time) ([]*workout.Session, error)
	GetSession(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*workout.Session, error)
	GenerateSession(ctx context.Context, userID uuid.UUID, date string) (*workout.Session, error)
	CreateSession(ctx context.Context, userID uuid.UUID, in workout.CreateSessionInput) (*workout.Session, error)
	UpdateSession(ctx context.Context, id uuid.UUID, userID uuid.UUID, in workout.UpdateSessionInput) (*workout.Session, error)
	UpdateSessionExercise(ctx context.Context, id uuid.UUID, sessionID uuid.UUID, userID uuid.UUID, in workout.UpdateSessionExerciseInput) (*workout.SessionExercise, error)
	BulkUpdateSessionExercises(ctx context.Context, sessionID uuid.UUID, userID uuid.UUID, items []workout.BulkUpdateSessionExerciseItem) ([]workout.SessionExercise, error)
	DeleteSession(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
	AddSessionExercise(ctx context.Context, sessionID uuid.UUID, userID uuid.UUID, in workout.AddSessionExerciseInput) (*workout.SessionExercise, error)
	DeleteSessionExercise(ctx context.Context, exID uuid.UUID, sessionID uuid.UUID, userID uuid.UUID) error
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
	// Presets
	router.Get("/presets", h.listPresets)
	router.Post("/presets", h.createPreset)
	router.Get("/presets/:id", h.getPreset)
	router.Put("/presets/:id", h.updatePreset)
	router.Delete("/presets/:id", h.deletePreset)

	// Schedule
	router.Get("/schedule", h.getSchedule)
	router.Put("/schedule", h.setSchedule)

	// Sessions — register /sessions/generate BEFORE /:id routes to ensure
	// Fiber's static segment takes priority over the param route.
	router.Get("/sessions", h.listSessions)
	router.Post("/sessions/generate", h.generateSession)
	router.Post("/sessions", h.createSession)
	router.Get("/sessions/:id", h.getSession)
	router.Patch("/sessions/:id", h.updateSession)
	router.Post("/sessions/:id/exercises", h.addSessionExercise)
	router.Patch("/sessions/:id/exercises", h.bulkUpdateSessionExercises)
	router.Patch("/sessions/:id/exercises/:exId", h.updateSessionExercise)
	router.Delete("/sessions/:id/exercises/:exId", h.deleteSessionExercise)
	router.Delete("/sessions/:id", h.deleteSession)
}

// ── Preset handlers ───────────────────────────────────────────────────────────

func (h *Handler) listPresets(c *fiber.Ctx) error {
	userID, err := h.resolveUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid user"})
	}
	presets, err := h.svc.ListPresets(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	if presets == nil {
		presets = []*workout.Preset{}
	}
	return c.JSON(fiber.Map{"data": presets})
}

func (h *Handler) getPreset(c *fiber.Ctx) error {
	userID, err := h.resolveUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid user"})
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid preset id"})
	}
	preset, err := h.svc.GetPreset(c.Context(), id, userID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "preset not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"data": preset})
}

func (h *Handler) createPreset(c *fiber.Ctx) error {
	userID, err := h.resolveUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid user"})
	}
	var in workout.CreatePresetInput
	if err := c.BodyParser(&in); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	if err := validatePresetInput(in.Name, in.Type, in.Exercises); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	preset, err := h.svc.CreatePreset(c.Context(), userID, in)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"data": preset})
}

func (h *Handler) updatePreset(c *fiber.Ctx) error {
	userID, err := h.resolveUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid user"})
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid preset id"})
	}
	var in workout.UpdatePresetInput
	if err := c.BodyParser(&in); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	if err := validatePresetInput(in.Name, in.Type, in.Exercises); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	preset, err := h.svc.UpdatePreset(c.Context(), id, userID, in)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "preset not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"data": preset})
}

func (h *Handler) deletePreset(c *fiber.Ctx) error {
	userID, err := h.resolveUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid user"})
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid preset id"})
	}
	if err := h.svc.DeletePreset(c.Context(), id, userID); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "preset not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// ── Schedule handlers ─────────────────────────────────────────────────────────

func (h *Handler) getSchedule(c *fiber.Ctx) error {
	userID, err := h.resolveUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid user"})
	}
	entries, err := h.svc.GetSchedule(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	if entries == nil {
		entries = []*workout.ScheduleEntry{}
	}
	return c.JSON(fiber.Map{"data": entries})
}

func (h *Handler) setSchedule(c *fiber.Ctx) error {
	userID, err := h.resolveUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid user"})
	}
	var in workout.SetScheduleInput
	if err := c.BodyParser(&in); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	// Validate day_of_week values and detect duplicates.
	seen := make(map[int]bool)
	for _, e := range in.Entries {
		if e.DayOfWeek < 0 || e.DayOfWeek > 6 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "day_of_week must be 0 (Sun) to 6 (Sat)"})
		}
		if seen[e.DayOfWeek] {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "duplicate day_of_week in entries"})
		}
		seen[e.DayOfWeek] = true
		if e.PresetID == uuid.Nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "preset_id is required for each entry"})
		}
	}

	entries, err := h.svc.SetSchedule(c.Context(), userID, in.Entries)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"data": entries})
}

// ── Session handlers ──────────────────────────────────────────────────────────

func (h *Handler) listSessions(c *fiber.Ctx) error {
	userID, err := h.resolveUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid user"})
	}

	// Default to the last 60 days.
	now := time.Now().UTC().Truncate(24 * time.Hour)
	from := now.AddDate(0, 0, -60)
	to := now

	if s := c.Query("from"); s != "" {
		t, err := time.Parse("2006-01-02", s)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid from date"})
		}
		from = t
	}
	if s := c.Query("to"); s != "" {
		t, err := time.Parse("2006-01-02", s)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid to date"})
		}
		to = t
	}

	sessions, err := h.svc.ListSessions(c.Context(), userID, from, to)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	if sessions == nil {
		sessions = []*workout.Session{}
	}
	return c.JSON(fiber.Map{"data": sessions})
}

func (h *Handler) generateSession(c *fiber.Ctx) error {
	userID, err := h.resolveUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid user"})
	}
	var in workout.GenerateSessionInput
	if err := c.BodyParser(&in); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	session, err := h.svc.GenerateSession(c.Context(), userID, in.Date)
	if err != nil {
		if strings.Contains(err.Error(), "no preset scheduled") {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
		}
		if strings.Contains(err.Error(), "invalid date") {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"data": session})
}

func (h *Handler) createSession(c *fiber.Ctx) error {
	userID, err := h.resolveUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid user"})
	}
	var in workout.CreateSessionInput
	if err := c.BodyParser(&in); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	// Quick start: no preset required, but type must be supplied and valid.
	if in.PresetID == nil {
		if in.Type == nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "preset_id or type is required"})
		}
		if !isValidType(*in.Type) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid workout type"})
		}
	}
	session, err := h.svc.CreateSession(c.Context(), userID, in)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
		}
		if strings.Contains(err.Error(), "invalid date") {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"data": session})
}

func (h *Handler) addSessionExercise(c *fiber.Ctx) error {
	userID, err := h.resolveUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid user"})
	}
	sessionID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid session id"})
	}
	var in workout.AddSessionExerciseInput
	if err := c.BodyParser(&in); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	if strings.TrimSpace(in.Name) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "exercise name is required"})
	}
	if in.Section == "" {
		in.Section = workout.SectionMain
	}
	switch in.Section {
	case workout.SectionWarmup, workout.SectionMain, workout.SectionCooldown:
	default:
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "exercise section must be warmup, main, or cooldown"})
	}
	ex, err := h.svc.AddSessionExercise(c.Context(), sessionID, userID, in)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"data": ex})
}

func (h *Handler) bulkUpdateSessionExercises(c *fiber.Ctx) error {
	userID, err := h.resolveUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid user"})
	}
	sessionID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid session id"})
	}
	var in workout.BulkUpdateSessionExercisesInput
	if err := c.BodyParser(&in); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	if len(in.Items) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "items must not be empty"})
	}
	exercises, err := h.svc.BulkUpdateSessionExercises(c.Context(), sessionID, userID, in.Items)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"data": exercises})
}

func (h *Handler) deleteSessionExercise(c *fiber.Ctx) error {
	userID, err := h.resolveUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid user"})
	}
	sessionID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid session id"})
	}
	exID, err := uuid.Parse(c.Params("exId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid exercise id"})
	}
	if err := h.svc.DeleteSessionExercise(c.Context(), exID, sessionID, userID); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "session exercise not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) getSession(c *fiber.Ctx) error {
	userID, err := h.resolveUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid user"})
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid session id"})
	}
	session, err := h.svc.GetSession(c.Context(), id, userID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "session not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"data": session})
}

func (h *Handler) updateSession(c *fiber.Ctx) error {
	userID, err := h.resolveUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid user"})
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid session id"})
	}
	var in workout.UpdateSessionInput
	if err := c.BodyParser(&in); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	if strings.TrimSpace(in.Name) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "name is required"})
	}
	session, err := h.svc.UpdateSession(c.Context(), id, userID, in)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "session not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"data": session})
}

func (h *Handler) updateSessionExercise(c *fiber.Ctx) error {
	userID, err := h.resolveUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid user"})
	}
	sessionID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid session id"})
	}
	exID, err := uuid.Parse(c.Params("exId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid exercise id"})
	}
	var in workout.UpdateSessionExerciseInput
	if err := c.BodyParser(&in); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	ex, err := h.svc.UpdateSessionExercise(c.Context(), exID, sessionID, userID, in)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "session exercise not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"data": ex})
}

func (h *Handler) deleteSession(c *fiber.Ctx) error {
	userID, err := h.resolveUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid user"})
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid session id"})
	}
	if err := h.svc.DeleteSession(c.Context(), id, userID); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "session not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// ── Helpers ───────────────────────────────────────────────────────────────────

// resolveUserID looks up the internal UUID for the Firebase UID in the request context.
func (h *Handler) resolveUserID(c *fiber.Ctx) (uuid.UUID, error) {
	firebaseUID := auth.GetUserID(c)
	if firebaseUID == "" {
		return uuid.UUID{}, fiber.ErrUnauthorized
	}
	return h.users.GetIDByFirebaseUID(c.Context(), firebaseUID)
}

// isValidType reports whether t is a known workout type.
func isValidType(t workout.Type) bool {
	switch t {
	case workout.TypeWeightTraining, workout.TypeBodyWeight, workout.TypeRunning,
		workout.TypeWalking, workout.TypeCardio, workout.TypeMobility, workout.TypeCustom:
		return true
	}
	return false
}

// validatePresetInput checks name, type enum, and per-exercise section enum.
// It also defaults empty section values to SectionMain in place.
// custom is excluded from presets (it is a quick-start-only type).
func validatePresetInput(name string, typ workout.Type, exercises []workout.PresetExerciseInput) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("name is required")
	}
	switch typ {
	case workout.TypeWeightTraining, workout.TypeBodyWeight, workout.TypeRunning,
		workout.TypeWalking, workout.TypeCardio, workout.TypeMobility:
	default:
		return fmt.Errorf("invalid preset type")
	}
	for i, ex := range exercises {
		if strings.TrimSpace(ex.Name) == "" {
			return fmt.Errorf("exercise name is required")
		}
		if ex.Section == "" {
			exercises[i].Section = workout.SectionMain
			continue
		}
		switch ex.Section {
		case workout.SectionWarmup, workout.SectionMain, workout.SectionCooldown:
		default:
			return fmt.Errorf("exercise section must be warmup, main, or cooldown")
		}
	}
	return nil
}
