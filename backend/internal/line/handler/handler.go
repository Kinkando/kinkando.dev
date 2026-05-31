package handler

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/kinkando/personal-dashboard/internal/gemini"
	linepkg "github.com/kinkando/personal-dashboard/internal/line"
	"go.uber.org/zap"
)

// Deps bundles dependencies for the LINE webhook handler.
type Deps struct {
	ChannelID     string
	ChannelSecret string
	Client        *linepkg.Client
	Gemini        *gemini.Client // required; routes messages through Gemini + MCP
	Logger        *zap.Logger
}

// Handler handles LINE Messaging API webhook events.
type Handler struct {
	channelID     string
	channelSecret string
	client        *linepkg.Client
	gemini        *gemini.Client
	logger        *zap.Logger
}

// New creates a Handler from the provided dependencies.
func New(d Deps) *Handler {
	return &Handler{
		channelID:     d.ChannelID,
		channelSecret: d.ChannelSecret,
		client:        d.Client,
		gemini:        d.Gemini,
		logger:        d.Logger,
	}
}

// Register mounts routes onto the given router (no auth middleware — webhook is
// self-authenticated via X-Line-Signature HMAC verification).
func (h *Handler) Register(router fiber.Router) {
	router.Post("/webhook", h.webhook)
}

func (h *Handler) webhook(c *fiber.Ctx) error {
	body := c.Body()

	sig := c.Get("X-Line-Signature")
	if !h.validSignature(body, sig) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid signature"})
	}

	var payload linepkg.WebhookBody
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	// Acknowledge 200 immediately — LINE retries on slow or non-2xx responses.
	// Each event is processed in its own goroutine with a detached context so
	// the reply is not cancelled when Fiber closes the request context.
	for _, ev := range payload.Events {
		if ev.Type != "message" || ev.Message == nil || ev.Message.Type != "text" || ev.ReplyToken == "" {
			continue
		}
		ev := ev
		go func() {
			bgCtx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()
			reply := h.handleText(bgCtx, ev.Message.Text)
			if err := h.client.Reply(bgCtx, ev.ReplyToken, []linepkg.ReplyMessage{linepkg.TextMessage(reply)}); err != nil {
				h.logger.Error("LINE reply failed", zap.String("replyToken", ev.ReplyToken), zap.Error(err))
			}
		}()
	}

	return c.SendStatus(fiber.StatusOK)
}

// handleText sends the text to Gemini (via MCP tool calls) and returns a
// human-readable reply. Always returns a non-empty string.
func (h *Handler) handleText(ctx context.Context, text string) string {
	if h.gemini == nil {
		return "Sorry, the assistant is unavailable right now."
	}
	reply, err := h.gemini.Chat(ctx, text)
	if err != nil {
		h.logger.Error("Gemini chat failed", zap.Error(err))
		return "Sorry, something went wrong. Please try again."
	}
	if strings.TrimSpace(reply) == "" {
		return "Sorry, I couldn't generate a response."
	}
	return reply
}

func (h *Handler) validSignature(body []byte, sig string) bool {
	mac := hmac.New(sha256.New, []byte(h.channelSecret))
	mac.Write(body) //nolint:errcheck
	expected := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(sig))
}
