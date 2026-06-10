package reminder

import (
	"context"
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

func bkkAt(year, month, day, hour, min int) time.Time {
	return time.Date(year, time.Month(month), day, hour, min, 0, 0, bangkokLoc)
}

// ── Fakes ─────────────────────────────────────────────────────────────────────

type fakeHealthRepo struct {
	candidates []uuid.UUID
	queryCalls int
}

func (f *fakeHealthRepo) UsersMissingWeightToday(_ context.Context, _ time.Time) ([]uuid.UUID, error) {
	f.queryCalls++
	return f.candidates, nil
}

type fakeRemLog struct {
	alwaysNew bool
}

func (f *fakeRemLog) Log(_ context.Context, _ uuid.UUID, _ reminderlog.Domain, _ string) (bool, error) {
	return f.alwaysNew, nil
}

type fakeNotifier struct {
	calls []notification.Message
	users []uuid.UUID
}

func (f *fakeNotifier) Notify(_ context.Context, userID uuid.UUID, msg notification.Message) *notification.DeliveryResult {
	f.calls = append(f.calls, msg)
	f.users = append(f.users, userID)
	return &notification.DeliveryResult{Attempted: 1, Delivered: 1}
}

func newSvc(nowFn func() time.Time, health *fakeHealthRepo, remLog *fakeRemLog, noti *fakeNotifier) *Service {
	return &Service{
		health: health,
		remLog: remLog,
		noti:   noti,
		log:    zap.NewNop(),
		now:    nowFn,
	}
}

// ── Tests ─────────────────────────────────────────────────────────────────────

func TestWeightNudge_HourGate_NotReached(t *testing.T) {
	// 07:30 BKK — before the weightNudgeHour (08:00) gate.
	health := &fakeHealthRepo{candidates: []uuid.UUID{uuid.New()}}
	noti := &fakeNotifier{}
	svc := newSvc(func() time.Time { return bkkAt(2026, 6, 10, 7, 30) },
		health, &fakeRemLog{alwaysNew: true}, noti)

	result, err := svc.Run(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if health.queryCalls != 0 {
		t.Errorf("UsersMissingWeightToday called %d times, want 0 (before hour gate)", health.queryCalls)
	}
	if result.UsersNotified != 0 {
		t.Errorf("UsersNotified = %d, want 0", result.UsersNotified)
	}
}

func TestWeightNudge_FiresAfterHour(t *testing.T) {
	user1, user2 := uuid.New(), uuid.New()
	health := &fakeHealthRepo{candidates: []uuid.UUID{user1, user2}}
	noti := &fakeNotifier{}
	svc := newSvc(func() time.Time { return bkkAt(2026, 6, 10, 8, 0) },
		health, &fakeRemLog{alwaysNew: true}, noti)

	result, err := svc.Run(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.UsersNotified != 2 {
		t.Errorf("UsersNotified = %d, want 2", result.UsersNotified)
	}
	if len(noti.calls) != 2 {
		t.Errorf("Notify calls = %d, want 2", len(noti.calls))
	}
	// Verify the notification content.
	for _, msg := range noti.calls {
		if msg.Title != "Weight log" {
			t.Errorf("notification title = %q, want \"Weight log\"", msg.Title)
		}
		if msg.Body == "" {
			t.Error("notification body is empty")
		}
	}
}

func TestWeightNudge_Dedup(t *testing.T) {
	// remLog returns false → already sent → no notification.
	health := &fakeHealthRepo{candidates: []uuid.UUID{uuid.New()}}
	noti := &fakeNotifier{}
	svc := newSvc(func() time.Time { return bkkAt(2026, 6, 10, 9, 0) },
		health, &fakeRemLog{alwaysNew: false}, noti)

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

func TestWeightNudge_NoCandidates(t *testing.T) {
	// All users logged weight today — empty candidate list.
	health := &fakeHealthRepo{candidates: nil}
	noti := &fakeNotifier{}
	svc := newSvc(func() time.Time { return bkkAt(2026, 6, 10, 9, 0) },
		health, &fakeRemLog{alwaysNew: true}, noti)

	result, err := svc.Run(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.UsersNotified != 0 {
		t.Errorf("UsersNotified = %d, want 0 (no candidates)", result.UsersNotified)
	}
}

// ── todayFor helper ───────────────────────────────────────────────────────────

func TestTodayFor(t *testing.T) {
	// 2026-06-10 at 23:30 BKK is still the same Bangkok calendar day.
	got := todayFor(bkkAt(2026, 6, 10, 23, 30))
	if got.Format(time.DateOnly) != "2026-06-10" {
		t.Errorf("todayFor(23:30 BKK 2026-06-10) = %s, want 2026-06-10", got.Format(time.DateOnly))
	}
	if got.Hour() != 0 || got.Minute() != 0 {
		t.Errorf("todayFor() = %v, want UTC midnight", got)
	}
}
