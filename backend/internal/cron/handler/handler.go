// Package handler exposes the cron-triggered batch endpoints.
// Routes are mounted outside Firebase auth and protected instead by the
// CronAuth shared-secret middleware.
package handler

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/kinkando/personal-dashboard/internal/medicine/reminder"
)

// ReminderRunner is the narrow interface the handler depends on.
type ReminderRunner interface {
	Run(ctx context.Context) (*reminder.RunResult, error)
}

type Handler struct {
	reminder ReminderRunner
}

func New(r ReminderRunner) *Handler {
	return &Handler{reminder: r}
}

func (h *Handler) Register(router fiber.Router) {
	router.Post("/medicine-reminders", h.runMedicineReminders)
}

// runMedicineReminders triggers the batch medicine reminder job.
// Called by the Cloudflare cron worker; protected by CronAuth middleware.
func (h *Handler) runMedicineReminders(c *fiber.Ctx) error {
	result, err := h.reminder.Run(c.Context())
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(result)
}
