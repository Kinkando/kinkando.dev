package middleware

import (
	"encoding/json"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

const maxBodyLog = 4096

// RequestLogger logs method, path, status, latency, IP, and request/response bodies (capped at 4KB).
func RequestLogger(log *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		var reqBody []byte
		if b := c.Body(); len(b) > 0 && len(b) <= maxBodyLog {
			reqBody = make([]byte, len(b))
			copy(reqBody, b)
		}

		chainErr := c.Next()

		latency := time.Since(start)
		status := c.Response().StatusCode()

		fields := []zap.Field{
			zap.String("method", c.Method()),
			zap.String("path", c.Path()),
			zap.Int("status", status),
			zap.Duration("latency", latency),
			zap.String("ip", c.IP()),
		}
		if qs := c.Request().URI().QueryString(); len(qs) > 0 {
			fields = append(fields, zap.ByteString("query", qs))
		}
		if len(reqBody) > 0 {
			var req any
			json.Unmarshal(reqBody, &req)
			fields = append(fields, zap.Any("request", req))
		}
		if rb := c.Response().Body(); len(rb) > 0 && len(rb) <= maxBodyLog {
			var res any
			json.Unmarshal(rb, &res)
			fields = append(fields, zap.Any("response", res))
		}

		switch {
		case status >= 500:
			log.Error("http", fields...)
		case status >= 400:
			log.Warn("http", fields...)
		default:
			log.Info("http", fields...)
		}

		return chainErr
	}
}
