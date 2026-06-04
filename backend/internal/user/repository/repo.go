package repository

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-jet/jet/v2/postgres"
	"github.com/go-jet/jet/v2/qrm"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
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

// GetByFirebaseUID returns the full user row (including line_id) for the given
// Firebase UID.  Returns nil, nil when no row is found.
func (r *Repository) GetByFirebaseUID(ctx context.Context, firebaseUID string) (*user.User, error) {
	stmt := postgres.SELECT(table.Users.AllColumns).
		FROM(table.Users).
		WHERE(table.Users.FirebaseUID.EQ(postgres.String(firebaseUID)))

	var dest model.Users
	if err := stmt.QueryContext(ctx, r.db, &dest); err != nil {
		if errors.Is(err, qrm.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get user by firebase uid: %w", err)
	}
	return toUser(dest), nil
}

// GetByID returns the full user row for the given internal UUID.
// Returns nil, nil when no row is found.
func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*user.User, error) {
	stmt := postgres.SELECT(table.Users.AllColumns).
		FROM(table.Users).
		WHERE(table.Users.ID.EQ(postgres.UUID(id)))

	var dest model.Users
	if err := stmt.QueryContext(ctx, r.db, &dest); err != nil {
		if errors.Is(err, qrm.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	return toUser(dest), nil
}

// GetByLineID returns the full user row for the given LINE user ID.
// Returns nil, nil when no row is found.
func (r *Repository) GetByLineID(ctx context.Context, lineUserID string) (*user.User, error) {
	stmt := postgres.SELECT(table.Users.AllColumns).
		FROM(table.Users).
		WHERE(table.Users.LineID.EQ(postgres.String(lineUserID)))

	var dest model.Users
	if err := stmt.QueryContext(ctx, r.db, &dest); err != nil {
		if errors.Is(err, qrm.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get user by line id: %w", err)
	}
	return toUser(dest), nil
}

// CreateLinkCode generates a one-time link code for the given user that expires
// in 10 minutes.  Any previous pending code for the same user is replaced.
func (r *Repository) CreateLinkCode(ctx context.Context, userID uuid.UUID) (code string, expiresAt time.Time, err error) {
	b := make([]byte, 3)
	if _, err = rand.Read(b); err != nil {
		return "", time.Time{}, fmt.Errorf("generate code: %w", err)
	}
	code = strings.ToUpper(hex.EncodeToString(b)) // exactly 6 uppercase hex chars
	expiresAt = time.Now().Add(10 * time.Minute)

	// Remove any existing code for this user, then insert the new one.
	delStmt := table.LineLinkCodes.DELETE().
		WHERE(table.LineLinkCodes.UserID.EQ(postgres.UUID(userID)))
	if _, err = delStmt.ExecContext(ctx, r.db); err != nil {
		return "", time.Time{}, fmt.Errorf("delete old link code: %w", err)
	}

	insStmt := table.LineLinkCodes.INSERT(
		table.LineLinkCodes.Code,
		table.LineLinkCodes.UserID,
		table.LineLinkCodes.ExpiresAt,
	).VALUES(code, userID, expiresAt)
	if _, err = insStmt.ExecContext(ctx, r.db); err != nil {
		return "", time.Time{}, fmt.Errorf("insert link code: %w", err)
	}
	return code, expiresAt, nil
}

// LinkByCode atomically validates a pending link code, stores the LINE user ID
// on the corresponding app user, and deletes the used code.
func (r *Repository) LinkByCode(ctx context.Context, code, lineUserID string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	// Look up a non-expired code row.
	sel := postgres.SELECT(table.LineLinkCodes.UserID).
		FROM(table.LineLinkCodes).
		WHERE(
			table.LineLinkCodes.Code.EQ(postgres.String(code)).
				AND(table.LineLinkCodes.ExpiresAt.GT(postgres.NOW())),
		).
		FOR(postgres.UPDATE())

	var codeRow model.LineLinkCodes
	if err := sel.QueryContext(ctx, tx, &codeRow); err != nil {
		if errors.Is(err, qrm.ErrNoRows) {
			return user.ErrLinkCodeInvalid
		}
		return fmt.Errorf("lookup link code: %w", err)
	}

	// Persist the LINE user ID on the matched app user.
	upd := table.Users.UPDATE(table.Users.LineID).
		SET(lineUserID).
		WHERE(table.Users.ID.EQ(postgres.UUID(codeRow.UserID)))
	if _, err := upd.ExecContext(ctx, tx); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return user.ErrLineAlreadyLinked
		}
		return fmt.Errorf("set line_id: %w", err)
	}

	// Delete the consumed code.
	del := table.LineLinkCodes.DELETE().
		WHERE(table.LineLinkCodes.Code.EQ(postgres.String(code)))
	if _, err := del.ExecContext(ctx, tx); err != nil {
		return fmt.Errorf("delete used code: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	return nil
}

// Unlink removes the LINE user ID from the given app user's row.
func (r *Repository) Unlink(ctx context.Context, userID uuid.UUID) error {
	stmt := table.Users.UPDATE(table.Users.LineID).
		SET((*string)(nil)).
		WHERE(table.Users.ID.EQ(postgres.UUID(userID)))
	if _, err := stmt.ExecContext(ctx, r.db); err != nil {
		return fmt.Errorf("unlink line: %w", err)
	}
	return nil
}

func toUser(m model.Users) *user.User {
	return &user.User{
		ID:          m.ID,
		FirebaseUID: m.FirebaseUID,
		Email:       m.Email,
		LineID:      m.LineID,
		CreatedAt:   m.CreatedAt,
	}
}
