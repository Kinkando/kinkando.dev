package handler

import (
	"context"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/kinkando/personal-dashboard/internal/auth"
	"github.com/kinkando/personal-dashboard/internal/workout"
	"github.com/kinkando/personal-dashboard/pkg/respond"
	"github.com/kinkando/personal-dashboard/pkg/validate"
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
	FinishSession(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*workout.Session, error)
	AddSessionExercise(ctx context.Context, sessionID uuid.UUID, userID uuid.UUID, in workout.AddSessionExerciseInput) (*workout.SessionExercise, error)
	DeleteSessionExercise(ctx context.Context, exID uuid.UUID, sessionID uuid.UUID, userID uuid.UUID) error
}

type Handler struct {
	svc   Service
	users auth.UserResolver
}

func New(svc Service, users auth.UserResolver) *Handler {
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
	router.Post("/sessions/:id/finish", h.finishSession)
	router.Post("/sessions/:id/exercises", h.addSessionExercise)
	router.Patch("/sessions/:id/exercises", h.bulkUpdateSessionExercises)
	router.Patch("/sessions/:id/exercises/:exId", h.updateSessionExercise)
	router.Delete("/sessions/:id/exercises/:exId", h.deleteSessionExercise)
	router.Delete("/sessions/:id", h.deleteSession)
}

// ── Preset handlers ───────────────────────────────────────────────────────────

func (h *Handler) listPresets(c *fiber.Ctx) error {
	userID, err := auth.ResolveUserID(c, h.users)
	if err != nil {
		return respond.Unauthorized(c, "invalid user")
	}
	presets, err := h.svc.ListPresets(c.Context(), userID)
	if err != nil {
		return respond.Internal(c, err)
	}
	if presets == nil {
		presets = []*workout.Preset{}
	}
	return respond.Data(c, presets)
}

func (h *Handler) getPreset(c *fiber.Ctx) error {
	userID, err := auth.ResolveUserID(c, h.users)
	if err != nil {
		return respond.Unauthorized(c, "invalid user")
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return respond.BadRequest(c, "invalid preset id")
	}
	preset, err := h.svc.GetPreset(c.Context(), id, userID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return respond.NotFound(c, "preset not found")
		}
		return respond.Internal(c, err)
	}
	return respond.Data(c, preset)
}

func (h *Handler) createPreset(c *fiber.Ctx) error {
	userID, err := auth.ResolveUserID(c, h.users)
	if err != nil {
		return respond.Unauthorized(c, "invalid user")
	}
	var in workout.CreatePresetInput
	if err := c.BodyParser(&in); err != nil {
		return respond.BadRequest(c, "invalid request body")
	}
	if err := validate.Struct(in); err != nil {
		return err
	}
	// Default empty section to "main" for each exercise.
	for i := range in.Exercises {
		if in.Exercises[i].Section == "" {
			in.Exercises[i].Section = workout.SectionMain
		}
	}
	preset, err := h.svc.CreatePreset(c.Context(), userID, in)
	if err != nil {
		return respond.Internal(c, err)
	}
	return respond.Created(c, preset)
}

func (h *Handler) updatePreset(c *fiber.Ctx) error {
	userID, err := auth.ResolveUserID(c, h.users)
	if err != nil {
		return respond.Unauthorized(c, "invalid user")
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return respond.BadRequest(c, "invalid preset id")
	}
	var in workout.UpdatePresetInput
	if err := c.BodyParser(&in); err != nil {
		return respond.BadRequest(c, "invalid request body")
	}
	if err := validate.Struct(in); err != nil {
		return err
	}
	// Default empty section to "main" for each exercise.
	for i := range in.Exercises {
		if in.Exercises[i].Section == "" {
			in.Exercises[i].Section = workout.SectionMain
		}
	}
	preset, err := h.svc.UpdatePreset(c.Context(), id, userID, in)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return respond.NotFound(c, "preset not found")
		}
		return respond.Internal(c, err)
	}
	return respond.Data(c, preset)
}

