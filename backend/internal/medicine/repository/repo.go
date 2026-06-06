package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-jet/jet/v2/postgres"
	"github.com/go-jet/jet/v2/qrm"
	"github.com/google/uuid"
	"github.com/kinkando/personal-dashboard/gen/kinkando/public/model"
	"github.com/kinkando/personal-dashboard/gen/kinkando/public/table"
	"github.com/kinkando/personal-dashboard/internal/medicine"
	"github.com/shopspring/decimal"
)

type Repository struct {
	db *sql.DB
}

func New(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// ── Medicines ─────────────────────────────────────────────────────────────────

func (r *Repository) ListMedicines(ctx context.Context, userID uuid.UUID, includeArchived bool, sourceType *medicine.SourceType) ([]*medicine.Medicine, error) {
	cond := table.Medicines.UserID.EQ(postgres.UUID(userID))
	if !includeArchived {
		cond = cond.AND(table.Medicines.ArchivedAt.IS_NULL())
	}
	if sourceType != nil {
		cond = cond.AND(table.Medicines.SourceType.EQ(postgres.String(string(*sourceType))))
	}

	stmt := postgres.SELECT(table.Medicines.AllColumns).
		FROM(table.Medicines).
		WHERE(cond).
		ORDER_BY(table.Medicines.CreatedAt.ASC())

	var dest []model.Medicines
	if err := stmt.QueryContext(ctx, r.db, &dest); err != nil {
		return nil, fmt.Errorf("list medicines: %w", err)
	}
	meds := make([]*medicine.Medicine, len(dest))
	for i, d := range dest {
		meds[i] = toMedicine(d)
	}
	return meds, nil
}

func (r *Repository) CreateMedicine(ctx context.Context, userID uuid.UUID, in medicine.CreateMedicineInput) (*medicine.Medicine, error) {
	dosageAmount := decimal.NewFromFloat(in.DosageAmount)
	stockQuantity := decimal.NewFromFloat(in.StockQuantity)

	threshold := decimal.NewFromFloat(7)
	if in.LowStockThreshold != nil {
		threshold = decimal.NewFromFloat(*in.LowStockThreshold)
	}

	var frequencyValue *int32
	if in.FrequencyValue != nil {
		v := int32(*in.FrequencyValue)
		frequencyValue = &v
	}

	var timing *string
	if in.Timing != nil {
		t := string(*in.Timing)
		timing = &t
	}

	var startDate, endDate *time.Time
	if in.StartDate != "" {
		t, err := time.Parse(time.DateOnly, in.StartDate)
		if err != nil {
			return nil, fmt.Errorf("invalid start_date format: %w", err)
		}
		startDate = &t
	}
	if in.EndDate != "" {
		t, err := time.Parse(time.DateOnly, in.EndDate)
		if err != nil {
			return nil, fmt.Errorf("invalid end_date format: %w", err)
		}
		endDate = &t
	}

	reminderTimes, err := marshalReminderTimes(in.ReminderTimes)
	if err != nil {
		return nil, fmt.Errorf("marshal reminder_times: %w", err)
	}

	cols := table.Medicines.INSERT(
		table.Medicines.UserID,
		table.Medicines.Name,
		table.Medicines.GenericName,
		table.Medicines.Description,
		table.Medicines.StockQuantity,
		table.Medicines.StockUnit,
		table.Medicines.DosageAmount,
		table.Medicines.DosageUnit,
		table.Medicines.FrequencyType,
		table.Medicines.FrequencyValue,
		table.Medicines.Timing,
		table.Medicines.StartDate,
		table.Medicines.EndDate,
		table.Medicines.LowStockThreshold,
		table.Medicines.Note,
		table.Medicines.SourceType,
		table.Medicines.ReminderEnabled,
		table.Medicines.ReminderTimes,
	).VALUES(
		postgres.UUID(userID),
		in.Name,
		in.GenericName,
		in.Description,
		stockQuantity,
		in.StockUnit,
		dosageAmount,
		in.DosageUnit,
		string(in.FrequencyType),
		frequencyValue,
		timing,
		startDateExpr(startDate),
		endDateExpr(endDate),
		threshold,
		in.Note,
		string(in.SourceType),
		in.ReminderEnabled,
		reminderTimes,
	).RETURNING(table.Medicines.AllColumns)

	var dest model.Medicines
	if err := cols.QueryContext(ctx, r.db, &dest); err != nil {
		return nil, fmt.Errorf("create medicine: %w", err)
	}
	return toMedicine(dest), nil
}

func (r *Repository) UpdateMedicine(ctx context.Context, id uuid.UUID, userID uuid.UUID, in medicine.UpdateMedicineInput) (*medicine.Medicine, error) {
	dosageAmount := decimal.NewFromFloat(in.DosageAmount)
	stockQuantity := decimal.NewFromFloat(in.StockQuantity)

	threshold := decimal.NewFromFloat(7)
	if in.LowStockThreshold != nil {
		threshold = decimal.NewFromFloat(*in.LowStockThreshold)
	}

	var frequencyValue *int32
	if in.FrequencyValue != nil {
		v := int32(*in.FrequencyValue)
		frequencyValue = &v
	}

	var timing *string
	if in.Timing != nil {
		t := string(*in.Timing)
		timing = &t
	}

	var startDate, endDate *time.Time
	if in.StartDate != "" {
		t, err := time.Parse(time.DateOnly, in.StartDate)
		if err != nil {
			return nil, fmt.Errorf("invalid start_date format: %w", err)
		}
		startDate = &t
	}
	if in.EndDate != "" {
		t, err := time.Parse(time.DateOnly, in.EndDate)
		if err != nil {
			return nil, fmt.Errorf("invalid end_date format: %w", err)
		}
		endDate = &t
	}

	reminderTimes, err := marshalReminderTimes(in.ReminderTimes)
	if err != nil {
		return nil, fmt.Errorf("marshal reminder_times: %w", err)
	}

	stmt := table.Medicines.UPDATE(
		table.Medicines.Name,
		table.Medicines.GenericName,
		table.Medicines.Description,
		table.Medicines.StockQuantity,
		table.Medicines.StockUnit,
		table.Medicines.DosageAmount,
		table.Medicines.DosageUnit,
		table.Medicines.FrequencyType,
		table.Medicines.FrequencyValue,
		table.Medicines.Timing,
		table.Medicines.StartDate,
		table.Medicines.EndDate,
		table.Medicines.LowStockThreshold,
		table.Medicines.Note,
		table.Medicines.SourceType,
		table.Medicines.ReminderEnabled,
		table.Medicines.ReminderTimes,
		table.Medicines.UpdatedAt,
	).SET(
		in.Name,
		in.GenericName,
		in.Description,
		stockQuantity,
		in.StockUnit,
		dosageAmount,
		in.DosageUnit,
		string(in.FrequencyType),
		frequencyValue,
		timing,
		startDateExpr(startDate),
		endDateExpr(endDate),
		threshold,
		in.Note,
		string(in.SourceType),
		in.ReminderEnabled,
		reminderTimes,
		postgres.NOW(),
	).WHERE(
		table.Medicines.ID.EQ(postgres.UUID(id)).
			AND(table.Medicines.UserID.EQ(postgres.UUID(userID))),
	).RETURNING(table.Medicines.AllColumns)

	var dest model.Medicines
	if err := stmt.QueryContext(ctx, r.db, &dest); err != nil {
		if err == qrm.ErrNoRows {
			return nil, fmt.Errorf("medicine not found")
		}
		return nil, fmt.Errorf("update medicine: %w", err)
	}
	return toMedicine(dest), nil
}

func (r *Repository) SetArchived(ctx context.Context, id uuid.UUID, userID uuid.UUID, archived bool) (*medicine.Medicine, error) {
	var archivedAt postgres.Expression
	if archived {
		archivedAt = postgres.NOW()
	} else {
		archivedAt = postgres.NULL
	}

	stmt := table.Medicines.UPDATE(
		table.Medicines.ArchivedAt,
		table.Medicines.UpdatedAt,
	).SET(
		archivedAt,
		postgres.NOW(),
	).WHERE(
		table.Medicines.ID.EQ(postgres.UUID(id)).
			AND(table.Medicines.UserID.EQ(postgres.UUID(userID))),
	).RETURNING(table.Medicines.AllColumns)

	var dest model.Medicines
	if err := stmt.QueryContext(ctx, r.db, &dest); err != nil {
		if err == qrm.ErrNoRows {
			return nil, fmt.Errorf("medicine not found")
		}
		return nil, fmt.Errorf("set archived medicine: %w", err)
	}
	return toMedicine(dest), nil
}

// ── Take (transactional) ──────────────────────────────────────────────────────

func (r *Repository) Take(ctx context.Context, userID uuid.UUID, medicineID uuid.UUID, in medicine.TakeMedicineInput) (*medicine.MedicineIntake, *medicine.Medicine, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	// Fetch current medicine (scoped to user for ownership check)
	selectStmt := postgres.SELECT(table.Medicines.AllColumns).
		FROM(table.Medicines).
		WHERE(
			table.Medicines.ID.EQ(postgres.UUID(medicineID)).
				AND(table.Medicines.UserID.EQ(postgres.UUID(userID))),
		)

	var med model.Medicines
	if err := selectStmt.QueryContext(ctx, tx, &med); err != nil {
		if err == qrm.ErrNoRows {
			return nil, nil, fmt.Errorf("medicine not found")
		}
		return nil, nil, fmt.Errorf("fetch medicine for take: %w", err)
	}

	stockBefore, _ := med.StockQuantity.Float64()
	quantityTaken := in.QuantityTaken
	stockAfter := stockBefore - quantityTaken

	if stockAfter < 0 && !in.AllowNegative {
		return nil, nil, medicine.ErrInsufficientStock
	}

	// Determine taken_at
	takenAt := time.Now().UTC()
	if in.TakenAt != nil && *in.TakenAt != "" {
		t, err := time.Parse(time.RFC3339, *in.TakenAt)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid taken_at format (RFC3339 required): %w", err)
		}
		takenAt = t.UTC()
	}

	// Determine status
	status := medicine.IntakeStatusTaken
	if in.Status != nil {
		status = *in.Status
	}

	// INSERT intake
	intakeStmt := table.MedicineIntakes.INSERT(
		table.MedicineIntakes.MedicineID,
		table.MedicineIntakes.UserID,
		table.MedicineIntakes.MedicineName,
		table.MedicineIntakes.TakenAt,
		table.MedicineIntakes.QuantityTaken,
		table.MedicineIntakes.StockBefore,
		table.MedicineIntakes.StockAfter,
		table.MedicineIntakes.Status,
		table.MedicineIntakes.Note,
	).VALUES(
		postgres.UUID(medicineID),
		postgres.UUID(userID),
		med.Name,
		takenAt,
		decimal.NewFromFloat(quantityTaken),
		decimal.NewFromFloat(stockBefore),
		decimal.NewFromFloat(stockAfter),
		string(status),
		in.Note,
	).RETURNING(table.MedicineIntakes.AllColumns)

	var intakeDest model.MedicineIntakes
	if err := intakeStmt.QueryContext(ctx, tx, &intakeDest); err != nil {
		return nil, nil, fmt.Errorf("insert intake: %w", err)
	}

	// UPDATE medicine stock
	updateStmt := table.Medicines.UPDATE(
		table.Medicines.StockQuantity,
		table.Medicines.UpdatedAt,
	).SET(
		decimal.NewFromFloat(stockAfter),
		postgres.NOW(),
	).WHERE(
		table.Medicines.ID.EQ(postgres.UUID(medicineID)),
	).RETURNING(table.Medicines.AllColumns)

	var updatedMed model.Medicines
	if err := updateStmt.QueryContext(ctx, tx, &updatedMed); err != nil {
		return nil, nil, fmt.Errorf("update medicine stock after take: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, nil, fmt.Errorf("commit take tx: %w", err)
	}

	return toIntake(intakeDest), toMedicine(updatedMed), nil
}

// ── Adjust stock (transactional) ──────────────────────────────────────────────

func (r *Repository) AdjustStock(ctx context.Context, userID uuid.UUID, medicineID uuid.UUID, in medicine.AdjustStockInput) (*medicine.MedicineStockAdjustment, *medicine.Medicine, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	// Fetch current medicine
	selectStmt := postgres.SELECT(table.Medicines.AllColumns).
		FROM(table.Medicines).
		WHERE(
			table.Medicines.ID.EQ(postgres.UUID(medicineID)).
				AND(table.Medicines.UserID.EQ(postgres.UUID(userID))),
		)

	var med model.Medicines
	if err := selectStmt.QueryContext(ctx, tx, &med); err != nil {
		if err == qrm.ErrNoRows {
			return nil, nil, fmt.Errorf("medicine not found")
		}
		return nil, nil, fmt.Errorf("fetch medicine for stock adjustment: %w", err)
	}

	stockBefore, _ := med.StockQuantity.Float64()
	var stockAfter float64

	switch in.Type {
	case medicine.AdjustmentTypeAdd:
		stockAfter = stockBefore + in.Quantity
	case medicine.AdjustmentTypeRemove:
		stockAfter = stockBefore - in.Quantity
		if stockAfter < 0 {
			stockAfter = 0
		}
	case medicine.AdjustmentTypeCorrection:
		stockAfter = in.Quantity
	}

	// INSERT adjustment
	adjStmt := table.MedicineStockAdjustments.INSERT(
		table.MedicineStockAdjustments.MedicineID,
		table.MedicineStockAdjustments.UserID,
		table.MedicineStockAdjustments.Type,
		table.MedicineStockAdjustments.Quantity,
		table.MedicineStockAdjustments.StockBefore,
		table.MedicineStockAdjustments.StockAfter,
		table.MedicineStockAdjustments.Reason,
	).VALUES(
		postgres.UUID(medicineID),
		postgres.UUID(userID),
		string(in.Type),
		decimal.NewFromFloat(in.Quantity),
		decimal.NewFromFloat(stockBefore),
		decimal.NewFromFloat(stockAfter),
		in.Reason,
	).RETURNING(table.MedicineStockAdjustments.AllColumns)

	var adjDest model.MedicineStockAdjustments
	if err := adjStmt.QueryContext(ctx, tx, &adjDest); err != nil {
		return nil, nil, fmt.Errorf("insert stock adjustment: %w", err)
	}

	// UPDATE medicine stock
	updateStmt := table.Medicines.UPDATE(
		table.Medicines.StockQuantity,
		table.Medicines.UpdatedAt,
	).SET(
		decimal.NewFromFloat(stockAfter),
		postgres.NOW(),
	).WHERE(
		table.Medicines.ID.EQ(postgres.UUID(medicineID)),
	).RETURNING(table.Medicines.AllColumns)

	var updatedMed model.Medicines
	if err := updateStmt.QueryContext(ctx, tx, &updatedMed); err != nil {
		return nil, nil, fmt.Errorf("update medicine stock after adjustment: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, nil, fmt.Errorf("commit adjust stock tx: %w", err)
	}

	return toAdjustment(adjDest), toMedicine(updatedMed), nil
}

// ── Intakes ───────────────────────────────────────────────────────────────────

func (r *Repository) ListIntakes(ctx context.Context, userID uuid.UUID, opts medicine.ListIntakeOpts) ([]*medicine.MedicineIntake, error) {
	limit := opts.Limit
	if limit <= 0 {
		limit = 50
	}

	cond := table.MedicineIntakes.UserID.EQ(postgres.UUID(userID))
	if opts.MedicineID != nil {
		cond = cond.AND(table.MedicineIntakes.MedicineID.EQ(postgres.UUID(*opts.MedicineID)))
	}
	if opts.Date != nil {
		startOfDay := opts.Date.UTC().Truncate(24 * time.Hour)
		endOfDay := startOfDay.Add(24 * time.Hour)
		cond = cond.AND(
			table.MedicineIntakes.TakenAt.GT_EQ(postgres.TimestampzT(startOfDay)).
				AND(table.MedicineIntakes.TakenAt.LT(postgres.TimestampzT(endOfDay))),
		)
	}

	// Filtering by source_type joins the parent medicine, whose source_type the
	// intake itself doesn't carry.
	var from postgres.ReadableTable = table.MedicineIntakes
	if opts.SourceType != nil {
		from = table.MedicineIntakes.INNER_JOIN(
			table.Medicines,
			table.Medicines.ID.EQ(table.MedicineIntakes.MedicineID),
		)
		cond = cond.AND(table.Medicines.SourceType.EQ(postgres.String(string(*opts.SourceType))))
	}

	stmt := postgres.SELECT(table.MedicineIntakes.AllColumns).
		FROM(from).
		WHERE(cond).
		ORDER_BY(table.MedicineIntakes.TakenAt.DESC()).
		LIMIT(int64(limit))

	var dest []model.MedicineIntakes
	if err := stmt.QueryContext(ctx, r.db, &dest); err != nil {
		return nil, fmt.Errorf("list intakes: %w", err)
	}
	intakes := make([]*medicine.MedicineIntake, len(dest))
	for i, d := range dest {
		intakes[i] = toIntake(d)
	}
	return intakes, nil
}

// ── Stock adjustments ─────────────────────────────────────────────────────────

func (r *Repository) ListStockAdjustments(ctx context.Context, userID uuid.UUID, opts medicine.ListAdjustmentOpts) ([]*medicine.MedicineStockAdjustment, error) {
	limit := opts.Limit
	if limit <= 0 {
		limit = 50
	}

	cond := table.MedicineStockAdjustments.UserID.EQ(postgres.UUID(userID))
	if opts.MedicineID != nil {
		cond = cond.AND(table.MedicineStockAdjustments.MedicineID.EQ(postgres.UUID(*opts.MedicineID)))
	}
	if opts.Date != nil {
		startOfDay := opts.Date.UTC().Truncate(24 * time.Hour)
		endOfDay := startOfDay.Add(24 * time.Hour)
		cond = cond.AND(
			table.MedicineStockAdjustments.CreatedAt.GT_EQ(postgres.TimestampzT(startOfDay)).
				AND(table.MedicineStockAdjustments.CreatedAt.LT(postgres.TimestampzT(endOfDay))),
		)
	}

	// Filtering by source_type joins the parent medicine, whose source_type the
	// adjustment itself doesn't carry.
	var from postgres.ReadableTable = table.MedicineStockAdjustments
	if opts.SourceType != nil {
		from = table.MedicineStockAdjustments.INNER_JOIN(
			table.Medicines,
			table.Medicines.ID.EQ(table.MedicineStockAdjustments.MedicineID),
		)
		cond = cond.AND(table.Medicines.SourceType.EQ(postgres.String(string(*opts.SourceType))))
	}

	stmt := postgres.SELECT(table.MedicineStockAdjustments.AllColumns).
		FROM(from).
		WHERE(cond).
		ORDER_BY(table.MedicineStockAdjustments.CreatedAt.DESC()).
		LIMIT(int64(limit))

	var dest []model.MedicineStockAdjustments
	if err := stmt.QueryContext(ctx, r.db, &dest); err != nil {
		return nil, fmt.Errorf("list stock adjustments: %w", err)
	}
	adjs := make([]*medicine.MedicineStockAdjustment, len(dest))
	for i, d := range dest {
		adjs[i] = toAdjustment(d)
	}
	return adjs, nil
}

// ── Helpers ───────────────────────────────────────────────────────────────────

// startDateExpr / endDateExpr return a jet expression for a nullable DATE column.
func startDateExpr(t *time.Time) postgres.Expression {
	if t == nil {
		return postgres.NULL
	}
	return postgres.DateT(*t)
}

func endDateExpr(t *time.Time) postgres.Expression {
	if t == nil {
		return postgres.NULL
	}
	return postgres.DateT(*t)
}

func toMedicine(m model.Medicines) *medicine.Medicine {
	stock, _ := m.StockQuantity.Float64()
	dosage, _ := m.DosageAmount.Float64()
	threshold, _ := m.LowStockThreshold.Float64()

	reminderTimes := unmarshalReminderTimes(m.ReminderTimes)

	med := &medicine.Medicine{
		ID:                m.ID,
		UserID:            m.UserID,
		Name:              m.Name,
		SourceType:        medicine.SourceType(m.SourceType),
		GenericName:       m.GenericName,
		Description:       m.Description,
		StockQuantity:     stock,
		StockUnit:         m.StockUnit,
		DosageAmount:      dosage,
		DosageUnit:        m.DosageUnit,
		FrequencyType:     medicine.FrequencyType(m.FrequencyType),
		Timing:            nil,
		StartDate:         m.StartDate,
		EndDate:           m.EndDate,
		LowStockThreshold: threshold,
		Note:              m.Note,
		CreatedAt:         m.CreatedAt,
		UpdatedAt:         m.UpdatedAt,
		ArchivedAt:        m.ArchivedAt,
		ReminderEnabled:   m.ReminderEnabled,
		ReminderTimes:     reminderTimes,
	}
	if m.FrequencyValue != nil {
		v := int(*m.FrequencyValue)
		med.FrequencyValue = &v
	}
	if m.Timing != nil {
		t := medicine.Timing(*m.Timing)
		med.Timing = &t
	}
	return med
}

// marshalReminderTimes JSON-encodes a slice of "HH:MM" strings for storage.
// A nil or empty slice is stored as the literal "[]".
func marshalReminderTimes(times []string) (string, error) {
	if len(times) == 0 {
		return "[]", nil
	}
	b, err := json.Marshal(times)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// unmarshalReminderTimes decodes the JSON-encoded reminder_times column value.
// Returns an empty (non-nil) slice on any parse error so callers can range safely.
func unmarshalReminderTimes(raw string) []string {
	if raw == "" || raw == "[]" {
		return []string{}
	}
	var times []string
	if err := json.Unmarshal([]byte(raw), &times); err != nil {
		return []string{}
	}
	return times
}

func toIntake(m model.MedicineIntakes) *medicine.MedicineIntake {
	qty, _ := m.QuantityTaken.Float64()
	before, _ := m.StockBefore.Float64()
	after, _ := m.StockAfter.Float64()
	return &medicine.MedicineIntake{
		ID:            m.ID,
		MedicineID:    m.MedicineID,
		UserID:        m.UserID,
		MedicineName:  m.MedicineName,
		TakenAt:       m.TakenAt,
		QuantityTaken: qty,
		StockBefore:   before,
		StockAfter:    after,
		Status:        medicine.IntakeStatus(m.Status),
		Note:          m.Note,
		CreatedAt:     m.CreatedAt,
	}
}

func toAdjustment(m model.MedicineStockAdjustments) *medicine.MedicineStockAdjustment {
	qty, _ := m.Quantity.Float64()
	before, _ := m.StockBefore.Float64()
	after, _ := m.StockAfter.Float64()
	return &medicine.MedicineStockAdjustment{
		ID:          m.ID,
		MedicineID:  m.MedicineID,
		UserID:      m.UserID,
		Type:        medicine.AdjustmentType(m.Type),
		Quantity:    qty,
		StockBefore: before,
		StockAfter:  after,
		Reason:      m.Reason,
		CreatedAt:   m.CreatedAt,
	}
}

// ── Reminder batch helpers ────────────────────────────────────────────────────

// ScanActiveMedicinesForReminders returns all non-archived medicines across all
// users that are currently active (within start_date/end_date if set). Used
// exclusively by the cron reminder job — not scoped to a single user.
func (r *Repository) ScanActiveMedicinesForReminders(ctx context.Context) ([]*medicine.Medicine, error) {
	now := time.Now()
	cond := table.Medicines.ArchivedAt.IS_NULL().
		AND(
			table.Medicines.StartDate.IS_NULL().
				OR(table.Medicines.StartDate.LT_EQ(postgres.DateT(now))),
		).
		AND(
			table.Medicines.EndDate.IS_NULL().
				OR(table.Medicines.EndDate.GT_EQ(postgres.DateT(now))),
		)

	stmt := postgres.SELECT(table.Medicines.AllColumns).
		FROM(table.Medicines).
		WHERE(cond).
		ORDER_BY(table.Medicines.UserID.ASC(), table.Medicines.CreatedAt.ASC())

	var dest []model.Medicines
	if err := stmt.QueryContext(ctx, r.db, &dest); err != nil {
		return nil, fmt.Errorf("scan active medicines: %w", err)
	}
	meds := make([]*medicine.Medicine, len(dest))
	for i, d := range dest {
		meds[i] = toMedicine(d)
	}
	return meds, nil
}

// LogReminder inserts a reminder-log entry for the given (medicineID,
// reminderType, reminderKey) triple. Returns true when the row was newly
// inserted (i.e. the reminder has not been sent yet), or false when the
// ON CONFLICT suppressed the insert (already sent).
func (r *Repository) LogReminder(ctx context.Context, userID, medicineID uuid.UUID, reminderType, reminderKey string) (bool, error) {
	stmt := table.MedicineReminderLog.INSERT(
		table.MedicineReminderLog.UserID,
		table.MedicineReminderLog.MedicineID,
		table.MedicineReminderLog.ReminderType,
		table.MedicineReminderLog.ReminderKey,
	).VALUES(
		postgres.UUID(userID),
		postgres.UUID(medicineID),
		reminderType,
		reminderKey,
	).ON_CONFLICT(
		table.MedicineReminderLog.MedicineID,
		table.MedicineReminderLog.ReminderType,
		table.MedicineReminderLog.ReminderKey,
	).DO_NOTHING()

	res, err := stmt.ExecContext(ctx, r.db)
	if err != nil {
		return false, fmt.Errorf("log reminder: %w", err)
	}
	n, _ := res.RowsAffected()
	return n > 0, nil
}

// ListIntakesInRange returns all intakes for a specific medicine whose
// taken_at falls within [from, to). Used to check whether a dose was taken
// near a scheduled reminder time (missed-dose detection).
func (r *Repository) ListIntakesInRange(ctx context.Context, medicineID uuid.UUID, from, to time.Time) ([]*medicine.MedicineIntake, error) {
	cond := table.MedicineIntakes.MedicineID.EQ(postgres.UUID(medicineID)).
		AND(table.MedicineIntakes.TakenAt.GT_EQ(postgres.TimestampzT(from))).
		AND(table.MedicineIntakes.TakenAt.LT(postgres.TimestampzT(to)))

	stmt := postgres.SELECT(table.MedicineIntakes.AllColumns).
		FROM(table.MedicineIntakes).
		WHERE(cond).
		ORDER_BY(table.MedicineIntakes.TakenAt.ASC())

	var dest []model.MedicineIntakes
	if err := stmt.QueryContext(ctx, r.db, &dest); err != nil {
		return nil, fmt.Errorf("list intakes in range: %w", err)
	}
	intakes := make([]*medicine.MedicineIntake, len(dest))
	for i, d := range dest {
		intakes[i] = toIntake(d)
	}
	return intakes, nil
}
