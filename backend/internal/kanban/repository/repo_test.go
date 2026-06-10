package repository

import (
	"testing"

	"github.com/kinkando/personal-dashboard/internal/kanban"
)

func TestInferColumnType(t *testing.T) {
	cases := []struct {
		name       string
		wantType   kanban.ColumnType
		wantSystem bool
	}{
		{"To Do", kanban.ColumnTypeTodo, true},
		{"In Progress", kanban.ColumnTypeInProgress, true},
		{"Done", kanban.ColumnTypeDone, true},
		{"My Column", kanban.ColumnTypeCustom, false},
		{"", kanban.ColumnTypeCustom, false},
	}
	for _, tc := range cases {
		gotType, gotSystem := inferColumnType(tc.name)
		if gotType != tc.wantType {
			t.Errorf("inferColumnType(%q) type = %q, want %q", tc.name, gotType, tc.wantType)
		}
		if gotSystem != tc.wantSystem {
			t.Errorf("inferColumnType(%q) isSystem = %v, want %v", tc.name, gotSystem, tc.wantSystem)
		}
	}
}
