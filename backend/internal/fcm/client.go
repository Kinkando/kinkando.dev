// Package fcm provides a thin wrapper around the Firebase Cloud Messaging API.
// It reuses the same Firebase service-account credentials used by the auth middleware.
package fcm

import (
	"context"
	"errors"
	"fmt"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"
)

// ErrTokenInvalid is returned when FCM reports that a device token is no longer
// registered.  Callers should remove the token from their store.
var ErrTokenInvalid = errors.New("fcm: device token is invalid or unregistered")

// Client wraps the FCM messaging client.
type Client struct {
	mc *messaging.Client
}

// NewClient initialises a new FCM client using a Firebase service-account
// credentials JSON string (the same value as FIREBASE_CREDENTIALS).
func NewClient(ctx context.Context, credentials string) (*Client, error) {
	app, err := firebase.NewApp(ctx, nil, option.WithCredentialsJSON([]byte(credentials)))
	if err != nil {
		return nil, fmt.Errorf("fcm: init firebase app: %w", err)
	}
	mc, err := app.Messaging(ctx)
	if err != nil {
		return nil, fmt.Errorf("fcm: get messaging client: %w", err)
	}
	return &Client{mc: mc}, nil
}

// Send delivers a web-push notification to the given device token.
// Returns ErrTokenInvalid when FCM tells us the token is stale so the caller
// can prune it from the database.
func (c *Client) Send(ctx context.Context, token, title, body string) error {
	// Data-only message: no Notification field so the browser never auto-displays
	// a duplicate notification. The service worker (onBackgroundMessage) and the
	// foreground onMessage handler both read payload.data and render via SW.showNotification.
	msg := &messaging.Message{
		Token: token,
		Data:  map[string]string{"title": title, "body": body},
	}
	if _, err := c.mc.Send(ctx, msg); err != nil {
		if messaging.IsUnregistered(err) {
			return ErrTokenInvalid
		}
		return fmt.Errorf("fcm: send: %w", err)
	}
	return nil
}
