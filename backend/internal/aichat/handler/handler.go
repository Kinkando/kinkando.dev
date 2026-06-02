package handler

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/kinkando/personal-dashboard/internal/gemini"
	"github.com/kinkando/personal-dashboard/pkg/validate"
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
	router.Post("/transcribe", h.transcribe)
	router.Post("/tts", h.tts)
}

// chatRequest is the JSON body sent by the client.
type chatRequest struct {
	Messages []chatMessage `json:"messages" validate:"required,min=1,dive"`
}

// chatMessage is a single turn in the conversation.
type chatMessage struct {
	Role    string `json:"role"    validate:"required,oneof=user assistant"` // "user" | "assistant"
	Content string `json:"content" validate:"required"`
}

func (h *Handler) chat(c *fiber.Ctx) error {
	var req chatRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	if err := validate.Struct(req); err != nil {
		return err
	}
	last := req.Messages[len(req.Messages)-1]
	// Cross-field rule: the final message must be from the user.
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

// maxAudioBytes caps audio uploads at 10 MiB to prevent oversized payloads.
const maxAudioBytes = 10 << 20

// transcribe accepts a multipart form upload with an "audio" field and returns
// the transcribed text via the Gemini model.
func (h *Handler) transcribe(c *fiber.Ctx) error {
	fh, err := c.FormFile("audio")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "audio file required"})
	}
	if fh.Size > maxAudioBytes {
		return c.Status(fiber.StatusRequestEntityTooLarge).JSON(fiber.Map{"error": "audio file too large (max 10 MiB)"})
	}

	f, err := fh.Open()
	if err != nil {
		h.logger.Error("transcribe: open upload", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to read audio"})
	}
	defer f.Close() //nolint:errcheck

	audio, err := io.ReadAll(f)
	if err != nil {
		h.logger.Error("transcribe: read upload", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to read audio"})
	}

	mimeType := fh.Header.Get("Content-Type")
	if mimeType == "" {
		mimeType = "audio/webm"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	text, err := h.gemini.Transcribe(ctx, audio, mimeType)
	if err != nil {
		h.logger.Error("transcribe: gemini error", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "transcription failed"})
	}

	return c.JSON(fiber.Map{"data": fiber.Map{"text": text}})
}

// ttsRequest is the JSON body for the TTS endpoint.
type ttsRequest struct {
	Text string `json:"text" validate:"required"`
}

// maxTTSChars caps TTS input to avoid excessively long synthesis requests.
const maxTTSChars = 4000

// tts synthesizes the provided text to WAV audio via Gemini and returns the
// raw bytes with Content-Type: audio/wav.
func (h *Handler) tts(c *fiber.Ctx) error {
	var req ttsRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	if err := validate.Struct(req); err != nil {
		return err
	}
	req.Text = strings.TrimSpace(req.Text)
	if req.Text == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "text must not be empty"})
	}
	if len([]rune(req.Text)) > maxTTSChars {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "text too long (max 4000 characters)"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	wav, err := h.gemini.Synthesize(ctx, req.Text)
	if err != nil {
		h.logger.Error("tts: gemini error", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "speech synthesis failed"})
	}

	c.Set("Content-Type", "audio/wav")
	return c.Send(wav)
}
