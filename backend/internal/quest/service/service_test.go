package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kinkando/personal-dashboard/internal/quest"
	"github.com/kinkando/personal-dashboard/pkg/event"
)

// ── fakes ─────────────────────────────────────────────────────────────────────

// fakeRepo satisfies Repository via function fields. Unset fields return zero
// values (nil, 0, no error) so tests only configure what they need.
type fakeRepo struct {
	getQuestFn         func(ctx context.Context, id, userID uuid.UUID) (*quest.Quest, error)
	incrementFn        func(ctx context.Context, userID, questID uuid.UUID, periodStart time.Time, source string) error
	decrementFn        func(ctx context.Context, userID, questID uuid.UUID, periodStart time.Time) error
	getQuestStatusFn   func(ctx context.Context, userID uuid.UUID, qType quest.QuestType, today time.Time) ([]*quest.QuestStatus, error)
	totalXPFn          func(ctx context.Context, userID uuid.UUID) (int, error)
	listDailyResultsFn func(ctx context.Context, userID uuid.UUID, from, to time.Time) ([]quest.PeriodResultRow, error)
}

func (f *fakeRepo) CreateQuest(ctx context.Context, userID uuid.UUID, in quest.CreateQuestInput) (*quest.Quest, error) {
	return nil, nil
}
func (f *fakeRepo) ListQuests(ctx context.Context, userID uuid.UUID, questType string) ([]*quest.Quest, error) {
	return nil, nil
}
func (f *fakeRepo) GetQuest(ctx context.Context, id, userID uuid.UUID) (*quest.Quest, error) {
	if f.getQuestFn != nil {
		return f.getQuestFn(ctx, id, userID)
	}
	return nil, nil
}
func (f *fakeRepo) UpdateQuest(ctx context.Context, id, userID uuid.UUID, in quest.UpdateQuestInput) (*quest.Quest, error) {
	return nil, nil
}
func (f *fakeRepo) DeleteQuest(ctx context.Context, id, userID uuid.UUID) error { return nil }
func (f *fakeRepo) SetActive(ctx context.Context, id, userID uuid.UUID, active bool) (*quest.Quest, error) {
	return nil, nil
}
func (f *fakeRepo) GetQuestStatus(ctx context.Context, userID uuid.UUID, qType quest.QuestType, today time.Time) ([]*quest.QuestStatus, error) {
	if f.getQuestStatusFn != nil {
		return f.getQuestStatusFn(ctx, userID, qType, today)
	}
	return nil, nil
}
func (f *fakeRepo) TotalXP(ctx context.Context, userID uuid.UUID) (int, error) {
	if f.totalXPFn != nil {
		return f.totalXPFn(ctx, userID)
	}
	return 0, nil
}
func (f *fakeRepo) ListDailyResults(ctx context.Context, userID uuid.UUID, from, to time.Time) ([]quest.PeriodResultRow, error) {
	if f.listDailyResultsFn != nil {
		return f.listDailyResultsFn(ctx, userID, from, to)
	}
	return nil, nil
}
func (f *fakeRepo) Increment(ctx context.Context, userID, questID uuid.UUID, periodStart time.Time, source string) error {
	if f.incrementFn != nil {
		return f.incrementFn(ctx, userID, questID, periodStart, source)
	}
	return nil
}
func (f *fakeRepo) Decrement(ctx context.Context, userID, questID uuid.UUID, periodStart time.Time) error {
	if f.decrementFn != nil {
		return f.decrementFn(ctx, userID, questID, periodStart)
	}
	return nil
}
func (f *fakeRepo) ProgressBySource(ctx context.Context, userID uuid.UUID, sourceType string, today, weekStart time.Time) error {
	return nil
}
func (f *fakeRepo) ListXPEvents(ctx context.Context, userID uuid.UUID, limit int) ([]*quest.XPEvent, error) {
	return nil, nil
}

// fakeEvents records every Publish call.
type fakeEvents struct {
	published []event.Event
}

func (f *fakeEvents) Publish(_ context.Context, e event.Event) {
	f.published = append(f.published, e)
}

