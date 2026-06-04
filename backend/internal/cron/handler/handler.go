// Package handler exposes the cron-triggered batch endpoints.
// Routes are mounted outside Firebase auth and protected instead by the
// CronAuth shared-secret middleware.
//
// Each entry in the runners map becomes POST /{name}. All runners return
// (any, error) so callers can use whatever result type they like — the handler
// JSON-serialises the value on success.
package handler

import (
	"context"

	"github.com/gofiber/fiber/v2"
)

// RunFunc is the common signature for a cron batch job.
type RunFunc func(ctx context.Context) (any, error)

// Handler holds a named set of batch runners, each exposed as POST /{name}.
type Handler struct {
	runners map[string]RunFunc
}

// New creates a Handler that registers one POST route per entry in runners.
func New(runners map[string]RunFunc) *Handler {
	return &Handler{runners: runners}
}

// Register mounts POST /{name} for every runner on the supplied router.
func (h *Handler) Register(router fiber.Router) {
	for name, run := range h.runners {
		run := run // capture loop var
		router.Post("/"+name, func(c *fiber.Ctx) error {
			result, err := run(c.Context())
			if err != nil {
				return fiber.NewError(fiber.StatusInternalServerError, err.Error())
			}
			return c.JSON(result)
		})
	}
}
