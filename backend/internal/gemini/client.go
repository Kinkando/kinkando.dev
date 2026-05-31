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

const systemInstruction = `You are a personal dashboard assistant for finance tracking and kanban task management.
Reply concisely in the same language the user writes in.
Always use tools to read or write data — never fabricate records or IDs.
When creating a finance record, call finance_list_categories first unless you already know the exact category name.
When creating a kanban card, call kanban_get_board first unless you already have the column ID.`

// maxToolRounds caps the number of tool-call iterations in a single Chat to
// prevent infinite loops from a misbehaving model.
const maxToolRounds = 8

// Deps bundles dependencies for the Gemini client.
type Deps struct {
	APIKey string
	Model  string
	MCP    *mcp.ClientSession // required; routes all tool calls through the MCP server
}

// Client wraps a Gemini generative model and dispatches tool calls via MCP.
type Client struct {
	gc         *genai.Client
	model      string
	config     *genai.GenerateContentConfig
	mcpSession *mcp.ClientSession
}

// New creates a Client. Returns an error if the Gemini API key is invalid or unreachable.
func New(ctx context.Context, d Deps) (*Client, error) {
	gc, err := genai.NewClient(ctx, &genai.ClientConfig{APIKey: d.APIKey})
	if err != nil {
		return nil, fmt.Errorf("gemini: new client: %w", err)
	}
	modelName := d.Model
	if modelName == "" {
		modelName = defaultModel
	}
	config := &genai.GenerateContentConfig{
		Tools:             []*genai.Tool{{FunctionDeclarations: AllTools()}},
		SystemInstruction: genai.NewContentFromText(systemInstruction, genai.RoleUser),
	}
	return &Client{
		gc:         gc,
		model:      modelName,
		config:     config,
		mcpSession: d.MCP,
	}, nil
}

// Chat sends userMsg to Gemini, executes any tool calls via the MCP session in a
// loop, and returns the final text reply.
func (c *Client) Chat(ctx context.Context, userMsg string) (string, error) {
	msg := fmt.Sprintf("[Today: %s]\n%s", time.Now().Format("2006-01-02"), userMsg)

	chat, err := c.gc.Chats.Create(ctx, c.model, c.config, nil)
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

// ChatStream sends userMsg to Gemini with optional conversation history, streams
// text tokens through emit as they arrive, executes any tool calls via MCP in a
// loop, and returns when the model produces its final reply (or on error).
func (c *Client) ChatStream(ctx context.Context, history []Message, userMsg string, emit func(token string) error) error {
	msg := fmt.Sprintf("[Today: %s]\n%s", time.Now().Format("2006-01-02"), userMsg)

	// Build history contents for the chat session.
	historyContents := make([]*genai.Content, 0, len(history))
	for _, m := range history {
		role := m.Role
		if role == "assistant" {
			role = genai.RoleModel
		}
		historyContents = append(historyContents, genai.NewContentFromText(m.Text, genai.Role(role)))
	}

	chat, err := c.gc.Chats.Create(ctx, c.model, c.config, historyContents)
	if err != nil {
		return fmt.Errorf("gemini: create chat: %w", err)
	}

	sendParts := []genai.Part{{Text: msg}}

	for round := range maxToolRounds {
		var fcParts []*genai.Part
		var streamErr error

		for resp, err := range chat.SendMessageStream(ctx, sendParts...) {
			if err != nil {
				streamErr = fmt.Errorf("gemini: stream: %w", err)
				break
			}
			if len(resp.Candidates) == 0 {
				continue
			}
			for _, p := range resp.Candidates[0].Content.Parts {
				if p.Text != "" {
					if emitErr := emit(p.Text); emitErr != nil {
						return emitErr
					}
				}
				if p.FunctionCall != nil {
					part := p
					fcParts = append(fcParts, part)
				}
			}
		}
		if streamErr != nil {
			return streamErr
		}

		if len(fcParts) == 0 {
			// No tool calls — final reply was already streamed.
			return nil
		}

		// Last round reached with unresolved tool calls.
		if round == maxToolRounds-1 {
			return errors.New("gemini: tool-call loop exceeded max rounds")
		}

		sendParts = c.executeToolCalls(ctx, fcParts)
	}

	// Unreachable, but satisfies the compiler.
	return errors.New("gemini: tool-call loop exceeded max rounds")
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
			response = map[string]any{"result": jsonFromContent(res.Content)}
		}
		// Echo ThoughtSignature back to satisfy thinking-model requirements.
		parts = append(parts, genai.Part{
			FunctionResponse: &genai.FunctionResponse{
				Name:     fc.Name,
				Response: response,
			},
			ThoughtSignature: p.ThoughtSignature,
		})
	}
	return parts
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