func (h *Handler) deletePreset(c *fiber.Ctx) error {
	userID, err := auth.ResolveUserID(c, h.users)
	if err != nil {
		return respond.Unauthorized(c, "invalid user")
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return respond.BadRequest(c, "invalid preset id")
	}
	if err := h.svc.DeletePreset(c.Context(), id, userID); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return respond.NotFound(c, "preset not found")
		}
		return respond.Internal(c, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// ── Schedule handlers ─────────────────────────────────────────────────────────

func (h *Handler) getSchedule(c *fiber.Ctx) error {
	userID, err := auth.ResolveUserID(c, h.users)
	if err != nil {
		return respond.Unauthorized(c, "invalid user")
	}
	entries, err := h.svc.GetSchedule(c.Context(), userID)
	if err != nil {
		return respond.Internal(c, err)
	}
	if entries == nil {
		entries = []*workout.ScheduleEntry{}
	}
	return respond.Data(c, entries)
}

func (h *Handler) setSchedule(c *fiber.Ctx) error {
	userID, err := auth.ResolveUserID(c, h.users)
	if err != nil {
		return respond.Unauthorized(c, "invalid user")
	}
	var in workout.SetScheduleInput
	if err := c.BodyParser(&in); err != nil {
		return respond.BadRequest(c, "invalid request body")
	}
	if err := validate.Struct(in); err != nil {
		return err
	}
	// Detect duplicate day_of_week values (cannot be expressed via struct tags).
	seen := make(map[int]bool)
	for _, e := range in.Entries {
		if seen[e.DayOfWeek] {
			return respond.BadRequest(c, "duplicate day_of_week in entries")
		}
		seen[e.DayOfWeek] = true
	}
	entries, err := h.svc.SetSchedule(c.Context(), userID, in.Entries)
	if err != nil {
		return respond.Internal(c, err)
	}
	return respond.Data(c, entries)
}

// ── Session handlers ──────────────────────────────────────────────────────────

func (h *Handler) listSessions(c *fiber.Ctx) error {
	userID, err := auth.ResolveUserID(c, h.users)
	if err != nil {
		return respond.Unauthorized(c, "invalid user")
	}

	// Default to the last 60 days.
	now := time.Now().UTC().Truncate(24 * time.Hour)
	from := now.AddDate(0, 0, -60)
	to := now

	if s := c.Query("from"); s != "" {
		t, err := time.Parse(time.DateOnly, s)
		if err != nil {
			return respond.BadRequest(c, "invalid from date")
		}
		from = t
	}
	if s := c.Query("to"); s != "" {
		t, err := time.Parse(time.DateOnly, s)
		if err != nil {
			return respond.BadRequest(c, "invalid to date")
		}
		to = t
	}

	sessions, err := h.svc.ListSessions(c.Context(), userID, from, to)
	if err != nil {
		return respond.Internal(c, err)
	}
	if sessions == nil {
		sessions = []*workout.Session{}
	}
	return respond.Data(c, sessions)
}

func (h *Handler) generateSession(c *fiber.Ctx) error {
	userID, err := auth.ResolveUserID(c, h.users)
	if err != nil {
		return respond.Unauthorized(c, "invalid user")
	}
	var in workout.GenerateSessionInput
	if err := c.BodyParser(&in); err != nil {
		return respond.BadRequest(c, "invalid request body")
	}
	session, err := h.svc.GenerateSession(c.Context(), userID, in.Date)
	if err != nil {
		if strings.Contains(err.Error(), "no preset scheduled") {
			return respond.NotFound(c, err.Error())
		}
		if strings.Contains(err.Error(), "invalid date") {
			return respond.BadRequest(c, err.Error())
		}
		return respond.Internal(c, err)
	}
	return respond.Created(c, session)
}

func (h *Handler) createSession(c *fiber.Ctx) error {
	userID, err := auth.ResolveUserID(c, h.users)
	if err != nil {
		return respond.Unauthorized(c, "invalid user")
	}
	var in workout.CreateSessionInput
	if err := c.BodyParser(&in); err != nil {
		return respond.BadRequest(c, "invalid request body")
	}
	if err := validate.Struct(in); err != nil {
		return err
	}
	// Cross-field rule: quick start requires type when no preset is selected.
	if in.PresetID == nil && in.Type == nil {
		return respond.BadRequest(c, "preset_id or type is required")
	}
	session, err := h.svc.CreateSession(c.Context(), userID, in)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return respond.NotFound(c, err.Error())
		}
		if strings.Contains(err.Error(), "invalid date") {
			return respond.BadRequest(c, err.Error())
		}
		return respond.Internal(c, err)
	}
	return respond.Created(c, session)
}

func (h *Handler) getSession(c *fiber.Ctx) error {
	userID, err := auth.ResolveUserID(c, h.users)
	if err != nil {
		return respond.Unauthorized(c, "invalid user")
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return respond.BadRequest(c, "invalid session id")
	}
	session, err := h.svc.GetSession(c.Context(), id, userID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return respond.NotFound(c, "session not found")
		}
		return respond.Internal(c, err)
	}
	return respond.Data(c, session)
}

func (h *Handler) updateSession(c *fiber.Ctx) error {
	userID, err := auth.ResolveUserID(c, h.users)
	if err != nil {
		return respond.Unauthorized(c, "invalid user")
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return respond.BadRequest(c, "invalid session id")
	}
	var in workout.UpdateSessionInput
	if err := c.BodyParser(&in); err != nil {
		return respond.BadRequest(c, "invalid request body")
	}
	if err := validate.Struct(in); err != nil {
		return err
	}
	session, err := h.svc.UpdateSession(c.Context(), id, userID, in)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return respond.NotFound(c, "session not found")
		}
		if strings.Contains(err.Error(), "already completed") {
			return respond.Conflict(c, err.Error())
		}
		return respond.Internal(c, err)
	}
	return respond.Data(c, session)
}

