package handler

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/kinkando/personal-dashboard/internal/news"
)

// Service is the domain operation the handler depends on.
type Service interface {
	List(ctx context.Context) ([]news.Item, error)
}

type Handler struct {
	svc Service
}

func New(svc Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Register(router fiber.Router) {
	router.Get("/", h.list)
}

// list returns the aggregated news feed. Public — no auth.
func (h *Handler) list(c *fiber.Ctx) error {
	items, err := h.svc.List(c.Context())
	if err != nil {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": err.Error()})
	}
	if items == nil {
		items = []news.Item{}
	}
	return c.JSON(fiber.Map{"data": items})
}
