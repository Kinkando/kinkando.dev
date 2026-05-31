package handler

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/kinkando/personal-dashboard/internal/gemini"
	"go.uber.org/zap"
)

// Deps bundles dependencies for the AI chat handler.
type Deps struct {
	Gemini *gemini.Client // required
	Logger *zap.Logger
}

// Handler handles AI chat requests from the web app.
type Handler struct {
	gemini *gemini.Client
	logger *zap.Logger
}

// New creates a Handler from the provided dependencies.
func New(d Deps) *Handler {
	return &Handler{
		gemini: d.Gemini,
		logger: d.Logger,
	}
}

// Register mounts routes onto the given router (auth middleware applied by caller).
func (h *Handler) Register(router fiber.Router) {
	router.Post("/", h.chat)
}

// chatRequest is the JSON body sent by the client.
type chatRequest struct {
	Messages []chatMessage `json:"messages"`
}

// chatMessage is a single turn in the conversation.
type chatMessage struct {
	Role    string `json:"role"`    // "user" | "assistant"
	Content string `json:"content"`
}

func (h *Handler) chat(c *fiber.Ctx) error {
	var req chatRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	if len(req.Messages) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "messages must not be empty"})
	}
	last := req.Messages[len(req.Messages)-1]
	if last.Role != "user" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "last message must have role \"user\""})
	}

	// Convert request messages to gemini.Message history (all except the last turn).
	history := make([]gemini.Message, 0, len(req.Messages)-1)
	for _, m := range req.Messages[:len(req.Messages)-1] {
		role := m.Role
		if role == "assistant" {
			role = "model"
		}
		history = append(history, gemini.Message{Role: role, Text: m.Content})
	}
	userMsg := last.Content

	// Set SSE headers before writing the body stream.
	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("X-Accel-Buffering", "no")

	// Stream the response. Fiber recycles the request context when the handler
	// returns, so we use a detached context — same pattern as the LINE handler.
	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		streamCtx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
		defer cancel()

		emit := func(token string) error {
			frame, _ := json.Marshal(map[string]string{"token": token})
			_, err := fmt.Fprintf(w, "data: %s\n\n", frame)
			if err != nil {
				return err
			}
			return w.Flush()
		}

		usage, err := h.gemini.ChatStream(streamCtx, history, userMsg, emit)
		if err != nil {
			if !strings.Contains(err.Error(), "context") {
				h.logger.Error("AI chat stream error", zap.Error(err))
			}
			frame, _ := json.Marshal(map[string]string{"error": "Something went wrong. Please try again."})
			fmt.Fprintf(w, "event: error\ndata: %s\n\n", frame) //nolint:errcheck
			w.Flush()                                            //nolint:errcheck
			return
		}

		usageFrame, _ := json.Marshal(map[string]int32{
			"inputTokens":  usage.InputTokens,
			"outputTokens": usage.OutputTokens,
		})
		fmt.Fprintf(w, "event: usage\ndata: %s\n\n", usageFrame) //nolint:errcheck
		fmt.Fprint(w, "event: done\ndata: {}\n\n")               //nolint:errcheck
		w.Flush()                                                  //nolint:errcheck
	})

	return nil
}
