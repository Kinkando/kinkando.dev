package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/kinkando/personal-dashboard/internal/achievement"
	"github.com/kinkando/personal-dashboard/internal/notification"
	"github.com/kinkando/personal-dashboard/internal/quest"
)

// Repository is the persistence + count data the service depends on.
type Repository interface {
	ListUnlocked(ctx context.Context, userID uuid.UUID) (map[string]time.Time, error)
	Unlock(ctx context.Context, userID uuid.UUID, code string, at time.Time) error
	Counts(ctx context.Context, userID uuid.UUID) (map[achievement.Metric]int, error)
}

// QuestStats supplies the gamification metrics owned by the quest module.
// *questservice.Service satisfies it.
type QuestStats interface {
	GetOverview(ctx context.Context, userID uuid.UUID) (*quest.Overview, error)
	GetStreaks(ctx context.Context, userID uuid.UUID) (*quest.StreakSummary, error)
}

// Notifier delivers a push when a badge unlocks. *notificationservice.Service satisfies it.
type Notifier interface {
	Notify(ctx context.Context, userID uuid.UUID, msg notification.Message) *notification.DeliveryResult
}

type Service struct {
	repo     Repository
	quest    QuestStats
	notifier Notifier
	now      func() time.Time // injectable clock; defaults to time.Now
}

func New(repo Repository, quest QuestStats, notifier Notifier) *Service {
	return &Service{repo: repo, quest: quest, notifier: notifier, now: time.Now}
}

// Evaluate computes progress for every badge, persists any newly-met unlocks,
// and returns the full summary. Idempotent — already-unlocked badges are left
// untouched. Does NOT send notifications (see EvaluateAndNotify).
func (s *Service) Evaluate(ctx context.Context, userID uuid.UUID) (*achievement.Summary, error) {
	overview, err := s.quest.GetOverview(ctx, userID)
	if err != nil {
		return nil, err
	}
	streaks, err := s.quest.GetStreaks(ctx, userID)
	if err != nil {
		return nil, err
	}
	counts, err := s.repo.Counts(ctx, userID)
	if err != nil {
		return nil, err
	}

	metrics := map[achievement.Metric]int{
		achievement.MetricLevel:         overview.XP.Level,
		achievement.MetricTotalXP:       overview.XP.TotalXP,
		achievement.MetricCurrentStreak: streaks.CurrentStreak,
		achievement.MetricLongestStreak: streaks.LongestStreak,
		achievement.MetricPerfectDays:   streaks.PerfectDays,
	}
	for k, v := range counts {
		metrics[k] = v
	}

	unlocked, err := s.repo.ListUnlocked(ctx, userID)
	if err != nil {
		return nil, err
	}

	now := s.now()
	achievements := make([]achievement.Achievement, 0, len(achievement.Catalog))
	var newly []string
	unlockedCount := 0

	for _, def := range achievement.Catalog {
		val := metrics[def.Metric]
		at, was := unlocked[def.Code]
		isUnlocked := was

		if !was && val >= def.Threshold {
			if err := s.repo.Unlock(ctx, userID, def.Code, now); err != nil {
				return nil, err
			}
			at = now
			isUnlocked = true
			newly = append(newly, def.Code)
		}

		var unlockedAt *time.Time
		if isUnlocked {
			unlockedCount++
			t := at
			unlockedAt = &t
		}

		progress := val
		if progress > def.Threshold {
			progress = def.Threshold
		}

		achievements = append(achievements, achievement.Achievement{
			Code:        def.Code,
			Title:       def.Title,
			Description: def.Description,
			Icon:        def.Icon,
			Category:    def.Category,
			Unlocked:    isUnlocked,
			UnlockedAt:  unlockedAt,
			Progress:    progress,
			Target:      def.Threshold,
		})
	}

	return &achievement.Summary{
		Achievements:  achievements,
		UnlockedCount: unlockedCount,
		Total:         len(achievement.Catalog),
		NewlyUnlocked: newly,
	}, nil
}

// EvaluateAndNotify evaluates then pushes a notification for each newly-unlocked
// badge. Used by event subscribers. Best-effort — notification failures are not
// surfaced. Safe to call repeatedly (unlock persistence is idempotent).
func (s *Service) EvaluateAndNotify(ctx context.Context, userID uuid.UUID) error {
	summary, err := s.Evaluate(ctx, userID)
	if err != nil {
		return err
	}
	if s.notifier == nil {
		return nil
	}
	for _, code := range summary.NewlyUnlocked {
		def, ok := achievement.FindDef(code)
		if !ok {
			continue
		}
		s.notifier.Notify(ctx, userID, notification.Message{
			Title: "Achievement unlocked",
			Body:  def.Icon + " " + def.Title + " — " + def.Description,
		})
	}
	return nil
}
