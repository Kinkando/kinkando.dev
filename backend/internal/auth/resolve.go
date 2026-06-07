package auth

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// UserResolver maps a Firebase UID to the internal user UUID stored in the users table.
type UserResolver interface {
	GetIDByFirebaseUID(ctx context.Context, firebaseUID string) (uuid.UUID, error)
}

// ResolveUserID extracts the Firebase UID from the request context and resolves it
// to the internal user UUID. Returns fiber.ErrUnauthorized when the UID is absent.
func ResolveUserID(c *fiber.Ctx, users UserResolver) (uuid.UUID, error) {
	uid := GetUserID(c)
	if uid == "" {
		return uuid.UUID{}, fiber.ErrUnauthorized
	}
	return users.GetIDByFirebaseUID(c.Context(), uid)
}
