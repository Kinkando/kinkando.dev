package gemini

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"google.golang.org/genai"
)

// Message represents a single turn in a chat conversation history.
type Message struct {
	Role string // "user" | "model" (or "assistant" — normalised to "model")
	Text string
}

const defaultModel = "gemini-2.0-flash"

// maxToolRounds caps the number of tool-call iterations in a single Chat to
// prevent infinite loops from a misbehaving model.
const maxToolRounds = 8

const defaultTTSModel = "gemini-2.5-flash-preview-tts"

// Deps bundles dependencies for the Gemini client.
type Deps struct {
	APIKey   string
	Model    string
	TTSModel string             // optional; defaults to defaultTTSModel
	MCP      *mcp.ClientSession // required; routes all tool calls through the MCP server
}

// Client wraps a Gemini generative model and dispatches tool calls via MCP.
// Each persona has its own pre-built GenerateContentConfig (system instruction + scoped
// tool declarations). The config is selected per-request based on the input text.
type Client struct {
	gc         *genai.Client
	model      string
	ttsModel   string
	configs    map[persona]*genai.GenerateContentConfig
	mcpSession *mcp.ClientSession
}

// New creates a Client. Returns an error if the Gemini API key is invalid or unreachable.
// Tool declarations and system instructions are built once here (not per-request) and
// stored per persona. The correct config is selected at call time by resolvePersona.
func New(ctx context.Context, d Deps) (*Client, error) {
	gc, err := genai.NewClient(ctx, &genai.ClientConfig{APIKey: d.APIKey})
	if err != nil {
		return nil, fmt.Errorf("gemini: new client: %w", err)
	}
	modelName := d.Model
	if modelName == "" {
		modelName = defaultModel
	}
	ttsModelName := d.TTSModel
	if ttsModelName == "" {
		ttsModelName = defaultTTSModel
	}
	configs := map[persona]*genai.GenerateContentConfig{
		personaAether: {
			SystemInstruction: genai.NewContentFromText(aetherInstruction, genai.RoleUser),
		},
		personaKaito: {
			Tools:             []*genai.Tool{{FunctionDeclarations: toolDecls("kanban_")}},
			SystemInstruction: genai.NewContentFromText(kaitoInstruction, genai.RoleUser),
		},
		personaMint: {
			Tools:             []*genai.Tool{{FunctionDeclarations: toolDecls("finance_")}},
			SystemInstruction: genai.NewContentFromText(mintInstruction, genai.RoleUser),
		},
		personaTensei: {
			Tools:             []*genai.Tool{{FunctionDeclarations: toolDecls("workout_", "food_", "sleep_")}},
			SystemInstruction: genai.NewContentFromText(tenseiInstruction, genai.RoleUser),
		},
		personaKusuri: {
			Tools:             []*genai.Tool{{FunctionDeclarations: toolDecls("medicine_")}},
			SystemInstruction: genai.NewContentFromText(kusuriInstruction, genai.RoleUser),
		},
	}
	return &Client{
		gc:         gc,
		model:      modelName,
		ttsModel:   ttsModelName,
		configs:    configs,
		mcpSession: d.MCP,
	}, nil
}

