package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/kinkando/personal-dashboard/internal/kanban"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Repository struct {
	boards  *mongo.Collection
	columns *mongo.Collection
	cards   *mongo.Collection
}

func New(db *mongo.Database) *Repository {
	return &Repository{
		boards:  db.Collection("boards"),
		columns: db.Collection("columns"),
		cards:   db.Collection("cards"),
	}
}

var defaultColumns = []string{"To Do", "In Progress", "Done"}

// ListBoards returns all boards for the user sorted by created_at.
// If none exist, one default board is auto-created (preserves original behavior).
func (r *Repository) ListBoards(ctx context.Context, userID string) ([]*kanban.Board, error) {
	cur, err := r.boards.Find(ctx, bson.M{"user_id": userID}, options.Find().SetSort(bson.M{"created_at": 1}))
	if err != nil {
		return nil, fmt.Errorf("find boards: %w", err)
	}
	defer cur.Close(ctx)
	var boards []*kanban.Board
	if err := cur.All(ctx, &boards); err != nil {
		return nil, err
	}
	if len(boards) == 0 {
		board, err := r.CreateBoard(ctx, userID, "My Board")
		if err != nil {
			return nil, err
		}
		return []*kanban.Board{board}, nil
	}
	return boards, nil
}

// GetBoardByID returns a board by ID enforcing user ownership.
// Returns "board not found" error when the board doesn't exist or belongs to another user.
func (r *Repository) GetBoardByID(ctx context.Context, boardID primitive.ObjectID, userID string) (*kanban.Board, error) {
	var board kanban.Board
	err := r.boards.FindOne(ctx, bson.M{"_id": boardID, "user_id": userID}).Decode(&board)
	if err == mongo.ErrNoDocuments {
		return nil, fmt.Errorf("board not found")
	}
	if err != nil {
		return nil, fmt.Errorf("find board: %w", err)
	}
	return &board, nil
}

// CreateBoard inserts a new board and seeds it with the default columns.
func (r *Repository) CreateBoard(ctx context.Context, userID, name string) (*kanban.Board, error) {
	if name == "" {
		name = "My Board"
	}
	board := kanban.Board{
		ID:        primitive.NewObjectID(),
		UserID:    userID,
		Name:      name,
		CreatedAt: time.Now(),
	}
	if _, err := r.boards.InsertOne(ctx, board); err != nil {
		return nil, fmt.Errorf("insert board: %w", err)
	}
	for i, colName := range defaultColumns {
		col := kanban.Column{
			ID:        primitive.NewObjectID(),
			BoardID:   board.ID,
			Name:      colName,
			Order:     i,
			CreatedAt: time.Now(),
		}
		if _, err := r.columns.InsertOne(ctx, col); err != nil {
			return nil, fmt.Errorf("insert default column %q: %w", colName, err)
		}
	}
	return &board, nil
}

// UpdateBoard renames a board.
func (r *Repository) UpdateBoard(ctx context.Context, boardID primitive.ObjectID, name string) error {
	res, err := r.boards.UpdateOne(ctx, bson.M{"_id": boardID}, bson.M{"$set": bson.M{"name": name}})
	if err != nil {
		return fmt.Errorf("update board: %w", err)
	}
	if res.MatchedCount == 0 {
		return fmt.Errorf("board not found")
	}
	return nil
}

// DeleteBoard deletes a board and cascades to its columns and cards.
func (r *Repository) DeleteBoard(ctx context.Context, boardID primitive.ObjectID) error {
	if _, err := r.cards.DeleteMany(ctx, bson.M{"board_id": boardID}); err != nil {
		return fmt.Errorf("delete cards: %w", err)
	}
	if _, err := r.columns.DeleteMany(ctx, bson.M{"board_id": boardID}); err != nil {
		return fmt.Errorf("delete columns: %w", err)
	}
	res, err := r.boards.DeleteOne(ctx, bson.M{"_id": boardID})
	if err != nil {
		return fmt.Errorf("delete board: %w", err)
	}
	if res.DeletedCount == 0 {
		return fmt.Errorf("board not found")
	}
	return nil
}

