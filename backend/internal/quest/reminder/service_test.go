package reminder

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kinkando/personal-dashboard/internal/notification"
	"github.com/kinkando/personal-dashboard/internal/reminderlog"
	"go.uber.org/zap"
)

// ── Helpers ───────────────────────────────────────────────────────────────────

var bangkokLoc *time.Location

func init() {
	loc, err := time.LoadLocation("Asia/Bangkok")
	if err != nil {
		panic(err)
	}
	bangkokLoc = loc
}

// bkkAt returns a time.Time in Asia/Bangkok at the given components.
func bkkAt(year, month, day, hour, min int) time.Time {
	return time.Date(year, time.Month(month), day, hour, min, 0, 0, bangkokLoc)
}

// ── Fakes ─────────────────────────────────────────────────────────────────────

type fakeQuestRepo struct {
	incompleteByUser map[uuid.UUID]int // returned for any call
	calls            []string          // records questType arguments
}

func (f *fakeQuestRepo) CountIncompleteByUser(_ context.Context, questType string, _ time.Time) (map[uuid.UUID]int, error) {
	f.calls = append(f.calls, questType)
	if f.incompleteByUser != nil {
		return f.incompleteByUser, nil
	}
	return map[uuid.UUID]int{}, nil
}

type fakeRemLog struct {
	// alwaysNew controls whether Log always returns (true, nil).
	// Set to false to simulate "already sent this period".
	alwaysNew bool
	logged    []string // records "domain:key" for every call
}

func newRemLog(alwaysNew bool) *fakeRemLog { return &fakeRemLog{alwaysNew: alwaysNew} }

func (f *fakeRemLog) Log(_ context.Context, _ uuid.UUID, domain reminderlog.Domain, key string) (bool, error) {
	f.logged = append(f.logged, string(domain)+":"+key)
	return f.alwaysNew, nil
}

type fakeNotifier struct {
	calls []notification.Message
}

func (f *fakeNotifier) Notify(_ context.Context, _ uuid.UUID, msg notification.Message) *notification.DeliveryResult {
	f.calls = append(f.calls, msg)
	return &notification.DeliveryResult{Attempted: 1, Delivered: 1}
}

// newSvc wires up a Service with a fixed clock.
func newSvc(nowFn func() time.Time, questRepo *fakeQuestRepo, remLog *fakeRemLog, noti *fakeNotifier) *Service {
	return &Service{
		quests: questRepo,
		remLog: remLog,
		noti:   noti,
		log:    noopLogger(),
		now:    nowFn,
	}
}

// noopLogger returns a no-op zap logger so test output stays clean.
func noopLogger() *zap.Logger {
	return zap.NewNop()
}

// ── Daily hour gate ───────────────────────────────────────────────────────────

