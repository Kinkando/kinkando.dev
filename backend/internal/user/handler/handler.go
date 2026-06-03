package handler

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/kinkando/personal-dashboard/internal/auth"
	"github.com/kinkando/personal-dashboard/internal/user"
)

// UserResolver looks up the internal UUID for a Firebase UID.  Matches the
// interface already used by other domain handlers (e.g. finance).
type UserResolver interface {
	GetIDByFirebaseUID(ctx context.Context, firebaseUID string) (uuid.UUID, error)
}

type Service interface {
	GetOrCreate(ctx context.Context, firebaseUID, email string) (*user.User, error)
	GetByFirebaseUID(ctx context.Context, firebaseUID string) (*user.User, error)
	CreateLinkCode(ctx context.Context, userID uuid.UUID) (code string, expiresAt time.Time, err error)
	Unlink(ctx context.Context, userID uuid.UUID) error
}

type Handler struct {
	svc      Service
	resolver UserResolver
}

func New(svc Service, resolver UserResolver) *Handler {
	return &Handler{svc: svc, resolver: resolver}
}

func (h *Handler) Register(router fiber.Router) {
	router.Post("", h.ensureUser)
	router.Get("/me", h.getMe)
	router.Post("/line/link-code", h.createLineLinkCode)
	router.Delete("/line", h.unlinkLine)
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

// getMe returns the full user row (including line_id) for the authenticated user.
func (h *Handler) getMe(c *fiber.Ctx) error {
	firebaseUID := auth.GetUserID(c)
	if firebaseUID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid user"})
	}

	u, err := h.svc.GetByFirebaseUID(c.Context(), firebaseUID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	if u == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "user not found"})
	}
	return c.JSON(fiber.Map{"data": u})
}

// createLineLinkCode generates a one-time verification code the user can send
// to the LINE bot to link their account.
func (h *Handler) createLineLinkCode(c *fiber.Ctx) error {
	userID, err := h.resolveUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid user"})
	}

	code, expiresAt, err := h.svc.CreateLinkCode(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"data": fiber.Map{"code": code, "expires_at": expiresAt}})
}

// unlinkLine removes the LINE account association for the authenticated user.
func (h *Handler) unlinkLine(c *fiber.Ctx) error {
	userID, err := h.resolveUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid user"})
	}

	if err := h.svc.Unlink(c.Context(), userID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) resolveUserID(c *fiber.Ctx) (uuid.UUID, error) {
	firebaseUID := auth.GetUserID(c)
	if firebaseUID == "" {
		return uuid.UUID{}, fiber.ErrUnauthorized
	}
	return h.resolver.GetIDByFirebaseUID(c.Context(), firebaseUID)
}
