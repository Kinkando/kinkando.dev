package handler

import "github.com/gofiber/fiber/v2"

type Handler struct{}

func New() *Handler { return &Handler{} }

func (h *Handler) Register(router fiber.Router) {
	router.Get("/projects", h.projects)
	router.Get("/skills", h.skills)
}

func (h *Handler) projects(c *fiber.Ctx) error {
	data := []fiber.Map{
		{
			"name":        "Personal Dashboard",
			"description": "A self-hosted personal productivity dashboard.",
			"url":         "https://github.com/kinkando/personal-dashboard",
			"tags":        []string{"Go", "React", "PostgreSQL", "MongoDB"},
		},
	}
	return c.JSON(fiber.Map{"data": data})
}

func (h *Handler) skills(c *fiber.Ctx) error {
	data := []fiber.Map{
		{"category": "Languages", "items": []string{"Go", "TypeScript", "Python"}},
		{"category": "Backend", "items": []string{"Fiber", "gRPC", "REST"}},
		{"category": "Databases", "items": []string{"PostgreSQL", "MongoDB", "Redis"}},
		{"category": "Infrastructure", "items": []string{"Docker", "Kubernetes", "GCP"}},
	}
	return c.JSON(fiber.Map{"data": data})
}
