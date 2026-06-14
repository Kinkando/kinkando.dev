package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/kinkando/personal-dashboard/internal/kanban"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ── Fixed IDs used across tests ───────────────────────────────────────────────

var (
	testBoardID  = primitive.NewObjectID()
	testColumnID = primitive.NewObjectID()
	testCardID   = primitive.NewObjectID()
)

// ── fakeRepo ──────────────────────────────────────────────────────────────────

// fakeRepo implements the handler's Repository interface. Configure only the
// fields relevant to each test; all others return zero/nil.
type fakeRepo struct {
	card    *kanban.Card
	cardErr error

	board    *kanban.Board
	boardErr error

	col    *kanban.Column
	colErr error

	columns    []*kanban.Column
	columnsErr error

	cards    []*kanban.Card
	cardsErr error

	archivedCard *kanban.Card
	archiveErr   error

	archivedCards  []*kanban.Card
	archivedErr    error
	capturedFilter kanban.ListArchivedFilter

	moveCardErr  error
	deleteColErr error
}

func (f *fakeRepo) ListBoards(_ context.Context, _ string) ([]*kanban.Board, error) {
	return nil, nil
}
func (f *fakeRepo) GetBoardByID(_ context.Context, _ primitive.ObjectID, _ string) (*kanban.Board, error) {
	return f.board, f.boardErr
}
func (f *fakeRepo) CreateBoard(_ context.Context, _, _ string) (*kanban.Board, error) {
	return nil, nil
}
func (f *fakeRepo) UpdateBoard(_ context.Context, _ primitive.ObjectID, _ string) error { return nil }
func (f *fakeRepo) DeleteBoard(_ context.Context, _ primitive.ObjectID) error            { return nil }
func (f *fakeRepo) GetColumns(_ context.Context, _ primitive.ObjectID) ([]*kanban.Column, error) {
	return f.columns, f.columnsErr
}
func (f *fakeRepo) GetColumn(_ context.Context, _ primitive.ObjectID) (*kanban.Column, error) {
	return f.col, f.colErr
}
func (f *fakeRepo) CreateColumn(_ context.Context, _ primitive.ObjectID, _ string) (*kanban.Column, error) {
	return nil, nil
}
func (f *fakeRepo) UpdateColumn(_ context.Context, _ primitive.ObjectID, _ string) error { return nil }
func (f *fakeRepo) ReorderColumns(_ context.Context, _ primitive.ObjectID, _ []string) error {
	return nil
}
func (f *fakeRepo) DeleteColumn(_ context.Context, _ primitive.ObjectID, _, _ string) error {
	return f.deleteColErr
}
func (f *fakeRepo) GetCards(_ context.Context, _ primitive.ObjectID) ([]*kanban.Card, error) {
	return f.cards, f.cardsErr
}
func (f *fakeRepo) GetCard(_ context.Context, _ primitive.ObjectID) (*kanban.Card, error) {
	return f.card, f.cardErr
}
func (f *fakeRepo) CreateCard(_ context.Context, _, _ primitive.ObjectID, _ kanban.CreateCardInput) (*kanban.Card, error) {
	return nil, nil
}
func (f *fakeRepo) UpdateCard(_ context.Context, _ primitive.ObjectID, _ kanban.UpdateCardInput) (*kanban.Card, error) {
	return nil, nil
}
func (f *fakeRepo) MoveCard(_ context.Context, _ primitive.ObjectID, _ kanban.MoveCardInput) error {
	return f.moveCardErr
}
func (f *fakeRepo) DeleteCard(_ context.Context, _ primitive.ObjectID) error { return nil }
func (f *fakeRepo) ArchiveCard(_ context.Context, _ primitive.ObjectID, _ kanban.ArchiveReason) (*kanban.Card, error) {
	return f.archivedCard, f.archiveErr
}
func (f *fakeRepo) UnarchiveCard(_ context.Context, _ primitive.ObjectID) (*kanban.Card, error) {
	return nil, nil
}
func (f *fakeRepo) ListArchivedCards(_ context.Context, _ primitive.ObjectID, filter kanban.ListArchivedFilter) ([]*kanban.Card, error) {
	f.capturedFilter = filter
	return f.archivedCards, f.archivedErr
}
func (f *fakeRepo) GetBoardStats(_ context.Context, _ primitive.ObjectID) (*kanban.BoardStats, error) {
	return nil, nil
}
func (f *fakeRepo) AddAttachment(_ context.Context, _ primitive.ObjectID, _ kanban.AddAttachmentInput) (*kanban.Attachment, error) {
	return nil, nil
}
func (f *fakeRepo) RemoveAttachment(_ context.Context, _, _ primitive.ObjectID) (*kanban.Attachment, error) {
	return nil, nil
}

