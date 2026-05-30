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

func (r *Repository) GetBoard(ctx context.Context, userID string) (*kanban.Board, error) {
	var board kanban.Board
	err := r.boards.FindOne(ctx, bson.M{"user_id": userID}).Decode(&board)
	if err == mongo.ErrNoDocuments {
		return r.createBoard(ctx, userID)
	}
	if err != nil {
		return nil, fmt.Errorf("find board: %w", err)
	}
	return &board, nil
}

func (r *Repository) createBoard(ctx context.Context, userID string) (*kanban.Board, error) {
	board := kanban.Board{
		ID:        primitive.NewObjectID(),
		UserID:    userID,
		CreatedAt: time.Now(),
	}
	if _, err := r.boards.InsertOne(ctx, board); err != nil {
		return nil, fmt.Errorf("insert board: %w", err)
	}
	for i, name := range defaultColumns {
		col := kanban.Column{
			ID:        primitive.NewObjectID(),
			BoardID:   board.ID,
			Name:      name,
			Order:     i,
			CreatedAt: time.Now(),
		}
		if _, err := r.columns.InsertOne(ctx, col); err != nil {
			return nil, fmt.Errorf("insert default column %q: %w", name, err)
		}
	}
	return &board, nil
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

func (r *Repository) CreateCard(ctx context.Context, boardID, columnID primitive.ObjectID, in kanban.CreateCardInput) (*kanban.Card, error) {
	count, err := r.cards.CountDocuments(ctx, bson.M{"board_id": boardID, "column_id": columnID})
	if err != nil {
		return nil, fmt.Errorf("count cards: %w", err)
	}
	card := kanban.Card{
		ID:        primitive.NewObjectID(),
		BoardID:   boardID,
		ColumnID:  columnID,
		Title:     in.Title,
		Content:   in.Content,
		Order:     int(count),
		CreatedAt: time.Now(),
	}
	if _, err := r.cards.InsertOne(ctx, card); err != nil {
		return nil, fmt.Errorf("insert card: %w", err)
	}
	return &card, nil
}

func (r *Repository) MoveCard(ctx context.Context, cardID primitive.ObjectID, in kanban.MoveCardInput) error {
	colID, err := primitive.ObjectIDFromHex(in.ColumnID)
	if err != nil {
		return fmt.Errorf("invalid column_id: %w", err)
	}
	_, err = r.cards.UpdateOne(ctx,
		bson.M{"_id": cardID},
		bson.M{"$set": bson.M{"column_id": colID, "order": in.Order}},
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
