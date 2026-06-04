package line

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	replyAPIURL = "https://api.line.me/v2/bot/message/reply"
	pushAPIURL  = "https://api.line.me/v2/bot/message/push"
)

// Client sends messages back to LINE users via the Messaging API.
type Client struct {
	accessToken string
	httpClient  *http.Client
}

// NewClient creates a Client that authenticates with the given channel access token.
func NewClient(accessToken string) *Client {
	return &Client{
		accessToken: accessToken,
		httpClient:  &http.Client{},
	}
}

// ReplyMessage is a single LINE text message payload.
type ReplyMessage struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// TextMessage returns a ReplyMessage of type "text".
func TextMessage(text string) ReplyMessage {
	return ReplyMessage{Type: "text", Text: text}
}

type replyRequest struct {
	ReplyToken string         `json:"replyToken"`
	Messages   []ReplyMessage `json:"messages"`
}

type pushRequest struct {
	To       string         `json:"to"`
	Messages []ReplyMessage `json:"messages"`
}

// Push sends messages to the LINE user identified by to (a LINE user ID).
func (c *Client) Push(ctx context.Context, to string, messages []ReplyMessage) error {
	body, err := json.Marshal(pushRequest{To: to, Messages: messages})
	if err != nil {
		return fmt.Errorf("marshal push: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, pushAPIURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create push request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.accessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("send push: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("LINE push API %d: %s", resp.StatusCode, string(b))
	}
	return nil
}

// Reply sends messages to the user identified by replyToken.
func (c *Client) Reply(ctx context.Context, replyToken string, messages []ReplyMessage) error {
	body, err := json.Marshal(replyRequest{ReplyToken: replyToken, Messages: messages})
	if err != nil {
		return fmt.Errorf("marshal reply: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, replyAPIURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create reply request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.accessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("send reply: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("LINE reply API %d: %s", resp.StatusCode, string(b))
	}
	return nil
}
