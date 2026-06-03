package user

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Sentinel errors for LINE account-linking operations.
var (
	ErrLinkCodeInvalid   = errors.New("link code is invalid or expired")
	ErrLineAlreadyLinked = errors.New("this LINE account is already linked to another user")
)

type User struct {
	ID          uuid.UUID `json:"id"`
	FirebaseUID string    `json:"firebase_uid"`
	Email       string    `json:"email"`
	LineID      *string   `json:"line_id"`
	CreatedAt   time.Time `json:"created_at"`
}
