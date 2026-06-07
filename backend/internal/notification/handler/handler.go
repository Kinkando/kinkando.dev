package handler

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/kinkando/personal-dashboard/internal/auth"
	"github.com/kinkando/personal-dashboard/internal/notification"
	"github.com/kinkando/personal-dashboard/pkg/respond"
	"github.com/kinkando/personal-dashboard/pkg/validate"
)

// Service is the business-logic interface the handler depends on.
type Service interface {
	GetSettings(ctx context.Context, userID uuid.UUID) (*notification.Settings, error)
	UpsertSettings(ctx context.Context, userID uuid.UUID, in notification.UpsertSettingsInput) (*notification.Settings, error)
	RegisterToken(ctx context.Context, userID uuid.UUID, token string) error
	RemoveToken(ctx context.Context, token string) error
	IsTokenRegistered(ctx context.Context, userID uuid.UUID, token string) (bool, error)
	SendTest(ctx context.Context, userID uuid.UUID) *notification.DeliveryResult
}

// Handler wires HTTP routes to the notification service.
type Handler struct {
	svc   Service
	users auth.UserResolver
}

// New constructs a Handler.
func New(svc Service, users auth.UserResolver) *Handler {
	return &Handler{svc: svc, users: users}
}

// Register mounts all notification routes on router.
func (h *Handler) Register(router fiber.Router) {
	router.Get("/settings", h.getSettings)
	router.Put("/settings", h.upsertSettings)
	router.Post("/tokens", h.registerToken)
	router.Delete("/tokens", h.removeToken)
	router.Post("/tokens/check", h.checkToken)
	router.Post("/test", h.sendTest)
}

// ── Handlers ──────────────────────────────────────────────────────────────────

func (h *Handler) getSettings(c *fiber.Ctx) error {
	userID, err := auth.ResolveUserID(c, h.users)
	if err != nil {
		return respond.Unauthorized(c, "invalid user")
	}
	settings, err := h.svc.GetSettings(c.Context(), userID)
	if err != nil {
		return respond.Internal(c, err)
	}
	return respond.Data(c, settings)
}

func (h *Handler) upsertSettings(c *fiber.Ctx) error {
	userID, err := auth.ResolveUserID(c, h.users)
	if err != nil {
		return respond.Unauthorized(c, "invalid user")
	}
	var in notification.UpsertSettingsInput
	if err := c.BodyParser(&in); err != nil {
		return respond.BadRequest(c, "invalid request body")
	}
	if err := validate.Struct(in); err != nil {
		return err
	}
	settings, err := h.svc.UpsertSettings(c.Context(), userID, in)
	if err != nil {
		return respond.Internal(c, err)
	}
	return respond.Data(c, settings)
}

func (h *Handler) registerToken(c *fiber.Ctx) error {
	userID, err := auth.ResolveUserID(c, h.users)
	if err != nil {
		return respond.Unauthorized(c, "invalid user")
	}
	var in notification.RegisterTokenInput
	if err := c.BodyParser(&in); err != nil {
		return respond.BadRequest(c, "invalid request body")
	}
	if err := validate.Struct(in); err != nil {
		return err
	}
	if err := h.svc.RegisterToken(c.Context(), userID, in.Token); err != nil {
		return respond.Internal(c, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) removeToken(c *fiber.Ctx) error {
	// No user-scope check needed — token uniqueness guarantees ownership.
	var in notification.RemoveTokenInput
	if err := c.BodyParser(&in); err != nil {
		return respond.BadRequest(c, "invalid request body")
	}
	if err := validate.Struct(in); err != nil {
		return err
	}
	if err := h.svc.RemoveToken(c.Context(), in.Token); err != nil {
		return respond.Internal(c, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) checkToken(c *fiber.Ctx) error {
	userID, err := auth.ResolveUserID(c, h.users)
	if err != nil {
		return respond.Unauthorized(c, "invalid user")
	}
	var in notification.RegisterTokenInput
	if err := c.BodyParser(&in); err != nil {
		return respond.BadRequest(c, "invalid request body")
	}
	if err := validate.Struct(in); err != nil {
		return err
	}
	registered, err := h.svc.IsTokenRegistered(c.Context(), userID, in.Token)
	if err != nil {
		return respond.Internal(c, err)
	}
	return respond.Data(c, fiber.Map{"registered": registered})
}

func (h *Handler) sendTest(c *fiber.Ctx) error {
	userID, err := auth.ResolveUserID(c, h.users)
	if err != nil {
		return respond.Unauthorized(c, "invalid user")
	}
	res := h.svc.SendTest(c.Context(), userID)
	return respond.Data(c, res)
}
