package kanban

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Priority represents the urgency level of a kanban card.
type Priority string

const (
	PriorityNone   Priority = "none"
	PriorityLow    Priority = "low"
	PriorityMedium Priority = "medium"
	PriorityHigh   Priority = "high"
	PriorityUrgent Priority = "urgent"
)

// Valid reports whether p is a recognised priority value.
func (p Priority) Valid() bool {
	switch p {
	case PriorityNone, PriorityLow, PriorityMedium, PriorityHigh, PriorityUrgent:
		return true
	}
	return false
}

// ColumnType identifies the semantic role of a kanban column.
type ColumnType string

const (
	ColumnTypeTodo       ColumnType = "todo"
	ColumnTypeInProgress ColumnType = "in_progress"
	ColumnTypeDone       ColumnType = "done"
	ColumnTypeCustom     ColumnType = "custom"
)


// ArchiveReason records why a card was archived.
type ArchiveReason string

const (
	ArchiveReasonCompleted ArchiveReason = "completed"
	ArchiveReasonCancelled ArchiveReason = "cancelled"
	ArchiveReasonDuplicate ArchiveReason = "duplicate"
	ArchiveReasonStale     ArchiveReason = "stale"
)

// ValidUserSupplied reports whether r is a reason the user may supply directly.
// ArchiveReasonCompleted is reserved for system use (auto-assigned when archiving
// from a Done column).
func (r ArchiveReason) ValidUserSupplied() bool {
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
	Type      ColumnType         `bson:"type"          json:"type"`
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
	Priority      Priority           `bson:"priority"                 json:"priority"`
	DueDate       *time.Time         `bson:"due_date,omitempty"       json:"due_date,omitempty"`
	Tags          []string           `bson:"tags"                     json:"tags"`
	Order         int                `bson:"order"                    json:"order"`
	CompletedAt   *time.Time         `bson:"completed_at,omitempty"   json:"completed_at,omitempty"`
	ArchivedAt    *time.Time         `bson:"archived_at,omitempty"    json:"archived_at,omitempty"`
	ArchiveReason ArchiveReason      `bson:"archive_reason,omitempty" json:"archive_reason,omitempty"`
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
	Reason ArchiveReason `json:"reason" validate:"omitempty,oneof=cancelled duplicate stale"` // cancelled | duplicate | stale; "completed" is system-assigned
}

type CreateCardInput struct {
	BoardID     string   `json:"board_id"  validate:"required"`
	ColumnID    string   `json:"column_id" validate:"required"`
	Title       string   `json:"title"     validate:"required"`
	Content     string   `json:"content"`
	Description string   `json:"description"`
	Priority    Priority `json:"priority"  validate:"omitempty,oneof=none low medium high urgent"`
	DueDate     *string  `json:"due_date"` // "YYYY-MM-DD"
	Tags        []string `json:"tags"`
}

// UpdateCardInput uses pointer fields so only explicitly sent fields are updated.
// For due_date: nil = don't update; "" = clear; "YYYY-MM-DD" = set new value.
type UpdateCardInput struct {
	Title       *string   `json:"title"`
	Description *string   `json:"description"`
	Priority    *Priority `json:"priority"  validate:"omitempty,oneof=none low medium high urgent"`
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
