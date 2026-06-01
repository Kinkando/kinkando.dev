package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/go-jet/jet/v2/postgres"
	"github.com/google/uuid"
	"github.com/kinkando/personal-dashboard/gen/kinkando/public/model"
	"github.com/kinkando/personal-dashboard/gen/kinkando/public/table"
	"github.com/kinkando/personal-dashboard/internal/finance"
	"github.com/shopspring/decimal"
)

type Repository struct {
	db *sql.DB
}

func New(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// ── Category CRUD ─────────────────────────────────────────────────────────────

func (r *Repository) CreateCategory(ctx context.Context, userID uuid.UUID, in finance.CreateCategoryInput) (*finance.Category, error) {
	stmt := table.FinanceCategories.INSERT(
		table.FinanceCategories.UserID,
		table.FinanceCategories.Name,
		table.FinanceCategories.Type,
		table.FinanceCategories.Icon,
		table.FinanceCategories.Color,
	).VALUES(
		postgres.UUID(userID),
		in.Name,
		string(in.Type),
		in.Icon,
		in.Color,
	).RETURNING(table.FinanceCategories.AllColumns)

	var dest model.FinanceCategories
	if err := stmt.QueryContext(ctx, r.db, &dest); err != nil {
		return nil, fmt.Errorf("create category: %w", err)
	}
	return toCategory(dest), nil
}

func (r *Repository) ListCategories(ctx context.Context, userID uuid.UUID) ([]*finance.Category, error) {
	stmt := postgres.SELECT(table.FinanceCategories.AllColumns).
		FROM(table.FinanceCategories).
		WHERE(table.FinanceCategories.UserID.EQ(postgres.UUID(userID))).
		ORDER_BY(table.FinanceCategories.Type.ASC(), table.FinanceCategories.Name.ASC())

	var dest []model.FinanceCategories
	if err := stmt.QueryContext(ctx, r.db, &dest); err != nil {
		return nil, fmt.Errorf("list categories: %w", err)
	}
	cats := make([]*finance.Category, len(dest))
	for i, d := range dest {
		cats[i] = toCategory(d)
	}
	return cats, nil
}

func (r *Repository) GetCategory(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*finance.Category, error) {
	stmt := postgres.SELECT(table.FinanceCategories.AllColumns).
		FROM(table.FinanceCategories).
		WHERE(
			table.FinanceCategories.ID.EQ(postgres.UUID(id)).
				AND(table.FinanceCategories.UserID.EQ(postgres.UUID(userID))),
		)

	var dest model.FinanceCategories
	if err := stmt.QueryContext(ctx, r.db, &dest); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("category not found")
		}
		return nil, fmt.Errorf("get category: %w", err)
	}
	return toCategory(dest), nil
}

func (r *Repository) UpdateCategory(ctx context.Context, id uuid.UUID, userID uuid.UUID, in finance.UpdateCategoryInput) (*finance.Category, error) {
	stmt := table.FinanceCategories.UPDATE(
		table.FinanceCategories.Name,
		table.FinanceCategories.Icon,
		table.FinanceCategories.Color,
	).SET(
		in.Name,
		in.Icon,
		in.Color,
	).WHERE(
		table.FinanceCategories.ID.EQ(postgres.UUID(id)).
			AND(table.FinanceCategories.UserID.EQ(postgres.UUID(userID))),
	).RETURNING(table.FinanceCategories.AllColumns)

	var dest model.FinanceCategories
	if err := stmt.QueryContext(ctx, r.db, &dest); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("category not found")
		}
		return nil, fmt.Errorf("update category: %w", err)
	}
	return toCategory(dest), nil
}

