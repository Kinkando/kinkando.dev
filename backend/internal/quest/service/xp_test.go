package service

import (
	"testing"
	"time"

	"github.com/kinkando/personal-dashboard/internal/quest"
)

// ── xpSummary ─────────────────────────────────────────────────────────────────

func TestXPSummary(t *testing.T) {
	cases := []struct {
		name    string
		totalXP int
		want    quest.XPSummary
	}{
		{
			name:    "zero XP",
			totalXP: 0,
			want:    quest.XPSummary{TotalXP: 0, Level: 1, XPIntoLevel: 0, XPForLevel: 100, XPToNext: 100},
		},
		{
			name:    "99 XP — last of level 1",
			totalXP: 99,
			want:    quest.XPSummary{TotalXP: 99, Level: 1, XPIntoLevel: 99, XPForLevel: 100, XPToNext: 1},
		},
		{
			name:    "100 XP — first of level 2",
			totalXP: 100,
			want:    quest.XPSummary{TotalXP: 100, Level: 2, XPIntoLevel: 0, XPForLevel: 100, XPToNext: 100},
		},
		{
			name:    "250 XP — level 3 halfway",
			totalXP: 250,
			want:    quest.XPSummary{TotalXP: 250, Level: 3, XPIntoLevel: 50, XPForLevel: 100, XPToNext: 50},
		},
		{
			name:    "1000 XP — exact level boundary",
			totalXP: 1000,
			want:    quest.XPSummary{TotalXP: 1000, Level: 11, XPIntoLevel: 0, XPForLevel: 100, XPToNext: 100},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := xpSummary(tc.totalXP)
			if got != tc.want {
				t.Errorf("xpSummary(%d):\n  got  %+v\n  want %+v", tc.totalXP, got, tc.want)
			}
		})
	}
}

// ── weekStart ─────────────────────────────────────────────────────────────────

// weekStartFixedNow returns a Service whose clock is pinned to a fixed date
// so weekStart() is deterministic regardless of when the test runs.
func weekStartFixedNow(dateStr string) *Service {
	return &Service{now: func() time.Time { return mustParseDate(dateStr) }}
}

func TestWeekStart(t *testing.T) {
	// All days of the week Mon 2026-06-08 .. Sun 2026-06-14 must return 2026-06-08.
	// The following Monday (2026-06-15) starts a new week.
	cases := []struct {
		name     string
		today    string
		wantDate string
	}{
		{"Monday", "2026-06-08", "2026-06-08"},
		{"Tuesday", "2026-06-09", "2026-06-08"},
		{"Wednesday", "2026-06-10", "2026-06-08"},
		{"Thursday", "2026-06-11", "2026-06-08"},
		{"Friday", "2026-06-12", "2026-06-08"},
		{"Saturday", "2026-06-13", "2026-06-08"},
		{"Sunday", "2026-06-14", "2026-06-08"},
		{"next Monday", "2026-06-15", "2026-06-15"},
		// Cross-month boundary: Friday 2026-05-29 is in the week Mon 2026-05-25.
		{"cross-month Friday", "2026-05-29", "2026-05-25"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc := weekStartFixedNow(tc.today)
			got := svc.weekStart().Format(time.DateOnly)
			if got != tc.wantDate {
				t.Errorf("weekStart() for %s = %s, want %s", tc.today, got, tc.wantDate)
			}
		})
	}
}

// mustParseDate parses a "YYYY-MM-DD" string as UTC midnight; panics on error.
func mustParseDate(s string) time.Time {
	t, err := time.Parse(time.DateOnly, s)
	if err != nil {
		panic("mustParseDate: " + err.Error())
	}
	return t
}
