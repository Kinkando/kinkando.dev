// Package respond provides small Fiber response helpers that enforce the
// project-wide JSON envelope shapes:
//
//	success: {"data": <payload>}
//	error:   {"error": "<message>"}
package respond

import (
	"github.com/gofiber/fiber/v2"
)

// Data sends HTTP 200 with {"data": v}.
func Data(c *fiber.Ctx, v any) error {
	return c.JSON(fiber.Map{"data": v})
}

// Created sends HTTP 201 with {"data": v}.
func Created(c *fiber.Ctx, v any) error {
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"data": v})
}

// Err sends the given HTTP status with {"error": msg}.
func Err(c *fiber.Ctx, status int, msg string) error {
	return c.Status(status).JSON(fiber.Map{"error": msg})
}

// BadRequest sends HTTP 400 with {"error": msg}.
func BadRequest(c *fiber.Ctx, msg string) error {
	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": msg})
}

// NotFound sends HTTP 404 with {"error": msg}.
func NotFound(c *fiber.Ctx, msg string) error {
	return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": msg})
}

// Unauthorized sends HTTP 401 with {"error": msg}.
func Unauthorized(c *fiber.Ctx, msg string) error {
	return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": msg})
}

// Conflict sends HTTP 409 with {"error": msg}.
func Conflict(c *fiber.Ctx, msg string) error {
	return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": msg})
}

// Internal sends HTTP 500 with {"error": err.Error()}.
func Internal(c *fiber.Ctx, err error) error {
	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
}
