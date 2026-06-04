package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/kinkando/personal-dashboard/internal/fcm"
	"github.com/kinkando/personal-dashboard/internal/line"
	"github.com/kinkando/personal-dashboard/internal/notification"
	"github.com/kinkando/personal-dashboard/internal/user"
	"go.uber.org/zap"
)

// Repository is the data-access interface the service depends on.
type Repository interface {
	GetSettings(ctx context.Context, userID uuid.UUID) (*notification.Settings, error)
	UpsertSettings(ctx context.Context, userID uuid.UUID, in notification.UpsertSettingsInput) (*notification.Settings, error)
	AddToken(ctx context.Context, userID uuid.UUID, token string) error
	DeleteToken(ctx context.Context, token string) error
	ListTokens(ctx context.Context, userID uuid.UUID) ([]string, error)
}

// LinePusher can send a push message to a LINE user.
type LinePusher interface {
	Push(ctx context.Context, to string, messages []line.ReplyMessage) error
}

// DiscordSender can post a message to a Discord incoming webhook.
type DiscordSender interface {
	PostWebhook(ctx context.Context, webhookURL, content string) error
}

// FCMSender can send a web-push notification via FCM.
type FCMSender interface {
	Send(ctx context.Context, token, title, body string) error
}

// UserLookup resolves an internal user ID to the full user row (including line_id).
type UserLookup interface {
	GetByID(ctx context.Context, id uuid.UUID) (*user.User, error)
}

// Service implements the notification business logic.
type Service struct {
	repo    Repository
	line    LinePusher
	discord DiscordSender
	fcm     FCMSender
	users   UserLookup
	log     *zap.Logger
}

// New constructs a Service.  Any sender may be nil; the corresponding channel
// will simply be skipped during fan-out.
func New(
	repo Repository,
	linePusher LinePusher,
	discord DiscordSender,
	fcmSender FCMSender,
	users UserLookup,
	log *zap.Logger,
) *Service {
	return &Service{
		repo:    repo,
		line:    linePusher,
		discord: discord,
		fcm:     fcmSender,
		users:   users,
		log:     log,
	}
}

// ── Settings ──────────────────────────────────────────────────────────────────

// GetSettings returns the user's notification settings, or an all-disabled
// default when no row exists yet.
func (s *Service) GetSettings(ctx context.Context, userID uuid.UUID) (*notification.Settings, error) {
	settings, err := s.repo.GetSettings(ctx, userID)
	if err != nil {
		return nil, err
	}
	if settings == nil {
		return notification.DefaultSettings(), nil
	}
	return settings, nil
}

// UpsertSettings saves the user's notification preferences.
func (s *Service) UpsertSettings(ctx context.Context, userID uuid.UUID, in notification.UpsertSettingsInput) (*notification.Settings, error) {
	return s.repo.UpsertSettings(ctx, userID, in)
}

// ── Tokens ────────────────────────────────────────────────────────────────────

// RegisterToken stores a new FCM device token for the user.
func (s *Service) RegisterToken(ctx context.Context, userID uuid.UUID, token string) error {
	return s.repo.AddToken(ctx, userID, token)
}

// RemoveToken deletes an FCM device token.
func (s *Service) RemoveToken(ctx context.Context, token string) error {
	return s.repo.DeleteToken(ctx, token)
}

// ── Notification fan-out ──────────────────────────────────────────────────────

// Notify delivers msg to every enabled channel for userID and returns a
// DeliveryResult so callers can surface real delivery feedback.
// Per-channel errors are logged and counted, but never fatal — fan-out is best-effort.
func (s *Service) Notify(ctx context.Context, userID uuid.UUID, msg notification.Message) *notification.DeliveryResult {
	res := &notification.DeliveryResult{}

	settings, err := s.repo.GetSettings(ctx, userID)
	if err != nil {
		s.log.Warn("notify: get settings", zap.String("user_id", userID.String()), zap.Error(err))
		return res
	}
	if settings == nil {
		return res // user has never configured notifications
	}

	// LINE
	if settings.LineEnabled && s.line != nil {
		u, err := s.users.GetByID(ctx, userID)
		if err != nil {
			s.log.Warn("notify: get user for line", zap.Error(err))
		} else if u != nil && u.LineID != nil {
			res.Attempted++
			text := fmt.Sprintf("%s\n%s", msg.Title, msg.Body)
			if err := s.line.Push(ctx, *u.LineID, []line.ReplyMessage{line.TextMessage(text)}); err != nil {
				s.log.Warn("notify: line push", zap.Error(err))
				res.Errors = append(res.Errors, "LINE: "+err.Error())
			} else {
				res.Delivered++
			}
		}
	}

	// Discord
	if settings.DiscordEnabled && settings.DiscordWebhookURL != nil && s.discord != nil {
		res.Attempted++
		content := fmt.Sprintf("**%s**\n%s", msg.Title, msg.Body)
		if err := s.discord.PostWebhook(ctx, *settings.DiscordWebhookURL, content); err != nil {
			s.log.Warn("notify: discord webhook", zap.Error(err))
			res.Errors = append(res.Errors, "Discord: "+err.Error())
		} else {
			res.Delivered++
		}
	}

	// Web Push (FCM)
	if settings.WebPushEnabled && s.fcm != nil {
		tokens, err := s.repo.ListTokens(ctx, userID)
		if err != nil {
			s.log.Warn("notify: list fcm tokens", zap.Error(err))
		}
		for _, tok := range tokens {
			res.Attempted++
			if err := s.fcm.Send(ctx, tok, msg.Title, msg.Body); err != nil {
				if errors.Is(err, fcm.ErrTokenInvalid) {
					// Prune the stale token; log but don't abort.
					if delErr := s.repo.DeleteToken(ctx, tok); delErr != nil {
						s.log.Warn("notify: delete stale fcm token", zap.Error(delErr))
					}
					res.Errors = append(res.Errors, "Web Push: token invalid, removed")
				} else {
					s.log.Warn("notify: fcm send", zap.Error(err))
					res.Errors = append(res.Errors, "Web Push: "+err.Error())
				}
			} else {
				res.Delivered++
			}
		}
	}

	return res
}

// SendTest fans out a test notification to all enabled channels for userID
// and returns the delivery result so the caller can show real feedback.
func (s *Service) SendTest(ctx context.Context, userID uuid.UUID) *notification.DeliveryResult {
	return s.Notify(ctx, userID, notification.Message{
		Title: "Test notification",
		Body:  "Your notifications are working correctly.",
	})
}
