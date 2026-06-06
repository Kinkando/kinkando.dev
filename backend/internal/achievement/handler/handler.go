package handler

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/kinkando/personal-dashboard/internal/achievement"
	"github.com/kinkando/personal-dashboard/internal/auth"
)

// Service is the domain operation the handler depends on.
type Service interface {
	Evaluate(ctx context.Context, userID uuid.UUID) (*achievement.Summary, error)
}

// UserResolver resolves a Firebase UID to the internal user UUID.
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
	router.Get("/", h.list)
}

// list evaluates and returns all badges with the user's progress.
func (h *Handler) list(c *fiber.Ctx) error {
	userID, err := h.resolveUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid user"})
	}
	summary, err := h.svc.Evaluate(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"data": summary})
}

func (h *Handler) resolveUserID(c *fiber.Ctx) (uuid.UUID, error) {
	firebaseUID := auth.GetUserID(c)
	if firebaseUID == "" {
		return uuid.UUID{}, fiber.ErrUnauthorized
	}
	return h.users.GetIDByFirebaseUID(c.Context(), firebaseUID)
}