func (r *Repository) DeleteCategory(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	stmt := table.FinanceCategories.DELETE().WHERE(
		table.FinanceCategories.ID.EQ(postgres.UUID(id)).
			AND(table.FinanceCategories.UserID.EQ(postgres.UUID(userID))),
	)
	res, err := stmt.ExecContext(ctx, r.db)
	if err != nil {
		return fmt.Errorf("delete category: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil || n == 0 {
		return fmt.Errorf("category not found")
	}
	return nil
}

// ── Record CRUD ───────────────────────────────────────────────────────────────

func (r *Repository) Create(ctx context.Context, userID uuid.UUID, in finance.CreateRecordInput) (*finance.Record, error) {
	date, err := time.Parse("2006-01-02", in.Date)
	if err != nil {
		return nil, fmt.Errorf("invalid date format: %w", err)
	}

	cat, err := r.GetCategory(ctx, in.CategoryID, userID)
	if err != nil {
		return nil, fmt.Errorf("category: %w", err)
	}
	if cat.Type != in.Type {
		return nil, fmt.Errorf("category type %q does not match record type %q", cat.Type, in.Type)
	}

	stmt := table.FinanceRecords.INSERT(
		table.FinanceRecords.UserID,
		table.FinanceRecords.Type,
		table.FinanceRecords.Amount,
		table.FinanceRecords.CategoryID,
		table.FinanceRecords.Note,
		table.FinanceRecords.Date,
	).VALUES(
		postgres.UUID(userID),
		string(in.Type),
		decimal.NewFromFloat(in.Amount),
		cat.Name,
		postgres.UUID(in.CategoryID),
		in.Note,
		date,
	).RETURNING(table.FinanceRecords.AllColumns)

	var dest model.FinanceRecords
	if err := stmt.QueryContext(ctx, r.db, &dest); err != nil {
		return nil, fmt.Errorf("create record: %w", err)
	}
	rec := toRecord(dest)
	rec.Category = &finance.CategoryRef{
		ID:    cat.ID,
		Name:  cat.Name,
		Icon:  cat.Icon,
		Color: cat.Color,
	}
	return rec, nil
}

func (r *Repository) List(ctx context.Context, userID uuid.UUID, month string) ([]*finance.Record, error) {
	start, end, err := monthRange(month)
	if err != nil {
		return nil, err
	}

	stmt := postgres.SELECT(
		table.FinanceRecords.AllColumns,
		table.FinanceCategories.AllColumns,
	).FROM(
		table.FinanceRecords.LEFT_JOIN(
			table.FinanceCategories,
			table.FinanceCategories.ID.EQ(table.FinanceRecords.CategoryID),
		),
	).WHERE(
		table.FinanceRecords.UserID.EQ(postgres.UUID(userID)).
			AND(table.FinanceRecords.Date.GT_EQ(postgres.DateT(start))).
			AND(table.FinanceRecords.Date.LT(postgres.DateT(end))),
	).ORDER_BY(
		table.FinanceRecords.Date.DESC(),
		table.FinanceRecords.CreatedAt.DESC(),
	)

	type listRow struct {
		model.FinanceRecords
		Category *model.FinanceCategories
	}

	var dest []listRow
	if err := stmt.QueryContext(ctx, r.db, &dest); err != nil {
		return nil, fmt.Errorf("list records: %w", err)
	}

	var records []*finance.Record
	for _, d := range dest {
		rec := toRecord(d.FinanceRecords)
		if d.Category != nil {
			rec.Category = &finance.CategoryRef{
				ID:    d.Category.ID,
				Name:  d.Category.Name,
				Icon:  d.Category.Icon,
				Color: d.Category.Color,
			}
		}
		records = append(records, rec)
	}
	return records, nil
}

func (r *Repository) MonthlySummary(ctx context.Context, userID uuid.UUID, month string) (*finance.MonthlySummary, error) {
	start, end, err := monthRange(month)
	if err != nil {
		return nil, err
	}

	catName := postgres.COALESCE(table.FinanceCategories.Name, postgres.String(""))
	catIcon := postgres.COALESCE(table.FinanceCategories.Icon, postgres.String(""))
	catColor := postgres.COALESCE(table.FinanceCategories.Color, postgres.String(""))
	sumAmount := postgres.SUM(table.FinanceRecords.Amount)

	stmt := postgres.SELECT(
		table.FinanceRecords.CategoryID.AS("cat_id"),
		catName.AS("cat_name"),
		catIcon.AS("cat_icon"),
		catColor.AS("cat_color"),
		table.FinanceRecords.Type.AS("rec_type"),
		sumAmount.AS("total"),
	).FROM(
		table.FinanceRecords.LEFT_JOIN(
			table.FinanceCategories,
			table.FinanceCategories.ID.EQ(table.FinanceRecords.CategoryID),
		),
	).WHERE(
		table.FinanceRecords.UserID.EQ(postgres.UUID(userID)).
			AND(table.FinanceRecords.Date.GT_EQ(postgres.DateT(start))).
			AND(table.FinanceRecords.Date.LT(postgres.DateT(end))),
	).GROUP_BY(
		table.FinanceRecords.CategoryID,
		catName,
		table.FinanceRecords.Type,
		catIcon,
		catColor,
	).ORDER_BY(
		sumAmount.DESC(),
	)

	var rows []struct {
		CategoryID *uuid.UUID      `alias:"cat_id"`
		CatName    string          `alias:"cat_name"`
		CatIcon    string          `alias:"cat_icon"`
		CatColor   string          `alias:"cat_color"`
		Type       string          `alias:"rec_type"`
		Total      decimal.Decimal `alias:"total"`
	}
	if err := stmt.QueryContext(ctx, r.db, &rows); err != nil {
		return nil, fmt.Errorf("monthly summary: %w", err)
	}

	summary := &finance.MonthlySummary{
		Month:      month,
		Categories: []finance.CategorySummary{},
	}
	for _, row := range rows {
		t, _ := row.Total.Float64()
		switch row.Type {
		case string(finance.RecordTypeIncome):
			summary.Income += t
		case string(finance.RecordTypeExpense):
			summary.Expense += t
		}
		summary.Categories = append(summary.Categories, finance.CategorySummary{
			CategoryID: row.CategoryID,
			Category:   row.CatName,
			Type:       finance.RecordType(row.Type),
			Total:      t,
			Icon:       row.CatIcon,
			Color:      row.CatColor,
		})
	}
	summary.Net = summary.Income - summary.Expense
	return summary, nil
}

func (r *Repository) Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	stmt := table.FinanceRecords.DELETE().WHERE(
		table.FinanceRecords.ID.EQ(postgres.UUID(id)).
			AND(table.FinanceRecords.UserID.EQ(postgres.UUID(userID))),
	)
	res, err := stmt.ExecContext(ctx, r.db)
	if err != nil {
		return fmt.Errorf("delete record: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil || n == 0 {
		return fmt.Errorf("record not found")
	}
	return nil
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func monthRange(month string) (time.Time, time.Time, error) {
	t, err := time.Parse("2006-01", month)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid month %q (expected YYYY-MM): %w", month, err)
	}
	start := time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
	return start, start.AddDate(0, 1, 0), nil
}

func toRecord(m model.FinanceRecords) *finance.Record {
	amount, _ := m.Amount.Float64()
	return &finance.Record{
		ID:         m.ID,
		UserID:     m.UserID,
		Type:       finance.RecordType(m.Type),
		Amount:     amount,
		CategoryID: m.CategoryID,
		Note:       m.Note,
		Date:       m.Date,
		CreatedAt:  m.CreatedAt,
	}
}

func toCategory(m model.FinanceCategories) *finance.Category {
	return &finance.Category{
		ID:        m.ID,
		UserID:    m.UserID,
		Name:      m.Name,
		Type:      finance.RecordType(m.Type),
		Icon:      m.Icon,
		Color:     m.Color,
		CreatedAt: m.CreatedAt,
	}
}