// ── Harness ───────────────────────────────────────────────────────────────────

// newTestApp wires a fresh Fiber app with the kanban handler routes.
// No auth middleware is needed: auth.GetUserID reads c.Locals, which returns ""
// in tests; the fakeRepo ignores the userID value.
func newTestApp(repo Repository) *fiber.App {
	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	New(repo, fakeStorage{}).Register(app)
	return app
}

// fakeStorage is a no-op Storage implementation for handler tests that don't
// exercise the attachment endpoints. It satisfies the Storage interface so the
// handler can be constructed without reaching for the network.
type fakeStorage struct{}

func (fakeStorage) Upload(_ context.Context, _, _ string, _ io.Reader) (string, error) {
	return "", nil
}
func (fakeStorage) Delete(_ context.Context, _ string) error { return nil }

// readBody reads the HTTP response body and unmarshals it into a
// map[string]interface{} for assertion.
func readBody(t *testing.T, resp *http.Response) map[string]interface{} {
	t.Helper()
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	var out map[string]interface{}
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatalf("unmarshal body %q: %v", b, err)
	}
	return out
}

// ── Tests ─────────────────────────────────────────────────────────────────────

// archiveCard: sending reason="completed" must be rejected before validation.
func TestArchiveCard_CompletedReasonRejected(t *testing.T) {
	repo := &fakeRepo{
		card:  &kanban.Card{ID: testCardID, BoardID: testBoardID},
		board: &kanban.Board{ID: testBoardID},
	}
	app := newTestApp(repo)

	req := httptest.NewRequest(http.MethodPatch, "/cards/"+testCardID.Hex()+"/archive",
		strings.NewReader(`{"reason":"completed"}`))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", resp.StatusCode)
	}
	body := readBody(t, resp)
	msg, _ := body["error"].(string)
	if !strings.Contains(msg, "reserved") {
		t.Errorf("error message %q should mention 'reserved'", msg)
	}
}

// deleteColumn: system columns must be rejected with 409 before body is parsed.
func TestDeleteColumn_SystemColumnReturns409(t *testing.T) {
	repo := &fakeRepo{
		col:   &kanban.Column{ID: testColumnID, BoardID: testBoardID, IsSystem: true},
		board: &kanban.Board{ID: testBoardID},
	}
	app := newTestApp(repo)

	req := httptest.NewRequest(http.MethodDelete, "/columns/"+testColumnID.Hex(), nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusConflict {
		t.Errorf("status = %d, want 409", resp.StatusCode)
	}
	body := readBody(t, resp)
	msg, _ := body["error"].(string)
	if !strings.Contains(msg, "system") {
		t.Errorf("error message %q should mention 'system'", msg)
	}
}

// deleteColumn: repo error strings are mapped to the correct HTTP status codes.
func TestDeleteColumn_RepoErrorMapping(t *testing.T) {
	cases := []struct {
		repoErr    string
		wantStatus int
	}{
		{"system column cannot be deleted", http.StatusConflict},
		{"invalid action value", http.StatusBadRequest},
		{"column not found", http.StatusNotFound},
	}
	for _, tc := range cases {
		t.Run(tc.repoErr, func(t *testing.T) {
			repo := &fakeRepo{
				col:          &kanban.Column{ID: testColumnID, BoardID: testBoardID},
				board:        &kanban.Board{ID: testBoardID},
				deleteColErr: errors.New(tc.repoErr),
			}
			app := newTestApp(repo)

			// action=archive needs no target_column_id, so validation passes.
			req := httptest.NewRequest(http.MethodDelete, "/columns/"+testColumnID.Hex(),
				strings.NewReader(`{"action":"archive"}`))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("app.Test: %v", err)
			}
			if resp.StatusCode != tc.wantStatus {
				t.Errorf("repoErr=%q: status = %d, want %d", tc.repoErr, resp.StatusCode, tc.wantStatus)
			}
		})
	}
}

