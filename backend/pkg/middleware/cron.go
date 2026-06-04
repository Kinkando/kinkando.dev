package middleware

import (
	"crypto/subtle"

	"github.com/gofiber/fiber/v2"
)

// CronAuth returns a Fiber middleware that validates the X-Cron-Secret request
// header against the configured secret. Comparison is constant-time to resist
// timing attacks. Mount this on any route driven by the external cron worker
// instead of the Firebase auth middleware.
func CronAuth(secret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		provided := c.Get("X-Cron-Secret")
		if subtle.ConstantTimeCompare([]byte(provided), []byte(secret)) != 1 {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid or missing cron secret")
		}
		return c.Next()
	}
}
