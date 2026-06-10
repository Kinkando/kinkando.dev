package snapshot

import (
	"context"
	"testing"
	"time"

	"github.com/kinkando/personal-dashboard/internal/quest"
	"go.uber.org/zap"
)

// ── Fakes ─────────────────────────────────────────────────────────────────────

type fakeQuestRepo struct {
	// calls records ("questType", periodStart) pairs in order.
	calls []recordCall
}

type recordCall struct {
	questType   string
	periodStart time.Time
}

func (f *fakeQuestRepo) RecordPeriodResults(_ context.Context, questType string, periodStart time.Time) (*quest.PeriodSnapshotResult, error) {
	f.calls = append(f.calls, recordCall{questType, periodStart})
	return &quest.PeriodSnapshotResult{Total: 1}, nil
}

func newSvc(nowFn func() time.Time, repo *fakeQuestRepo) *Service {
	return &Service{quests: repo, log: zap.NewNop(), now: nowFn}
}

// bkkAt returns a Bangkok-timezone time at the given components.
func bkkAt(year, month, day, hour, min int) time.Time {
	loc, _ := time.LoadLocation("Asia/Bangkok")
	return time.Date(year, time.Month(month), day, hour, min, 0, 0, loc)
}

// ── Run: daily always records yesterday ───────────────────────────────────────

func TestRun_DailyAlwaysRecordsYesterday(t *testing.T) {
	// 2026-06-10 (Wednesday) 00:30 BKK — only daily should run.
	repo := &fakeQuestRepo{}
	svc := newSvc(func() time.Time { return bkkAt(2026, 6, 10, 0, 30) }, repo)

	result, err := svc.Run(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Daily == nil {
		t.Fatal("Daily result should not be nil")
	}
	if result.Weekly != nil {
		t.Error("Weekly result should be nil on a non-Monday")
	}

	// Verify the daily call recorded yesterday (2026-06-09 UTC midnight).
	wantYesterday := time.Date(2026, 6, 9, 0, 0, 0, 0, time.UTC)
	var dailyCall *recordCall
	for i := range repo.calls {
		if repo.calls[i].questType == "daily" {
			dailyCall = &repo.calls[i]
		}
	}
	if dailyCall == nil {
		t.Fatal("no daily RecordPeriodResults call found")
	}
	if !dailyCall.periodStart.Equal(wantYesterday) {
		t.Errorf("daily periodStart = %v, want %v", dailyCall.periodStart, wantYesterday)
	}
}

// ── Run: weekly only on Monday ────────────────────────────────────────────────

func TestRun_WeeklyOnlyOnMonday(t *testing.T) {
	// 2026-06-15 (Monday) 00:30 BKK — both daily and weekly should run.
	repo := &fakeQuestRepo{}
	svc := newSvc(func() time.Time { return bkkAt(2026, 6, 15, 0, 30) }, repo)

	result, err := svc.Run(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Daily == nil {
		t.Error("Daily result should not be nil on Monday")
	}
	if result.Weekly == nil {
		t.Error("Weekly result should not be nil on Monday")
	}

	// On Monday 2026-06-15, yesterday = Sunday 2026-06-14.
	// The week that ended = Mon 2026-06-08 through Sun 2026-06-14 → weekStart = 2026-06-08.
	wantWeekStart := time.Date(2026, 6, 8, 0, 0, 0, 0, time.UTC)
	var weeklyCall *recordCall
	for i := range repo.calls {
		if repo.calls[i].questType == "weekly" {
			weeklyCall = &repo.calls[i]
		}
	}
	if weeklyCall == nil {
		t.Fatal("no weekly RecordPeriodResults call found")
	}
	if !weeklyCall.periodStart.Equal(wantWeekStart) {
		t.Errorf("weekly periodStart = %v, want %v", weeklyCall.periodStart, wantWeekStart)
	}
}

func TestRun_WeeklySkippedOnNonMonday(t *testing.T) {
	days := []struct {
		name string
		t    time.Time
	}{
		{"Tuesday", bkkAt(2026, 6, 9, 0, 30)},
		{"Sunday", bkkAt(2026, 6, 14, 0, 30)},
		{"Saturday", bkkAt(2026, 6, 13, 0, 30)},
	}

	for _, tc := range days {
		t.Run(tc.name, func(t *testing.T) {
			repo := &fakeQuestRepo{}
			svc := newSvc(func() time.Time { return tc.t }, repo)

			_, err := svc.Run(context.Background())
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			for _, c := range repo.calls {
				if c.questType == "weekly" {
					t.Errorf("weekly RecordPeriodResults called on %s", tc.name)
				}
			}
		})
	}
}

// ── weekStartFor pure helper ──────────────────────────────────────────────────

func TestWeekStartFor(t *testing.T) {
	cases := []struct {
		day       time.Time
		wantStart time.Time
	}{
		// Monday — weekStart is itself.
		{
			time.Date(2026, 6, 8, 0, 0, 0, 0, time.UTC), // Monday
			time.Date(2026, 6, 8, 0, 0, 0, 0, time.UTC),
		},
		// Sunday — weekStart is the Monday 6 days earlier.
		{
			time.Date(2026, 6, 14, 0, 0, 0, 0, time.UTC), // Sunday
			time.Date(2026, 6, 8, 0, 0, 0, 0, time.UTC),
		},
		// Wednesday — weekStart is Monday 2 days earlier.
		{
			time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC), // Wednesday
			time.Date(2026, 6, 8, 0, 0, 0, 0, time.UTC),
		},
		// Cross-month: Friday 2026-05-29 → weekStart Monday 2026-05-25.
		{
			time.Date(2026, 5, 29, 0, 0, 0, 0, time.UTC), // Friday
			time.Date(2026, 5, 25, 0, 0, 0, 0, time.UTC),
		},
	}

	for _, tc := range cases {
		got := weekStartFor(tc.day)
		if !got.Equal(tc.wantStart) {
			t.Errorf("weekStartFor(%s) = %v, want %v",
				tc.day.Weekday(), got, tc.wantStart)
		}
	}
}

// ── todayFor helper ───────────────────────────────────────────────────────────

func TestTodayFor(t *testing.T) {
	// 23:30 BKK on 2026-06-10 is still 2026-06-10 Bangkok date, midnight UTC.
	loc, _ := time.LoadLocation("Asia/Bangkok")
	in := time.Date(2026, 6, 10, 23, 30, 0, 0, loc)
	got := todayFor(in)
	want := time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Errorf("todayFor(23:30 BKK 2026-06-10) = %v, want %v", got, want)
	}
}
