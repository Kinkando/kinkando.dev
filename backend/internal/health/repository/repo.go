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
		if err == sql.ErrNoRows {
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
	var age *int32
	if in.Age != nil {
		a := int32(*in.Age)
		age = &a
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
		table.HealthProfiles.Age,
		table.HealthProfiles.Gender,
		table.HealthProfiles.Goal,
	).VALUES(
		postgres.UUID(userID),
		height,
		age,
		gender,
		goal,
	).ON_CONFLICT(table.HealthProfiles.UserID).
		DO_UPDATE(postgres.SET(
			table.HealthProfiles.Height.SET(table.HealthProfiles.EXCLUDED.Height),
			table.HealthProfiles.Age.SET(table.HealthProfiles.EXCLUDED.Age),
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

func (r *Repository) ListWeightLogs(ctx context.Context, userID uuid.UUID) ([]*health.WeightLog, error) {
	stmt := postgres.SELECT(table.HealthWeightLogs.AllColumns).
		FROM(table.HealthWeightLogs).
		WHERE(table.HealthWeightLogs.UserID.EQ(postgres.UUID(userID))).
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
		t, err := time.Parse("2006-01-02", in.LoggedAt)
		if err != nil {
			return nil, fmt.Errorf("invalid logged_at date format: %w", err)
		}
		loggedAt = t
	}

	stmt := table.HealthWeightLogs.INSERT(
		table.HealthWeightLogs.UserID,
		table.HealthWeightLogs.Weight,
		table.HealthWeightLogs.LoggedAt,
	).VALUES(
		postgres.UUID(userID),
		decimal.NewFromFloat(in.Weight),
		postgres.DateT(loggedAt),
	).RETURNING(table.HealthWeightLogs.AllColumns)

	var dest model.HealthWeightLogs
	if err := stmt.QueryContext(ctx, r.db, &dest); err != nil {
		return nil, fmt.Errorf("create weight log: %w", err)
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

// ── Exercises ─────────────────────────────────────────────────────────────────

func (r *Repository) ListExercises(ctx context.Context, userID uuid.UUID) ([]*health.Exercise, error) {
	stmt := postgres.SELECT(table.HealthExercises.AllColumns).
		FROM(table.HealthExercises).
		WHERE(table.HealthExercises.UserID.EQ(postgres.UUID(userID))).
		ORDER_BY(table.HealthExercises.PerformedAt.DESC(), table.HealthExercises.CreatedAt.DESC())

	var dest []model.HealthExercises
	if err := stmt.QueryContext(ctx, r.db, &dest); err != nil {
		return nil, fmt.Errorf("list exercises: %w", err)
	}
	exercises := make([]*health.Exercise, len(dest))
	for i, d := range dest {
		exercises[i] = toExercise(d)
	}
	return exercises, nil
}

func (r *Repository) CreateExercise(ctx context.Context, userID uuid.UUID, in health.CreateExerciseInput) (*health.Exercise, error) {
	performedAt := time.Now().UTC().Truncate(24 * time.Hour)
	if in.PerformedAt != "" {
		t, err := time.Parse("2006-01-02", in.PerformedAt)
		if err != nil {
			return nil, fmt.Errorf("invalid performed_at date format: %w", err)
		}
		performedAt = t
	}

	var durationMinutes *int32
	if in.DurationMinutes != nil {
		d := int32(*in.DurationMinutes)
		durationMinutes = &d
	}
	var calories *int32
	if in.Calories != nil {
		cal := int32(*in.Calories)
		calories = &cal
	}

	stmt := table.HealthExercises.INSERT(
		table.HealthExercises.UserID,
		table.HealthExercises.Name,
		table.HealthExercises.Type,
		table.HealthExercises.DurationMinutes,
		table.HealthExercises.Calories,
		table.HealthExercises.Notes,
		table.HealthExercises.PerformedAt,
	).VALUES(
		postgres.UUID(userID),
		in.Name,
		string(in.Type),
		durationMinutes,
		calories,
		in.Notes,
		postgres.DateT(performedAt),
	).RETURNING(table.HealthExercises.AllColumns)

	var dest model.HealthExercises
	if err := stmt.QueryContext(ctx, r.db, &dest); err != nil {
		return nil, fmt.Errorf("create exercise: %w", err)
	}
	return toExercise(dest), nil
}

func (r *Repository) UpdateExercise(ctx context.Context, id uuid.UUID, userID uuid.UUID, in health.UpdateExerciseInput) (*health.Exercise, error) {
	performedAt := time.Now().UTC().Truncate(24 * time.Hour)
	if in.PerformedAt != "" {
		t, err := time.Parse("2006-01-02", in.PerformedAt)
		if err != nil {
			return nil, fmt.Errorf("invalid performed_at date format: %w", err)
		}
		performedAt = t
	}

	var durationMinutes *int32
	if in.DurationMinutes != nil {
		d := int32(*in.DurationMinutes)
		durationMinutes = &d
	}
	var calories *int32
	if in.Calories != nil {
		cal := int32(*in.Calories)
		calories = &cal
	}

	stmt := table.HealthExercises.UPDATE(
		table.HealthExercises.Name,
		table.HealthExercises.Type,
		table.HealthExercises.DurationMinutes,
		table.HealthExercises.Calories,
		table.HealthExercises.Notes,
		table.HealthExercises.PerformedAt,
	).SET(
		in.Name,
		string(in.Type),
		durationMinutes,
		calories,
		in.Notes,
		postgres.DateT(performedAt),
	).WHERE(
		table.HealthExercises.ID.EQ(postgres.UUID(id)).
			AND(table.HealthExercises.UserID.EQ(postgres.UUID(userID))),
	).RETURNING(table.HealthExercises.AllColumns)

	var dest model.HealthExercises
	if err := stmt.QueryContext(ctx, r.db, &dest); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("exercise not found")
		}
		return nil, fmt.Errorf("update exercise: %w", err)
	}
	return toExercise(dest), nil
}

func (r *Repository) DeleteExercise(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	stmt := table.HealthExercises.DELETE().WHERE(
		table.HealthExercises.ID.EQ(postgres.UUID(id)).
			AND(table.HealthExercises.UserID.EQ(postgres.UUID(userID))),
	)
	res, err := stmt.ExecContext(ctx, r.db)
	if err != nil {
		return fmt.Errorf("delete exercise: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil || n == 0 {
		return fmt.Errorf("exercise not found")
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
	if m.Age != nil {
		a := int(*m.Age)
		p.Age = &a
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
		LoggedAt: m.LoggedAt,
	}
}

func toExercise(m model.HealthExercises) *health.Exercise {
	e := &health.Exercise{
		ID:          m.ID,
		UserID:      m.UserID,
		Name:        m.Name,
		Type:        health.ExerciseType(m.Type),
		Notes:       m.Notes,
		PerformedAt: m.PerformedAt,
		CreatedAt:   m.CreatedAt,
	}
	if m.DurationMinutes != nil {
		d := int(*m.DurationMinutes)
		e.DurationMinutes = &d
	}
	if m.Calories != nil {
		cal := int(*m.Calories)
		e.Calories = &cal
	}
	return e
}
