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

// defaultColumns seeds a new board with system columns.
var defaultColumns = []struct {
	Name     string
	Type     kanban.ColumnType
	IsSystem bool
}{
	{Name: "To Do", Type: kanban.ColumnTypeTodo, IsSystem: true},
	{Name: "In Progress", Type: kanban.ColumnTypeInProgress, IsSystem: true},
	{Name: "Done", Type: kanban.ColumnTypeDone, IsSystem: true},
}

// inferColumnType maps a legacy column name to its canonical type and system flag.
func inferColumnType(name string) (kanban.ColumnType, bool) {
	switch name {
	case "To Do":
		return kanban.ColumnTypeTodo, true
	case "In Progress":
		return kanban.ColumnTypeInProgress, true
	case "Done":
		return kanban.ColumnTypeDone, true
	default:
		return kanban.ColumnTypeCustom, false
	}
}

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

// CreateBoard inserts a new board and seeds it with the default typed system columns.
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
	for i, seed := range defaultColumns {
		col := kanban.Column{
			ID:        primitive.NewObjectID(),
			BoardID:   board.ID,
			Name:      seed.Name,
			Type:      seed.Type,
			IsSystem:  seed.IsSystem,
			Order:     i,
			CreatedAt: time.Now(),
		}
		if _, err := r.columns.InsertOne(ctx, col); err != nil {
			return nil, fmt.Errorf("insert default column %q: %w", seed.Name, err)
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

// GetColumns returns all columns for a board sorted by order. A lazy backfill is applied
// to columns that predate the type/is_system fields.
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

	// Lazy backfill: columns created before type/is_system fields were added have Type == "".
	var toUpdate []mongo.WriteModel
	for _, col := range cols {
		if col.Type == "" {
			colType, isSystem := inferColumnType(col.Name)
			col.Type = colType
			col.IsSystem = isSystem
			toUpdate = append(toUpdate, mongo.NewUpdateOneModel().
				SetFilter(bson.M{"_id": col.ID}).
				SetUpdate(bson.M{"$set": bson.M{"type": colType, "is_system": isSystem}}))
		}
	}
	if len(toUpdate) > 0 {
		_, _ = r.columns.BulkWrite(ctx, toUpdate) // best-effort; non-fatal
	}

	return cols, nil
}

// GetColumn fetches a single column by ID, applying backfill if needed.
func (r *Repository) GetColumn(ctx context.Context, columnID primitive.ObjectID) (*kanban.Column, error) {
	var col kanban.Column
	err := r.columns.FindOne(ctx, bson.M{"_id": columnID}).Decode(&col)
	if err == mongo.ErrNoDocuments {
		return nil, fmt.Errorf("column not found")
	}
	if err != nil {
		return nil, fmt.Errorf("find column: %w", err)
	}
	if col.Type == "" {
		col.Type, col.IsSystem = inferColumnType(col.Name)
		_, _ = r.columns.UpdateOne(ctx,
			bson.M{"_id": col.ID},
			bson.M{"$set": bson.M{"type": col.Type, "is_system": col.IsSystem}},
		)
	}
	return &col, nil
}

// CreateColumn appends a new custom column to a board.
func (r *Repository) CreateColumn(ctx context.Context, boardID primitive.ObjectID, name string) (*kanban.Column, error) {
	count, err := r.columns.CountDocuments(ctx, bson.M{"board_id": boardID})
	if err != nil {
		return nil, fmt.Errorf("count columns: %w", err)
	}
	col := kanban.Column{
		ID:        primitive.NewObjectID(),
		BoardID:   boardID,
		Name:      name,
		Type:      kanban.ColumnTypeCustom,
		IsSystem:  false,
		Order:     int(count),
		CreatedAt: time.Now(),
	}
	if _, err := r.columns.InsertOne(ctx, col); err != nil {
		return nil, fmt.Errorf("insert column: %w", err)
	}
	return &col, nil
}

// UpdateColumn renames a column. Type and IsSystem are immutable.
func (r *Repository) UpdateColumn(ctx context.Context, columnID primitive.ObjectID, name string) error {
	res, err := r.columns.UpdateOne(ctx, bson.M{"_id": columnID}, bson.M{"$set": bson.M{"name": name}})
	if err != nil {
		return fmt.Errorf("update column: %w", err)
	}
	if res.MatchedCount == 0 {
		return fmt.Errorf("column not found")
	}
	return nil
}

// ReorderColumns sets each column's order to its index in the provided slice.
func (r *Repository) ReorderColumns(ctx context.Context, boardID primitive.ObjectID, columnIDs []string) error {
	var models []mongo.WriteModel
	for i, idStr := range columnIDs {
		colID, err := primitive.ObjectIDFromHex(idStr)
		if err != nil {
			return fmt.Errorf("invalid column_id %q: %w", idStr, err)
		}
		models = append(models, mongo.NewUpdateOneModel().
			SetFilter(bson.M{"_id": colID, "board_id": boardID}).
			SetUpdate(bson.M{"$set": bson.M{"order": i}}))
	}
	if len(models) > 0 {
		if _, err := r.columns.BulkWrite(ctx, models); err != nil {
			return fmt.Errorf("reorder columns: %w", err)
		}
	}
	return nil
}

// DeleteColumn deletes a non-system column. Cards are either moved to targetColumnID
// (action="move") or bulk-archived as stale (action="archive").
func (r *Repository) DeleteColumn(ctx context.Context, columnID primitive.ObjectID, action, targetColumnID string) error {
	col, err := r.GetColumn(ctx, columnID)
	if err != nil {
		return err
	}
	if col.IsSystem {
		return fmt.Errorf("system column cannot be deleted")
	}

	switch action {
	case "move":
		targetID, err := primitive.ObjectIDFromHex(targetColumnID)
		if err != nil {
			return fmt.Errorf("invalid target_column_id: %w", err)
		}
		targetCol, err := r.GetColumn(ctx, targetID)
		if err != nil {
			return fmt.Errorf("get target column: %w", err)
		}

		// Base order = current non-archived card count in target column.
		baseOrder, err := r.cards.CountDocuments(ctx,
			bson.M{"column_id": targetID, "archived_at": bson.M{"$exists": false}})
		if err != nil {
			return fmt.Errorf("count target cards: %w", err)
		}

		cur, err := r.cards.Find(ctx, bson.M{"column_id": columnID}, options.Find().SetSort(bson.M{"order": 1}))
		if err != nil {
			return fmt.Errorf("find source cards: %w", err)
		}
		defer cur.Close(ctx)
		var sourceCards []*kanban.Card
		if err := cur.All(ctx, &sourceCards); err != nil {
			return fmt.Errorf("decode source cards: %w", err)
		}

		if len(sourceCards) > 0 {
			now := time.Now()
			var bulk []mongo.WriteModel
			for i, card := range sourceCards {
				set := bson.M{
					"column_id": targetID,
					"order":     int(baseOrder) + i,
				}
				if targetCol.Type == kanban.ColumnTypeDone && card.CompletedAt == nil {
					set["completed_at"] = now
				}
				bulk = append(bulk, mongo.NewUpdateOneModel().
					SetFilter(bson.M{"_id": card.ID}).
					SetUpdate(bson.M{"$set": set}))
			}
			if _, err := r.cards.BulkWrite(ctx, bulk); err != nil {
				return fmt.Errorf("move cards to target: %w", err)
			}
		}

	case "archive":
		now := time.Now()
		_, err := r.cards.UpdateMany(ctx,
			bson.M{"column_id": columnID, "archived_at": bson.M{"$exists": false}},
			bson.M{"$set": bson.M{
				"archived_at":    now,
				"archive_reason": kanban.ArchiveReasonStale,
			}},
		)
		if err != nil {
			return fmt.Errorf("archive cards: %w", err)
		}

	default:
		return fmt.Errorf("invalid action %q: must be 'move' or 'archive'", action)
	}

	res, err := r.columns.DeleteOne(ctx, bson.M{"_id": columnID})
	if err != nil {
		return fmt.Errorf("delete column: %w", err)
	}
	if res.DeletedCount == 0 {
		return fmt.Errorf("column not found")
	}
	return nil
}

// columnTypeByID returns a map from column hex ID to column type for a board.
func (r *Repository) columnTypeByID(ctx context.Context, boardID primitive.ObjectID) (map[string]kanban.ColumnType, error) {
	cols, err := r.GetColumns(ctx, boardID)
	if err != nil {
		return nil, err
	}
	m := make(map[string]kanban.ColumnType, len(cols))
	for _, c := range cols {
		m[c.ID.Hex()] = c.Type
	}
	return m, nil
}

// GetCards returns all non-archived cards for a board sorted by column_id then order.
func (r *Repository) GetCards(ctx context.Context, boardID primitive.ObjectID) ([]*kanban.Card, error) {
	filter := bson.M{
		"board_id":    boardID,
		"archived_at": bson.M{"$exists": false},
	}
	cur, err := r.cards.Find(ctx, filter, options.Find().SetSort(bson.D{{Key: "column_id", Value: 1}, {Key: "order", Value: 1}}))
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

// GetCard fetches a single card by ID (including archived cards).
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
	count, err := r.cards.CountDocuments(ctx, bson.M{
		"board_id":    boardID,
		"column_id":   columnID,
		"archived_at": bson.M{"$exists": false},
	})
	if err != nil {
		return nil, fmt.Errorf("count cards: %w", err)
	}

	priority := in.Priority
	if !priority.Valid() {
		priority = kanban.PriorityNone
	}

	tags := in.Tags
	if tags == nil {
		tags = []string{}
	}

	var dueDate *time.Time
	if in.DueDate != nil && *in.DueDate != "" {
		t, parseErr := time.Parse(time.DateOnly, *in.DueDate)
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
			t, parseErr := time.Parse(time.DateOnly, *in.DueDate)
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

// MoveCard repositions a card within or across columns, toggling completed_at when the
// card enters or leaves a "done" column.
func (r *Repository) MoveCard(ctx context.Context, cardID primitive.ObjectID, in kanban.MoveCardInput) error {
	colID, err := primitive.ObjectIDFromHex(in.ColumnID)
	if err != nil {
		return fmt.Errorf("invalid column_id: %w", err)
	}

	card, err := r.GetCard(ctx, cardID)
	if err != nil {
		return err
	}

	colTypes, err := r.columnTypeByID(ctx, card.BoardID)
	if err != nil {
		return fmt.Errorf("get column types: %w", err)
	}
	sourceType := colTypes[card.ColumnID.Hex()]
	targetType := colTypes[colID.Hex()]

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

	finalSet := bson.M{"column_id": colID, "order": newOrder}
	finalUpdate := bson.M{}

	if sourceType == kanban.ColumnTypeDone && targetType != kanban.ColumnTypeDone {
		// Moving out of Done: clear completed_at.
		finalUpdate["$set"] = finalSet
		finalUpdate["$unset"] = bson.M{"completed_at": ""}
	} else if targetType == kanban.ColumnTypeDone && card.CompletedAt == nil {
		// Moving into Done: set completed_at.
		finalSet["completed_at"] = time.Now()
		finalUpdate["$set"] = finalSet
	} else {
		finalUpdate["$set"] = finalSet
	}

	_, err = r.cards.UpdateOne(ctx, bson.M{"_id": cardID}, finalUpdate)
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

// ArchiveCard soft-deletes a card. The archive_reason is forced to "completed" when the
// card's column type is "done"; otherwise the caller-supplied reason is used.
func (r *Repository) ArchiveCard(ctx context.Context, cardID primitive.ObjectID, reason kanban.ArchiveReason) (*kanban.Card, error) {
	card, err := r.GetCard(ctx, cardID)
	if err != nil {
		return nil, err
	}

	colTypes, err := r.columnTypeByID(ctx, card.BoardID)
	if err != nil {
		return nil, fmt.Errorf("get column types: %w", err)
	}
	effectiveReason := reason
	if colTypes[card.ColumnID.Hex()] == kanban.ColumnTypeDone {
		effectiveReason = kanban.ArchiveReasonCompleted // system-assigned
	}

	// Shift remaining sibling cards' orders down.
	if _, err := r.cards.UpdateMany(ctx,
		bson.M{
			"column_id":   card.ColumnID,
			"order":       bson.M{"$gt": card.Order},
			"archived_at": bson.M{"$exists": false},
		},
		bson.M{"$inc": bson.M{"order": -1}},
	); err != nil {
		return nil, fmt.Errorf("reorder after archive: %w", err)
	}

	archiveSet := bson.M{
		"archived_at":    time.Now(),
		"archive_reason": effectiveReason,
	}
	// Ensure completed_at is set when archiving from a Done column.
	if effectiveReason == kanban.ArchiveReasonCompleted && card.CompletedAt == nil {
		archiveSet["completed_at"] = time.Now()
	}
	if _, err := r.cards.UpdateOne(ctx, bson.M{"_id": cardID}, bson.M{"$set": archiveSet}); err != nil {
		return nil, fmt.Errorf("archive card: %w", err)
	}

	return r.GetCard(ctx, cardID)
}

// UnarchiveCard restores an archived card to the end of its original column.
// completed_at is set/cleared based on the column type.
func (r *Repository) UnarchiveCard(ctx context.Context, cardID primitive.ObjectID) (*kanban.Card, error) {
	card, err := r.GetCard(ctx, cardID)
	if err != nil {
		return nil, err
	}
	if card.ArchivedAt == nil {
		return nil, fmt.Errorf("card is not archived")
	}

	colTypes, err := r.columnTypeByID(ctx, card.BoardID)
	if err != nil {
		return nil, fmt.Errorf("get column types: %w", err)
	}

	newOrder, err := r.cards.CountDocuments(ctx,
		bson.M{"column_id": card.ColumnID, "archived_at": bson.M{"$exists": false}})
	if err != nil {
		return nil, fmt.Errorf("count column cards: %w", err)
	}

	unset := bson.M{"archived_at": "", "archive_reason": ""}
	set := bson.M{"order": int(newOrder)}

	if colTypes[card.ColumnID.Hex()] == kanban.ColumnTypeDone {
		if card.CompletedAt == nil {
			set["completed_at"] = time.Now()
		}
	} else {
		unset["completed_at"] = ""
	}

	if _, err := r.cards.UpdateOne(ctx, bson.M{"_id": cardID}, bson.M{"$set": set, "$unset": unset}); err != nil {
		return nil, fmt.Errorf("unarchive card: %w", err)
	}

	return r.GetCard(ctx, cardID)
}

// ListArchivedCards returns archived cards for a board.
// filter.Reason: "completed" → archive_reason=completed sorted by completed_at DESC;
//
//	"general" → archive_reason!=completed sorted by archived_at DESC;
//	"" → all archived sorted by archived_at DESC.
//
// filter.Month/Year further restrict by the relevant date field.
func (r *Repository) ListArchivedCards(ctx context.Context, boardID primitive.ObjectID, filter kanban.ListArchivedFilter) ([]*kanban.Card, error) {
	q := bson.M{
		"board_id":    boardID,
		"archived_at": bson.M{"$exists": true},
	}

	sortField := "archived_at"
	switch filter.Reason {
	case string(kanban.ArchiveReasonCompleted):
		q["archive_reason"] = kanban.ArchiveReasonCompleted
		sortField = "completed_at"
	case "general":
		q["archive_reason"] = bson.M{"$ne": kanban.ArchiveReasonCompleted}
	case "":
		// no reason filter
	default:
		q["archive_reason"] = filter.Reason
	}

	if filter.Year > 0 {
		var from, to time.Time
		if filter.Month > 0 {
			from = time.Date(filter.Year, time.Month(filter.Month), 1, 0, 0, 0, 0, time.UTC)
			to = from.AddDate(0, 1, 0)
		} else {
			from = time.Date(filter.Year, 1, 1, 0, 0, 0, 0, time.UTC)
			to = from.AddDate(1, 0, 0)
		}
		q[sortField] = bson.M{"$gte": from, "$lt": to}
	}

	cur, err := r.cards.Find(ctx, q, options.Find().SetSort(bson.M{sortField: -1}))
	if err != nil {
		return nil, fmt.Errorf("find archived cards: %w", err)
	}
	defer cur.Close(ctx)
	var cards []*kanban.Card
	if err := cur.All(ctx, &cards); err != nil {
		return nil, err
	}
	return cards, nil
}

// GetBoardStats computes card statistics for a board (archived cards are excluded).
func (r *Repository) GetBoardStats(ctx context.Context, boardID primitive.ObjectID) (*kanban.BoardStats, error) {
	cols, err := r.GetColumns(ctx, boardID)
	if err != nil {
		return nil, err
	}
	colName := make(map[string]string, len(cols))
	for _, c := range cols {
		colName[c.ID.Hex()] = c.Name
	}

	cards, err := r.GetCards(ctx, boardID) // already excludes archived
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
		stats.ByPriority[string(p)]++

		if card.DueDate == nil {
			stats.NoDueDate++
		} else if card.DueDate.Before(now) {
			stats.Overdue++
		}
	}

	return stats, nil
}
