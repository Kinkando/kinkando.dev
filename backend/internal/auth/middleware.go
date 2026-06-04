package auth

import (
	"context"
	"strings"

	firebase "firebase.google.com/go/v4"
	firebaseauth "firebase.google.com/go/v4/auth"
	"github.com/gofiber/fiber/v2"
	"google.golang.org/api/option"
)

const (
	localKeyUserID = "user_id"
	localKeyEmail  = "user_email"
)

type Middleware struct {
	client *firebaseauth.Client
}

func NewMiddleware(ctx context.Context, credentials string) (*Middleware, error) {
	app, err := firebase.NewApp(ctx, nil, option.WithCredentialsJSON([]byte(credentials)))
	if err != nil {
		return nil, err
	}
	client, err := app.Auth(ctx)
	if err != nil {
		return nil, err
	}
	return &Middleware{client: client}, nil
}

func (m *Middleware) Require() fiber.Handler {
	return func(c *fiber.Ctx) error {
		header := c.Get("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "missing bearer token"})
		}
		idToken := strings.TrimPrefix(header, "Bearer ")
		token, err := m.client.VerifyIDToken(context.Background(), idToken)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid or expired token"})
		}
		c.Locals(localKeyUserID, token.UID)
		if email, ok := token.Claims["email"].(string); ok {
			c.Locals(localKeyEmail, email)
		}
		return c.Next()
	}
}

func GetUserID(c *fiber.Ctx) string {
	uid, _ := c.Locals(localKeyUserID).(string)
	return uid
}

func GetEmail(c *fiber.Ctx) string {
	email, _ := c.Locals(localKeyEmail).(string)
	return email
}

// Verify verifies a Firebase ID token and returns the UID and email it encodes.
// Use this outside of Fiber handlers where a fiber.Ctx is not available.
func (m *Middleware) Verify(ctx context.Context, idToken string) (uid, email string, err error) {
	token, err := m.client.VerifyIDToken(ctx, idToken)
	if err != nil {
		return "", "", err
	}
	uid = token.UID
	if e, ok := token.Claims["email"].(string); ok {
		email = e
	}
	return uid, email, nil
}
