package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kinkando/personal-dashboard/internal/achievement"
	"github.com/kinkando/personal-dashboard/internal/notification"
	"github.com/kinkando/personal-dashboard/internal/quest"
)

// ── Fakes ─────────────────────────────────────────────────────────────────────

type fakeRepo struct {
	unlocked    map[string]time.Time
	counts      map[achievement.Metric]int
	unlockCalls []string // codes passed to Unlock, in order
}

func (f *fakeRepo) ListUnlocked(_ context.Context, _ uuid.UUID) (map[string]time.Time, error) {
	if f.unlocked != nil {
		return f.unlocked, nil
	}
	return map[string]time.Time{}, nil
}

func (f *fakeRepo) Unlock(_ context.Context, _ uuid.UUID, code string, _ time.Time) error {
	f.unlockCalls = append(f.unlockCalls, code)
	if f.unlocked == nil {
		f.unlocked = map[string]time.Time{}
	}
	f.unlocked[code] = time.Now()
	return nil
}

func (f *fakeRepo) Counts(_ context.Context, _ uuid.UUID) (map[achievement.Metric]int, error) {
	if f.counts != nil {
		return f.counts, nil
	}
	return map[achievement.Metric]int{}, nil
}

type fakeQuestStats struct {
	overview *quest.Overview
	streaks  *quest.StreakSummary
}

func (f *fakeQuestStats) GetOverview(_ context.Context, _ uuid.UUID) (*quest.Overview, error) {
	if f.overview != nil {
		return f.overview, nil
	}
	return &quest.Overview{XP: quest.XPSummary{Level: 1, TotalXP: 0}}, nil
}

func (f *fakeQuestStats) GetStreaks(_ context.Context, _ uuid.UUID) (*quest.StreakSummary, error) {
	if f.streaks != nil {
		return f.streaks, nil
	}
	return &quest.StreakSummary{}, nil
}

type fakeNotifier struct {
	calls []notification.Message
}

func (f *fakeNotifier) Notify(_ context.Context, _ uuid.UUID, msg notification.Message) *notification.DeliveryResult {
	f.calls = append(f.calls, msg)
	return &notification.DeliveryResult{Attempted: 1, Delivered: 1}
}

func newSvc(repo *fakeRepo, qs *fakeQuestStats, noti *fakeNotifier) *Service {
	svc := New(repo, qs, noti)
	return svc
}

// ── Tests ─────────────────────────────────────────────────────────────────────

func TestEvaluate_BadgeUnlocked_WhenMetricMeetsThreshold(t *testing.T) {
	// level_5 threshold = 5; set level to 5 → should unlock.
	repo := &fakeRepo{}
	qs := &fakeQuestStats{overview: &quest.Overview{XP: quest.XPSummary{Level: 5}}}
	svc := newSvc(repo, qs, &fakeNotifier{})

	summary, err := svc.Evaluate(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Find the level_5 achievement in the response.
	var found *achievement.Achievement
	for i := range summary.Achievements {
		if summary.Achievements[i].Code == "level_5" {
			found = &summary.Achievements[i]
			break
		}
	}
	if found == nil {
		t.Fatal("level_5 not found in achievements")
	}
	if !found.Unlocked {
		t.Error("level_5 should be unlocked at level 5")
	}
	if found.UnlockedAt == nil {
		t.Error("level_5 UnlockedAt should not be nil")
	}

	// Verify Unlock was called exactly once for level_5.
	if len(repo.unlockCalls) != 1 || repo.unlockCalls[0] != "level_5" {
		t.Errorf("unlockCalls = %v, want [level_5]", repo.unlockCalls)
	}

	// NewlyUnlocked should include level_5.
	found2 := false
	for _, code := range summary.NewlyUnlocked {
		if code == "level_5" {
			found2 = true
		}
	}
	if !found2 {
		t.Error("level_5 should be in NewlyUnlocked")
	}
}

func TestEvaluate_ProgressCappedAtThreshold(t *testing.T) {
	// level_5 threshold = 5; set level to 50 — progress should cap at 5.
	repo := &fakeRepo{}
	qs := &fakeQuestStats{overview: &quest.Overview{XP: quest.XPSummary{Level: 50}}}
	svc := newSvc(repo, qs, &fakeNotifier{})

	summary, err := svc.Evaluate(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, a := range summary.Achievements {
		if a.Progress > a.Target {
			t.Errorf("badge %s: Progress %d > Target %d (should be capped)", a.Code, a.Progress, a.Target)
		}
	}
}

func TestEvaluate_AlreadyUnlockedBadgeNotReLocked(t *testing.T) {
	// level_5 is already unlocked; metric still at 5. Unlock must not be called again.
	alreadyAt := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	repo := &fakeRepo{
		unlocked: map[string]time.Time{"level_5": alreadyAt},
	}
	qs := &fakeQuestStats{overview: &quest.Overview{XP: quest.XPSummary{Level: 5}}}
	svc := newSvc(repo, qs, &fakeNotifier{})

	summary, err := svc.Evaluate(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(repo.unlockCalls) != 0 {
		t.Errorf("Unlock called %d times for already-unlocked badge, want 0", len(repo.unlockCalls))
	}
	if len(summary.NewlyUnlocked) != 0 {
		t.Errorf("NewlyUnlocked = %v, want empty (badge already unlocked)", summary.NewlyUnlocked)
	}
}

func TestEvaluate_TotalsAreCorrect(t *testing.T) {
	repo := &fakeRepo{}
	qs := &fakeQuestStats{} // level 1, 0 XP → no badges unlocked
	svc := newSvc(repo, qs, &fakeNotifier{})

	summary, err := svc.Evaluate(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if summary.Total != len(achievement.Catalog) {
		t.Errorf("Total = %d, want %d", summary.Total, len(achievement.Catalog))
	}
	if summary.UnlockedCount != 0 {
		t.Errorf("UnlockedCount = %d, want 0 (no thresholds met)", summary.UnlockedCount)
	}
}

func TestEvaluateAndNotify_SendsNotificationForNewlyUnlocked(t *testing.T) {
	// Trigger level_5 unlock (threshold = 5).
	repo := &fakeRepo{}
	qs := &fakeQuestStats{overview: &quest.Overview{XP: quest.XPSummary{Level: 5}}}
	noti := &fakeNotifier{}
	svc := newSvc(repo, qs, noti)

	if err := svc.EvaluateAndNotify(context.Background(), uuid.New()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// At least one notification for "level_5" must have been sent.
	found := false
	for _, msg := range noti.calls {
		if msg.Title == "Achievement unlocked" {
			found = true
		}
	}
	if !found {
		t.Error("expected at least one 'Achievement unlocked' notification")
	}
}

func TestEvaluateAndNotify_NilNotifierIsSafe(t *testing.T) {
	// Notifier is nil — must not panic even when badges unlock.
	repo := &fakeRepo{}
	qs := &fakeQuestStats{overview: &quest.Overview{XP: quest.XPSummary{Level: 5}}}
	svc := New(repo, qs, nil) // nil notifier

	if err := svc.EvaluateAndNotify(context.Background(), uuid.New()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
