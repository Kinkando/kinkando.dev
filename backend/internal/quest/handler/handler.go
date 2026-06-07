package handler

import (
	"context"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/kinkando/personal-dashboard/internal/auth"
	"github.com/kinkando/personal-dashboard/internal/quest"
	"github.com/kinkando/personal-dashboard/pkg/respond"
	"github.com/kinkando/personal-dashboard/pkg/validate"
)

// Service is the domain operations the handler depends on.
type Service interface {
	CreateQuest(ctx context.Context, userID uuid.UUID, in quest.CreateQuestInput) (*quest.Quest, error)
	ListQuests(ctx context.Context, userID uuid.UUID, questType string) ([]*quest.Quest, error)
	UpdateQuest(ctx context.Context, id uuid.UUID, userID uuid.UUID, in quest.UpdateQuestInput) (*quest.Quest, error)
	DeleteQuest(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
	SetActive(ctx context.Context, id uuid.UUID, userID uuid.UUID, active bool) (*quest.Quest, error)

	GetOverview(ctx context.Context, userID uuid.UUID) (*quest.Overview, error)
	GetStreaks(ctx context.Context, userID uuid.UUID) (*quest.StreakSummary, error)
	IncrementQuest(ctx context.Context, userID uuid.UUID, questID uuid.UUID) error
	DecrementQuest(ctx context.Context, userID uuid.UUID, questID uuid.UUID) error

	ListXPEvents(ctx context.Context, userID uuid.UUID, limit int) ([]*quest.XPEvent, error)
}

type Handler struct {
	svc   Service
	users auth.UserResolver
}

func New(svc Service, users auth.UserResolver) *Handler {
	return &Handler{svc: svc, users: users}
}

func (h *Handler) Register(router fiber.Router) {
	router.Get("/quests", h.listQuests)
	router.Post("/quests", h.createQuest)
	router.Patch("/quests/:id", h.updateQuest)
	router.Delete("/quests/:id", h.deleteQuest)
	router.Post("/quests/:id/activate", h.activateQuest)
	router.Post("/quests/:id/deactivate", h.deactivateQuest)

	router.Get("/overview", h.getOverview)
	router.Get("/streaks", h.getStreaks)

	router.Post("/quests/:id/increment", h.incrementQuest)
	router.Post("/quests/:id/decrement", h.decrementQuest)

	router.Get("/history", h.listHistory)
}

// ── Quest CRUD ────────────────────────────────────────────────────────────────

func (h *Handler) listQuests(c *fiber.Ctx) error {
	userID, err := auth.ResolveUserID(c, h.users)
	if err != nil {
		return respond.Unauthorized(c, "invalid user")
	}
	questType := c.Query("type")
	quests, err := h.svc.ListQuests(c.Context(), userID, questType)
	if err != nil {
		return respond.Internal(c, err)
	}
	if quests == nil {
		quests = []*quest.Quest{}
	}
	return respond.Data(c, quests)
}

func (h *Handler) createQuest(c *fiber.Ctx) error {
	userID, err := auth.ResolveUserID(c, h.users)
	if err != nil {
		return respond.Unauthorized(c, "invalid user")
	}
	var in quest.CreateQuestInput
	if err := c.BodyParser(&in); err != nil {
		return respond.BadRequest(c, "invalid request body")
	}
	if err := validate.Struct(in); err != nil {
		return err
	}
	q, err := h.svc.CreateQuest(c.Context(), userID, in)
	if err != nil {
		return respond.BadRequest(c, err.Error())
	}
	return respond.Created(c, q)
}

func (h *Handler) updateQuest(c *fiber.Ctx) error {
	userID, err := auth.ResolveUserID(c, h.users)
	if err != nil {
		return respond.Unauthorized(c, "invalid user")
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return respond.BadRequest(c, "invalid quest id")
	}
	var in quest.UpdateQuestInput
	if err := c.BodyParser(&in); err != nil {
		return respond.BadRequest(c, "invalid request body")
	}
	if err := validate.Struct(in); err != nil {
		return err
	}
	q, err := h.svc.UpdateQuest(c.Context(), id, userID, in)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return respond.NotFound(c, "quest not found")
		}
		return respond.BadRequest(c, err.Error())
	}
	return respond.Data(c, q)
}

