package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/go-jet/jet/v2/postgres"
	"github.com/go-jet/jet/v2/qrm"
	"github.com/google/uuid"
	"github.com/kinkando/personal-dashboard/gen/kinkando/public/model"
	"github.com/kinkando/personal-dashboard/gen/kinkando/public/table"
	"github.com/kinkando/personal-dashboard/internal/notification"
)

type Repository struct {
	db *sql.DB
}

func New(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// ── Settings ──────────────────────────────────────────────────────────────────

// GetSettings returns the notification settings for the given user.
// Returns nil, nil when no row exists yet.
func (r *Repository) GetSettings(ctx context.Context, userID uuid.UUID) (*notification.Settings, error) {
	stmt := postgres.SELECT(table.NotificationSettings.AllColumns).
		FROM(table.NotificationSettings).
		WHERE(table.NotificationSettings.UserID.EQ(postgres.UUID(userID)))

	var dest model.NotificationSettings
	if err := stmt.QueryContext(ctx, r.db, &dest); err != nil {
		if errors.Is(err, qrm.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get notification settings: %w", err)
	}
	return toSettings(dest), nil
}

// UpsertSettings creates or updates the notification settings for the given user.
func (r *Repository) UpsertSettings(ctx context.Context, userID uuid.UUID, in notification.UpsertSettingsInput) (*notification.Settings, error) {
	stmt := table.NotificationSettings.INSERT(
		table.NotificationSettings.UserID,
		table.NotificationSettings.LineEnabled,
		table.NotificationSettings.DiscordEnabled,
		table.NotificationSettings.DiscordWebhookURL,
		table.NotificationSettings.WebPushEnabled,
	).VALUES(
		postgres.UUID(userID),
		in.LineEnabled,
		in.DiscordEnabled,
		in.DiscordWebhookURL,
		in.WebPushEnabled,
	).ON_CONFLICT(table.NotificationSettings.UserID).
		DO_UPDATE(postgres.SET(
			table.NotificationSettings.LineEnabled.SET(table.NotificationSettings.EXCLUDED.LineEnabled),
			table.NotificationSettings.DiscordEnabled.SET(table.NotificationSettings.EXCLUDED.DiscordEnabled),
			table.NotificationSettings.DiscordWebhookURL.SET(table.NotificationSettings.EXCLUDED.DiscordWebhookURL),
			table.NotificationSettings.WebPushEnabled.SET(table.NotificationSettings.EXCLUDED.WebPushEnabled),
			table.NotificationSettings.UpdatedAt.SET(postgres.NOW()),
		)).RETURNING(table.NotificationSettings.AllColumns)

	var dest model.NotificationSettings
	if err := stmt.QueryContext(ctx, r.db, &dest); err != nil {
		return nil, fmt.Errorf("upsert notification settings: %w", err)
	}
	return toSettings(dest), nil
}

// ── FCM tokens ────────────────────────────────────────────────────────────────

// AddToken stores a new FCM device token for the user.
// If the token already exists (owned by anyone), the insert is silently skipped.
func (r *Repository) AddToken(ctx context.Context, userID uuid.UUID, token string) error {
	stmt := table.FcmTokens.INSERT(
		table.FcmTokens.UserID,
		table.FcmTokens.Token,
	).VALUES(
		postgres.UUID(userID),
		token,
	).ON_CONFLICT(table.FcmTokens.Token).DO_NOTHING()

	if _, err := stmt.ExecContext(ctx, r.db); err != nil {
		return fmt.Errorf("add fcm token: %w", err)
	}
	return nil
}

// DeleteToken removes an FCM device token.
// Returns notification.ErrTokenNotFound when the token doesn't exist.
func (r *Repository) DeleteToken(ctx context.Context, token string) error {
	stmt := table.FcmTokens.DELETE().
		WHERE(table.FcmTokens.Token.EQ(postgres.String(token)))

	res, err := stmt.ExecContext(ctx, r.db)
	if err != nil {
		return fmt.Errorf("delete fcm token: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete fcm token rows affected: %w", err)
	}
	if n == 0 {
		return notification.ErrTokenNotFound
	}
	return nil
}

// HasToken reports whether the given FCM device token is registered for userID.
func (r *Repository) HasToken(ctx context.Context, userID uuid.UUID, token string) (bool, error) {
	stmt := postgres.SELECT(postgres.COUNT(postgres.STAR)).
		FROM(table.FcmTokens).
		WHERE(
			table.FcmTokens.UserID.EQ(postgres.UUID(userID)).
				AND(table.FcmTokens.Token.EQ(postgres.String(token))),
		)

	var count struct{ Count int64 }
	if err := stmt.QueryContext(ctx, r.db, &count); err != nil {
		return false, fmt.Errorf("has fcm token: %w", err)
	}
	return count.Count > 0, nil
}

// ListTokens returns all FCM device tokens for the given user.
func (r *Repository) ListTokens(ctx context.Context, userID uuid.UUID) ([]string, error) {
	stmt := postgres.SELECT(table.FcmTokens.Token).
		FROM(table.FcmTokens).
		WHERE(table.FcmTokens.UserID.EQ(postgres.UUID(userID)))

	var dest []model.FcmTokens
	if err := stmt.QueryContext(ctx, r.db, &dest); err != nil {
		return nil, fmt.Errorf("list fcm tokens: %w", err)
	}
	tokens := make([]string, len(dest))
	for i, t := range dest {
		tokens[i] = t.Token
	}
	return tokens, nil
}

// ── mapping ───────────────────────────────────────────────────────────────────

func toSettings(m model.NotificationSettings) *notification.Settings {
	return &notification.Settings{
		ID:                m.ID,
		UserID:            m.UserID,
		LineEnabled:       m.LineEnabled,
		DiscordEnabled:    m.DiscordEnabled,
		DiscordWebhookURL: m.DiscordWebhookURL,
		WebPushEnabled:    m.WebPushEnabled,
		CreatedAt:         m.CreatedAt,
		UpdatedAt:         m.UpdatedAt,
	}
}
