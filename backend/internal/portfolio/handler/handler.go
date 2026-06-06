package handler

import "github.com/gofiber/fiber/v2"

type Handler struct{}

func New() *Handler { return &Handler{} }

func (h *Handler) Register(router fiber.Router) {
	router.Get("/profile", h.profile)
	router.Get("/experience", h.experience)
	router.Get("/education", h.education)
	router.Get("/projects", h.projects)
	router.Get("/skills", h.skills)
}

func (h *Handler) profile(c *fiber.Ctx) error {
	data := fiber.Map{
		"name":    "Thanawat Yuwansiri",
		"title":   "Backend Developer",
		"summary": "I have over two years of experience developing RESTful APIs with Go, along with foundational skills in architecture design and solutions. I enjoy writing documentation on workflows and architecture to foster collaboration and mutual understanding.",
		"email":   "tanawat.yuwansiri@gmail.com",
		"github":  "https://github.com/kinkando",
	}
	return c.JSON(fiber.Map{"data": data})
}

func (h *Handler) experience(c *fiber.Ctx) error {
	data := []fiber.Map{
		{
			"role":    "Backend Developer",
			"company": "Nexter Digital & Solution Co., Ltd.",
			"period":  "May 2022 – Present",
			"highlights": []string{
				"Developed RESTful APIs using Go with a microservices architecture, enabling scalability via gRPC for inter-service communication.",
				"Implemented an automated payout system integrating with KGP, reducing payout time from 3 days to 1 day.",
				"Developed a payment system integrating with 2C2P — processing payments, voiding transactions, managing refunds, and real-time status updates via Firebase Realtime Database.",
				"Built a Customer Service Sentiment Solution: distributed rate limiter, call fetching/analysis services, and prompt tuning to reduce false positives in critical-call detection.",
				"Implemented CI/CD pipelines with GitHub Actions for automated deployment, release notes, and Discord notifications.",
				"Created a report generator to automate critical call reporting for company admins and supervisors.",
				"Developed a media processing service to convert images to WebP format using Pub/Sub and gRPC for async/sync request support.",
				"Implemented an API Gateway using KrakenD — flexible multi-project configuration, single-subdomain consolidation, and BFF-pattern response transformation.",
				"Improved query performance by optimizing indexes and query statements.",
				"Maintained and enhanced a shared Go microservices library focused on reusable functions.",
			},
		},
	}
	return c.JSON(fiber.Map{"data": data})
}

func (h *Handler) education(c *fiber.Ctx) error {
	data := []fiber.Map{
		{
			"school": "Silpakorn University",
			"degree": "Bachelor of Science, Computer Science",
			"detail": "GPA 3.79",
			"period": "2019 – 2023",
		},
	}
	return c.JSON(fiber.Map{"data": data})
}

func (h *Handler) projects(c *fiber.Ctx) error {
	data := []fiber.Map{
		{
			"name":        "Personal Dashboard",
			"description": "A self-hosted personal productivity dashboard with finance tracking, health logging, quest system, and more.",
			"url":         "https://github.com/kinkando/kinkando.dev",
			"tags":        []string{"Go", "React", "PostgreSQL", "MongoDB"},
		},
	}
	return c.JSON(fiber.Map{"data": data})
}

func (h *Handler) skills(c *fiber.Ctx) error {
	data := []fiber.Map{
		{"category": "Language", "items": []string{"Go", "JavaScript", "TypeScript"}},
		{"category": "Frontend", "items": []string{"ReactJS", "NextJS", "VueJS", "SvelteJS"}},
		{"category": "Backend", "items": []string{"Echo (Go)", "Node.js", "NestJS"}},
		{"category": "Database", "items": []string{"PostgreSQL", "MySQL", "MongoDB", "Redis"}},
		{"category": "DevOps", "items": []string{"Docker", "Kubernetes", "GitHub Actions", "Circle CI", "Argo CD"}},
		{"category": "Other", "items": []string{"Firebase", "Google Cloud", "gRPC", "WebSocket"}},
	}
	return c.JSON(fiber.Map{"data": data})
}
