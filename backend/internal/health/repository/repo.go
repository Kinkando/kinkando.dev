package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/go-jet/jet/v2/postgres"
	"github.com/go-jet/jet/v2/qrm"
	"github.com/google/uuid"
	"github.com/kinkando/personal-dashboard/gen/kinkando/public/model"
	"github.com/kinkando/personal-dashboard/gen/kinkando/public/table"
	"github.com/kinkando/personal-dashboard/internal/health"
	"github.com/shopspring/decimal"
)

type Repository struct {
	db *sql.DB
}

func New(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// ── Profile ───────────────────────────────────────────────────────────────────

func (r *Repository) GetProfile(ctx context.Context, userID uuid.UUID) (*health.Profile, error) {
	stmt := postgres.SELECT(table.HealthProfiles.AllColumns).
		FROM(table.HealthProfiles).
		WHERE(table.HealthProfiles.UserID.EQ(postgres.UUID(userID)))

	var dest model.HealthProfiles
	if err := stmt.QueryContext(ctx, r.db, &dest); err != nil {
		if err == qrm.ErrNoRows {
			return nil, nil // no profile yet — not an error
		}
		return nil, fmt.Errorf("get profile: %w", err)
	}
	return toProfile(dest), nil
}

func (r *Repository) UpsertProfile(ctx context.Context, userID uuid.UUID, in health.UpsertProfileInput) (*health.Profile, error) {
	var height *decimal.Decimal
	if in.Height != nil {
		h := decimal.NewFromFloat(*in.Height)
		height = &h
	}
	var birthdate *time.Time
	if in.Birthdate != nil {
		t, err := time.Parse(time.DateOnly, *in.Birthdate)
		if err != nil {
			return nil, fmt.Errorf("invalid birthdate format: %w", err)
		}
		birthdate = &t
	}
	var gender *string
	if in.Gender != nil {
		g := string(*in.Gender)
		gender = &g
	}
	var goal *string
	if in.Goal != nil {
		g := string(*in.Goal)
		goal = &g
	}

	stmt := table.HealthProfiles.INSERT(
		table.HealthProfiles.UserID,
		table.HealthProfiles.Height,
		table.HealthProfiles.Birthdate,
		table.HealthProfiles.Gender,
		table.HealthProfiles.Goal,
	).VALUES(
		postgres.UUID(userID),
		height,
		birthdate,
		gender,
		goal,
	).ON_CONFLICT(table.HealthProfiles.UserID).
		DO_UPDATE(postgres.SET(
			table.HealthProfiles.Height.SET(table.HealthProfiles.EXCLUDED.Height),
			table.HealthProfiles.Birthdate.SET(table.HealthProfiles.EXCLUDED.Birthdate),
			table.HealthProfiles.Gender.SET(table.HealthProfiles.EXCLUDED.Gender),
			table.HealthProfiles.Goal.SET(table.HealthProfiles.EXCLUDED.Goal),
			table.HealthProfiles.UpdatedAt.SET(postgres.NOW()),
		)).RETURNING(table.HealthProfiles.AllColumns)

	var dest model.HealthProfiles
	if err := stmt.QueryContext(ctx, r.db, &dest); err != nil {
		return nil, fmt.Errorf("upsert profile: %w", err)
	}
	return toProfile(dest), nil
}

// ── Weight logs ───────────────────────────────────────────────────────────────

func (r *Repository) ListWeightLogs(ctx context.Context, userID uuid.UUID, from, to time.Time) ([]*health.WeightLog, error) {
	cond := table.HealthWeightLogs.UserID.EQ(postgres.UUID(userID))
	if !from.IsZero() {
		cond = cond.AND(table.HealthWeightLogs.LoggedAt.GT_EQ(postgres.DateT(from)))
	}
	if !to.IsZero() {
		cond = cond.AND(table.HealthWeightLogs.LoggedAt.LT_EQ(postgres.DateT(to)))
	}

	stmt := postgres.SELECT(table.HealthWeightLogs.AllColumns).
		FROM(table.HealthWeightLogs).
		WHERE(cond).
		ORDER_BY(table.HealthWeightLogs.LoggedAt.ASC(), table.HealthWeightLogs.CreatedAt.ASC())

	var dest []model.HealthWeightLogs
	if err := stmt.QueryContext(ctx, r.db, &dest); err != nil {
		return nil, fmt.Errorf("list weight logs: %w", err)
	}
	logs := make([]*health.WeightLog, len(dest))
	for i, d := range dest {
		logs[i] = toWeightLog(d)
	}
	return logs, nil
}

func (r *Repository) CreateWeightLog(ctx context.Context, userID uuid.UUID, in health.CreateWeightInput) (*health.WeightLog, error) {
	loggedAt := time.Now().UTC().Truncate(24 * time.Hour)
	if in.LoggedAt != "" {
		t, err := time.Parse(time.DateOnly, in.LoggedAt)
		if err != nil {
			return nil, fmt.Errorf("invalid logged_at date format: %w", err)
		}
		loggedAt = t
	}

	stmt := table.HealthWeightLogs.INSERT(
		table.HealthWeightLogs.UserID,
		table.HealthWeightLogs.Weight,
		table.HealthWeightLogs.Note,
		table.HealthWeightLogs.LoggedAt,
	).VALUES(
		postgres.UUID(userID),
		decimal.NewFromFloat(in.Weight),
		in.Note,
		postgres.DateT(loggedAt),
	).ON_CONFLICT(table.HealthWeightLogs.UserID, table.HealthWeightLogs.LoggedAt).
		DO_UPDATE(postgres.SET(
			table.HealthWeightLogs.Weight.SET(table.HealthWeightLogs.EXCLUDED.Weight),
			table.HealthWeightLogs.Note.SET(table.HealthWeightLogs.EXCLUDED.Note),
		)).RETURNING(table.HealthWeightLogs.AllColumns)

	var dest model.HealthWeightLogs
	if err := stmt.QueryContext(ctx, r.db, &dest); err != nil {
		return nil, fmt.Errorf("create weight log: %w", err)
	}
	return toWeightLog(dest), nil
}

func (r *Repository) UpdateWeightLog(ctx context.Context, id uuid.UUID, userID uuid.UUID, in health.UpdateWeightInput) (*health.WeightLog, error) {
	loggedAt := time.Now().UTC().Truncate(24 * time.Hour)
	if in.LoggedAt != "" {
		t, err := time.Parse(time.DateOnly, in.LoggedAt)
		if err != nil {
			return nil, fmt.Errorf("invalid logged_at date format: %w", err)
		}
		loggedAt = t
	}

	stmt := table.HealthWeightLogs.UPDATE(
		table.HealthWeightLogs.Weight,
		table.HealthWeightLogs.Note,
		table.HealthWeightLogs.LoggedAt,
	).SET(
		decimal.NewFromFloat(in.Weight),
		in.Note,
		postgres.DateT(loggedAt),
	).WHERE(
		table.HealthWeightLogs.ID.EQ(postgres.UUID(id)).
			AND(table.HealthWeightLogs.UserID.EQ(postgres.UUID(userID))),
	).RETURNING(table.HealthWeightLogs.AllColumns)

	var dest model.HealthWeightLogs
	if err := stmt.QueryContext(ctx, r.db, &dest); err != nil {
		if err == qrm.ErrNoRows {
			return nil, fmt.Errorf("weight log not found")
		}
		return nil, fmt.Errorf("update weight log: %w", err)
	}
	return toWeightLog(dest), nil
}

func (r *Repository) DeleteWeightLog(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	stmt := table.HealthWeightLogs.DELETE().WHERE(
		table.HealthWeightLogs.ID.EQ(postgres.UUID(id)).
			AND(table.HealthWeightLogs.UserID.EQ(postgres.UUID(userID))),
	)
	res, err := stmt.ExecContext(ctx, r.db)
	if err != nil {
		return fmt.Errorf("delete weight log: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil || n == 0 {
		return fmt.Errorf("weight log not found")
	}
	return nil
}

// ── Food logs ─────────────────────────────────────────────────────────────────

func (r *Repository) ListFoodLogs(ctx context.Context, userID uuid.UUID) ([]*health.FoodLog, error) {
	stmt := postgres.SELECT(table.HealthFoodLogs.AllColumns).
		FROM(table.HealthFoodLogs).
		WHERE(table.HealthFoodLogs.UserID.EQ(postgres.UUID(userID))).
		ORDER_BY(table.HealthFoodLogs.ConsumedAt.DESC(), table.HealthFoodLogs.CreatedAt.DESC())

	var dest []model.HealthFoodLogs
	if err := stmt.QueryContext(ctx, r.db, &dest); err != nil {
		return nil, fmt.Errorf("list food logs: %w", err)
	}
	logs := make([]*health.FoodLog, len(dest))
	for i, d := range dest {
		logs[i] = toFoodLog(d)
	}
	return logs, nil
}

func (r *Repository) CreateFoodLog(ctx context.Context, userID uuid.UUID, in health.CreateFoodInput) (*health.FoodLog, error) {
	consumedAt := time.Now().UTC().Truncate(24 * time.Hour)
	if in.ConsumedAt != "" {
		t, err := time.Parse(time.DateOnly, in.ConsumedAt)
		if err != nil {
			return nil, fmt.Errorf("invalid consumed_at date format: %w", err)
		}
		consumedAt = t
	}

	var calories *int32
	if in.Calories != nil {
		c := int32(*in.Calories)
		calories = &c
	}
	var proteinG, carbsG, fatG *decimal.Decimal
	if in.ProteinG != nil {
		d := decimal.NewFromFloat(*in.ProteinG)
		proteinG = &d
	}
	if in.CarbsG != nil {
		d := decimal.NewFromFloat(*in.CarbsG)
		carbsG = &d
	}
	if in.FatG != nil {
		d := decimal.NewFromFloat(*in.FatG)
		fatG = &d
	}

	stmt := table.HealthFoodLogs.INSERT(
		table.HealthFoodLogs.UserID,
		table.HealthFoodLogs.Name,
		table.HealthFoodLogs.MealType,
		table.HealthFoodLogs.Calories,
		table.HealthFoodLogs.ProteinG,
		table.HealthFoodLogs.CarbsG,
		table.HealthFoodLogs.FatG,
		table.HealthFoodLogs.Notes,
		table.HealthFoodLogs.ConsumedAt,
	).VALUES(
		postgres.UUID(userID),
		in.Name,
		string(in.MealType),
		calories,
		proteinG,
		carbsG,
		fatG,
		in.Notes,
		postgres.DateT(consumedAt),
	).RETURNING(table.HealthFoodLogs.AllColumns)

	var dest model.HealthFoodLogs
	if err := stmt.QueryContext(ctx, r.db, &dest); err != nil {
		return nil, fmt.Errorf("create food log: %w", err)
	}
	return toFoodLog(dest), nil
}

func (r *Repository) UpdateFoodLog(ctx context.Context, id uuid.UUID, userID uuid.UUID, in health.UpdateFoodInput) (*health.FoodLog, error) {
	consumedAt := time.Now().UTC().Truncate(24 * time.Hour)
	if in.ConsumedAt != "" {
		t, err := time.Parse(time.DateOnly, in.ConsumedAt)
		if err != nil {
			return nil, fmt.Errorf("invalid consumed_at date format: %w", err)
		}
		consumedAt = t
	}

	var calories *int32
	if in.Calories != nil {
		c := int32(*in.Calories)
		calories = &c
	}
	var proteinG, carbsG, fatG *decimal.Decimal
	if in.ProteinG != nil {
		d := decimal.NewFromFloat(*in.ProteinG)
		proteinG = &d
	}
	if in.CarbsG != nil {
		d := decimal.NewFromFloat(*in.CarbsG)
		carbsG = &d
	}
	if in.FatG != nil {
		d := decimal.NewFromFloat(*in.FatG)
		fatG = &d
	}

	stmt := table.HealthFoodLogs.UPDATE(
		table.HealthFoodLogs.Name,
		table.HealthFoodLogs.MealType,
		table.HealthFoodLogs.Calories,
		table.HealthFoodLogs.ProteinG,
		table.HealthFoodLogs.CarbsG,
		table.HealthFoodLogs.FatG,
		table.HealthFoodLogs.Notes,
		table.HealthFoodLogs.ConsumedAt,
	).SET(
		in.Name,
		string(in.MealType),
		calories,
		proteinG,
		carbsG,
		fatG,
		in.Notes,
		postgres.DateT(consumedAt),
	).WHERE(
		table.HealthFoodLogs.ID.EQ(postgres.UUID(id)).
			AND(table.HealthFoodLogs.UserID.EQ(postgres.UUID(userID))),
	).RETURNING(table.HealthFoodLogs.AllColumns)

	var dest model.HealthFoodLogs
	if err := stmt.QueryContext(ctx, r.db, &dest); err != nil {
		if err == qrm.ErrNoRows {
			return nil, fmt.Errorf("food log not found")
		}
		return nil, fmt.Errorf("update food log: %w", err)
	}
	return toFoodLog(dest), nil
}

func (r *Repository) DeleteFoodLog(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	stmt := table.HealthFoodLogs.DELETE().WHERE(
		table.HealthFoodLogs.ID.EQ(postgres.UUID(id)).
			AND(table.HealthFoodLogs.UserID.EQ(postgres.UUID(userID))),
	)
	res, err := stmt.ExecContext(ctx, r.db)
	if err != nil {
		return fmt.Errorf("delete food log: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil || n == 0 {
		return fmt.Errorf("food log not found")
	}
	return nil
}

// ── Sleep logs ────────────────────────────────────────────────────────────────

func (r *Repository) ListSleepLogs(ctx context.Context, userID uuid.UUID, from, to time.Time) ([]*health.SleepLog, error) {
	cond := table.HealthSleepLogs.UserID.EQ(postgres.UUID(userID))
	if !from.IsZero() {
		cond = cond.AND(table.HealthSleepLogs.LoggedAt.GT_EQ(postgres.DateT(from)))
	}
	if !to.IsZero() {
		cond = cond.AND(table.HealthSleepLogs.LoggedAt.LT_EQ(postgres.DateT(to)))
	}

	stmt := postgres.SELECT(table.HealthSleepLogs.AllColumns).
		FROM(table.HealthSleepLogs).
		WHERE(cond).
		ORDER_BY(table.HealthSleepLogs.LoggedAt.DESC(), table.HealthSleepLogs.CreatedAt.DESC())

	var dest []model.HealthSleepLogs
	if err := stmt.QueryContext(ctx, r.db, &dest); err != nil {
		return nil, fmt.Errorf("list sleep logs: %w", err)
	}
	logs := make([]*health.SleepLog, len(dest))
	for i, d := range dest {
		logs[i] = toSleepLog(d)
	}
	return logs, nil
}

func (r *Repository) CreateSleepLog(ctx context.Context, userID uuid.UUID, in health.CreateSleepInput) (*health.SleepLog, error) {
	startedAt, err := time.Parse(time.RFC3339, in.StartedAt)
	if err != nil {
		return nil, fmt.Errorf("invalid started_at format (RFC3339 required): %w", err)
	}
	endedAt, err := time.Parse(time.RFC3339, in.EndedAt)
	if err != nil {
		return nil, fmt.Errorf("invalid ended_at format (RFC3339 required): %w", err)
	}
	if !endedAt.After(startedAt) {
		return nil, fmt.Errorf("ended_at must be after started_at")
	}

	loggedAt := startedAt.UTC().Truncate(24 * time.Hour)
	if in.LoggedAt != "" {
		t, err := time.Parse(time.DateOnly, in.LoggedAt)
		if err != nil {
			return nil, fmt.Errorf("invalid logged_at date format: %w", err)
		}
		loggedAt = t
	}

	var score *int32
	if in.Score != nil {
		s := int32(*in.Score)
		score = &s
	}

	stmt := table.HealthSleepLogs.INSERT(
		table.HealthSleepLogs.UserID,
		table.HealthSleepLogs.StartedAt,
		table.HealthSleepLogs.EndedAt,
		table.HealthSleepLogs.Score,
		table.HealthSleepLogs.Notes,
		table.HealthSleepLogs.LoggedAt,
	).VALUES(
		postgres.UUID(userID),
		startedAt.UTC(),
		endedAt.UTC(),
		score,
		in.Notes,
		postgres.DateT(loggedAt),
	).ON_CONFLICT(table.HealthSleepLogs.UserID, table.HealthSleepLogs.LoggedAt).
		DO_UPDATE(postgres.SET(
			table.HealthSleepLogs.StartedAt.SET(table.HealthSleepLogs.EXCLUDED.StartedAt),
			table.HealthSleepLogs.EndedAt.SET(table.HealthSleepLogs.EXCLUDED.EndedAt),
			table.HealthSleepLogs.Score.SET(table.HealthSleepLogs.EXCLUDED.Score),
			table.HealthSleepLogs.Notes.SET(table.HealthSleepLogs.EXCLUDED.Notes),
		)).RETURNING(table.HealthSleepLogs.AllColumns)

	var dest model.HealthSleepLogs
	if err := stmt.QueryContext(ctx, r.db, &dest); err != nil {
		return nil, fmt.Errorf("create sleep log: %w", err)
	}
	return toSleepLog(dest), nil
}

func (r *Repository) UpdateSleepLog(ctx context.Context, id uuid.UUID, userID uuid.UUID, in health.UpdateSleepInput) (*health.SleepLog, error) {
	startedAt, err := time.Parse(time.RFC3339, in.StartedAt)
	if err != nil {
		return nil, fmt.Errorf("invalid started_at format (RFC3339 required): %w", err)
	}
	endedAt, err := time.Parse(time.RFC3339, in.EndedAt)
	if err != nil {
		return nil, fmt.Errorf("invalid ended_at format (RFC3339 required): %w", err)
	}
	if !endedAt.After(startedAt) {
		return nil, fmt.Errorf("ended_at must be after started_at")
	}

	loggedAt := startedAt.UTC().Truncate(24 * time.Hour)
	if in.LoggedAt != "" {
		t, err := time.Parse(time.DateOnly, in.LoggedAt)
		if err != nil {
			return nil, fmt.Errorf("invalid logged_at date format: %w", err)
		}
		loggedAt = t
	}

	var score *int32
	if in.Score != nil {
		s := int32(*in.Score)
		score = &s
	}

	stmt := table.HealthSleepLogs.UPDATE(
		table.HealthSleepLogs.StartedAt,
		table.HealthSleepLogs.EndedAt,
		table.HealthSleepLogs.Score,
		table.HealthSleepLogs.Notes,
		table.HealthSleepLogs.LoggedAt,
	).SET(
		startedAt.UTC(),
		endedAt.UTC(),
		score,
		in.Notes,
		postgres.DateT(loggedAt),
	).WHERE(
		table.HealthSleepLogs.ID.EQ(postgres.UUID(id)).
			AND(table.HealthSleepLogs.UserID.EQ(postgres.UUID(userID))),
	).RETURNING(table.HealthSleepLogs.AllColumns)

	var dest model.HealthSleepLogs
	if err := stmt.QueryContext(ctx, r.db, &dest); err != nil {
		if err == qrm.ErrNoRows {
			return nil, fmt.Errorf("sleep log not found")
		}
		return nil, fmt.Errorf("update sleep log: %w", err)
	}
	return toSleepLog(dest), nil
}

func (r *Repository) DeleteSleepLog(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	stmt := table.HealthSleepLogs.DELETE().WHERE(
		table.HealthSleepLogs.ID.EQ(postgres.UUID(id)).
			AND(table.HealthSleepLogs.UserID.EQ(postgres.UUID(userID))),
	)
	res, err := stmt.ExecContext(ctx, r.db)
	if err != nil {
		return fmt.Errorf("delete sleep log: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil || n == 0 {
		return fmt.Errorf("sleep log not found")
	}
	return nil
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func toProfile(m model.HealthProfiles) *health.Profile {
	p := &health.Profile{
		ID:        m.ID,
		UserID:    m.UserID,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
	if m.Height != nil {
		h, _ := m.Height.Float64()
		p.Height = &h
	}
	if m.Birthdate != nil {
		s := m.Birthdate.Format(time.DateOnly)
		p.Birthdate = &s
	}
	if m.Gender != nil {
		g := health.Gender(*m.Gender)
		p.Gender = &g
	}
	if m.Goal != nil {
		g := health.Goal(*m.Goal)
		p.Goal = &g
	}
	return p
}

func toWeightLog(m model.HealthWeightLogs) *health.WeightLog {
	w, _ := m.Weight.Float64()
	return &health.WeightLog{
		ID:       m.ID,
		Weight:   w,
		Note:     m.Note,
		LoggedAt: m.LoggedAt,
	}
}

func toFoodLog(m model.HealthFoodLogs) *health.FoodLog {
	f := &health.FoodLog{
		ID:         m.ID,
		UserID:     m.UserID,
		Name:       m.Name,
		MealType:   health.MealType(m.MealType),
		Notes:      m.Notes,
		ConsumedAt: m.ConsumedAt,
		CreatedAt:  m.CreatedAt,
	}
	if m.Calories != nil {
		c := int(*m.Calories)
		f.Calories = &c
	}
	if m.ProteinG != nil {
		v, _ := m.ProteinG.Float64()
		f.ProteinG = &v
	}
	if m.CarbsG != nil {
		v, _ := m.CarbsG.Float64()
		f.CarbsG = &v
	}
	if m.FatG != nil {
		v, _ := m.FatG.Float64()
		f.FatG = &v
	}
	return f
}

func toSleepLog(m model.HealthSleepLogs) *health.SleepLog {
	s := &health.SleepLog{
		ID:              m.ID,
		UserID:          m.UserID,
		StartedAt:       m.StartedAt,
		EndedAt:         m.EndedAt,
		DurationMinutes: int(m.EndedAt.Sub(m.StartedAt).Minutes()),
		Notes:           m.Notes,
		LoggedAt:        m.LoggedAt,
		CreatedAt:       m.CreatedAt,
	}
	if m.Score != nil {
		sc := int(*m.Score)
		s.Score = &sc
	}
	return s
}

// ── Batch reminder helpers ────────────────────────────────────────────────────

// UsersMissingWeightToday returns the IDs of users who have logged weight at
// least once in the past but have not yet logged for the given date.
// These are the candidates for the morning weight-log nudge. Users who have
// never logged weight are excluded (they haven't opted into weight tracking).
func (r *Repository) UsersMissingWeightToday(ctx context.Context, today time.Time) ([]uuid.UUID, error) {
	loggedToday := postgres.SELECT(table.HealthWeightLogs.UserID).
		FROM(table.HealthWeightLogs).
		WHERE(table.HealthWeightLogs.LoggedAt.EQ(postgres.DateT(today)))

	stmt := postgres.SELECT(table.HealthWeightLogs.UserID).
		FROM(table.HealthWeightLogs).
		WHERE(table.HealthWeightLogs.UserID.NOT_IN(loggedToday)).
		GROUP_BY(table.HealthWeightLogs.UserID).
		ORDER_BY(table.HealthWeightLogs.UserID.ASC())

	var rows []struct {
		UserID uuid.UUID `alias:"health_weight_logs.user_id"`
	}
	if err := stmt.QueryContext(ctx, r.db, &rows); err != nil {
		return nil, fmt.Errorf("users missing weight today: %w", err)
	}

	ids := make([]uuid.UUID, len(rows))
	for i, row := range rows {
		ids[i] = row.UserID
	}
	return ids, nil
}
