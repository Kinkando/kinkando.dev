package notification

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// ErrTokenNotFound is returned when a device token to be removed does not exist.
var ErrTokenNotFound = errors.New("notification: fcm token not found")

// Channel identifies a notification delivery channel.
type Channel string

const (
	ChannelLINE     Channel = "line"
	ChannelDiscord  Channel = "discord"
	ChannelWebPush  Channel = "web_push"
)

// Settings holds the per-user notification preferences.
type Settings struct {
	ID                 uuid.UUID `json:"id"`
	UserID             uuid.UUID `json:"user_id"`
	LineEnabled        bool      `json:"line_enabled"`
	DiscordEnabled     bool      `json:"discord_enabled"`
	DiscordWebhookURL  *string   `json:"discord_webhook_url"`
	WebPushEnabled     bool      `json:"web_push_enabled"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// DefaultSettings returns an all-disabled settings object (no DB row yet).
func DefaultSettings() *Settings {
	return &Settings{
		LineEnabled:    false,
		DiscordEnabled: false,
		WebPushEnabled: false,
	}
}

// UpsertSettingsInput is the request body for PUT /notifications/settings.
type UpsertSettingsInput struct {
	LineEnabled        bool    `json:"line_enabled"`
	DiscordEnabled     bool    `json:"discord_enabled"`
	DiscordWebhookURL  *string `json:"discord_webhook_url"  validate:"omitempty,url"`
	WebPushEnabled     bool    `json:"web_push_enabled"`
}

// RegisterTokenInput is the request body for POST /notifications/tokens.
type RegisterTokenInput struct {
	Token string `json:"token" validate:"required"`
}

// RemoveTokenInput is the request body for DELETE /notifications/tokens.
type RemoveTokenInput struct {
	Token string `json:"token" validate:"required"`
}

// Message is the channel-agnostic notification payload used by Notify.
type Message struct {
	Title string
	Body  string
}
