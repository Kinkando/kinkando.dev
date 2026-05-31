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
		table.FinanceRecords.Category,
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

	const q = `
		SELECT
			r.id::text, r.user_id::text, r.type, r.amount, r.category, r.note, r.date, r.created_at,
			r.category_id::text,
			c.id::text AS cat_id, c.name AS cat_name, c.icon AS cat_icon, c.color AS cat_color
		FROM finance_records r
		LEFT JOIN finance_categories c ON c.id = r.category_id
		WHERE r.user_id = $1::uuid AND r.date >= $2 AND r.date < $3
		ORDER BY r.date DESC, r.created_at DESC`

	rows, err := r.db.QueryContext(ctx, q, userID.String(), start, end)
	if err != nil {
		return nil, fmt.Errorf("list records: %w", err)
	}
	defer rows.Close()

	var records []*finance.Record
	for rows.Next() {
		var (
			idStr       string
			userIDStr   string
			typeStr     string
			amount      decimal.Decimal
			catName     string
			note        string
			date        time.Time
			createdAt   time.Time
			catIDStr    *string
			catRefIDStr *string
			catRefName  *string
			catRefIcon  *string
			catRefColor *string
		)
		if err := rows.Scan(
			&idStr, &userIDStr, &typeStr, &amount, &catName, &note, &date, &createdAt,
			&catIDStr,
			&catRefIDStr, &catRefName, &catRefIcon, &catRefColor,
		); err != nil {
			return nil, fmt.Errorf("scan record: %w", err)
		}
		rec := &finance.Record{
			ID:           uuid.MustParse(idStr),
			UserID:       uuid.MustParse(userIDStr),
			Type:         finance.RecordType(typeStr),
			CategoryName: catName,
			Note:         note,
			Date:         date,
			CreatedAt:    createdAt,
		}
		rec.Amount, _ = amount.Float64()
		if catIDStr != nil {
			id := uuid.MustParse(*catIDStr)
			rec.CategoryID = &id
		}
		if catRefIDStr != nil && catRefName != nil {
			id := uuid.MustParse(*catRefIDStr)
			rec.Category = &finance.CategoryRef{
				ID:    id,
				Name:  *catRefName,
				Icon:  derefStr(catRefIcon),
				Color: derefStr(catRefColor),
			}
		}
		records = append(records, rec)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate records: %w", err)
	}
	return records, nil
}

func (r *Repository) MonthlySummary(ctx context.Context, userID uuid.UUID, month string) (*finance.MonthlySummary, error) {
	start, end, err := monthRange(month)
	if err != nil {
		return nil, err
	}

	const q = `
		SELECT
			r.category_id::text,
			COALESCE(c.name, r.category)  AS cat_name,
			COALESCE(c.icon, '')           AS cat_icon,
			COALESCE(c.color, '')          AS cat_color,
			r.type,
			SUM(r.amount)                  AS total
		FROM finance_records r
		LEFT JOIN finance_categories c ON c.id = r.category_id
		WHERE r.user_id = $1::uuid AND r.date >= $2 AND r.date < $3
		GROUP BY r.category_id, COALESCE(c.name, r.category), r.type,
		         COALESCE(c.icon, ''), COALESCE(c.color, '')
		ORDER BY SUM(r.amount) DESC`

	rows, err := r.db.QueryContext(ctx, q, userID.String(), start, end)
	if err != nil {
		return nil, fmt.Errorf("monthly summary: %w", err)
	}
	defer rows.Close()

	summary := &finance.MonthlySummary{
		Month:      month,
		Categories: []finance.CategorySummary{},
	}
	for rows.Next() {
		var (
			catIDStr string
			catName  string
			catIcon  string
			catColor string
			recType  string
			total    decimal.Decimal
		)
		if err := rows.Scan(&catIDStr, &catName, &catIcon, &catColor, &recType, &total); err != nil {
			return nil, fmt.Errorf("scan summary row: %w", err)
		}
		t, _ := total.Float64()
		switch recType {
		case string(finance.RecordTypeIncome):
			summary.Income += t
		case string(finance.RecordTypeExpense):
			summary.Expense += t
		}
		var catID *uuid.UUID
		if catIDStr != "" {
			id := uuid.MustParse(catIDStr)
			catID = &id
		}
		summary.Categories = append(summary.Categories, finance.CategorySummary{
			CategoryID: catID,
			Category:   catName,
			Type:       finance.RecordType(recType),
			Total:      t,
			Icon:       catIcon,
			Color:      catColor,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate summary rows: %w", err)
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
		ID:           m.ID,
		UserID:       m.UserID,
		Type:         finance.RecordType(m.Type),
		Amount:       amount,
		CategoryID:   m.CategoryID,
		CategoryName: m.Category,
		Note:         m.Note,
		Date:         m.Date,
		CreatedAt:    m.CreatedAt,
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

func derefStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