// Chat sends userMsg to Gemini, executes any tool calls via the MCP session in a
// loop, and returns the final text reply. The persona (and its scoped tools) is
// resolved from the message text.
func (c *Client) Chat(ctx context.Context, userMsg string) (string, error) {
	msg := fmt.Sprintf("[Today: %s]\n%s", time.Now().Format(time.DateOnly), userMsg)

	p := resolvePersona(nil, userMsg)
	chat, err := c.gc.Chats.Create(ctx, c.model, c.configs[p], nil)
	if err != nil {
		return "", fmt.Errorf("gemini: create chat: %w", err)
	}

	sendParts := []genai.Part{{Text: msg}}

	for round := range maxToolRounds {
		resp, err := chat.SendMessage(ctx, sendParts...)
		if err != nil {
			return "", fmt.Errorf("gemini: send message: %w", err)
		}
		if len(resp.Candidates) == 0 {
			return "", errors.New("gemini: no candidates in response")
		}

		var text string
		// Collect parts that carry a FunctionCall; they also carry ThoughtSignature.
		var fcParts []*genai.Part
		for _, p := range resp.Candidates[0].Content.Parts {
			if p.Text != "" {
				text = p.Text
			}
			if p.FunctionCall != nil {
				fcParts = append(fcParts, p)
			}
		}

		if len(fcParts) == 0 {
			return text, nil
		}

		// Last round reached with unresolved tool calls — return whatever text we have.
		if round == maxToolRounds-1 {
			if text != "" {
				return text, nil
			}
			return "", errors.New("gemini: tool-call loop exceeded max rounds")
		}

		sendParts = c.executeToolCalls(ctx, fcParts)
	}

	// Unreachable, but satisfies the compiler.
	return "", errors.New("gemini: tool-call loop exceeded max rounds")
}

// Usage holds token counts for a ChatStream call, summed across all tool-call rounds.
// Each round is a separate Gemini API call, so costs accumulate across rounds.
type Usage struct {
	InputTokens  int32
	OutputTokens int32
}

// ChatStream sends userMsg to Gemini with optional conversation history, streams
// text tokens through emit as they arrive, executes any tool calls via MCP in a
// loop, and returns when the model produces its final reply (or on error).
//
// Note: we drive GenerateContentStream directly rather than using gc.Chats, because
// Chat.SendStream records history only when every streaming chunk passes
// validateResponse — but metadata-only chunks (no candidates) always fail that
// check, making finalIsValid=false and dropping the model's function-call turn from
// curatedHistory. Round 2 would then send a function response with no preceding
// function call, causing the model to reply with a confused error. Managing the
// contents slice ourselves sidesteps that SDK bug entirely.
func (c *Client) ChatStream(ctx context.Context, history []Message, userMsg string, emit func(token string) error) (Usage, error) {
	msg := fmt.Sprintf("[Today: %s]\n%s", time.Now().Format(time.DateOnly), userMsg)

	// Build initial contents: prior history + current user message.
	contents := make([]*genai.Content, 0, len(history)+1)
	for _, m := range history {
		role := m.Role
		if role == "assistant" {
			role = genai.RoleModel
		}
		contents = append(contents, genai.NewContentFromText(m.Text, genai.Role(role)))
	}
	contents = append(contents, &genai.Content{
		Parts: []*genai.Part{{Text: msg}},
		Role:  genai.RoleUser,
	})

	p := resolvePersona(history, userMsg)

	var cumulative Usage

	for round := range maxToolRounds {
		var fcParts []*genai.Part
		var streamErr error
		var roundUsage *genai.GenerateContentResponseUsageMetadata
		var modelParts []*genai.Part

		for resp, err := range c.gc.Models.GenerateContentStream(ctx, c.model, contents, c.configs[p]) {
			if err != nil {
				streamErr = fmt.Errorf("gemini: stream: %w", err)
				break
			}
			if resp.UsageMetadata != nil {
				roundUsage = resp.UsageMetadata
			}
			if len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil {
				continue
			}
			for _, part := range resp.Candidates[0].Content.Parts {
				if part.Text != "" {
					if emitErr := emit(part.Text); emitErr != nil {
						return cumulative, emitErr
					}
				}
				if part.FunctionCall != nil {
					fcParts = append(fcParts, part)
				}
				// Only keep parts that have a data field in the API's oneof.
				// Bare ThoughtSignature parts (thinking chunks with no text/function call)
				// carry no data field and the API returns 400 when they are echoed back
				// in conversation history. ThoughtSignature on FunctionCall parts is
				// preserved and echoed on the corresponding FunctionResponse.
				if partHasData(part) {
					modelParts = append(modelParts, part)
				}
			}
		}
		if streamErr != nil {
			return cumulative, streamErr
		}

		if roundUsage != nil {
			cumulative.InputTokens += roundUsage.PromptTokenCount
			cumulative.OutputTokens += roundUsage.CandidatesTokenCount
		}

		if len(fcParts) == 0 {
			// No tool calls — final reply was already streamed.
			return cumulative, nil
		}

		// Last round reached with unresolved tool calls.
		if round == maxToolRounds-1 {
			return cumulative, errors.New("gemini: tool-call loop exceeded max rounds")
		}

		// Append the model's function-call turn so round 2 sees a complete conversation.
		if len(modelParts) > 0 {
			contents = append(contents, &genai.Content{
				Parts: modelParts,
				Role:  genai.RoleModel,
			})
		}

		// Execute tool calls and append the function-response turn.
		toolParts := c.executeToolCalls(ctx, fcParts)
		frParts := make([]*genai.Part, len(toolParts))
		for i := range toolParts {
			pp := toolParts[i]
			frParts[i] = &pp
		}
		contents = append(contents, &genai.Content{
			Parts: frParts,
			Role:  genai.RoleUser,
		})
	}

	// Unreachable, but satisfies the compiler.
	return cumulative, errors.New("gemini: tool-call loop exceeded max rounds")
}