func (r *Repository) GetColumns(ctx context.Context, boardID primitive.ObjectID) ([]*kanban.Column, error) {
	cur, err := r.columns.Find(ctx, bson.M{"board_id": boardID}, options.Find().SetSort(bson.M{"order": 1}))
	if err != nil {
		return nil, fmt.Errorf("find columns: %w", err)
	}
	defer cur.Close(ctx)
	var cols []*kanban.Column
	if err := cur.All(ctx, &cols); err != nil {
		return nil, err
	}
	return cols, nil
}

func (r *Repository) GetCards(ctx context.Context, boardID primitive.ObjectID) ([]*kanban.Card, error) {
	cur, err := r.cards.Find(ctx, bson.M{"board_id": boardID}, options.Find().SetSort(bson.D{{Key: "column_id", Value: 1}, {Key: "order", Value: 1}}))
	if err != nil {
		return nil, fmt.Errorf("find cards: %w", err)
	}
	defer cur.Close(ctx)
	var cards []*kanban.Card
	if err := cur.All(ctx, &cards); err != nil {
		return nil, err
	}
	return cards, nil
}

// GetCard fetches a single card by ID.
func (r *Repository) GetCard(ctx context.Context, cardID primitive.ObjectID) (*kanban.Card, error) {
	var card kanban.Card
	err := r.cards.FindOne(ctx, bson.M{"_id": cardID}).Decode(&card)
	if err == mongo.ErrNoDocuments {
		return nil, fmt.Errorf("card not found")
	}
	if err != nil {
		return nil, fmt.Errorf("find card: %w", err)
	}
	return &card, nil
}

func (r *Repository) CreateCard(ctx context.Context, boardID, columnID primitive.ObjectID, in kanban.CreateCardInput) (*kanban.Card, error) {
	count, err := r.cards.CountDocuments(ctx, bson.M{"board_id": boardID, "column_id": columnID})
	if err != nil {
		return nil, fmt.Errorf("count cards: %w", err)
	}

	priority := in.Priority
	if !kanban.ValidPriority(priority) {
		priority = kanban.PriorityNone
	}

	tags := in.Tags
	if tags == nil {
		tags = []string{}
	}

	var dueDate *time.Time
	if in.DueDate != nil && *in.DueDate != "" {
		t, parseErr := time.Parse("2006-01-02", *in.DueDate)
		if parseErr != nil {
			return nil, fmt.Errorf("invalid due_date: %w", parseErr)
		}
		dueDate = &t
	}

	card := kanban.Card{
		ID:          primitive.NewObjectID(),
		BoardID:     boardID,
		ColumnID:    columnID,
		Title:       in.Title,
		Content:     in.Content,
		Description: in.Description,
		Priority:    priority,
		DueDate:     dueDate,
		Tags:        tags,
		Order:       int(count),
		CreatedAt:   time.Now(),
	}
	if _, err := r.cards.InsertOne(ctx, card); err != nil {
		return nil, fmt.Errorf("insert card: %w", err)
	}
	return &card, nil
}

// UpdateCard applies a partial update to a card. Only non-nil fields in UpdateCardInput are written.
// For due_date: nil = skip; "" = clear; "YYYY-MM-DD" = set.
func (r *Repository) UpdateCard(ctx context.Context, cardID primitive.ObjectID, in kanban.UpdateCardInput) (*kanban.Card, error) {
	set := bson.M{}
	unset := bson.M{}

	if in.Title != nil {
		set["title"] = *in.Title
	}
	if in.Description != nil {
		set["description"] = *in.Description
	}
	if in.Priority != nil {
		set["priority"] = *in.Priority
	}
	if in.Tags != nil {
		set["tags"] = *in.Tags
	}
	if in.DueDate != nil {
		if *in.DueDate == "" {
			unset["due_date"] = ""
		} else {
			t, parseErr := time.Parse("2006-01-02", *in.DueDate)
			if parseErr != nil {
				return nil, fmt.Errorf("invalid due_date: %w", parseErr)
			}
			set["due_date"] = t
		}
	}

	if len(set) == 0 && len(unset) == 0 {
		return r.GetCard(ctx, cardID)
	}

	update := bson.M{}
	if len(set) > 0 {
		update["$set"] = set
	}
	if len(unset) > 0 {
		update["$unset"] = unset
	}

	res, err := r.cards.UpdateOne(ctx, bson.M{"_id": cardID}, update)
	if err != nil {
		return nil, fmt.Errorf("update card: %w", err)
	}
	if res.MatchedCount == 0 {
		return nil, fmt.Errorf("card not found")
	}
	return r.GetCard(ctx, cardID)
}