func (h *Handler) finishSession(c *fiber.Ctx) error {
	userID, err := auth.ResolveUserID(c, h.users)
	if err != nil {
		return respond.Unauthorized(c, "invalid user")
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return respond.BadRequest(c, "invalid session id")
	}
	session, err := h.svc.FinishSession(c.Context(), id, userID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return respond.NotFound(c, "session not found")
		}
		if strings.Contains(err.Error(), "already completed") {
			return respond.Conflict(c, err.Error())
		}
		return respond.Internal(c, err)
	}
	return respond.Data(c, session)
}

func (h *Handler) addSessionExercise(c *fiber.Ctx) error {
	userID, err := auth.ResolveUserID(c, h.users)
	if err != nil {
		return respond.Unauthorized(c, "invalid user")
	}
	sessionID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return respond.BadRequest(c, "invalid session id")
	}
	var in workout.AddSessionExerciseInput
	if err := c.BodyParser(&in); err != nil {
		return respond.BadRequest(c, "invalid request body")
	}
	if err := validate.Struct(in); err != nil {
		return err
	}
	// Default empty section to "main".
	if in.Section == "" {
		in.Section = workout.SectionMain
	}
	ex, err := h.svc.AddSessionExercise(c.Context(), sessionID, userID, in)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return respond.NotFound(c, err.Error())
		}
		if strings.Contains(err.Error(), "already completed") {
			return respond.Conflict(c, err.Error())
		}
		return respond.Internal(c, err)
	}
	return respond.Created(c, ex)
}

func (h *Handler) bulkUpdateSessionExercises(c *fiber.Ctx) error {
	userID, err := auth.ResolveUserID(c, h.users)
	if err != nil {
		return respond.Unauthorized(c, "invalid user")
	}
	sessionID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return respond.BadRequest(c, "invalid session id")
	}
	var in workout.BulkUpdateSessionExercisesInput
	if err := c.BodyParser(&in); err != nil {
		return respond.BadRequest(c, "invalid request body")
	}
	if err := validate.Struct(in); err != nil {
		return err
	}
	exercises, err := h.svc.BulkUpdateSessionExercises(c.Context(), sessionID, userID, in.Items)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return respond.NotFound(c, err.Error())
		}
		if strings.Contains(err.Error(), "already completed") {
			return respond.Conflict(c, err.Error())
		}
		return respond.Internal(c, err)
	}
	return respond.Data(c, exercises)
}

func (h *Handler) updateSessionExercise(c *fiber.Ctx) error {
	userID, err := auth.ResolveUserID(c, h.users)
	if err != nil {
		return respond.Unauthorized(c, "invalid user")
	}
	sessionID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return respond.BadRequest(c, "invalid session id")
	}
	exID, err := uuid.Parse(c.Params("exId"))
	if err != nil {
		return respond.BadRequest(c, "invalid exercise id")
	}
	var in workout.UpdateSessionExerciseInput
	if err := c.BodyParser(&in); err != nil {
		return respond.BadRequest(c, "invalid request body")
	}
	ex, err := h.svc.UpdateSessionExercise(c.Context(), exID, sessionID, userID, in)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return respond.NotFound(c, "session exercise not found")
		}
		if strings.Contains(err.Error(), "already completed") {
			return respond.Conflict(c, err.Error())
		}
		return respond.Internal(c, err)
	}
	return respond.Data(c, ex)
}

func (h *Handler) deleteSessionExercise(c *fiber.Ctx) error {
	userID, err := auth.ResolveUserID(c, h.users)
	if err != nil {
		return respond.Unauthorized(c, "invalid user")
	}
	sessionID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return respond.BadRequest(c, "invalid session id")
	}
	exID, err := uuid.Parse(c.Params("exId"))
	if err != nil {
		return respond.BadRequest(c, "invalid exercise id")
	}
	if err := h.svc.DeleteSessionExercise(c.Context(), exID, sessionID, userID); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return respond.NotFound(c, "session exercise not found")
		}
		if strings.Contains(err.Error(), "already completed") {
			return respond.Conflict(c, err.Error())
		}
		return respond.Internal(c, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) deleteSession(c *fiber.Ctx) error {
	userID, err := auth.ResolveUserID(c, h.users)
	if err != nil {
		return respond.Unauthorized(c, "invalid user")
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return respond.BadRequest(c, "invalid session id")
	}
	if err := h.svc.DeleteSession(c.Context(), id, userID); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return respond.NotFound(c, "session not found")
		}
		if strings.Contains(err.Error(), "already completed") {
			return respond.Conflict(c, err.Error())
		}
		return respond.Internal(c, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}
