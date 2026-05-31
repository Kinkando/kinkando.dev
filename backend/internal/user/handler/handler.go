package handler

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/kinkando/personal-dashboard/internal/auth"
	"github.com/kinkando/personal-dashboard/internal/user"
)

type Service interface {
	GetOrCreate(ctx context.Context, firebaseUID, email string) (*user.User, error)
}

type Handler struct {
	svc Service
}

func New(svc Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Register(router fiber.Router) {
	router.Post("", h.ensureUser)
}

// ensureUser idempotently provisions a users row for the authenticated Firebase
// account.  Called by the frontend on sign-in and sign-up.
func (h *Handler) ensureUser(c *fiber.Ctx) error {
	firebaseUID := auth.GetUserID(c)
	email := auth.GetEmail(c)
	if firebaseUID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid user"})
	}

	u, err := h.svc.GetOrCreate(c.Context(), firebaseUID, email)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"data": u})
}
