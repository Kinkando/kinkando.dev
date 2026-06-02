package kanban

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	PriorityNone   = "none"
	PriorityLow    = "low"
	PriorityMedium = "medium"
	PriorityHigh   = "high"
	PriorityUrgent = "urgent"
)

func ValidPriority(p string) bool {
	switch p {
	case PriorityNone, PriorityLow, PriorityMedium, PriorityHigh, PriorityUrgent:
		return true
	}
	return false
}

const (
	ColumnTypeTodo       = "todo"
	ColumnTypeInProgress = "in_progress"
	ColumnTypeDone       = "done"
	ColumnTypeCustom     = "custom"
)

func ValidColumnType(t string) bool {
	switch t {
	case ColumnTypeTodo, ColumnTypeInProgress, ColumnTypeDone, ColumnTypeCustom:
		return true
	}
	return false
}

const (
	ArchiveReasonCompleted = "completed"
	ArchiveReasonCancelled = "cancelled"
	ArchiveReasonDuplicate = "duplicate"
	ArchiveReasonStale     = "stale"
)

// ValidUserArchiveReason returns true for reasons the user may supply directly.
// "completed" is reserved for system use (auto-assigned when archiving from a Done column).
func ValidUserArchiveReason(r string) bool {
	switch r {
	case ArchiveReasonCancelled, ArchiveReasonDuplicate, ArchiveReasonStale:
		return true
	}
	return false
}

type Board struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID    string             `bson:"user_id"       json:"user_id"`
	Name      string             `bson:"name"          json:"name"`
	CreatedAt time.Time          `bson:"created_at"    json:"created_at"`
}

type Column struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	BoardID   primitive.ObjectID `bson:"board_id"      json:"board_id"`
	Name      string             `bson:"name"          json:"name"`
	Type      string             `bson:"type"          json:"type"`      // todo | in_progress | done | custom
	IsSystem  bool               `bson:"is_system"     json:"is_system"` // system columns cannot be deleted
	Order     int                `bson:"order"         json:"order"`
	CreatedAt time.Time          `bson:"created_at"    json:"created_at"`
}

type Card struct {
	ID            primitive.ObjectID `bson:"_id,omitempty"            json:"id"`
	BoardID       primitive.ObjectID `bson:"board_id"                 json:"board_id"`
	ColumnID      primitive.ObjectID `bson:"column_id"                json:"column_id"`
	Title         string             `bson:"title"                    json:"title"`
	Content       string             `bson:"content"                  json:"content"`
	Description   string             `bson:"description"              json:"description"`
	Priority      string             `bson:"priority"                 json:"priority"`
	DueDate       *time.Time         `bson:"due_date,omitempty"       json:"due_date,omitempty"`
	Tags          []string           `bson:"tags"                     json:"tags"`
	Order         int                `bson:"order"                    json:"order"`
	CompletedAt   *time.Time         `bson:"completed_at,omitempty"   json:"completed_at,omitempty"`
	ArchivedAt    *time.Time         `bson:"archived_at,omitempty"    json:"archived_at,omitempty"`
	ArchiveReason string             `bson:"archive_reason,omitempty" json:"archive_reason,omitempty"`
	CreatedAt     time.Time          `bson:"created_at"               json:"created_at"`
}

type CreateBoardInput struct {
	Name string `json:"name" validate:"required"`
}

type UpdateBoardInput struct {
	Name string `json:"name" validate:"required"`
}

type CreateColumnInput struct {
	BoardID string `json:"board_id" validate:"required"`
	Name    string `json:"name"     validate:"required"`
}

type UpdateColumnInput struct {
	Name string `json:"name" validate:"required"`
}

type ReorderColumnsInput struct {
	ColumnIDs []string `json:"column_ids" validate:"required,min=1,dive,required"`
}

type DeleteColumnInput struct {
	Action         string `json:"action"           validate:"required,oneof=move archive"` // "move" | "archive"
	TargetColumnID string `json:"target_column_id" validate:"required_if=Action move"`     // required when action is "move"
}

type ArchiveCardInput struct {
	Reason string `json:"reason" validate:"omitempty,oneof=cancelled duplicate stale"` // cancelled | duplicate | stale; "completed" is system-assigned
}

type CreateCardInput struct {
	BoardID     string   `json:"board_id"  validate:"required"`
	ColumnID    string   `json:"column_id" validate:"required"`
	Title       string   `json:"title"     validate:"required"`
	Content     string   `json:"content"`
	Description string   `json:"description"`
	Priority    string   `json:"priority"  validate:"omitempty,oneof=none low medium high urgent"`
	DueDate     *string  `json:"due_date"` // "YYYY-MM-DD"
	Tags        []string `json:"tags"`
}

// UpdateCardInput uses pointer fields so only explicitly sent fields are updated.
// For due_date: nil = don't update; "" = clear; "YYYY-MM-DD" = set new value.
type UpdateCardInput struct {
	Title       *string   `json:"title"`
	Description *string   `json:"description"`
	Priority    *string   `json:"priority"  validate:"omitempty,oneof=none low medium high urgent"`
	DueDate     *string   `json:"due_date"`
	Tags        *[]string `json:"tags"`
}

type MoveCardInput struct {
	ColumnID string `json:"column_id" validate:"required"`
	Order    int    `json:"order"     validate:"min=0"`
}

type BoardStats struct {
	Total      int            `json:"total"`
	ByColumn   map[string]int `json:"by_column"`
	ByPriority map[string]int `json:"by_priority"`
	Overdue    int            `json:"overdue"`
	NoDueDate  int            `json:"no_due_date"`
}

// ListArchivedFilter restricts the archive listing.
// Reason: "completed" (done-archive tab) | "general" (non-completed archive tab) | "" (all archived).
type ListArchivedFilter struct {
	Reason string // "completed" | "general" | ""
	Month  int    // 0 = all months; 1–12 = specific month
	Year   int    // 0 = all years
}
