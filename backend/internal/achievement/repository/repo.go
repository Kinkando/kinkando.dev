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
	"github.com/kinkando/personal-dashboard/internal/achievement"
)

type Repository struct {
	db *sql.DB
}

func New(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// ListUnlocked returns the user's unlocked badge codes mapped to their unlock time.
func (r *Repository) ListUnlocked(ctx context.Context, userID uuid.UUID) (map[string]time.Time, error) {
	stmt := postgres.SELECT(
		table.UserAchievements.Code,
		table.UserAchievements.UnlockedAt,
	).FROM(table.UserAchievements).
		WHERE(table.UserAchievements.UserID.EQ(postgres.UUID(userID)))

	var dest []model.UserAchievements
	if err := stmt.QueryContext(ctx, r.db, &dest); err != nil {
		return nil, fmt.Errorf("list unlocked achievements: %w", err)
	}
	out := make(map[string]time.Time, len(dest))
	for _, d := range dest {
		out[d.Code] = d.UnlockedAt
	}
	return out, nil
}

// Unlock records a badge unlock. Idempotent via the (user_id, code) unique
// constraint — repeated calls are no-ops.
func (r *Repository) Unlock(ctx context.Context, userID uuid.UUID, code string, at time.Time) error {
	stmt := table.UserAchievements.INSERT(
		table.UserAchievements.UserID,
		table.UserAchievements.Code,
		table.UserAchievements.UnlockedAt,
	).VALUES(
		postgres.UUID(userID),
		code,
		at,
	).ON_CONFLICT(table.UserAchievements.UserID, table.UserAchievements.Code).DO_NOTHING()

	if _, err := stmt.ExecContext(ctx, r.db); err != nil {
		return fmt.Errorf("unlock achievement %s: %w", code, err)
	}
	return nil
}

// Counts returns the count-based metric values for the user (the metrics not
// derived from the quest service).
func (r *Repository) Counts(ctx context.Context, userID uuid.UUID) (map[achievement.Metric]int, error) {
	uid := postgres.UUID(userID)

	workouts, err := r.count(ctx, table.WorkoutSessions, table.WorkoutSessions.ID,
		table.WorkoutSessions.UserID.EQ(uid).AND(table.WorkoutSessions.CompletedAt.IS_NOT_NULL()))
	if err != nil {
		return nil, err
	}
	weight, err := r.count(ctx, table.HealthWeightLogs, table.HealthWeightLogs.ID,
		table.HealthWeightLogs.UserID.EQ(uid))
	if err != nil {
		return nil, err
	}
	sleep, err := r.count(ctx, table.HealthSleepLogs, table.HealthSleepLogs.ID,
		table.HealthSleepLogs.UserID.EQ(uid))
	if err != nil {
		return nil, err
	}
	meds, err := r.count(ctx, table.MedicineIntakes, table.MedicineIntakes.ID,
		table.MedicineIntakes.UserID.EQ(uid))
	if err != nil {
		return nil, err
	}
	quests, err := r.count(ctx, table.QuestCompletions, table.QuestCompletions.ID,
		table.QuestCompletions.UserID.EQ(uid))
	if err != nil {
		return nil, err
	}

	return map[achievement.Metric]int{
		achievement.MetricWorkouts:        workouts,
		achievement.MetricWeightLogs:      weight,
		achievement.MetricSleepLogs:       sleep,
		achievement.MetricMedicineIntakes: meds,
		achievement.MetricQuestsCompleted: quests,
	}, nil
}

// count runs COUNT(idCol) over from filtered by cond and returns the result.
func (r *Repository) count(ctx context.Context, from postgres.ReadableTable, idCol postgres.Column, cond postgres.BoolExpression) (int, error) {
	stmt := postgres.SELECT(postgres.COUNT(idCol).AS("cnt")).FROM(from).WHERE(cond)
	var result struct {
		Cnt int64 `alias:"cnt"`
	}
	if err := stmt.QueryContext(ctx, r.db, &result); err != nil {
		return 0, fmt.Errorf("count: %w", err)
	}
	return int(result.Cnt), nil
}
