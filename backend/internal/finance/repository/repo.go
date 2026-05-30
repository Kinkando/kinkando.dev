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

// ── Create ────────────────────────────────────────────────────────────────────

func (r *Repository) Create(ctx context.Context, userID uuid.UUID, in finance.CreateRecordInput) (*finance.Record, error) {
	date, err := time.Parse("2006-01-02", in.Date)
	if err != nil {
		return nil, fmt.Errorf("invalid date format: %w", err)
	}

	stmt := table.FinanceRecords.INSERT(
		table.FinanceRecords.UserID,
		table.FinanceRecords.Type,
		table.FinanceRecords.Amount,
		table.FinanceRecords.Category,
		table.FinanceRecords.Note,
		table.FinanceRecords.Date,
	).VALUES(
		userID.String(),
		string(in.Type),
		decimal.NewFromFloat(in.Amount),
		in.Category,
		in.Note,
		date,
	).RETURNING(table.FinanceRecords.AllColumns)

	var dest model.FinanceRecords
	if err := stmt.QueryContext(ctx, r.db, &dest); err != nil {
		return nil, fmt.Errorf("create record: %w", err)
	}
	return toRecord(dest), nil
}

// ── List ─────────────────────────────────────────────────────────────────────

func (r *Repository) List(ctx context.Context, userID uuid.UUID, month string) ([]*finance.Record, error) {
	start, end, err := monthRange(month)
	if err != nil {
		return nil, err
	}

	stmt := postgres.SELECT(table.FinanceRecords.AllColumns).
		FROM(table.FinanceRecords).
		WHERE(
			table.FinanceRecords.UserID.EQ(postgres.String(userID.String())).
				AND(table.FinanceRecords.Date.GT_EQ(postgres.Date(start.Year(), start.Month(), start.Day()))).
				AND(table.FinanceRecords.Date.LT(postgres.Date(end.Year(), end.Month(), end.Day()))),
		).
		ORDER_BY(table.FinanceRecords.Date.DESC(), table.FinanceRecords.CreatedAt.DESC())

	var dest []model.FinanceRecords
	if err := stmt.QueryContext(ctx, r.db, &dest); err != nil {
		return nil, fmt.Errorf("list records: %w", err)
	}

	records := make([]*finance.Record, len(dest))
	for i, d := range dest {
		records[i] = toRecord(d)
	}
	return records, nil
}

// ── MonthlySummary ────────────────────────────────────────────────────────────

func (r *Repository) MonthlySummary(ctx context.Context, userID uuid.UUID, month string) (*finance.MonthlySummary, error) {
	start, end, err := monthRange(month)
	if err != nil {
		return nil, err
	}

	whereClause := table.FinanceRecords.UserID.EQ(postgres.String(userID.String())).
		AND(table.FinanceRecords.Date.GT_EQ(postgres.Date(start.Year(), start.Month(), start.Day()))).
		AND(table.FinanceRecords.Date.LT(postgres.Date(end.Year(), end.Month(), end.Day())))

	// One query: GROUP BY (category, type) gives us both the per-category breakdown
	// and, by summing those rows in Go, the overall income / expense totals.
	type categoryRow struct {
		Category string
		Type     string
		Total    decimal.Decimal
	}

	stmt := postgres.SELECT(
		table.FinanceRecords.Category,
		table.FinanceRecords.Type,
		postgres.SUM(table.FinanceRecords.Amount).AS("total"),
	).FROM(table.FinanceRecords).
		WHERE(whereClause).
		GROUP_BY(table.FinanceRecords.Category, table.FinanceRecords.Type).
		ORDER_BY(postgres.SUM(table.FinanceRecords.Amount).DESC())

	var rows []categoryRow
	if err := stmt.QueryContext(ctx, r.db, &rows); err != nil {
		return nil, fmt.Errorf("monthly summary: %w", err)
	}

	summary := &finance.MonthlySummary{
		Month:      month,
		Categories: []finance.CategorySummary{},
	}
	for _, row := range rows {
		total, _ := row.Total.Float64()
		switch row.Type {
		case string(finance.RecordTypeIncome):
			summary.Income += total
		case string(finance.RecordTypeExpense):
			summary.Expense += total
		}
		summary.Categories = append(summary.Categories, finance.CategorySummary{
			Category: row.Category,
			Type:     finance.RecordType(row.Type),
			Total:    total,
		})
	}
	summary.Net = summary.Income - summary.Expense

	return summary, nil
}

// ── Delete ────────────────────────────────────────────────────────────────────

func (r *Repository) Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	stmt := table.FinanceRecords.DELETE().WHERE(
		table.FinanceRecords.ID.EQ(postgres.String(id.String())).
			AND(table.FinanceRecords.UserID.EQ(postgres.String(userID.String()))),
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

// monthRange returns the [start, end) half-open interval for a "YYYY-MM" string.
func monthRange(month string) (time.Time, time.Time, error) {
	t, err := time.Parse("2006-01", month)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid month %q (expected YYYY-MM): %w", month, err)
	}
	start := time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
	return start, start.AddDate(0, 1, 0), nil
}

// toRecord maps the jet-generated DB model to the domain Record type.
func toRecord(m model.FinanceRecords) *finance.Record {
	amount, _ := m.Amount.Float64()
	return &finance.Record{
		ID:        m.ID,
		UserID:    m.UserID,
		Type:      finance.RecordType(m.Type),
		Amount:    amount,
		Category:  m.Category,
		Note:      m.Note,
		Date:      m.Date,
		CreatedAt: m.CreatedAt,
	}
}