func (r *Repository) MoveCard(ctx context.Context, cardID primitive.ObjectID, in kanban.MoveCardInput) error {
	colID, err := primitive.ObjectIDFromHex(in.ColumnID)
	if err != nil {
		return fmt.Errorf("invalid column_id: %w", err)
	}

	card, err := r.GetCard(ctx, cardID)
	if err != nil {
		return err
	}

	oldOrder := card.Order
	newOrder := in.Order

	if card.ColumnID == colID {
		if oldOrder == newOrder {
			return nil
		}
		if oldOrder < newOrder {
			_, err = r.cards.UpdateMany(ctx,
				bson.M{"_id": bson.M{"$ne": cardID}, "column_id": colID, "order": bson.M{"$gt": oldOrder, "$lte": newOrder}},
				bson.M{"$inc": bson.M{"order": -1}},
			)
		} else {
			_, err = r.cards.UpdateMany(ctx,
				bson.M{"_id": bson.M{"$ne": cardID}, "column_id": colID, "order": bson.M{"$gte": newOrder, "$lt": oldOrder}},
				bson.M{"$inc": bson.M{"order": 1}},
			)
		}
		if err != nil {
			return fmt.Errorf("reorder column: %w", err)
		}
	} else {
		_, err = r.cards.UpdateMany(ctx,
			bson.M{"column_id": card.ColumnID, "order": bson.M{"$gt": oldOrder}},
			bson.M{"$inc": bson.M{"order": -1}},
		)
		if err != nil {
			return fmt.Errorf("reorder source column: %w", err)
		}
		_, err = r.cards.UpdateMany(ctx,
			bson.M{"column_id": colID, "order": bson.M{"$gte": newOrder}},
			bson.M{"$inc": bson.M{"order": 1}},
		)
		if err != nil {
			return fmt.Errorf("reorder target column: %w", err)
		}
	}

	_, err = r.cards.UpdateOne(ctx,
		bson.M{"_id": cardID},
		bson.M{"$set": bson.M{"column_id": colID, "order": newOrder}},
	)
	return err
}

func (r *Repository) DeleteCard(ctx context.Context, cardID primitive.ObjectID) error {
	res, err := r.cards.DeleteOne(ctx, bson.M{"_id": cardID})
	if err != nil {
		return fmt.Errorf("delete card: %w", err)
	}
	if res.DeletedCount == 0 {
		return fmt.Errorf("card not found")
	}
	return nil
}

// GetBoardStats computes card statistics for a board by fetching all cards and tallying in Go.
func (r *Repository) GetBoardStats(ctx context.Context, boardID primitive.ObjectID) (*kanban.BoardStats, error) {
	cols, err := r.GetColumns(ctx, boardID)
	if err != nil {
		return nil, err
	}
	colName := make(map[string]string, len(cols))
	for _, c := range cols {
		colName[c.ID.Hex()] = c.Name
	}

	cards, err := r.GetCards(ctx, boardID)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	stats := &kanban.BoardStats{
		Total:      len(cards),
		ByColumn:   make(map[string]int),
		ByPriority: make(map[string]int),
	}

	for _, card := range cards {
		label := colName[card.ColumnID.Hex()]
		if label == "" {
			label = card.ColumnID.Hex()
		}
		stats.ByColumn[label]++

		p := card.Priority
		if p == "" {
			p = kanban.PriorityNone
		}
		stats.ByPriority[p]++

		if card.DueDate == nil {
			stats.NoDueDate++
		} else if card.DueDate.Before(now) {
			stats.Overdue++
		}
	}

	return stats, nil
}
