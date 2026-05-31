package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/go-jet/jet/v2/postgres"
	"github.com/google/uuid"
	"github.com/kinkando/personal-dashboard/gen/kinkando/public/model"
	"github.com/kinkando/personal-dashboard/gen/kinkando/public/table"
	"github.com/kinkando/personal-dashboard/internal/user"
)

type Repository struct {
	db *sql.DB
}

func New(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// GetOrCreate inserts a users row for the given Firebase UID, or updates the
// email if the row already exists.  Returns the persisted user.
func (r *Repository) GetOrCreate(ctx context.Context, firebaseUID, email string) (*user.User, error) {
	stmt := table.Users.INSERT(table.Users.FirebaseUID, table.Users.Email).
		VALUES(firebaseUID, email).
		ON_CONFLICT(table.Users.FirebaseUID).
		DO_UPDATE(postgres.SET(
			table.Users.Email.SET(table.Users.EXCLUDED.Email),
		)).
		RETURNING(table.Users.AllColumns)

	var dest model.Users
	if err := stmt.QueryContext(ctx, r.db, &dest); err != nil {
		return nil, fmt.Errorf("get or create user: %w", err)
	}
	return toUser(dest), nil
}

// GetIDByFirebaseUID looks up the internal UUID primary key for a Firebase UID.
func (r *Repository) GetIDByFirebaseUID(ctx context.Context, firebaseUID string) (uuid.UUID, error) {
	stmt := postgres.SELECT(table.Users.ID).
		FROM(table.Users).
		WHERE(table.Users.FirebaseUID.EQ(postgres.String(firebaseUID)))

	var dest model.Users
	if err := stmt.QueryContext(ctx, r.db, &dest); err != nil {
		return uuid.UUID{}, fmt.Errorf("get user by firebase uid: %w", err)
	}
	return dest.ID, nil
}

func toUser(m model.Users) *user.User {
	return &user.User{
		ID:          m.ID,
		FirebaseUID: m.FirebaseUID,
		Email:       m.Email,
		CreatedAt:   m.CreatedAt,
	}
}