func TestQuestReminder_DailyHourGate_NotReached(t *testing.T) {
	// 19:30 BKK — one tick before the daily gate (20:00).
	questRepo := &fakeQuestRepo{incompleteByUser: map[uuid.UUID]int{uuid.New(): 1}}
	noti := &fakeNotifier{}
	svc := newSvc(
		func() time.Time { return bkkAt(2026, 6, 10, 19, 30) },
		questRepo, newRemLog(true), noti,
	)

	result, err := svc.Run(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(questRepo.calls) > 0 {
		t.Errorf("CountIncompleteByUser called %d times, want 0 (before hour gate)", len(questRepo.calls))
	}
	if result.UsersNotified != 0 {
		t.Errorf("UsersNotified = %d, want 0", result.UsersNotified)
	}
}

func TestQuestReminder_DailyFires_AfterHour(t *testing.T) {
	userID := uuid.New()
	questRepo := &fakeQuestRepo{incompleteByUser: map[uuid.UUID]int{userID: 3}}
	noti := &fakeNotifier{}
	svc := newSvc(
		func() time.Time { return bkkAt(2026, 6, 10, 20, 30) },
		questRepo, newRemLog(true), noti,
	)

	result, err := svc.Run(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.UsersNotified != 1 {
		t.Errorf("UsersNotified = %d, want 1", result.UsersNotified)
	}
	if result.ItemsByType["quest_daily"] != 1 {
		t.Errorf("quest_daily count = %d, want 1", result.ItemsByType["quest_daily"])
	}
	if len(noti.calls) != 1 {
		t.Fatalf("Notify calls = %d, want 1", len(noti.calls))
	}
	if !strings.Contains(noti.calls[0].Body, "3 daily") {
		t.Errorf("notification body %q should mention count 3", noti.calls[0].Body)
	}
}

func TestQuestReminder_DailyDedup(t *testing.T) {
	// remLog returns false → already sent → no notification.
	questRepo := &fakeQuestRepo{incompleteByUser: map[uuid.UUID]int{uuid.New(): 2}}
	noti := &fakeNotifier{}
	svc := newSvc(
		func() time.Time { return bkkAt(2026, 6, 10, 20, 30) },
		questRepo, newRemLog(false), noti, // alwaysNew=false → deduped
	)

	result, err := svc.Run(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.UsersNotified != 0 {
		t.Errorf("UsersNotified = %d, want 0 (deduped)", result.UsersNotified)
	}
	if len(noti.calls) != 0 {
		t.Errorf("Notify calls = %d, want 0 (deduped)", len(noti.calls))
	}
}

// ── Weekly gate ───────────────────────────────────────────────────────────────

func TestQuestReminder_WeeklyNotOnSunday(t *testing.T) {
	// Wednesday: weekly should never fire.
	questRepo := &fakeQuestRepo{incompleteByUser: map[uuid.UUID]int{uuid.New(): 1}}
	noti := &fakeNotifier{}
	svc := newSvc(
		func() time.Time { return bkkAt(2026, 6, 10, 20, 30) }, // Wednesday 20:30
		questRepo, newRemLog(true), noti,
	)

	_, err := svc.Run(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, c := range questRepo.calls {
		if c == "weekly" {
			t.Error("CountIncompleteByUser called for weekly on a non-Sunday")
		}
	}
}

func TestQuestReminder_WeeklyHourGate_NotReached(t *testing.T) {
	// Sunday 17:30 — before weeklyNudgeHour (18:00).
	questRepo := &fakeQuestRepo{incompleteByUser: map[uuid.UUID]int{uuid.New(): 1}}
	noti := &fakeNotifier{}
	svc := newSvc(
		func() time.Time { return bkkAt(2026, 6, 14, 17, 30) }, // Sunday 17:30
		questRepo, newRemLog(true), noti,
	)

	_, err := svc.Run(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, c := range questRepo.calls {
		if c == "weekly" {
			t.Error("CountIncompleteByUser called for weekly before hour gate on Sunday")
		}
	}
}

func TestQuestReminder_WeeklyFires_OnSundayAfterHour(t *testing.T) {
	userID := uuid.New()
	questRepo := &fakeQuestRepo{incompleteByUser: map[uuid.UUID]int{userID: 1}}
	noti := &fakeNotifier{}
	svc := newSvc(
		func() time.Time { return bkkAt(2026, 6, 14, 18, 30) }, // Sunday 18:30
		questRepo, newRemLog(true), noti,
	)

	result, err := svc.Run(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ItemsByType["quest_weekly"] != 1 {
		t.Errorf("quest_weekly count = %d, want 1", result.ItemsByType["quest_weekly"])
	}
	if len(noti.calls) != 1 {
		t.Fatalf("Notify calls = %d, want 1", len(noti.calls))
	}
	if !strings.Contains(noti.calls[0].Body, "weekly") {
		t.Errorf("notification body %q should mention weekly", noti.calls[0].Body)
	}
}

// ── Combined daily + weekly ───────────────────────────────────────────────────

func TestQuestReminder_CombinesDailyAndWeekly(t *testing.T) {
	// Sunday 20:30 BKK — both daily (≥20) and weekly (≥18) gates open.
	// Same user has both incomplete daily and weekly quests.
	userID := uuid.New()
	questRepo := &fakeQuestRepo{incompleteByUser: map[uuid.UUID]int{userID: 2}}
	noti := &fakeNotifier{}
	svc := newSvc(
		func() time.Time { return bkkAt(2026, 6, 14, 20, 30) }, // Sunday 20:30
		questRepo, newRemLog(true), noti,
	)

	result, err := svc.Run(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Both quest_daily and quest_weekly should fire.
	if result.ItemsByType["quest_daily"] != 1 {
		t.Errorf("quest_daily = %d, want 1", result.ItemsByType["quest_daily"])
	}
	if result.ItemsByType["quest_weekly"] != 1 {
		t.Errorf("quest_weekly = %d, want 1", result.ItemsByType["quest_weekly"])
	}
	// Both nudges go to the same user → exactly one notification combining both lines.
	if result.UsersNotified != 1 {
		t.Errorf("UsersNotified = %d, want 1 (combined notification)", result.UsersNotified)
	}
	if len(noti.calls) != 1 {
		t.Fatalf("Notify calls = %d, want 1", len(noti.calls))
	}
	body := noti.calls[0].Body
	if !strings.Contains(body, "daily") {
		t.Errorf("combined body %q should mention daily", body)
	}
	if !strings.Contains(body, "weekly") {
		t.Errorf("combined body %q should mention weekly", body)
	}
}

// ── todayFor helper ───────────────────────────────────────────────────────────

func TestTodayFor(t *testing.T) {
	// 2026-06-10 at 23:30 BKK is still 2026-06-10 Bangkok date.
	got := todayFor(bkkAt(2026, 6, 10, 23, 30))
	if got.Format(time.DateOnly) != "2026-06-10" {
		t.Errorf("todayFor(23:30 BKK 2026-06-10) = %s, want 2026-06-10", got.Format(time.DateOnly))
	}
	// Must be UTC midnight.
	if got.Hour() != 0 || got.Minute() != 0 {
		t.Errorf("todayFor() = %v, want UTC midnight", got)
	}
}