// listArchive: valid month/year query params populate the filter; invalid values
// are silently ignored (filter stays zero).
func TestListArchive_FilterParsing(t *testing.T) {
	t.Run("valid month and year", func(t *testing.T) {
		repo := &fakeRepo{board: &kanban.Board{ID: testBoardID}}
		app := newTestApp(repo)

		req := httptest.NewRequest(http.MethodGet,
			"/boards/"+testBoardID.Hex()+"/archive?month=3&year=2025", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("app.Test: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Errorf("status = %d, want 200", resp.StatusCode)
		}
		resp.Body.Close()
		if repo.capturedFilter.Month != 3 {
			t.Errorf("filter.Month = %d, want 3", repo.capturedFilter.Month)
		}
		if repo.capturedFilter.Year != 2025 {
			t.Errorf("filter.Year = %d, want 2025", repo.capturedFilter.Year)
		}
	})

	t.Run("invalid month and year are ignored", func(t *testing.T) {
		repo := &fakeRepo{board: &kanban.Board{ID: testBoardID}}
		app := newTestApp(repo)

		req := httptest.NewRequest(http.MethodGet,
			"/boards/"+testBoardID.Hex()+"/archive?month=13&year=0", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("app.Test: %v", err)
		}
		resp.Body.Close()
		if repo.capturedFilter.Month != 0 {
			t.Errorf("filter.Month = %d after invalid month=13, want 0", repo.capturedFilter.Month)
		}
		if repo.capturedFilter.Year != 0 {
			t.Errorf("filter.Year = %d after year=0, want 0", repo.capturedFilter.Year)
		}
	})
}

// getBoard: nil Tags become [], empty Priority becomes "none", empty column Type
// becomes "custom" — legacy documents are normalised before the response.
func TestGetBoard_LegacyNormalization(t *testing.T) {
	repo := &fakeRepo{
		board: &kanban.Board{ID: testBoardID},
		columns: []*kanban.Column{
			{ID: testColumnID, BoardID: testBoardID, Type: ""}, // empty type → custom
		},
		cards: []*kanban.Card{
			{ID: testCardID, BoardID: testBoardID, Tags: nil, Priority: ""}, // nil tags, empty priority
		},
	}
	app := newTestApp(repo)

	req := httptest.NewRequest(http.MethodGet, "/boards/"+testBoardID.Hex(), nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}

	body := readBody(t, resp)
	data, _ := body["data"].(map[string]interface{})

	// Column type must be defaulted to "custom".
	cols, _ := data["columns"].([]interface{})
	if len(cols) != 1 {
		t.Fatalf("columns len = %d, want 1", len(cols))
	}
	col := cols[0].(map[string]interface{})
	if col["type"] != string(kanban.ColumnTypeCustom) {
		t.Errorf("column.type = %q, want %q", col["type"], kanban.ColumnTypeCustom)
	}

	// Card priority must be defaulted to "none"; tags must be [].
	cards, _ := data["cards"].([]interface{})
	if len(cards) != 1 {
		t.Fatalf("cards len = %d, want 1", len(cards))
	}
	card := cards[0].(map[string]interface{})
	if card["priority"] != string(kanban.PriorityNone) {
		t.Errorf("card.priority = %q, want %q", card["priority"], kanban.PriorityNone)
	}
	tags, _ := card["tags"].([]interface{})
	if tags == nil || len(tags) != 0 {
		t.Errorf("card.tags = %v, want []", tags)
	}
}

// moveCard: a repo error containing "invalid column_id" must map to 400.
func TestMoveCard_InvalidColumnIDReturns400(t *testing.T) {
	repo := &fakeRepo{
		card:        &kanban.Card{ID: testCardID, BoardID: testBoardID},
		board:       &kanban.Board{ID: testBoardID},
		moveCardErr: errors.New("invalid column_id: not found in board"),
	}
	app := newTestApp(repo)

	// Any non-empty column_id hex string + order>=0 passes validation.
	body := `{"column_id":"` + testColumnID.Hex() + `","order":0}`
	req := httptest.NewRequest(http.MethodPatch, "/cards/"+testCardID.Hex()+"/move",
		strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", resp.StatusCode)
	}
}
