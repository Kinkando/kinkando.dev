package handler

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/kinkando/personal-dashboard/internal/auth"
	"github.com/kinkando/personal-dashboard/internal/user"
	"github.com/kinkando/personal-dashboard/pkg/respond"
)

type Service interface {
	GetOrCreate(ctx context.Context, firebaseUID, email string) (*user.User, error)
	GetByFirebaseUID(ctx context.Context, firebaseUID string) (*user.User, error)
	CreateLinkCode(ctx context.Context, userID uuid.UUID) (code string, expiresAt time.Time, err error)
	Unlink(ctx context.Context, userID uuid.UUID) error
}

type Handler struct {
	svc      Service
	resolver auth.UserResolver
}

func New(svc Service, resolver auth.UserResolver) *Handler {
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
		return respond.Unauthorized(c, "invalid user")
	}
	u, err := h.svc.GetOrCreate(c.Context(), firebaseUID, email)
	if err != nil {
		return respond.Internal(c, err)
	}
	return respond.Data(c, u)
}

// getMe returns the full user row (including line_id) for the authenticated user.
func (h *Handler) getMe(c *fiber.Ctx) error {
	firebaseUID := auth.GetUserID(c)
	if firebaseUID == "" {
		return respond.Unauthorized(c, "invalid user")
	}
	u, err := h.svc.GetByFirebaseUID(c.Context(), firebaseUID)
	if err != nil {
		return respond.Internal(c, err)
	}
	if u == nil {
		return respond.NotFound(c, "user not found")
	}
	return respond.Data(c, u)
}

// createLineLinkCode generates a one-time verification code the user can send
// to the LINE bot to link their account.
func (h *Handler) createLineLinkCode(c *fiber.Ctx) error {
	userID, err := auth.ResolveUserID(c, h.resolver)
	if err != nil {
		return respond.Unauthorized(c, "invalid user")
	}
	code, expiresAt, err := h.svc.CreateLinkCode(c.Context(), userID)
	if err != nil {
		return respond.Internal(c, err)
	}
	return respond.Data(c, fiber.Map{"code": code, "expires_at": expiresAt})
}

// unlinkLine removes the LINE account association for the authenticated user.
func (h *Handler) unlinkLine(c *fiber.Ctx) error {
	userID, err := auth.ResolveUserID(c, h.resolver)
	if err != nil {
		return respond.Unauthorized(c, "invalid user")
	}
	if err := h.svc.Unlink(c.Context(), userID); err != nil {
		return respond.Internal(c, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}
