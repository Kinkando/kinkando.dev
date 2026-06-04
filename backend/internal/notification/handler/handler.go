package handler

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/kinkando/personal-dashboard/internal/auth"
	"github.com/kinkando/personal-dashboard/internal/notification"
	"github.com/kinkando/personal-dashboard/pkg/validate"
)

// Service is the business-logic interface the handler depends on.
type Service interface {
	GetSettings(ctx context.Context, userID uuid.UUID) (*notification.Settings, error)
	UpsertSettings(ctx context.Context, userID uuid.UUID, in notification.UpsertSettingsInput) (*notification.Settings, error)
	RegisterToken(ctx context.Context, userID uuid.UUID, token string) error
	RemoveToken(ctx context.Context, token string) error
	SendTest(ctx context.Context, userID uuid.UUID)
}

// UserResolver resolves a Firebase UID to the internal UUID.
type UserResolver interface {
	GetIDByFirebaseUID(ctx context.Context, firebaseUID string) (uuid.UUID, error)
}

// Handler wires HTTP routes to the notification service.
type Handler struct {
	svc   Service
	users UserResolver
}

// New constructs a Handler.
func New(svc Service, users UserResolver) *Handler {
	return &Handler{svc: svc, users: users}
}

// Register mounts all notification routes on router.
func (h *Handler) Register(router fiber.Router) {
	router.Get("/settings", h.getSettings)
	router.Put("/settings", h.upsertSettings)
	router.Post("/tokens", h.registerToken)
	router.Delete("/tokens", h.removeToken)
	router.Post("/test", h.sendTest)
}

// ── Handlers ──────────────────────────────────────────────────────────────────

func (h *Handler) getSettings(c *fiber.Ctx) error {
	userID, err := h.resolveUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid user"})
	}
	settings, err := h.svc.GetSettings(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"data": settings})
}

func (h *Handler) upsertSettings(c *fiber.Ctx) error {
	userID, err := h.resolveUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid user"})
	}
	var in notification.UpsertSettingsInput
	if err := c.BodyParser(&in); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	if err := validate.Struct(in); err != nil {
		return err
	}
	settings, err := h.svc.UpsertSettings(c.Context(), userID, in)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"data": settings})
}

func (h *Handler) registerToken(c *fiber.Ctx) error {
	userID, err := h.resolveUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid user"})
	}
	var in notification.RegisterTokenInput
	if err := c.BodyParser(&in); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	if err := validate.Struct(in); err != nil {
		return err
	}
	if err := h.svc.RegisterToken(c.Context(), userID, in.Token); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) removeToken(c *fiber.Ctx) error {
	// No user-scope check needed — token uniqueness guarantees ownership.
	var in notification.RemoveTokenInput
	if err := c.BodyParser(&in); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	if err := validate.Struct(in); err != nil {
		return err
	}
	if err := h.svc.RemoveToken(c.Context(), in.Token); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) sendTest(c *fiber.Ctx) error {
	userID, err := h.resolveUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid user"})
	}
	h.svc.SendTest(c.Context(), userID)
	return c.JSON(fiber.Map{"data": "test notification sent"})
}

// resolveUserID looks up the internal UUID for the Firebase UID in the request context.
func (h *Handler) resolveUserID(c *fiber.Ctx) (uuid.UUID, error) {
	firebaseUID := auth.GetUserID(c)
	if firebaseUID == "" {
		return uuid.UUID{}, fiber.ErrUnauthorized
	}
	return h.users.GetIDByFirebaseUID(c.Context(), firebaseUID)
}
