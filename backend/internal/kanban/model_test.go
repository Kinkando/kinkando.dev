package kanban

import "testing"

func TestPriority_Valid(t *testing.T) {
	cases := []struct {
		p    Priority
		want bool
	}{
		{PriorityNone, true},
		{PriorityLow, true},
		{PriorityMedium, true},
		{PriorityHigh, true},
		{PriorityUrgent, true},
		{"", false},
		{"extreme", false},
		{"NONE", false}, // case-sensitive
	}
	for _, tc := range cases {
		if got := tc.p.Valid(); got != tc.want {
			t.Errorf("Priority(%q).Valid() = %v, want %v", tc.p, got, tc.want)
		}
	}
}

func TestArchiveReason_ValidUserSupplied(t *testing.T) {
	cases := []struct {
		r    ArchiveReason
		want bool
	}{
		{ArchiveReasonCancelled, true},
		{ArchiveReasonDuplicate, true},
		{ArchiveReasonStale, true},
		{ArchiveReasonCompleted, false}, // system-reserved; clients must not supply it
		{"", false},
		{"unknown", false},
		{"CANCELLED", false}, // case-sensitive
	}
	for _, tc := range cases {
		if got := tc.r.ValidUserSupplied(); got != tc.want {
			t.Errorf("ArchiveReason(%q).ValidUserSupplied() = %v, want %v", tc.r, got, tc.want)
		}
	}
}
