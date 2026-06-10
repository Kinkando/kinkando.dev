package helper

import (
	"testing"
	"time"
)

func TestToday(t *testing.T) {
	got := Today()

	// Must be UTC midnight — no time component.
	if h, m, s, ns := got.Hour(), got.Minute(), got.Second(), got.Nanosecond(); h != 0 || m != 0 || s != 0 || ns != 0 {
		t.Errorf("Today() has time component %02d:%02d:%02d.%d, want 00:00:00.0", h, m, s, ns)
	}
	if got.Location() != time.UTC {
		t.Errorf("Today() location = %v, want UTC", got.Location())
	}

	// Date must match Bangkok's current calendar day.
	loc, err := time.LoadLocation("Asia/Bangkok")
	if err != nil {
		t.Fatalf("load Asia/Bangkok: %v", err)
	}
	bangkokNow := time.Now().In(loc)
	want := time.Date(bangkokNow.Year(), bangkokNow.Month(), bangkokNow.Day(), 0, 0, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Errorf("Today() = %s, want %s (Bangkok wall-clock date %s)",
			got.Format(time.DateOnly), want.Format(time.DateOnly), bangkokNow.Format(time.DateOnly))
	}
}
