// Package repository provides Postgres-backed persistence for the workout module.
//
// NOTE: The generated go-jet types (table.Workout*, model.Workout*) are produced
// by running:
//
//	make run-db-migrations-windows && make gen-sql-builder-windows
//
// The package will not compile until that step has been executed.
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
	"github.com/kinkando/personal-dashboard/internal/workout"
	"github.com/shopspring/decimal"
)

type Repository struct {
	db *sql.DB
}

func New(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// ── Presets ───────────────────────────────────────────────────────────────────

func (r *Repository) ListPresets(ctx context.Context, userID uuid.UUID) ([]*workout.Preset, error) {
	stmt := postgres.SELECT(table.WorkoutPresets.AllColumns).
		FROM(table.WorkoutPresets).
		WHERE(table.WorkoutPresets.UserID.EQ(postgres.UUID(userID))).
		ORDER_BY(table.WorkoutPresets.CreatedAt.ASC())

	var presetRows []model.WorkoutPresets
	if err := stmt.QueryContext(ctx, r.db, &presetRows); err != nil {
		return nil, fmt.Errorf("list presets: %w", err)
	}

	if len(presetRows) == 0 {
		return []*workout.Preset{}, nil
	}

	// Batch-load exercises for all presets in a single query.
	presetIDs := make([]postgres.Expression, len(presetRows))
	for i, p := range presetRows {
		presetIDs[i] = postgres.UUID(p.ID)
	}

	exStmt := postgres.SELECT(table.WorkoutPresetExercises.AllColumns).
		FROM(table.WorkoutPresetExercises).
		WHERE(table.WorkoutPresetExercises.PresetID.IN(presetIDs...)).
		ORDER_BY(table.WorkoutPresetExercises.OrderIndex.ASC())

	var exRows []model.WorkoutPresetExercises
	if err := exStmt.QueryContext(ctx, r.db, &exRows); err != nil {
		return nil, fmt.Errorf("list preset exercises: %w", err)
	}

	exByPreset := make(map[uuid.UUID][]model.WorkoutPresetExercises)
	for _, ex := range exRows {
		exByPreset[ex.PresetID] = append(exByPreset[ex.PresetID], ex)
	}

	presets := make([]*workout.Preset, len(presetRows))
	for i, p := range presetRows {
		presets[i] = toPreset(p, exByPreset[p.ID])
	}
	return presets, nil
}

func (r *Repository) GetPreset(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*workout.Preset, error) {
	stmt := postgres.SELECT(table.WorkoutPresets.AllColumns).
		FROM(table.WorkoutPresets).
		WHERE(
			table.WorkoutPresets.ID.EQ(postgres.UUID(id)).
				AND(table.WorkoutPresets.UserID.EQ(postgres.UUID(userID))),
		)

	var dest model.WorkoutPresets
	if err := stmt.QueryContext(ctx, r.db, &dest); err != nil {
		if err == qrm.ErrNoRows {
			return nil, fmt.Errorf("preset not found")
		}
		return nil, fmt.Errorf("get preset: %w", err)
	}

	exRows, err := r.fetchPresetExercises(ctx, r.db, id)
	if err != nil {
		return nil, err
	}
	return toPreset(dest, exRows), nil
}

func (r *Repository) CreatePreset(ctx context.Context, userID uuid.UUID, in workout.CreatePresetInput) (*workout.Preset, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	// Insert preset header.
	insertStmt := table.WorkoutPresets.INSERT(
		table.WorkoutPresets.UserID,
		table.WorkoutPresets.Name,
		table.WorkoutPresets.Type,
		table.WorkoutPresets.Description,
	).VALUES(
		postgres.UUID(userID),
		in.Name,
		string(in.Type),
		in.Description,
	).RETURNING(table.WorkoutPresets.AllColumns)

	var presetRow model.WorkoutPresets
	if err = insertStmt.QueryContext(ctx, tx, &presetRow); err != nil {
		return nil, fmt.Errorf("create preset: %w", err)
	}

	// Insert exercises (if any).
	if err = r.insertPresetExercises(ctx, tx, presetRow.ID, in.Exercises); err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	exRows, err := r.fetchPresetExercises(ctx, r.db, presetRow.ID)
	if err != nil {
		return nil, err
	}
	return toPreset(presetRow, exRows), nil
}

func (r *Repository) UpdatePreset(ctx context.Context, id uuid.UUID, userID uuid.UUID, in workout.UpdatePresetInput) (*workout.Preset, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	// Update preset header.
	updateStmt := table.WorkoutPresets.UPDATE(
		table.WorkoutPresets.Name,
		table.WorkoutPresets.Type,
		table.WorkoutPresets.Description,
		table.WorkoutPresets.UpdatedAt,
	).SET(
		in.Name,
		string(in.Type),
		in.Description,
		postgres.NOW(),
	).WHERE(
		table.WorkoutPresets.ID.EQ(postgres.UUID(id)).
			AND(table.WorkoutPresets.UserID.EQ(postgres.UUID(userID))),
	).RETURNING(table.WorkoutPresets.AllColumns)

	var presetRow model.WorkoutPresets
	if err = updateStmt.QueryContext(ctx, tx, &presetRow); err != nil {
		if err == qrm.ErrNoRows {
			err = fmt.Errorf("preset not found")
			return nil, err
		}
		return nil, fmt.Errorf("update preset: %w", err)
	}

	// Replace exercises.
	delStmt := table.WorkoutPresetExercises.DELETE().
		WHERE(table.WorkoutPresetExercises.PresetID.EQ(postgres.UUID(id)))
	if _, err = delStmt.ExecContext(ctx, tx); err != nil {
		return nil, fmt.Errorf("delete preset exercises: %w", err)
	}

	if err = r.insertPresetExercises(ctx, tx, id, in.Exercises); err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	exRows, err := r.fetchPresetExercises(ctx, r.db, id)
	if err != nil {
		return nil, err
	}
	return toPreset(presetRow, exRows), nil
}

func (r *Repository) DeletePreset(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	stmt := table.WorkoutPresets.DELETE().WHERE(
		table.WorkoutPresets.ID.EQ(postgres.UUID(id)).
			AND(table.WorkoutPresets.UserID.EQ(postgres.UUID(userID))),
	)
	res, err := stmt.ExecContext(ctx, r.db)
	if err != nil {
		return fmt.Errorf("delete preset: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil || n == 0 {
		return fmt.Errorf("preset not found")
	}
	return nil
}

// ── Schedule ──────────────────────────────────────────────────────────────────

func (r *Repository) GetSchedule(ctx context.Context, userID uuid.UUID) ([]*workout.ScheduleEntry, error) {
	schedStmt := postgres.SELECT(table.WorkoutSchedule.AllColumns).
		FROM(table.WorkoutSchedule).
		WHERE(table.WorkoutSchedule.UserID.EQ(postgres.UUID(userID))).
		ORDER_BY(table.WorkoutSchedule.DayOfWeek.ASC())

	var schedRows []model.WorkoutSchedule
	if err := schedStmt.QueryContext(ctx, r.db, &schedRows); err != nil {
		return nil, fmt.Errorf("get schedule: %w", err)
	}

	if len(schedRows) == 0 {
		return []*workout.ScheduleEntry{}, nil
	}

	// Fetch preset names/types for the schedule entries.
	presetIDs := make([]postgres.Expression, 0, len(schedRows))
	seen := map[uuid.UUID]bool{}
	for _, s := range schedRows {
		if !seen[s.PresetID] {
			seen[s.PresetID] = true
			presetIDs = append(presetIDs, postgres.UUID(s.PresetID))
		}
	}

	pStmt := postgres.SELECT(table.WorkoutPresets.AllColumns).
		FROM(table.WorkoutPresets).
		WHERE(table.WorkoutPresets.ID.IN(presetIDs...))

	var presetRows []model.WorkoutPresets
	if err := pStmt.QueryContext(ctx, r.db, &presetRows); err != nil {
		return nil, fmt.Errorf("get schedule presets: %w", err)
	}

	presetByID := make(map[uuid.UUID]model.WorkoutPresets, len(presetRows))
	for _, p := range presetRows {
		presetByID[p.ID] = p
	}

	entries := make([]*workout.ScheduleEntry, len(schedRows))
	for i, s := range schedRows {
		p := presetByID[s.PresetID]
		entries[i] = toScheduleEntry(s, p.Name, p.Type)
	}
	return entries, nil
}

func (r *Repository) SetSchedule(ctx context.Context, userID uuid.UUID, entries []workout.ScheduleEntryInput) ([]*workout.ScheduleEntry, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	// Delete all existing schedule entries for this user.
	delStmt := table.WorkoutSchedule.DELETE().
		WHERE(table.WorkoutSchedule.UserID.EQ(postgres.UUID(userID)))
	if _, err = delStmt.ExecContext(ctx, tx); err != nil {
		return nil, fmt.Errorf("clear schedule: %w", err)
	}

	// Insert new entries (if any).
	if len(entries) > 0 {
		insertStmt := table.WorkoutSchedule.INSERT(
			table.WorkoutSchedule.UserID,
			table.WorkoutSchedule.DayOfWeek,
			table.WorkoutSchedule.PresetID,
		)
		for _, e := range entries {
			insertStmt = insertStmt.VALUES(
				postgres.UUID(userID),
				int32(e.DayOfWeek),
				postgres.UUID(e.PresetID),
			)
		}
		if _, err = insertStmt.ExecContext(ctx, tx); err != nil {
			return nil, fmt.Errorf("insert schedule: %w", err)
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	return r.GetSchedule(ctx, userID)
}

// ── Sessions ──────────────────────────────────────────────────────────────────

func (r *Repository) ListSessions(ctx context.Context, userID uuid.UUID, from, to time.Time) ([]*workout.Session, error) {
	stmt := postgres.SELECT(table.WorkoutSessions.AllColumns).
		FROM(table.WorkoutSessions).
		WHERE(
			table.WorkoutSessions.UserID.EQ(postgres.UUID(userID)).
				AND(table.WorkoutSessions.PerformedAt.GT_EQ(postgres.DateT(from))).
				AND(table.WorkoutSessions.PerformedAt.LT_EQ(postgres.DateT(to))),
		).
		ORDER_BY(table.WorkoutSessions.PerformedAt.DESC(), table.WorkoutSessions.CreatedAt.DESC())

	var sessionRows []model.WorkoutSessions
	if err := stmt.QueryContext(ctx, r.db, &sessionRows); err != nil {
		return nil, fmt.Errorf("list sessions: %w", err)
	}

	if len(sessionRows) == 0 {
		return []*workout.Session{}, nil
	}

	// Batch-load exercises.
	sessionIDs := make([]postgres.Expression, len(sessionRows))
	for i, s := range sessionRows {
		sessionIDs[i] = postgres.UUID(s.ID)
	}

	exStmt := postgres.SELECT(table.WorkoutSessionExercises.AllColumns).
		FROM(table.WorkoutSessionExercises).
		WHERE(table.WorkoutSessionExercises.SessionID.IN(sessionIDs...)).
		ORDER_BY(table.WorkoutSessionExercises.OrderIndex.ASC())

	var exRows []model.WorkoutSessionExercises
	if err := exStmt.QueryContext(ctx, r.db, &exRows); err != nil {
		return nil, fmt.Errorf("list session exercises: %w", err)
	}

	exBySession := make(map[uuid.UUID][]model.WorkoutSessionExercises)
	for _, ex := range exRows {
		exBySession[ex.SessionID] = append(exBySession[ex.SessionID], ex)
	}

	sessions := make([]*workout.Session, len(sessionRows))
	for i, s := range sessionRows {
		sessions[i] = toSession(s, exBySession[s.ID])
	}
	return sessions, nil
}

func (r *Repository) GetSession(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*workout.Session, error) {
	stmt := postgres.SELECT(table.WorkoutSessions.AllColumns).
		FROM(table.WorkoutSessions).
		WHERE(
			table.WorkoutSessions.ID.EQ(postgres.UUID(id)).
				AND(table.WorkoutSessions.UserID.EQ(postgres.UUID(userID))),
		)

	var dest model.WorkoutSessions
	if err := stmt.QueryContext(ctx, r.db, &dest); err != nil {
		if err == qrm.ErrNoRows {
			return nil, fmt.Errorf("session not found")
		}
		return nil, fmt.Errorf("get session: %w", err)
	}

	exRows, err := r.fetchSessionExercises(ctx, r.db, id)
	if err != nil {
		return nil, err
	}
	return toSession(dest, exRows), nil
}

func (r *Repository) GenerateSession(ctx context.Context, userID uuid.UUID, dateStr string) (*workout.Session, error) {
	// Resolve target date.
	date := time.Now().UTC().Truncate(24 * time.Hour)
	if dateStr != "" {
		t, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			return nil, fmt.Errorf("invalid date format: %w", err)
		}
		date = t
	}

	// Find the preset scheduled for that weekday.
	dayOfWeek := int32(date.Weekday()) // 0=Sun … 6=Sat

	schedStmt := postgres.SELECT(table.WorkoutSchedule.AllColumns).
		FROM(table.WorkoutSchedule).
		WHERE(
			table.WorkoutSchedule.UserID.EQ(postgres.UUID(userID)).
				AND(table.WorkoutSchedule.DayOfWeek.EQ(postgres.Int32(dayOfWeek))),
		)

	var schedRow model.WorkoutSchedule
	if err := schedStmt.QueryContext(ctx, r.db, &schedRow); err != nil {
		if err == qrm.ErrNoRows {
			return nil, fmt.Errorf("no preset scheduled for this day")
		}
		return nil, fmt.Errorf("get schedule: %w", err)
	}

	// Fetch the preset (scoped to this user, no user_id check on preset since
	// the schedule entry guarantees it belongs to the user).
	presetStmt := postgres.SELECT(table.WorkoutPresets.AllColumns).
		FROM(table.WorkoutPresets).
		WHERE(table.WorkoutPresets.ID.EQ(postgres.UUID(schedRow.PresetID)))

	var presetRow model.WorkoutPresets
	if err := presetStmt.QueryContext(ctx, r.db, &presetRow); err != nil {
		return nil, fmt.Errorf("get scheduled preset: %w", err)
	}

	presetExRows, err := r.fetchPresetExercises(ctx, r.db, presetRow.ID)
	if err != nil {
		return nil, err
	}

	return r.createSessionFromPreset(ctx, userID, presetRow, presetExRows, date, nil)
}

func (r *Repository) CreateSession(ctx context.Context, userID uuid.UUID, in workout.CreateSessionInput) (*workout.Session, error) {
	date := time.Now().UTC().Truncate(24 * time.Hour)
	if in.Date != "" {
		t, err := time.Parse("2006-01-02", in.Date)
		if err != nil {
			return nil, fmt.Errorf("invalid date format: %w", err)
		}
		date = t
	}

	// Quick start: no preset — create a standalone empty session.
	if in.PresetID == nil {
		return r.createQuickSession(ctx, userID, *in.Type, in.Name, date)
	}

	// Start from preset: fetch and copy.
	presetStmt := postgres.SELECT(table.WorkoutPresets.AllColumns).
		FROM(table.WorkoutPresets).
		WHERE(
			table.WorkoutPresets.ID.EQ(postgres.UUID(*in.PresetID)).
				AND(table.WorkoutPresets.UserID.EQ(postgres.UUID(userID))),
		)

	var presetRow model.WorkoutPresets
	if err := presetStmt.QueryContext(ctx, r.db, &presetRow); err != nil {
		if err == qrm.ErrNoRows {
			return nil, fmt.Errorf("preset not found")
		}
		return nil, fmt.Errorf("get preset: %w", err)
	}

	presetExRows, err := r.fetchPresetExercises(ctx, r.db, presetRow.ID)
	if err != nil {
		return nil, err
	}

	return r.createSessionFromPreset(ctx, userID, presetRow, presetExRows, date, in.Name)
}

// createQuickSession inserts a standalone workout_sessions row with no preset and no exercises.
func (r *Repository) createQuickSession(
	ctx context.Context,
	userID uuid.UUID,
	typ workout.Type,
	nameOverride *string,
	date time.Time,
) (*workout.Session, error) {
	sessionName := string(typ)
	if nameOverride != nil && *nameOverride != "" {
		sessionName = *nameOverride
	}

	insertStmt := table.WorkoutSessions.INSERT(
		table.WorkoutSessions.UserID,
		table.WorkoutSessions.Name,
		table.WorkoutSessions.Type,
		table.WorkoutSessions.PerformedAt,
	).VALUES(
		postgres.UUID(userID),
		sessionName,
		string(typ),
		postgres.DateT(date),
	).RETURNING(table.WorkoutSessions.AllColumns)

	var sessionRow model.WorkoutSessions
	if err := insertStmt.QueryContext(ctx, r.db, &sessionRow); err != nil {
		return nil, fmt.Errorf("create quick session: %w", err)
	}

	return toSession(sessionRow, nil), nil
}

func (r *Repository) AddSessionExercise(ctx context.Context, sessionID uuid.UUID, userID uuid.UUID, in workout.AddSessionExerciseInput) (*workout.SessionExercise, error) {
	if err := r.assertSessionMutable(ctx, sessionID, userID); err != nil {
		return nil, err
	}

	// Compute next order_index.
	type maxRow struct {
		Max *int32
	}
	maxStmt := postgres.SELECT(
		postgres.COALESCE(
			postgres.MAX(table.WorkoutSessionExercises.OrderIndex),
			postgres.Int32(-1),
		).AS("max"),
	).FROM(table.WorkoutSessionExercises).
		WHERE(table.WorkoutSessionExercises.SessionID.EQ(postgres.UUID(sessionID)))

	var mr maxRow
	if err := maxStmt.QueryContext(ctx, r.db, &mr); err != nil {
		return nil, fmt.Errorf("compute order_index: %w", err)
	}
	nextIndex := int32(0)
	if mr.Max != nil {
		nextIndex = *mr.Max + 1
	}

	section := string(in.Section)
	if section == "" {
		section = string(workout.SectionMain)
	}

	var targetSets *int32
	if in.TargetSets != nil {
		s := int32(*in.TargetSets)
		targetSets = &s
	}
	var targetReps *int32
	if in.TargetReps != nil {
		rp := int32(*in.TargetReps)
		targetReps = &rp
	}
	var targetDuration *int32
	if in.TargetDurationSeconds != nil {
		d := int32(*in.TargetDurationSeconds)
		targetDuration = &d
	}
	var restSeconds *int32
	if in.RestSeconds != nil {
		rs := int32(*in.RestSeconds)
		restSeconds = &rs
	}

	insertStmt := table.WorkoutSessionExercises.INSERT(
		table.WorkoutSessionExercises.SessionID,
		table.WorkoutSessionExercises.Section,
		table.WorkoutSessionExercises.OrderIndex,
		table.WorkoutSessionExercises.Name,
		table.WorkoutSessionExercises.TargetMuscles,
		table.WorkoutSessionExercises.Instructions,
		table.WorkoutSessionExercises.TargetSets,
		table.WorkoutSessionExercises.TargetReps,
		table.WorkoutSessionExercises.TargetDurationSeconds,
		table.WorkoutSessionExercises.RestSeconds,
	).VALUES(
		postgres.UUID(sessionID),
		section,
		nextIndex,
		in.Name,
		in.TargetMuscles,
		in.Instructions,
		targetSets,
		targetReps,
		targetDuration,
		restSeconds,
	).RETURNING(table.WorkoutSessionExercises.AllColumns)

	var dest model.WorkoutSessionExercises
	if err := insertStmt.QueryContext(ctx, r.db, &dest); err != nil {
		return nil, fmt.Errorf("add session exercise: %w", err)
	}

	ex := toSessionExercise(dest)
	return &ex, nil
}

func (r *Repository) DeleteSessionExercise(ctx context.Context, exID uuid.UUID, sessionID uuid.UUID, userID uuid.UUID) error {
	if err := r.assertSessionMutable(ctx, sessionID, userID); err != nil {
		return err
	}

	stmt := table.WorkoutSessionExercises.DELETE().
		WHERE(
			table.WorkoutSessionExercises.ID.EQ(postgres.UUID(exID)).
				AND(
					table.WorkoutSessionExercises.SessionID.IN(
						postgres.SELECT(table.WorkoutSessions.ID).
							FROM(table.WorkoutSessions).
							WHERE(
								table.WorkoutSessions.ID.EQ(postgres.UUID(sessionID)).
									AND(table.WorkoutSessions.UserID.EQ(postgres.UUID(userID))),
							),
					),
				),
		)
	res, err := stmt.ExecContext(ctx, r.db)
	if err != nil {
		return fmt.Errorf("delete session exercise: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil || n == 0 {
		return fmt.Errorf("session exercise not found")
	}
	return nil
}

func (r *Repository) UpdateSession(ctx context.Context, id uuid.UUID, userID uuid.UUID, in workout.UpdateSessionInput) (*workout.Session, error) {
	if err := r.assertSessionMutable(ctx, id, userID); err != nil {
		return nil, err
	}

	var durationMinutes *int32
	if in.DurationMinutes != nil {
		d := int32(*in.DurationMinutes)
		durationMinutes = &d
	}

	stmt := table.WorkoutSessions.UPDATE(
		table.WorkoutSessions.Name,
		table.WorkoutSessions.DurationMinutes,
		table.WorkoutSessions.Notes,
		table.WorkoutSessions.UpdatedAt,
	).SET(
		in.Name,
		durationMinutes,
		in.Notes,
		postgres.NOW(),
	).WHERE(
		table.WorkoutSessions.ID.EQ(postgres.UUID(id)).
			AND(table.WorkoutSessions.UserID.EQ(postgres.UUID(userID))),
	).RETURNING(table.WorkoutSessions.AllColumns)

	var dest model.WorkoutSessions
	if err := stmt.QueryContext(ctx, r.db, &dest); err != nil {
		if err == qrm.ErrNoRows {
			return nil, fmt.Errorf("session not found")
		}
		return nil, fmt.Errorf("update session: %w", err)
	}

	exRows, err := r.fetchSessionExercises(ctx, r.db, id)
	if err != nil {
		return nil, err
	}
	return toSession(dest, exRows), nil
}

func (r *Repository) UpdateSessionExercise(ctx context.Context, id uuid.UUID, sessionID uuid.UUID, userID uuid.UUID, in workout.UpdateSessionExerciseInput) (*workout.SessionExercise, error) {
	if err := r.assertSessionMutable(ctx, sessionID, userID); err != nil {
		return nil, err
	}

	var actualSets *int32
	if in.ActualSets != nil {
		s := int32(*in.ActualSets)
		actualSets = &s
	}
	var actualReps *int32
	if in.ActualReps != nil {
		r := int32(*in.ActualReps)
		actualReps = &r
	}
	var actualDurationSeconds *int32
	if in.ActualDurationSeconds != nil {
		d := int32(*in.ActualDurationSeconds)
		actualDurationSeconds = &d
	}
	var weightKg *decimal.Decimal
	if in.WeightKg != nil {
		w := decimal.NewFromFloat(*in.WeightKg)
		weightKg = &w
	}

	// Update only if the exercise belongs to a session owned by this user.
	stmt := table.WorkoutSessionExercises.UPDATE(
		table.WorkoutSessionExercises.ActualSets,
		table.WorkoutSessionExercises.ActualReps,
		table.WorkoutSessionExercises.ActualDurationSeconds,
		table.WorkoutSessionExercises.WeightKg,
		table.WorkoutSessionExercises.Completed,
		table.WorkoutSessionExercises.Notes,
	).SET(
		actualSets,
		actualReps,
		actualDurationSeconds,
		weightKg,
		in.Completed,
		in.Notes,
	).WHERE(
		table.WorkoutSessionExercises.ID.EQ(postgres.UUID(id)).
			AND(
				table.WorkoutSessionExercises.SessionID.IN(
					postgres.SELECT(table.WorkoutSessions.ID).
						FROM(table.WorkoutSessions).
						WHERE(
							table.WorkoutSessions.ID.EQ(postgres.UUID(sessionID)).
								AND(table.WorkoutSessions.UserID.EQ(postgres.UUID(userID))),
						),
				),
			),
	).RETURNING(table.WorkoutSessionExercises.AllColumns)

	var dest model.WorkoutSessionExercises
	if err := stmt.QueryContext(ctx, r.db, &dest); err != nil {
		if err == qrm.ErrNoRows {
			return nil, fmt.Errorf("session exercise not found")
		}
		return nil, fmt.Errorf("update session exercise: %w", err)
	}

	ex := toSessionExercise(dest)
	return &ex, nil
}

func (r *Repository) BulkUpdateSessionExercises(ctx context.Context, sessionID uuid.UUID, userID uuid.UUID, items []workout.BulkUpdateSessionExerciseItem) ([]workout.SessionExercise, error) {
	if len(items) == 0 {
		return []workout.SessionExercise{}, nil
	}

	if err := r.assertSessionMutable(ctx, sessionID, userID); err != nil {
		return nil, err
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	results := make([]workout.SessionExercise, 0, len(items))
	for _, item := range items {
		var actualSets *int32
		if item.ActualSets != nil {
			s := int32(*item.ActualSets)
			actualSets = &s
		}
		var actualReps *int32
		if item.ActualReps != nil {
			rp := int32(*item.ActualReps)
			actualReps = &rp
		}
		var actualDurationSeconds *int32
		if item.ActualDurationSeconds != nil {
			d := int32(*item.ActualDurationSeconds)
			actualDurationSeconds = &d
		}
		var weightKg *decimal.Decimal
		if item.WeightKg != nil {
			w := decimal.NewFromFloat(*item.WeightKg)
			weightKg = &w
		}

		stmt := table.WorkoutSessionExercises.UPDATE(
			table.WorkoutSessionExercises.ActualSets,
			table.WorkoutSessionExercises.ActualReps,
			table.WorkoutSessionExercises.ActualDurationSeconds,
			table.WorkoutSessionExercises.WeightKg,
			table.WorkoutSessionExercises.Completed,
			table.WorkoutSessionExercises.Notes,
		).SET(
			actualSets,
			actualReps,
			actualDurationSeconds,
			weightKg,
			item.Completed,
			item.Notes,
		).WHERE(
			table.WorkoutSessionExercises.ID.EQ(postgres.UUID(item.ID)).
				AND(
					table.WorkoutSessionExercises.SessionID.IN(
						postgres.SELECT(table.WorkoutSessions.ID).
							FROM(table.WorkoutSessions).
							WHERE(
								table.WorkoutSessions.ID.EQ(postgres.UUID(sessionID)).
									AND(table.WorkoutSessions.UserID.EQ(postgres.UUID(userID))),
							),
					),
				),
		).RETURNING(table.WorkoutSessionExercises.AllColumns)

		var dest model.WorkoutSessionExercises
		if err = stmt.QueryContext(ctx, tx, &dest); err != nil {
			if err == qrm.ErrNoRows {
				err = fmt.Errorf("session exercise %s not found", item.ID)
				return nil, err
			}
			return nil, fmt.Errorf("update session exercise %s: %w", item.ID, err)
		}
		ex := toSessionExercise(dest)
		results = append(results, ex)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}
	return results, nil
}

func (r *Repository) DeleteSession(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	if err := r.assertSessionMutable(ctx, id, userID); err != nil {
		return err
	}
	stmt := table.WorkoutSessions.DELETE().WHERE(
		table.WorkoutSessions.ID.EQ(postgres.UUID(id)).
			AND(table.WorkoutSessions.UserID.EQ(postgres.UUID(userID))),
	)
	res, err := stmt.ExecContext(ctx, r.db)
	if err != nil {
		return fmt.Errorf("delete session: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil || n == 0 {
		return fmt.Errorf("session not found")
	}
	return nil
}

func (r *Repository) FinishSession(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*workout.Session, error) {
	if err := r.assertSessionMutable(ctx, id, userID); err != nil {
		return nil, err
	}

	stmt := table.WorkoutSessions.UPDATE(
		table.WorkoutSessions.CompletedAt,
		table.WorkoutSessions.UpdatedAt,
	).SET(
		postgres.NOW(),
		postgres.NOW(),
	).WHERE(
		table.WorkoutSessions.ID.EQ(postgres.UUID(id)).
			AND(table.WorkoutSessions.UserID.EQ(postgres.UUID(userID))),
	).RETURNING(table.WorkoutSessions.AllColumns)

	var dest model.WorkoutSessions
	if err := stmt.QueryContext(ctx, r.db, &dest); err != nil {
		if err == qrm.ErrNoRows {
			return nil, fmt.Errorf("session not found")
		}
		return nil, fmt.Errorf("finish session: %w", err)
	}

	exRows, err := r.fetchSessionExercises(ctx, r.db, id)
	if err != nil {
		return nil, err
	}
	return toSession(dest, exRows), nil
}

// ── Helpers ───────────────────────────────────────────────────────────────────

// db is the interface satisfied by both *sql.DB and *sql.Tx, allowing go-jet
// statements to run inside a transaction or directly against the pool.
type db interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
}

// assertSessionMutable returns an error if the session doesn't exist or is already completed.
func (r *Repository) assertSessionMutable(ctx context.Context, sessionID uuid.UUID, userID uuid.UUID) error {
	stmt := postgres.SELECT(table.WorkoutSessions.ID, table.WorkoutSessions.CompletedAt).
		FROM(table.WorkoutSessions).
		WHERE(
			table.WorkoutSessions.ID.EQ(postgres.UUID(sessionID)).
				AND(table.WorkoutSessions.UserID.EQ(postgres.UUID(userID))),
		)
	var dest model.WorkoutSessions
	if err := stmt.QueryContext(ctx, r.db, &dest); err != nil {
		if err == qrm.ErrNoRows {
			return fmt.Errorf("session not found")
		}
		return fmt.Errorf("check session: %w", err)
	}
	if dest.CompletedAt != nil {
		return fmt.Errorf("session is already completed")
	}
	return nil
}

func (r *Repository) fetchPresetExercises(ctx context.Context, qdb db, presetID uuid.UUID) ([]model.WorkoutPresetExercises, error) {
	stmt := postgres.SELECT(table.WorkoutPresetExercises.AllColumns).
		FROM(table.WorkoutPresetExercises).
		WHERE(table.WorkoutPresetExercises.PresetID.EQ(postgres.UUID(presetID))).
		ORDER_BY(table.WorkoutPresetExercises.OrderIndex.ASC())

	var rows []model.WorkoutPresetExercises
	if err := stmt.QueryContext(ctx, qdb, &rows); err != nil {
		return nil, fmt.Errorf("fetch preset exercises: %w", err)
	}
	return rows, nil
}

func (r *Repository) fetchSessionExercises(ctx context.Context, qdb db, sessionID uuid.UUID) ([]model.WorkoutSessionExercises, error) {
	stmt := postgres.SELECT(table.WorkoutSessionExercises.AllColumns).
		FROM(table.WorkoutSessionExercises).
		WHERE(table.WorkoutSessionExercises.SessionID.EQ(postgres.UUID(sessionID))).
		ORDER_BY(table.WorkoutSessionExercises.OrderIndex.ASC())

	var rows []model.WorkoutSessionExercises
	if err := stmt.QueryContext(ctx, qdb, &rows); err != nil {
		return nil, fmt.Errorf("fetch session exercises: %w", err)
	}
	return rows, nil
}

// insertPresetExercises bulk-inserts exercises for a preset inside qdb (tx or db).
// order_index is derived from the slice position.
func (r *Repository) insertPresetExercises(ctx context.Context, qdb db, presetID uuid.UUID, exercises []workout.PresetExerciseInput) error {
	if len(exercises) == 0 {
		return nil
	}

	insertStmt := table.WorkoutPresetExercises.INSERT(
		table.WorkoutPresetExercises.PresetID,
		table.WorkoutPresetExercises.Section,
		table.WorkoutPresetExercises.OrderIndex,
		table.WorkoutPresetExercises.Name,
		table.WorkoutPresetExercises.TargetMuscles,
		table.WorkoutPresetExercises.Instructions,
		table.WorkoutPresetExercises.Sets,
		table.WorkoutPresetExercises.Reps,
		table.WorkoutPresetExercises.DurationSeconds,
		table.WorkoutPresetExercises.RestSeconds,
		table.WorkoutPresetExercises.WeightKg,
		table.WorkoutPresetExercises.Equipment,
		table.WorkoutPresetExercises.Notes,
	)

	for i, ex := range exercises {
		section := string(ex.Section)
		if section == "" {
			section = string(workout.SectionMain)
		}
		var sets *int32
		if ex.Sets != nil {
			s := int32(*ex.Sets)
			sets = &s
		}
		var reps *int32
		if ex.Reps != nil {
			rp := int32(*ex.Reps)
			reps = &rp
		}
		var durationSeconds *int32
		if ex.DurationSeconds != nil {
			d := int32(*ex.DurationSeconds)
			durationSeconds = &d
		}
		var restSeconds *int32
		if ex.RestSeconds != nil {
			rs := int32(*ex.RestSeconds)
			restSeconds = &rs
		}
		var weightKg *decimal.Decimal
		if ex.WeightKg != nil {
			w := decimal.NewFromFloat(*ex.WeightKg)
			weightKg = &w
		}

		insertStmt = insertStmt.VALUES(
			postgres.UUID(presetID),
			section,
			int32(i),
			ex.Name,
			ex.TargetMuscles,
			ex.Instructions,
			sets,
			reps,
			durationSeconds,
			restSeconds,
			weightKg,
			ex.Equipment,
			ex.Notes,
		)
	}

	if _, err := insertStmt.ExecContext(ctx, qdb); err != nil {
		return fmt.Errorf("insert preset exercises: %w", err)
	}
	return nil
}

// createSessionFromPreset inserts a workout_sessions row and copies all preset
// exercises into workout_session_exercises (targets from preset, actuals blank).
func (r *Repository) createSessionFromPreset(
	ctx context.Context,
	userID uuid.UUID,
	presetRow model.WorkoutPresets,
	presetExRows []model.WorkoutPresetExercises,
	date time.Time,
	nameOverride *string,
) (*workout.Session, error) {
	sessionName := presetRow.Name
	if nameOverride != nil && *nameOverride != "" {
		sessionName = *nameOverride
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	presetID := presetRow.ID
	insertSession := table.WorkoutSessions.INSERT(
		table.WorkoutSessions.UserID,
		table.WorkoutSessions.PresetID,
		table.WorkoutSessions.Name,
		table.WorkoutSessions.Type,
		table.WorkoutSessions.PerformedAt,
	).VALUES(
		postgres.UUID(userID),
		postgres.UUID(presetID),
		sessionName,
		presetRow.Type,
		postgres.DateT(date),
	).RETURNING(table.WorkoutSessions.AllColumns)

	var sessionRow model.WorkoutSessions
	if err = insertSession.QueryContext(ctx, tx, &sessionRow); err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}

	// Copy preset exercises → session exercises.
	if len(presetExRows) > 0 {
		insertEx := table.WorkoutSessionExercises.INSERT(
			table.WorkoutSessionExercises.SessionID,
			table.WorkoutSessionExercises.Section,
			table.WorkoutSessionExercises.OrderIndex,
			table.WorkoutSessionExercises.Name,
			table.WorkoutSessionExercises.TargetMuscles,
			table.WorkoutSessionExercises.Instructions,
			table.WorkoutSessionExercises.TargetSets,
			table.WorkoutSessionExercises.TargetReps,
			table.WorkoutSessionExercises.TargetDurationSeconds,
			table.WorkoutSessionExercises.RestSeconds,
		)
		for _, ex := range presetExRows {
			insertEx = insertEx.VALUES(
				postgres.UUID(sessionRow.ID),
				ex.Section,
				ex.OrderIndex,
				ex.Name,
				ex.TargetMuscles,
				ex.Instructions,
				ex.Sets,            // sets → target_sets
				ex.Reps,            // reps → target_reps
				ex.DurationSeconds, // duration_seconds → target_duration_seconds
				ex.RestSeconds,
			)
		}
		if _, err = insertEx.ExecContext(ctx, tx); err != nil {
			return nil, fmt.Errorf("insert session exercises: %w", err)
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	exRows, err := r.fetchSessionExercises(ctx, r.db, sessionRow.ID)
	if err != nil {
		return nil, err
	}
	return toSession(sessionRow, exRows), nil
}

// ── Model → domain mappers ────────────────────────────────────────────────────

func toPreset(m model.WorkoutPresets, exercises []model.WorkoutPresetExercises) *workout.Preset {
	p := &workout.Preset{
		ID:          m.ID,
		UserID:      m.UserID,
		Name:        m.Name,
		Type:        workout.Type(m.Type),
		Description: m.Description,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
	if exercises == nil {
		p.Exercises = []workout.PresetExercise{}
	} else {
		p.Exercises = make([]workout.PresetExercise, len(exercises))
		for i, ex := range exercises {
			p.Exercises[i] = toPresetExercise(ex)
		}
	}
	return p
}

func toPresetExercise(m model.WorkoutPresetExercises) workout.PresetExercise {
	ex := workout.PresetExercise{
		ID:            m.ID,
		PresetID:      m.PresetID,
		Section:       workout.Section(m.Section),
		OrderIndex:    int(m.OrderIndex),
		Name:          m.Name,
		TargetMuscles: m.TargetMuscles,
		Instructions:  m.Instructions,
		Equipment:     m.Equipment,
		Notes:         m.Notes,
	}
	if m.Sets != nil {
		s := int(*m.Sets)
		ex.Sets = &s
	}
	if m.Reps != nil {
		rp := int(*m.Reps)
		ex.Reps = &rp
	}
	if m.DurationSeconds != nil {
		d := int(*m.DurationSeconds)
		ex.DurationSeconds = &d
	}
	if m.RestSeconds != nil {
		rs := int(*m.RestSeconds)
		ex.RestSeconds = &rs
	}
	if m.WeightKg != nil {
		w, _ := m.WeightKg.Float64()
		ex.WeightKg = &w
	}
	return ex
}

func toScheduleEntry(m model.WorkoutSchedule, presetName, presetType string) *workout.ScheduleEntry {
	return &workout.ScheduleEntry{
		ID:         m.ID,
		UserID:     m.UserID,
		DayOfWeek:  int(m.DayOfWeek),
		PresetID:   m.PresetID,
		PresetName: presetName,
		PresetType: workout.Type(presetType),
		CreatedAt:  m.CreatedAt,
	}
}

func toSession(m model.WorkoutSessions, exercises []model.WorkoutSessionExercises) *workout.Session {
	s := &workout.Session{
		ID:          m.ID,
		UserID:      m.UserID,
		PresetID:    m.PresetID,
		Name:        m.Name,
		Type:        workout.Type(m.Type),
		PerformedAt: m.PerformedAt,
		Notes:       m.Notes,
		CompletedAt: m.CompletedAt,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
	if m.DurationMinutes != nil {
		d := int(*m.DurationMinutes)
		s.DurationMinutes = &d
	}
	if exercises == nil {
		s.Exercises = []workout.SessionExercise{}
	} else {
		s.Exercises = make([]workout.SessionExercise, len(exercises))
		for i, ex := range exercises {
			s.Exercises[i] = toSessionExercise(ex)
		}
	}
	return s
}

func toSessionExercise(m model.WorkoutSessionExercises) workout.SessionExercise {
	ex := workout.SessionExercise{
		ID:            m.ID,
		SessionID:     m.SessionID,
		Section:       workout.Section(m.Section),
		OrderIndex:    int(m.OrderIndex),
		Name:          m.Name,
		TargetMuscles: m.TargetMuscles,
		Instructions:  m.Instructions,
		Completed:     m.Completed,
		Notes:         m.Notes,
	}
	if m.TargetSets != nil {
		s := int(*m.TargetSets)
		ex.TargetSets = &s
	}
	if m.TargetReps != nil {
		rp := int(*m.TargetReps)
		ex.TargetReps = &rp
	}
	if m.TargetDurationSeconds != nil {
		d := int(*m.TargetDurationSeconds)
		ex.TargetDurationSeconds = &d
	}
	if m.RestSeconds != nil {
		rs := int(*m.RestSeconds)
		ex.RestSeconds = &rs
	}
	if m.ActualSets != nil {
		s := int(*m.ActualSets)
		ex.ActualSets = &s
	}
	if m.ActualReps != nil {
		rp := int(*m.ActualReps)
		ex.ActualReps = &rp
	}
	if m.ActualDurationSeconds != nil {
		d := int(*m.ActualDurationSeconds)
		ex.ActualDurationSeconds = &d
	}
	if m.WeightKg != nil {
		w, _ := m.WeightKg.Float64()
		ex.WeightKg = &w
	}
	return ex
}