// executeToolCalls dispatches each function call through MCP and returns the
// corresponding FunctionResponse parts (with ThoughtSignature echoed back).
func (c *Client) executeToolCalls(ctx context.Context, fcParts []*genai.Part) []genai.Part {
	parts := make([]genai.Part, 0, len(fcParts))
	for _, p := range fcParts {
		fc := p.FunctionCall
		res, callErr := c.mcpSession.CallTool(ctx, &mcp.CallToolParams{
			Name:      fc.Name,
			Arguments: fc.Args,
		})
		var response map[string]any
		switch {
		case callErr != nil:
			response = map[string]any{"error": callErr.Error()}
		case res.IsError:
			response = map[string]any{"error": contentText(res.Content)}
		default:
			// Use the "output" key as documented by Gemini; the whole map is
			// treated as function output when neither "output" nor "error" is
			// present, but being explicit avoids model confusion.
			response = map[string]any{"output": jsonFromContent(res.Content)}
		}
		// Echo ThoughtSignature back to satisfy thinking-model requirements.
		parts = append(parts, genai.Part{
			FunctionResponse: &genai.FunctionResponse{
				ID:       fc.ID,
				Name:     fc.Name,
				Response: response,
			},
			ThoughtSignature: p.ThoughtSignature,
		})
	}
	return parts
}

// partHasData reports whether p has at least one field in the API's Part.data
// oneof (text, function_call, inline_data, etc.). Bare ThoughtSignature parts
// have no data field and are rejected by the API with INVALID_ARGUMENT when
// echoed back in conversation history.
func partHasData(p *genai.Part) bool {
	return p.Text != "" ||
		p.FunctionCall != nil ||
		p.FunctionResponse != nil ||
		p.InlineData != nil ||
		p.FileData != nil ||
		p.ExecutableCode != nil ||
		p.CodeExecutionResult != nil
}

// contentText extracts the text from the first TextContent in a Content slice.
func contentText(contents []mcp.Content) string {
	for _, c := range contents {
		if tc, ok := c.(*mcp.TextContent); ok {
			return tc.Text
		}
	}
	return "tool error"
}

// jsonFromContent JSON-parses the text from the first TextContent into a
// JSON-native value (map[string]any, []any, string, float64, bool, nil).
func jsonFromContent(contents []mcp.Content) any {
	text := contentText(contents)
	var v any
	if err := json.Unmarshal([]byte(text), &v); err != nil {
		return text // fallback: return as plain string
	}
	return v
}
