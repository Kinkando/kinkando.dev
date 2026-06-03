package helper

import "time"

// today returns midnight UTC for the current date in Asia/Bangkok timezone.
// Matches how ProgressTab sends logged_at (Bangkok todayDate()) and how the
// quest service computes its period_start, so "today" aligns on all three sides.
func Today() time.Time {
	loc, _ := time.LoadLocation("Asia/Bangkok")
	now := time.Now().In(loc)
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
}
