package handler

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/kinkando/personal-dashboard/internal/achievement"
	"github.com/kinkando/personal-dashboard/internal/auth"
	"github.com/kinkando/personal-dashboard/pkg/respond"
)

// Service is the domain operation the handler depends on.
type Service interface {
	Evaluate(ctx context.Context, userID uuid.UUID) (*achievement.Summary, error)
}

type Handler struct {
	svc   Service
	users auth.UserResolver
}

func New(svc Service, users auth.UserResolver) *Handler {
	return &Handler{svc: svc, users: users}
}

func (h *Handler) Register(router fiber.Router) {
	router.Get("/", h.list)
}

// list evaluates and returns all badges with the user's progress.
func (h *Handler) list(c *fiber.Ctx) error {
	userID, err := auth.ResolveUserID(c, h.users)
	if err != nil {
		return respond.Unauthorized(c, "invalid user")
	}
	summary, err := h.svc.Evaluate(c.Context(), userID)
	if err != nil {
		return respond.Internal(c, err)
	}
	return respond.Data(c, summary)
}
