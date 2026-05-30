package kanban

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Board struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID    string             `bson:"user_id"       json:"user_id"`
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
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	BoardID   primitive.ObjectID `bson:"board_id"      json:"board_id"`
	ColumnID  primitive.ObjectID `bson:"column_id"     json:"column_id"`
	Title     string             `bson:"title"         json:"title"`
	Content   string             `bson:"content"       json:"content"`
	Order     int                `bson:"order"         json:"order"`
	CreatedAt time.Time          `bson:"created_at"    json:"created_at"`
}

type CreateCardInput struct {
	ColumnID string `json:"column_id"`
	Title    string `json:"title"`
	Content  string `json:"content"`
}

type MoveCardInput struct {
	ColumnID string `json:"column_id"`
	Order    int    `json:"order"`
}