func (h *Handler) deleteQuest(c *fiber.Ctx) error {
	userID, err := auth.ResolveUserID(c, h.users)
	if err != nil {
		return respond.Unauthorized(c, "invalid user")
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return respond.BadRequest(c, "invalid quest id")
	}
	if err := h.svc.DeleteQuest(c.Context(), id, userID); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return respond.NotFound(c, "quest not found")
		}
		return respond.Internal(c, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) activateQuest(c *fiber.Ctx) error {
	return h.setActive(c, true)
}

func (h *Handler) deactivateQuest(c *fiber.Ctx) error {
	return h.setActive(c, false)
}

func (h *Handler) setActive(c *fiber.Ctx, active bool) error {
	userID, err := auth.ResolveUserID(c, h.users)
	if err != nil {
		return respond.Unauthorized(c, "invalid user")
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return respond.BadRequest(c, "invalid quest id")
	}
	q, err := h.svc.SetActive(c.Context(), id, userID, active)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return respond.NotFound(c, "quest not found")
		}
		return respond.Internal(c, err)
	}
	return respond.Data(c, q)
}

// ── Overview ──────────────────────────────────────────────────────────────────

func (h *Handler) getOverview(c *fiber.Ctx) error {
	userID, err := auth.ResolveUserID(c, h.users)
	if err != nil {
		return respond.Unauthorized(c, "invalid user")
	}
	overview, err := h.svc.GetOverview(c.Context(), userID)
	if err != nil {
		return respond.Internal(c, err)
	}
	return respond.Data(c, overview)
}

// ── Streaks ───────────────────────────────────────────────────────────────────

func (h *Handler) getStreaks(c *fiber.Ctx) error {
	userID, err := auth.ResolveUserID(c, h.users)
	if err != nil {
		return respond.Unauthorized(c, "invalid user")
	}
	streaks, err := h.svc.GetStreaks(c.Context(), userID)
	if err != nil {
		return respond.Internal(c, err)
	}
	return respond.Data(c, streaks)
}

// ── Actions ───────────────────────────────────────────────────────────────────

func (h *Handler) incrementQuest(c *fiber.Ctx) error {
	userID, err := auth.ResolveUserID(c, h.users)
	if err != nil {
		return respond.Unauthorized(c, "invalid user")
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return respond.BadRequest(c, "invalid quest id")
	}
	if err := h.svc.IncrementQuest(c.Context(), userID, id); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return respond.NotFound(c, "quest not found")
		}
		return respond.BadRequest(c, err.Error())
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) decrementQuest(c *fiber.Ctx) error {
	userID, err := auth.ResolveUserID(c, h.users)
	if err != nil {
		return respond.Unauthorized(c, "invalid user")
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return respond.BadRequest(c, "invalid quest id")
	}
	if err := h.svc.DecrementQuest(c.Context(), userID, id); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return respond.NotFound(c, "quest not found")
		}
		return respond.BadRequest(c, err.Error())
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// ── History ───────────────────────────────────────────────────────────────────

func (h *Handler) listHistory(c *fiber.Ctx) error {
	userID, err := auth.ResolveUserID(c, h.users)
	if err != nil {
		return respond.Unauthorized(c, "invalid user")
	}
	limit := 50
	if s := c.Query("limit"); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n > 0 {
			limit = n
		}
	}
	events, err := h.svc.ListXPEvents(c.Context(), userID, limit)
	if err != nil {
		return respond.Internal(c, err)
	}
	if events == nil {
		events = []*quest.XPEvent{}
	}
	return respond.Data(c, events)
}
