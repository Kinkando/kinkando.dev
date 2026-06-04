// Package reminder implements the batch cron reminder job for health tracking.
// It is invoked via POST /api/v1/cron/weight-nudge and sends a morning nudge
// to users who have logged weight before but haven't logged yet today.
//
// Timing gate: ≥ 08:00 Asia/Bangkok.
// Audience:    past weight-loggers only (users who never track weight are skipped).
// Dedup:       once per day per user via the generic reminder_log table.
package reminder

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/kinkando/personal-dashboard/internal/notification"
	"github.com/kinkando/personal-dashboard/pkg/helper"
	"go.uber.org/zap"
)

// Recommended crontab (Cloudflare Worker, UTC): 0,30 1-3 * * *
// Fires at :00 and :30 of 01:00–03:00 UTC = 08:00–10:00 BKK.
// Narrower window than the other jobs because weight logging is a morning habit;
// nudging later in the day would be disruptive.
// The weightNudgeHour gate ensures no notification fires before 08:00 BKK even
// if the cron is triggered early.

const (
	weightNudgeHour = 8
	domainWeight    = "weight"
)

// ── Repository interfaces ─────────────────────────────────────────────────────

// HealthRepository is the narrow data-access interface the service depends on.
// *healthrepository.Repository satisfies it.
type HealthRepository interface {
	UsersMissingWeightToday(ctx context.Context, today time.Time) ([]uuid.UUID, error)
}

// ReminderLog persists dedup entries. *reminderlog.Repository satisfies it.
type ReminderLog interface {
	Log(ctx context.Context, userID uuid.UUID, domain, key string) (bool, error)
}

// Notifier fans out a notification. *notificationSvc.Service satisfies it.
type Notifier interface {
	Notify(ctx context.Context, userID uuid.UUID, msg notification.Message) *notification.DeliveryResult
}

// ── Result ────────────────────────────────────────────────────────────────────

// RunResult summarises one cron run for observability.
type RunResult struct {
	UsersNotified int `json:"users_notified"`
}

// ── Service ───────────────────────────────────────────────────────────────────

type Service struct {
	health HealthRepository
	remLog ReminderLog
	noti   Notifier
	log    *zap.Logger
}

func New(health HealthRepository, remLog ReminderLog, noti Notifier, log *zap.Logger) *Service {
	return &Service{health: health, remLog: remLog, noti: noti, log: log}
}

// Run executes the weight-log nudge batch job.
// Safe to call concurrently — idempotency is guaranteed by the reminder_log
// unique constraint.
func (s *Service) Run(ctx context.Context) (*RunResult, error) {
	now := time.Now() // Asia/Bangkok via time.Local set in main.go
	result := &RunResult{}

	if now.Hour() < weightNudgeHour {
		s.log.Debug("weight nudge: before hour gate, skipping", zap.Int("hour", now.Hour()))
		return result, nil
	}

	today := helper.Today()
	todayKey := now.Format(time.DateOnly)

	candidates, err := s.health.UsersMissingWeightToday(ctx, today)
	if err != nil {
		return nil, fmt.Errorf("weight nudge: scan candidates: %w", err)
	}

	for _, userID := range candidates {
		logged, logErr := s.remLog.Log(ctx, userID, domainWeight, todayKey)
		if logErr != nil {
			s.log.Warn("weight nudge: log reminder", zap.String("user_id", userID.String()), zap.Error(logErr))
		}
		if !logged {
			continue // already sent today
		}
		s.noti.Notify(ctx, userID, notification.Message{
			Title: "Weight log",
			Body:  "Don't forget to log your weight today.",
		})
		result.UsersNotified++
	}

	s.log.Info("weight nudge run complete", zap.Int("users_notified", result.UsersNotified))
	return result, nil
}
