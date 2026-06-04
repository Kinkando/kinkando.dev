// Package reminder implements the batch cron reminder job for quests.
// It is invoked via POST /api/v1/cron/quest-reminders and sends a single
// notification per user that combines any due daily and/or weekly nudges:
//
//   - Daily quest nudge  — ≥ 20:00 Bangkok, if the user has incomplete daily quests
//   - Weekly quest nudge — Sunday ≥ 18:00 Bangkok, if weekly quests are incomplete
//
// Both types are deduped per-period via the generic reminder_log table.
package reminder

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/kinkando/personal-dashboard/internal/notification"
	"github.com/kinkando/personal-dashboard/pkg/helper"
	"go.uber.org/zap"
)

const (
	dailyNudgeHour  = 20
	weeklyNudgeHour = 18

	domainQuestDaily  = "quest_daily"
	domainQuestWeekly = "quest_weekly"
)

// ── Repository interfaces ─────────────────────────────────────────────────────

// QuestRepository is the narrow data-access interface the service depends on.
// *questrepository.Repository satisfies it.
type QuestRepository interface {
	CountIncompleteByUser(ctx context.Context, questType string, periodStart time.Time) (map[uuid.UUID]int, error)
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
	UsersNotified int            `json:"users_notified"`
	ItemsByType   map[string]int `json:"items_by_type"`
}

// ── Service ───────────────────────────────────────────────────────────────────

type Service struct {
	quests  QuestRepository
	remLog  ReminderLog
	noti    Notifier
	log     *zap.Logger
}

func New(quests QuestRepository, remLog ReminderLog, noti Notifier, log *zap.Logger) *Service {
	return &Service{quests: quests, remLog: remLog, noti: noti, log: log}
}

// Run executes the quest-reminder batch job. Safe to call concurrently —
// idempotency is guaranteed by the reminder_log unique constraint.
func (s *Service) Run(ctx context.Context) (*RunResult, error) {
	now := time.Now() // Asia/Bangkok via time.Local set in main.go
	today := helper.Today()
	todayKey := now.Format("2006-01-02")

	result := &RunResult{ItemsByType: make(map[string]int)}

	// Collect per-user nudge lines to combine into one notification.
	type userNudge struct {
		daily  int // incomplete daily count, 0 if not due or already sent
		weekly int // incomplete weekly count, 0 if not due or already sent
	}
	nudges := make(map[uuid.UUID]*userNudge)

	// ── Daily quest nudge ──────────────────────────────────────────────────
	if now.Hour() >= dailyNudgeHour {
		incomplete, err := s.quests.CountIncompleteByUser(ctx, "daily", today)
		if err != nil {
			return nil, fmt.Errorf("quest reminder: count daily incomplete: %w", err)
		}
		for userID, count := range incomplete {
			logged, logErr := s.remLog.Log(ctx, userID, domainQuestDaily, todayKey)
			if logErr != nil {
				s.log.Warn("quest reminder: log daily", zap.String("user_id", userID.String()), zap.Error(logErr))
			}
			if logged {
				if nudges[userID] == nil {
					nudges[userID] = &userNudge{}
				}
				nudges[userID].daily = count
				result.ItemsByType["quest_daily"]++
			}
		}
	}

	// ── Weekly quest nudge (Sundays only) ─────────────────────────────────
	if now.Weekday() == time.Sunday && now.Hour() >= weeklyNudgeHour {
		weekStart := weekStartFor(today)
		weekKey := weekStart.Format("2006-01-02")

		incomplete, err := s.quests.CountIncompleteByUser(ctx, "weekly", weekStart)
		if err != nil {
			return nil, fmt.Errorf("quest reminder: count weekly incomplete: %w", err)
		}
		for userID, count := range incomplete {
			logged, logErr := s.remLog.Log(ctx, userID, domainQuestWeekly, weekKey)
			if logErr != nil {
				s.log.Warn("quest reminder: log weekly", zap.String("user_id", userID.String()), zap.Error(logErr))
			}
			if logged {
				if nudges[userID] == nil {
					nudges[userID] = &userNudge{}
				}
				nudges[userID].weekly = count
				result.ItemsByType["quest_weekly"]++
			}
		}
	}

	// ── Fan out — one notification per user ───────────────────────────────
	for userID, n := range nudges {
		var parts []string
		if n.daily > 0 {
			parts = append(parts, fmt.Sprintf("You have %d daily quest(s) left to complete today.", n.daily))
		}
		if n.weekly > 0 {
			parts = append(parts, fmt.Sprintf("You have %d weekly quest(s) left to finish before the week ends.", n.weekly))
		}
		if len(parts) == 0 {
			continue
		}
		s.noti.Notify(ctx, userID, notification.Message{
			Title: "Quest reminders",
			Body:  strings.Join(parts, " "),
		})
		result.UsersNotified++
	}

	s.log.Info("quest reminder run complete",
		zap.Int("users_notified", result.UsersNotified),
		zap.Any("items_by_type", result.ItemsByType),
	)
	return result, nil
}

// ── Helpers ───────────────────────────────────────────────────────────────────

// weekStartFor returns midnight UTC for the Monday that starts the week
// containing the given day (Bangkok-midnight-UTC). Mirrors quest/service weekStart().
func weekStartFor(today time.Time) time.Time {
	weekday := today.Weekday()
	daysFromMonday := int(weekday-time.Monday+7) % 7
	return today.AddDate(0, 0, -daysFromMonday)
}
