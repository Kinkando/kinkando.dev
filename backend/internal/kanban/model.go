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
	Order     int                `bson:"order"         json:"order"`
	CreatedAt time.Time          `bson:"created_at"    json:"created_at"`
}

type Card struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"      json:"id"`
	BoardID     primitive.ObjectID `bson:"board_id"           json:"board_id"`
	ColumnID    primitive.ObjectID `bson:"column_id"          json:"column_id"`
	Title       string             `bson:"title"              json:"title"`
	Content     string             `bson:"content"            json:"content"`
	Description string             `bson:"description"        json:"description"`
	Priority    string             `bson:"priority"           json:"priority"`
	DueDate     *time.Time         `bson:"due_date,omitempty" json:"due_date,omitempty"`
	Tags        []string           `bson:"tags"               json:"tags"`
	Order       int                `bson:"order"              json:"order"`
	CreatedAt   time.Time          `bson:"created_at"         json:"created_at"`
}

type CreateBoardInput struct {
	Name string `json:"name"`
}

type UpdateBoardInput struct {
	Name string `json:"name"`
}

type CreateCardInput struct {
	BoardID     string  `json:"board_id"`
	ColumnID    string  `json:"column_id"`
	Title       string  `json:"title"`
	Content     string  `json:"content"`
	Description string  `json:"description"`
	Priority    string  `json:"priority"`
	DueDate     *string `json:"due_date"` // "YYYY-MM-DD"
	Tags        []string `json:"tags"`
}

// UpdateCardInput uses pointer fields so only explicitly sent fields are updated.
// For due_date: nil = don't update; "" = clear; "YYYY-MM-DD" = set new value.
type UpdateCardInput struct {
	Title       *string   `json:"title"`
	Description *string   `json:"description"`
	Priority    *string   `json:"priority"`
	DueDate     *string   `json:"due_date"`
	Tags        *[]string `json:"tags"`
}

type MoveCardInput struct {
	ColumnID string `json:"column_id"`
	Order    int    `json:"order"`
}

type BoardStats struct {
	Total      int            `json:"total"`
	ByColumn   map[string]int `json:"by_column"`
	ByPriority map[string]int `json:"by_priority"`
	Overdue    int            `json:"overdue"`
	NoDueDate  int            `json:"no_due_date"`
}