// newSvc builds a Service with the given repo + clock and no event bus.
func newSvc(repo Repository, nowDate string) (*Service, *fakeEvents) {
	ev := &fakeEvents{}
	svc := &Service{
		repo:   repo,
		events: ev,
		now:    func() time.Time { return mustParseDate(nowDate) },
	}
	return svc, ev
}

// ── IncrementQuest ────────────────────────────────────────────────────────────

func TestIncrementQuest_RejectsNonManual(t *testing.T) {
	auto := &quest.Quest{SourceType: quest.SourceTypeMedicine}
	repo := &fakeRepo{getQuestFn: func(_ context.Context, _, _ uuid.UUID) (*quest.Quest, error) {
		return auto, nil
	}}
	svc, _ := newSvc(repo, "2026-06-10")

	err := svc.IncrementQuest(context.Background(), uuid.New(), uuid.New())
	if err == nil {
		t.Fatal("want error for non-manual quest, got nil")
	}
	if !errors.Is(err, err) { // always true; just verify it has a message
		t.Fatal("error should not be nil")
	}
	const wantMsg = "this quest is auto-managed and cannot be manually incremented"
	if err.Error() != wantMsg {
		t.Errorf("error = %q, want %q", err.Error(), wantMsg)
	}
}

func TestIncrementQuest_ManualDailySuccess(t *testing.T) {
	userID := uuid.New()
	questID := uuid.New()
	manual := &quest.Quest{ID: questID, UserID: userID, Type: quest.QuestTypeDaily, SourceType: quest.SourceTypeManual}

	var gotPeriod time.Time
	var gotSource string
	var incrementCalled bool

	repo := &fakeRepo{
		getQuestFn: func(_ context.Context, _, _ uuid.UUID) (*quest.Quest, error) {
			return manual, nil
		},
		incrementFn: func(_ context.Context, _, _ uuid.UUID, periodStart time.Time, source string) error {
			incrementCalled = true
			gotPeriod = periodStart
			gotSource = source
			return nil
		},
	}
	svc, ev := newSvc(repo, "2026-06-10")

	if err := svc.IncrementQuest(context.Background(), userID, questID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !incrementCalled {
		t.Error("repo.Increment was not called")
	}
	if gotPeriod.Format(time.DateOnly) != "2026-06-10" {
		t.Errorf("periodStart = %s, want 2026-06-10", gotPeriod.Format(time.DateOnly))
	}
	if gotSource != "daily" {
		t.Errorf("source = %q, want \"daily\"", gotSource)
	}
	if len(ev.published) != 1 || ev.published[0].Type != event.QuestCompleted {
		t.Errorf("expected one QuestCompleted event, got %+v", ev.published)
	}
}

func TestIncrementQuest_ManualWeeklyUsesWeekStart(t *testing.T) {
	manual := &quest.Quest{Type: quest.QuestTypeWeekly, SourceType: quest.SourceTypeManual}

	var gotPeriod time.Time
	repo := &fakeRepo{
		getQuestFn: func(_ context.Context, _, _ uuid.UUID) (*quest.Quest, error) { return manual, nil },
		incrementFn: func(_ context.Context, _, _ uuid.UUID, p time.Time, _ string) error {
			gotPeriod = p
			return nil
		},
	}
	// Wednesday — week start should be Monday 2026-06-08
	svc, _ := newSvc(repo, "2026-06-10")
	if err := svc.IncrementQuest(context.Background(), uuid.New(), uuid.New()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotPeriod.Format(time.DateOnly) != "2026-06-08" {
		t.Errorf("weekly periodStart = %s, want 2026-06-08 (Monday)", gotPeriod.Format(time.DateOnly))
	}
}

// ── DecrementQuest ────────────────────────────────────────────────────────────

func TestDecrementQuest_RejectsNonManual(t *testing.T) {
	repo := &fakeRepo{getQuestFn: func(_ context.Context, _, _ uuid.UUID) (*quest.Quest, error) {
		return &quest.Quest{SourceType: quest.SourceTypeWorkout}, nil
	}}
	svc, _ := newSvc(repo, "2026-06-10")

	err := svc.DecrementQuest(context.Background(), uuid.New(), uuid.New())
	if err == nil {
		t.Fatal("want error for non-manual quest, got nil")
	}
	const wantMsg = "this quest is auto-managed and cannot be manually decremented"
	if err.Error() != wantMsg {
		t.Errorf("error = %q, want %q", err.Error(), wantMsg)
	}
}

func TestDecrementQuest_ManualSuccess(t *testing.T) {
	decrementCalled := false
	repo := &fakeRepo{
		getQuestFn: func(_ context.Context, _, _ uuid.UUID) (*quest.Quest, error) {
			return &quest.Quest{Type: quest.QuestTypeDaily, SourceType: quest.SourceTypeManual}, nil
		},
		decrementFn: func(_ context.Context, _, _ uuid.UUID, _ time.Time) error {
			decrementCalled = true
			return nil
		},
	}
	svc, _ := newSvc(repo, "2026-06-10")
	if err := svc.DecrementQuest(context.Background(), uuid.New(), uuid.New()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !decrementCalled {
		t.Error("repo.Decrement was not called")
	}
}

// ── GetOverview ───────────────────────────────────────────────────────────────

func TestGetOverview_CountsDoneQuests(t *testing.T) {
	daily := []*quest.QuestStatus{
		{Completed: true},
		{Completed: false},
		{Completed: true},
	}
	weekly := []*quest.QuestStatus{
		{Completed: true},
		{Completed: false},
	}

	repo := &fakeRepo{
		getQuestStatusFn: func(_ context.Context, _ uuid.UUID, qType quest.QuestType, _ time.Time) ([]*quest.QuestStatus, error) {
			if qType == quest.QuestTypeDaily {
				return daily, nil
			}
			return weekly, nil
		},
		totalXPFn: func(_ context.Context, _ uuid.UUID) (int, error) { return 250, nil },
	}
	svc, _ := newSvc(repo, "2026-06-10")

	ov, err := svc.GetOverview(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ov.DailyDone != 2 {
		t.Errorf("DailyDone = %d, want 2", ov.DailyDone)
	}
	if ov.DailyTotal != 3 {
		t.Errorf("DailyTotal = %d, want 3", ov.DailyTotal)
	}
	if ov.WeeklyDone != 1 {
		t.Errorf("WeeklyDone = %d, want 1", ov.WeeklyDone)
	}
	if ov.WeeklyTotal != 2 {
		t.Errorf("WeeklyTotal = %d, want 2", ov.WeeklyTotal)
	}
}

func TestGetOverview_NilSlicesCoerced(t *testing.T) {
	// GetQuestStatus returns nil — service must convert to empty (not nil) slices.
	repo := &fakeRepo{
		getQuestStatusFn: func(_ context.Context, _ uuid.UUID, _ quest.QuestType, _ time.Time) ([]*quest.QuestStatus, error) {
			return nil, nil
		},
		totalXPFn: func(_ context.Context, _ uuid.UUID) (int, error) { return 0, nil },
	}
	svc, _ := newSvc(repo, "2026-06-10")

	ov, err := svc.GetOverview(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ov.Daily == nil {
		t.Error("Overview.Daily is nil, want empty slice")
	}
	if ov.Weekly == nil {
		t.Error("Overview.Weekly is nil, want empty slice")
	}
}

func TestGetOverview_XPSummaryEmbedded(t *testing.T) {
	repo := &fakeRepo{
		getQuestStatusFn: func(_ context.Context, _ uuid.UUID, _ quest.QuestType, _ time.Time) ([]*quest.QuestStatus, error) {
			return nil, nil
		},
		totalXPFn: func(_ context.Context, _ uuid.UUID) (int, error) { return 150, nil },
	}
	svc, _ := newSvc(repo, "2026-06-10")

	ov, err := svc.GetOverview(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ov.XP.Level != 2 {
		t.Errorf("XP.Level = %d, want 2 (150 XP / 100 per level + 1)", ov.XP.Level)
	}
	if ov.XP.XPIntoLevel != 50 {
		t.Errorf("XP.XPIntoLevel = %d, want 50", ov.XP.XPIntoLevel)
	}
}

// ── GetStreaks ─────────────────────────────────────────────────────────────────

// row constructs a PeriodResultRow quickly.
func row(dateStr string, completed bool) quest.PeriodResultRow {
	return quest.PeriodResultRow{PeriodStart: mustParseDate(dateStr), Completed: completed}
}

func TestGetStreaks_CurrentAndLongest(t *testing.T) {
	// Fixed today: Wednesday 2026-06-03.
	// Snapshot rows (two quests per day simulated by two rows per date):
	//   2026-05-30: both completed → perfect
	//   2026-05-31: both completed → perfect
	//   2026-06-01: 1 completed, 1 not → imperfect (breaks streak)
	//   2026-06-02: both completed → perfect
	// Live today (2026-06-03): 0 quests returned → today not in heatmap → current streak from yesterday.
	// 2026-06-02 is perfect → current streak = 1.
	// Longest: 05-30,05-31 = 2 (06-01 breaks), then 06-02 = 1 → longest = 2.

	snapRows := []quest.PeriodResultRow{
		row("2026-05-30", true), row("2026-05-30", true),
		row("2026-05-31", true), row("2026-05-31", true),
		row("2026-06-01", true), row("2026-06-01", false),
		row("2026-06-02", true), row("2026-06-02", true),
	}

	repo := &fakeRepo{
		listDailyResultsFn: func(_ context.Context, _ uuid.UUID, _, _ time.Time) ([]quest.PeriodResultRow, error) {
			return snapRows, nil
		},
		getQuestStatusFn: func(_ context.Context, _ uuid.UUID, _ quest.QuestType, _ time.Time) ([]*quest.QuestStatus, error) {
			return nil, nil // no live quests → today absent from heatmap
		},
	}
	svc, _ := newSvc(repo, "2026-06-03")

	got, err := svc.GetStreaks(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.CurrentStreak != 1 {
		t.Errorf("CurrentStreak = %d, want 1", got.CurrentStreak)
	}
	if got.LongestStreak != 2 {
		t.Errorf("LongestStreak = %d, want 2", got.LongestStreak)
	}
	if got.PerfectDays != 3 { // 05-30, 05-31, 06-02
		t.Errorf("PerfectDays = %d, want 3", got.PerfectDays)
	}
}

func TestGetStreaks_TodayPerfectExtendsCurrentStreak(t *testing.T) {
	// Same history; today (2026-06-03) is perfect via live status → current streak = 2.
	snapRows := []quest.PeriodResultRow{
		row("2026-06-01", true), row("2026-06-01", false), // imperfect
		row("2026-06-02", true), row("2026-06-02", true),  // perfect
	}

	repo := &fakeRepo{
		listDailyResultsFn: func(_ context.Context, _ uuid.UUID, _, _ time.Time) ([]quest.PeriodResultRow, error) {
			return snapRows, nil
		},
		getQuestStatusFn: func(_ context.Context, _ uuid.UUID, _ quest.QuestType, _ time.Time) ([]*quest.QuestStatus, error) {
			return []*quest.QuestStatus{
				{Completed: true},
				{Completed: true},
			}, nil
		},
	}
	svc, _ := newSvc(repo, "2026-06-03")

	got, err := svc.GetStreaks(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 06-02 perfect + 06-03 perfect (today, live) → current = 2
	if got.CurrentStreak != 2 {
		t.Errorf("CurrentStreak = %d, want 2", got.CurrentStreak)
	}
}

func TestGetStreaks_NoQuests_EmptyHeatmap(t *testing.T) {
	repo := &fakeRepo{
		listDailyResultsFn: func(_ context.Context, _ uuid.UUID, _, _ time.Time) ([]quest.PeriodResultRow, error) {
			return nil, nil
		},
		getQuestStatusFn: func(_ context.Context, _ uuid.UUID, _ quest.QuestType, _ time.Time) ([]*quest.QuestStatus, error) {
			return nil, nil
		},
	}
	svc, _ := newSvc(repo, "2026-06-10")

	got, err := svc.GetStreaks(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.CurrentStreak != 0 || got.LongestStreak != 0 || got.PerfectDays != 0 {
		t.Errorf("expected all-zero streaks for empty history, got %+v", got)
	}
	if len(got.Days) != 0 {
		t.Errorf("expected empty heatmap, got %d days", len(got.Days))
	}
}
