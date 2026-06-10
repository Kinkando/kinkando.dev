// Package snapshot implements the batch cron job that records the final status
// of every active daily/weekly quest at the end of each period.
// It is invoked via POST /api/v1/cron/quest-period-snapshot and writes one row
// per active quest per period into quest_period_results.
//
//   - Daily snapshot  — runs every day just after midnight Bangkok, records yesterday
//   - Weekly snapshot — runs every Monday just after midnight Bangkok, records the week that ended
//
// Both writes are idempotent: ON CONFLICT (quest_id, period_start) DO NOTHING.
package snapshot

// Recommended crontab (Cloudflare Worker, UTC): 0,30 17 * * *
// Fires at 17:00 and 17:30 UTC = 00:00 and 00:30 Bangkok.
// The weekly gate below makes most Monday runs record both daily + weekly results;
// all other days only the daily snapshot is written.

import (
	"context"
	"fmt"
	"time"

	"github.com/kinkando/personal-dashboard/internal/quest"
	"go.uber.org/zap"
)

// ── Repository interface ──────────────────────────────────────────────────────

// QuestRepository is the narrow data-access interface the service depends on.
// *questrepository.Repository satisfies it.
type QuestRepository interface {
	RecordPeriodResults(ctx context.Context, questType string, periodStart time.Time) (*quest.PeriodSnapshotResult, error)
}

// ── Result ────────────────────────────────────────────────────────────────────

// RunResult summarises one cron run for observability.
type RunResult struct {
	Daily  *quest.PeriodSnapshotResult `json:"daily,omitempty"`
	Weekly *quest.PeriodSnapshotResult `json:"weekly,omitempty"`
}

// ── Service ───────────────────────────────────────────────────────────────────

type Service struct {
	quests QuestRepository
	log    *zap.Logger
	now    func() time.Time // injectable clock; defaults to time.Now (Asia/Bangkok via time.Local)
}

func New(quests QuestRepository, log *zap.Logger) *Service {
	return &Service{quests: quests, log: log, now: time.Now}
}

// Run executes the quest-period-snapshot batch job. Safe to call concurrently —
// idempotency is guaranteed by the quest_period_results unique constraint.
func (s *Service) Run(ctx context.Context) (*RunResult, error) {
	now := s.now()
	today := todayFor(now)
	yesterday := today.AddDate(0, 0, -1)

	result := &RunResult{}

	// ── Daily snapshot (every day) ────────────────────────────────────────────
	daily, err := s.quests.RecordPeriodResults(ctx, "daily", yesterday)
	if err != nil {
		return nil, fmt.Errorf("quest period snapshot: daily: %w", err)
	}
	result.Daily = daily

	// ── Weekly snapshot (Mondays only — the week that just ended) ─────────────
	if now.Weekday() == time.Monday {
		weekStart := weekStartFor(yesterday) // yesterday = Sunday → its week's Monday
		weekly, err := s.quests.RecordPeriodResults(ctx, "weekly", weekStart)
		if err != nil {
			return nil, fmt.Errorf("quest period snapshot: weekly: %w", err)
		}
		result.Weekly = weekly
	}

	s.log.Info("quest period snapshot run complete",
		zap.Any("daily", result.Daily),
		zap.Any("weekly", result.Weekly),
	)
	return result, nil
}

// ── Helpers ───────────────────────────────────────────────────────────────────

// todayFor returns midnight UTC for the Bangkok calendar day of t.
func todayFor(t time.Time) time.Time {
	loc, _ := time.LoadLocation("Asia/Bangkok")
	bkk := t.In(loc)
	return time.Date(bkk.Year(), bkk.Month(), bkk.Day(), 0, 0, 0, 0, time.UTC)
}

// weekStartFor returns midnight UTC for the Monday that starts the week
// containing the given day (Bangkok-midnight-UTC). Mirrors quest/service weekStart().
func weekStartFor(day time.Time) time.Time {
	weekday := day.Weekday()
	daysFromMonday := int(weekday-time.Monday+7) % 7
	return day.AddDate(0, 0, -daysFromMonday)
}
