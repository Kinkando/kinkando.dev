package gemini

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/generative-ai-go/genai"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
)

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
	model      *genai.GenerativeModel
	mcpSession *mcp.ClientSession
}

// New creates a Client. Returns an error if the Gemini API key is invalid or unreachable.
func New(ctx context.Context, d Deps) (*Client, error) {
	gc, err := genai.NewClient(ctx, option.WithAPIKey(d.APIKey))
	if err != nil {
		return nil, fmt.Errorf("gemini: new client: %w", err)
	}
	modelName := d.Model
	if modelName == "" {
		modelName = defaultModel
	}
	model := gc.GenerativeModel(modelName)
	model.Tools = []*genai.Tool{{FunctionDeclarations: AllTools()}}
	model.SystemInstruction = &genai.Content{
		Parts: []genai.Part{genai.Text(systemInstruction)},
	}
	return &Client{
		gc:         gc,
		model:      model,
		mcpSession: d.MCP,
	}, nil
}

// Close releases resources held by the underlying Gemini API client.
func (c *Client) Close() error { return c.gc.Close() }

// Chat sends userMsg to Gemini, executes any tool calls via the MCP session in a
// loop, and returns the final text reply.
func (c *Client) Chat(ctx context.Context, userMsg string) (string, error) {
	msg := fmt.Sprintf("[Today: %s]\n%s", time.Now().Format("2006-01-02"), userMsg)
	session := c.model.StartChat()
	parts := []genai.Part{genai.Text(msg)}

	for round := range maxToolRounds {
		resp, err := session.SendMessage(ctx, parts...)
		if err != nil {
			var gErr *googleapi.Error
			if errors.As(err, &gErr) {
				return "", fmt.Errorf("gemini: send message (code=%d body=%s): %w", gErr.Code, gErr.Body, err)
			}
			return "", fmt.Errorf("gemini: send message: %w", err)
		}
		if len(resp.Candidates) == 0 {
			return "", errors.New("gemini: no candidates in response")
		}

		var text string
		var calls []genai.FunctionCall
		for _, p := range resp.Candidates[0].Content.Parts {
			switch v := p.(type) {
			case genai.Text:
				text = string(v)
			case genai.FunctionCall:
				calls = append(calls, v)
			}
		}

		if len(calls) == 0 {
			return text, nil
		}

		// Last round reached with unresolved tool calls — return whatever text we have.
		if round == maxToolRounds-1 {
			if text != "" {
				return text, nil
			}
			return "", errors.New("gemini: tool-call loop exceeded max rounds")
		}

		parts = nil
		for _, fc := range calls {
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
				// Parse the JSON text from MCP Content (always set by the SDK) to
				// get JSON-native types (map[string]any etc.) that proto.Struct accepts.
				response = map[string]any{"result": jsonFromContent(res.Content)}
			}
			parts = append(parts, genai.FunctionResponse{
				Name:     fc.Name,
				Response: response,
			})
		}
	}

	// Unreachable, but satisfies the compiler.
	return "", errors.New("gemini: tool-call loop exceeded max rounds")
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
// The MCP SDK always serializes structured output to a TextContent JSON string,
// so this is equivalent to the old toJSONValue round-trip and produces types
// that structpb.NewStruct accepts.
func jsonFromContent(contents []mcp.Content) any {
	text := contentText(contents)
	var v any
	if err := json.Unmarshal([]byte(text), &v); err != nil {
		return text // fallback: return as plain string
	}
	return v
}
