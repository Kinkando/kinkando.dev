package notification

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// DiscordClient delivers messages to a Discord channel via an incoming webhook URL.
type DiscordClient struct {
	httpClient *http.Client
}

// NewDiscordClient returns a DiscordClient ready to use.
func NewDiscordClient() *DiscordClient {
	return &DiscordClient{httpClient: &http.Client{}}
}

type discordWebhookPayload struct {
	Content string `json:"content"`
}

// PostWebhook sends content as a Discord message to the given webhook URL.
func (d *DiscordClient) PostWebhook(ctx context.Context, webhookURL, content string) error {
	payload, err := json.Marshal(discordWebhookPayload{Content: content})
	if err != nil {
		return fmt.Errorf("discord: marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, webhookURL, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("discord: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := d.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("discord: send: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("discord webhook %d: %s", resp.StatusCode, string(b))
	}
	return nil
}
