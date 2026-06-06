package repository

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"time"

	"github.com/go-jet/jet/v2/postgres"
	"github.com/go-jet/jet/v2/qrm"
	"github.com/google/uuid"
	"github.com/kinkando/personal-dashboard/gen/kinkando/public/model"
	"github.com/kinkando/personal-dashboard/gen/kinkando/public/table"
	"github.com/kinkando/personal-dashboard/internal/quest"
)

type Repository struct {
	db *sql.DB
}

func New(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// ── Quest CRUD ────────────────────────────────────────────────────────────────

func (r *Repository) CreateQuest(ctx context.Context, userID uuid.UUID, in quest.CreateQuestInput) (*quest.Quest, error) {
	stmt := table.QuestDefinitions.INSERT(
		table.QuestDefinitions.UserID,
		table.QuestDefinitions.Type,
		table.QuestDefinitions.SourceType,
		table.QuestDefinitions.Title,
		table.QuestDefinitions.Description,
		table.QuestDefinitions.XpReward,
		table.QuestDefinitions.TargetCount,
	).VALUES(
		postgres.UUID(userID),
		string(in.Type),
		string(in.SourceType),
		in.Title,
		in.Description,
		in.XPReward,
		in.TargetCount,
	).RETURNING(table.QuestDefinitions.AllColumns)

	var dest model.QuestDefinitions
	if err := stmt.QueryContext(ctx, r.db, &dest); err != nil {
		return nil, fmt.Errorf("create quest: %w", err)
	}
	q := toQuest(dest)
	return &q, nil
}

func (r *Repository) ListQuests(ctx context.Context, userID uuid.UUID, questType string) ([]*quest.Quest, error) {
	cond := table.QuestDefinitions.UserID.EQ(postgres.UUID(userID))
	if questType != "" {
		cond = cond.AND(table.QuestDefinitions.Type.EQ(postgres.String(questType)))
	}

	stmt := postgres.SELECT(table.QuestDefinitions.AllColumns).
		FROM(table.QuestDefinitions).
		WHERE(cond).
		ORDER_BY(table.QuestDefinitions.CreatedAt.ASC())

	var dest []model.QuestDefinitions
	if err := stmt.QueryContext(ctx, r.db, &dest); err != nil {
		return nil, fmt.Errorf("list quests: %w", err)
	}
	quests := make([]*quest.Quest, len(dest))
	for i, d := range dest {
		q := toQuest(d)
		quests[i] = &q
	}
	return quests, nil
}

func (r *Repository) GetQuest(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*quest.Quest, error) {
	stmt := postgres.SELECT(table.QuestDefinitions.AllColumns).
		FROM(table.QuestDefinitions).
		WHERE(
			table.QuestDefinitions.ID.EQ(postgres.UUID(id)).
				AND(table.QuestDefinitions.UserID.EQ(postgres.UUID(userID))),
		)

	var dest model.QuestDefinitions
	if err := stmt.QueryContext(ctx, r.db, &dest); err != nil {
		if err == qrm.ErrNoRows {
			return nil, fmt.Errorf("quest not found")
		}
		return nil, fmt.Errorf("get quest: %w", err)
	}
	q := toQuest(dest)
	return &q, nil
}

func (r *Repository) UpdateQuest(ctx context.Context, id uuid.UUID, userID uuid.UUID, in quest.UpdateQuestInput) (*quest.Quest, error) {
	stmt := table.QuestDefinitions.UPDATE(
		table.QuestDefinitions.SourceType,
		table.QuestDefinitions.Title,
		table.QuestDefinitions.Description,
		table.QuestDefinitions.XpReward,
		table.QuestDefinitions.TargetCount,
		table.QuestDefinitions.IsActive,
		table.QuestDefinitions.UpdatedAt,
	).SET(
		string(in.SourceType),
		in.Title,
		in.Description,
		in.XPReward,
		in.TargetCount,
		in.IsActive,
		time.Now().UTC(),
	).WHERE(
		table.QuestDefinitions.ID.EQ(postgres.UUID(id)).
			AND(table.QuestDefinitions.UserID.EQ(postgres.UUID(userID))),
	).RETURNING(table.QuestDefinitions.AllColumns)

	var dest model.QuestDefinitions
	if err := stmt.QueryContext(ctx, r.db, &dest); err != nil {
		if err == qrm.ErrNoRows {
			return nil, fmt.Errorf("quest not found")
		}
		return nil, fmt.Errorf("update quest: %w", err)
	}
	q := toQuest(dest)
	return &q, nil
}

func (r *Repository) DeleteQuest(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	stmt := table.QuestDefinitions.DELETE().WHERE(
		table.QuestDefinitions.ID.EQ(postgres.UUID(id)).
			AND(table.QuestDefinitions.UserID.EQ(postgres.UUID(userID))),
	)
	res, err := stmt.ExecContext(ctx, r.db)
	if err != nil {
		return fmt.Errorf("delete quest: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil || n == 0 {
		return fmt.Errorf("quest not found")
	}
	return nil
}

func (r *Repository) SetActive(ctx context.Context, id uuid.UUID, userID uuid.UUID, active bool) (*quest.Quest, error) {
	stmt := table.QuestDefinitions.UPDATE(
		table.QuestDefinitions.IsActive,
		table.QuestDefinitions.UpdatedAt,
	).SET(
		active,
		time.Now().UTC(),
	).WHERE(
		table.QuestDefinitions.ID.EQ(postgres.UUID(id)).
			AND(table.QuestDefinitions.UserID.EQ(postgres.UUID(userID))),
	).RETURNING(table.QuestDefinitions.AllColumns)

	var dest model.QuestDefinitions
	if err := stmt.QueryContext(ctx, r.db, &dest); err != nil {
		if err == qrm.ErrNoRows {
			return nil, fmt.Errorf("quest not found")
		}
		return nil, fmt.Errorf("set active: %w", err)
	}
	q := toQuest(dest)
	return &q, nil
}

// ── Overview queries ──────────────────────────────────────────────────────────

func (r *Repository) GetQuestStatus(ctx context.Context, userID uuid.UUID, questType quest.QuestType, today time.Time) ([]*quest.QuestStatus, error) {
	countExpr := postgres.COUNT(table.QuestCompletions.ID)

	stmt := postgres.SELECT(
		table.QuestDefinitions.AllColumns,
		countExpr.AS("current_count"),
	).FROM(
		table.QuestDefinitions.LEFT_JOIN(
			table.QuestCompletions,
			table.QuestCompletions.QuestID.EQ(table.QuestDefinitions.ID).
				AND(table.QuestCompletions.PeriodStart.EQ(postgres.DateT(today))).
				AND(table.QuestCompletions.UserID.EQ(postgres.UUID(userID))),
		),
	).WHERE(
		table.QuestDefinitions.UserID.EQ(postgres.UUID(userID)).
			AND(table.QuestDefinitions.Type.EQ(postgres.String(string(questType)))),
	).GROUP_BY(
		table.QuestDefinitions.ID,
		table.QuestDefinitions.UserID,
		table.QuestDefinitions.Type,
		table.QuestDefinitions.SourceType,
		table.QuestDefinitions.Title,
		table.QuestDefinitions.Description,
		table.QuestDefinitions.XpReward,
		table.QuestDefinitions.TargetCount,
		table.QuestDefinitions.IsActive,
		table.QuestDefinitions.CreatedAt,
		table.QuestDefinitions.UpdatedAt,
	).ORDER_BY(table.QuestDefinitions.CreatedAt.ASC())

	var rows []struct {
		model.QuestDefinitions
		CurrentCount int64 `alias:"current_count"`
	}
	if err := stmt.QueryContext(ctx, r.db, &rows); err != nil {
		return nil, fmt.Errorf("quest status: %w", err)
	}

	result := make([]*quest.QuestStatus, len(rows))
	for i, row := range rows {
		q := toQuest(row.QuestDefinitions)
		count := int(row.CurrentCount)
		result[i] = &quest.QuestStatus{
			Quest:        q,
			CurrentCount: count,
			Completed:    count >= q.TargetCount,
		}
	}

	// Sink completed quests to the bottom while preserving created_at ASC within each group.
	sort.SliceStable(result, func(i, j int) bool {
		return !result[i].Completed && result[j].Completed
	})

	return result, nil
}

func (r *Repository) TotalXP(ctx context.Context, userID uuid.UUID) (int, error) {
	sumExpr := postgres.COALESCE(postgres.SUM(table.UserXpEvents.Xp), postgres.Int(0))
	stmt := postgres.SELECT(sumExpr.AS("total_xp")).
		FROM(table.UserXpEvents).
		WHERE(table.UserXpEvents.UserID.EQ(postgres.UUID(userID)))

	var result struct {
		TotalXP int64 `alias:"total_xp"`
	}
	if err := stmt.QueryContext(ctx, r.db, &result); err != nil {
		return 0, fmt.Errorf("total xp: %w", err)
	}
	return int(result.TotalXP), nil
}

// ── Completion ────────────────────────────────────────────────────────────────

// Increment adds one completion row for the given period and grants XP once when
// the target is first reached. source is the XP-event label ("daily" or "weekly").
// It applies identically to both daily (periodStart = today) and weekly (periodStart = weekStart).
func (r *Repository) Increment(ctx context.Context, userID uuid.UUID, questID uuid.UUID, periodStart time.Time, source string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	// Get quest to validate ownership, get target + XP reward + title.
	q, err := getQuestTx(ctx, tx, questID, userID)
	if err != nil {
		return err
	}

	// Insert a completion row (no uniqueness constraint — multiple allowed per period).
	insComp := table.QuestCompletions.INSERT(
		table.QuestCompletions.UserID,
		table.QuestCompletions.QuestID,
		table.QuestCompletions.PeriodStart,
	).VALUES(postgres.UUID(userID), postgres.UUID(questID), postgres.DateT(periodStart))
	if _, err := insComp.ExecContext(ctx, tx); err != nil {
		return fmt.Errorf("insert completion: %w", err)
	}

	// Recount.
	count, err := countCompletionsTx(ctx, tx, questID, periodStart)
	if err != nil {
		return err
	}

	// Grant XP once when target is first reached (idempotent via ON CONFLICT DO NOTHING).
	if q.XPReward > 0 && count >= q.TargetCount {
		if err := insertXPEvent(ctx, tx, userID, questID, q.Title, source, periodStart, q.XPReward); err != nil {
			return err
		}
	}

	// Reconcile all-set bonus for this quest type and period.
	if err := reconcileBonusTx(ctx, tx, userID, q.Type, periodStart); err != nil {
		return err
	}

	return tx.Commit()
}

// Decrement removes the most recent completion row for the given period and revokes
// XP if the count drops below the target. source mirrors Increment's label.
func (r *Repository) Decrement(ctx context.Context, userID uuid.UUID, questID uuid.UUID, periodStart time.Time) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	// Validate ownership (also get target for XP reconciliation).
	q, err := getQuestTx(ctx, tx, questID, userID)
	if err != nil {
		return err
	}

	// Check current count before decrementing.
	count, err := countCompletionsTx(ctx, tx, questID, periodStart)
	if err != nil {
		return err
	}
	if count == 0 {
		return fmt.Errorf("cannot decrement below 0")
	}

	// Delete the most recent completion row for this period.
	selectLatest := postgres.SELECT(table.QuestCompletions.ID).FROM(table.QuestCompletions).WHERE(
		table.QuestCompletions.QuestID.EQ(postgres.UUID(questID)).
			AND(table.QuestCompletions.PeriodStart.EQ(postgres.DateT(periodStart))).
			AND(table.QuestCompletions.UserID.EQ(postgres.UUID(userID))),
	).ORDER_BY(table.QuestCompletions.CreatedAt.DESC()).LIMIT(1)

	var latestRow model.QuestCompletions
	if err := selectLatest.QueryContext(ctx, tx, &latestRow); err != nil {
		return fmt.Errorf("find latest completion: %w", err)
	}
	delComp := table.QuestCompletions.DELETE().WHERE(
		table.QuestCompletions.ID.EQ(postgres.UUID(latestRow.ID)),
	)
	if _, err := delComp.ExecContext(ctx, tx); err != nil {
		return fmt.Errorf("delete completion: %w", err)
	}

	// Remaining count after decrement.
	remaining := count - 1

	// Revoke XP if progress dropped below target.
	if q.XPReward > 0 && remaining < q.TargetCount {
		if err := deleteXPEvent(ctx, tx, questID, periodStart); err != nil {
			return err
		}
	}

	// Reconcile all-set bonus for this quest type and period.
	if err := reconcileBonusTx(ctx, tx, userID, q.Type, periodStart); err != nil {
		return err
	}

	return tx.Commit()
}

// ── Source-driven auto-progress ───────────────────────────────────────────────

// ProgressBySource advances all active quests linked to sourceType for the user.
// For daily quests, it marks them completed today (idempotent).
// For weekly quests, it increments the count by one and grants XP when the target is reached.
func (r *Repository) ProgressBySource(ctx context.Context, userID uuid.UUID, sourceType string, today, weekStart time.Time) error {
	// Fetch all active quests matching this source.
	stmt := postgres.SELECT(table.QuestDefinitions.AllColumns).
		FROM(table.QuestDefinitions).
		WHERE(
			table.QuestDefinitions.UserID.EQ(postgres.UUID(userID)).
				AND(table.QuestDefinitions.SourceType.EQ(postgres.String(sourceType))).
				AND(table.QuestDefinitions.IsActive.IS_TRUE()),
		)

	var defs []model.QuestDefinitions
	if err := stmt.QueryContext(ctx, r.db, &defs); err != nil {
		return fmt.Errorf("progress by source: fetch quests: %w", err)
	}

	for _, def := range defs {
		q := toQuest(def)
		switch q.Type {
		case quest.QuestTypeDaily:
			if err := r.Increment(ctx, userID, q.ID, today, string(q.Type)); err != nil {
				return fmt.Errorf("progress by source: increment daily %s: %w", q.ID, err)
			}
		case quest.QuestTypeWeekly:
			if err := r.Increment(ctx, userID, q.ID, weekStart, string(q.Type)); err != nil {
				return fmt.Errorf("progress by source: increment weekly %s: %w", q.ID, err)
			}
		}
	}
	return nil
}

// ── Batch reminder helpers ────────────────────────────────────────────────────

// CountIncompleteByUser returns a map of userID → number of active quests of
// the given type that still have fewer completions than their target for the
// supplied period. It scans across all users — used exclusively by the cron
// reminder job, never by the per-user API.
func (r *Repository) CountIncompleteByUser(ctx context.Context, questType string, periodStart time.Time) (map[uuid.UUID]int, error) {
	stmt := postgres.SELECT(
		table.QuestDefinitions.UserID,
		table.QuestDefinitions.TargetCount,
		postgres.COUNT(table.QuestCompletions.ID).AS("current_count"),
	).FROM(
		table.QuestDefinitions.LEFT_JOIN(
			table.QuestCompletions,
			table.QuestCompletions.QuestID.EQ(table.QuestDefinitions.ID).
				AND(table.QuestCompletions.UserID.EQ(table.QuestDefinitions.UserID)).
				AND(table.QuestCompletions.PeriodStart.EQ(postgres.DateT(periodStart))),
		),
	).WHERE(
		table.QuestDefinitions.IsActive.IS_TRUE().
			AND(table.QuestDefinitions.Type.EQ(postgres.String(questType))),
	).GROUP_BY(
		table.QuestDefinitions.UserID,
		table.QuestDefinitions.ID,
		table.QuestDefinitions.TargetCount,
	)

	var rows []struct {
		UserID       uuid.UUID `alias:"quest_definitions.user_id"`
		TargetCount  int32     `alias:"quest_definitions.target_count"`
		CurrentCount int64     `alias:"current_count"`
	}
	if err := stmt.QueryContext(ctx, r.db, &rows); err != nil {
		return nil, fmt.Errorf("count incomplete quests: %w", err)
	}

	result := make(map[uuid.UUID]int)
	for _, row := range rows {
		if int(row.CurrentCount) < int(row.TargetCount) {
			result[row.UserID]++
		}
	}
	return result, nil
}

// RecordPeriodResults snapshots the final status of every active quest of
// questType for the given periodStart. It is called by the quest-period-snapshot
// cron job once per period (daily: yesterday; weekly: Monday for the prior week).
// Each row is inserted idempotent via ON CONFLICT (quest_id, period_start) DO NOTHING,
// so repeated invocations within the same period are safe.
func (r *Repository) RecordPeriodResults(ctx context.Context, questType string, periodStart time.Time) (*quest.PeriodSnapshotResult, error) {
	// Query every active quest of questType alongside its completion count for the period.
	stmt := postgres.SELECT(
		table.QuestDefinitions.UserID,
		table.QuestDefinitions.ID,
		table.QuestDefinitions.Title,
		table.QuestDefinitions.TargetCount,
		table.QuestDefinitions.XpReward,
		postgres.COUNT(table.QuestCompletions.ID).AS("current_count"),
	).FROM(
		table.QuestDefinitions.LEFT_JOIN(
			table.QuestCompletions,
			table.QuestCompletions.QuestID.EQ(table.QuestDefinitions.ID).
				AND(table.QuestCompletions.UserID.EQ(table.QuestDefinitions.UserID)).
				AND(table.QuestCompletions.PeriodStart.EQ(postgres.DateT(periodStart))),
		),
	).WHERE(
		table.QuestDefinitions.IsActive.IS_TRUE().
			AND(table.QuestDefinitions.Type.EQ(postgres.String(questType))),
	).GROUP_BY(
		table.QuestDefinitions.UserID,
		table.QuestDefinitions.ID,
		table.QuestDefinitions.Title,
		table.QuestDefinitions.TargetCount,
		table.QuestDefinitions.XpReward,
	)

	var rows []struct {
		UserID       uuid.UUID `alias:"quest_definitions.user_id"`
		ID           uuid.UUID `alias:"quest_definitions.id"`
		Title        string    `alias:"quest_definitions.title"`
		TargetCount  int32     `alias:"quest_definitions.target_count"`
		XpReward     int32     `alias:"quest_definitions.xp_reward"`
		CurrentCount int64     `alias:"current_count"`
	}
	if err := stmt.QueryContext(ctx, r.db, &rows); err != nil {
		return nil, fmt.Errorf("record period results: fetch quests: %w", err)
	}

	res := &quest.PeriodSnapshotResult{Total: len(rows)}
	if len(rows) == 0 {
		return res, nil
	}

	// Build a bulk INSERT with all rows; idempotent via ON CONFLICT DO NOTHING.
	ins := table.QuestPeriodResults.INSERT(
		table.QuestPeriodResults.UserID,
		table.QuestPeriodResults.QuestID,
		table.QuestPeriodResults.QuestTitle,
		table.QuestPeriodResults.Type,
		table.QuestPeriodResults.PeriodStart,
		table.QuestPeriodResults.TargetCount,
		table.QuestPeriodResults.CompletedCount,
		table.QuestPeriodResults.Completed,
		table.QuestPeriodResults.XpReward,
	)
	for _, row := range rows {
		completed := int(row.CurrentCount) >= int(row.TargetCount)
		if completed {
			res.Completed++
		} else {
			res.Incomplete++
		}
		qid := row.ID // local copy for pointer
		ins = ins.VALUES(
			postgres.UUID(row.UserID),
			postgres.UUID(qid),
			row.Title,
			questType,
			postgres.DateT(periodStart),
			row.TargetCount,
			int32(row.CurrentCount),
			completed,
			row.XpReward,
		)
	}
	ins = ins.ON_CONFLICT(table.QuestPeriodResults.QuestID, table.QuestPeriodResults.PeriodStart).DO_NOTHING()

	sqlRes, err := ins.ExecContext(ctx, r.db)
	if err != nil {
		return nil, fmt.Errorf("record period results: insert: %w", err)
	}
	affected, _ := sqlRes.RowsAffected()
	res.Inserted = int(affected)
	return res, nil
}

// ListDailyResults returns the finalized status of every daily quest snapshot in
// [from, to] for the user — one row per quest per day. The streaks service
// aggregates these by day. Uses idx_quest_period_results_user_period.
func (r *Repository) ListDailyResults(ctx context.Context, userID uuid.UUID, from, to time.Time) ([]quest.PeriodResultRow, error) {
	stmt := postgres.SELECT(
		table.QuestPeriodResults.PeriodStart,
		table.QuestPeriodResults.Completed,
	).FROM(table.QuestPeriodResults).
		WHERE(
			table.QuestPeriodResults.UserID.EQ(postgres.UUID(userID)).
				AND(table.QuestPeriodResults.Type.EQ(postgres.String(string(quest.QuestTypeDaily)))).
				AND(table.QuestPeriodResults.PeriodStart.GT_EQ(postgres.DateT(from))).
				AND(table.QuestPeriodResults.PeriodStart.LT_EQ(postgres.DateT(to))),
		).ORDER_BY(table.QuestPeriodResults.PeriodStart.ASC())

	var dest []model.QuestPeriodResults
	if err := stmt.QueryContext(ctx, r.db, &dest); err != nil {
		return nil, fmt.Errorf("list daily results: %w", err)
	}
	rows := make([]quest.PeriodResultRow, len(dest))
	for i, d := range dest {
		rows[i] = quest.PeriodResultRow{PeriodStart: d.PeriodStart, Completed: d.Completed}
	}
	return rows, nil
}

// ── History ───────────────────────────────────────────────────────────────────

func (r *Repository) ListXPEvents(ctx context.Context, userID uuid.UUID, limit int) ([]*quest.XPEvent, error) {
	stmt := postgres.SELECT(table.UserXpEvents.AllColumns).
		FROM(table.UserXpEvents).
		WHERE(table.UserXpEvents.UserID.EQ(postgres.UUID(userID))).
		ORDER_BY(table.UserXpEvents.CreatedAt.DESC())

	if limit > 0 {
		stmt = stmt.LIMIT(int64(limit))
	}

	var dest []model.UserXpEvents
	if err := stmt.QueryContext(ctx, r.db, &dest); err != nil {
		return nil, fmt.Errorf("list xp events: %w", err)
	}
	events := make([]*quest.XPEvent, len(dest))
	for i, d := range dest {
		events[i] = toXPEvent(d)
	}
	return events, nil
}

// ── Transaction helpers ───────────────────────────────────────────────────────

func getQuestTx(ctx context.Context, db qrm.DB, questID, userID uuid.UUID) (*quest.Quest, error) {
	stmt := postgres.SELECT(table.QuestDefinitions.AllColumns).FROM(table.QuestDefinitions).WHERE(
		table.QuestDefinitions.ID.EQ(postgres.UUID(questID)).
			AND(table.QuestDefinitions.UserID.EQ(postgres.UUID(userID))),
	)
	var dest model.QuestDefinitions
	if err := stmt.QueryContext(ctx, db, &dest); err != nil {
		if err == qrm.ErrNoRows {
			return nil, fmt.Errorf("quest not found")
		}
		return nil, fmt.Errorf("get quest: %w", err)
	}
	q := toQuest(dest)
	return &q, nil
}

func countCompletionsTx(ctx context.Context, db qrm.DB, questID uuid.UUID, periodStart time.Time) (int, error) {
	countExpr := postgres.COUNT(table.QuestCompletions.ID)
	stmt := postgres.SELECT(countExpr.AS("cnt")).FROM(table.QuestCompletions).WHERE(
		table.QuestCompletions.QuestID.EQ(postgres.UUID(questID)).
			AND(table.QuestCompletions.PeriodStart.EQ(postgres.DateT(periodStart))),
	)
	var result struct {
		Cnt int64 `alias:"cnt"`
	}
	if err := stmt.QueryContext(ctx, db, &result); err != nil {
		return 0, fmt.Errorf("count completions: %w", err)
	}
	return int(result.Cnt), nil
}

func insertXPEvent(ctx context.Context, db qrm.DB, userID, questID uuid.UUID, title, source string, periodStart time.Time, xp int) error {
	stmt := table.UserXpEvents.INSERT(
		table.UserXpEvents.UserID,
		table.UserXpEvents.QuestID,
		table.UserXpEvents.QuestTitle,
		table.UserXpEvents.Source,
		table.UserXpEvents.PeriodStart,
		table.UserXpEvents.Xp,
	).VALUES(
		postgres.UUID(userID),
		postgres.UUID(questID),
		title,
		source,
		postgres.DateT(periodStart),
		xp,
	).ON_CONFLICT(table.UserXpEvents.QuestID, table.UserXpEvents.PeriodStart).DO_NOTHING()

	if _, err := stmt.ExecContext(ctx, db); err != nil {
		return fmt.Errorf("insert xp event: %w", err)
	}
	return nil
}

func deleteXPEvent(ctx context.Context, db qrm.DB, questID uuid.UUID, periodStart time.Time) error {
	stmt := table.UserXpEvents.DELETE().WHERE(
		table.UserXpEvents.QuestID.EQ(postgres.UUID(questID)).
			AND(table.UserXpEvents.PeriodStart.EQ(postgres.DateT(periodStart))),
	)
	if _, err := stmt.ExecContext(ctx, db); err != nil {
		return fmt.Errorf("delete xp event: %w", err)
	}
	return nil
}

// ── Bonus XP helpers ──────────────────────────────────────────────────────────

// reconcileBonusTx grants or revokes the all-set bonus XP for questType/periodStart
// inside an existing transaction. Called at the tail of Increment and Decrement so both
// the manual and the event-driven paths stay in sync.
func reconcileBonusTx(ctx context.Context, db qrm.DB, userID uuid.UUID, questType quest.QuestType, periodStart time.Time) error {
	// Count active quests of this type and how many are still incomplete.
	stmt := postgres.SELECT(
		table.QuestDefinitions.ID,
		table.QuestDefinitions.TargetCount,
		postgres.COUNT(table.QuestCompletions.ID).AS("current_count"),
	).FROM(
		table.QuestDefinitions.LEFT_JOIN(
			table.QuestCompletions,
			table.QuestCompletions.QuestID.EQ(table.QuestDefinitions.ID).
				AND(table.QuestCompletions.UserID.EQ(table.QuestDefinitions.UserID)).
				AND(table.QuestCompletions.PeriodStart.EQ(postgres.DateT(periodStart))),
		),
	).WHERE(
		table.QuestDefinitions.UserID.EQ(postgres.UUID(userID)).
			AND(table.QuestDefinitions.Type.EQ(postgres.String(string(questType)))).
			AND(table.QuestDefinitions.IsActive.IS_TRUE()),
	).GROUP_BY(
		table.QuestDefinitions.ID,
		table.QuestDefinitions.TargetCount,
	)

	var rows []struct {
		ID           uuid.UUID `alias:"quest_definitions.id"`
		TargetCount  int32     `alias:"quest_definitions.target_count"`
		CurrentCount int64     `alias:"current_count"`
	}
	if err := stmt.QueryContext(ctx, db, &rows); err != nil {
		return fmt.Errorf("reconcile bonus: fetch quest status: %w", err)
	}

	total := len(rows)
	incomplete := 0
	for _, r := range rows {
		if int(r.CurrentCount) < int(r.TargetCount) {
			incomplete++
		}
	}

	var bonusSource string
	var bonusXP int
	var bonusTitle string
	if questType == quest.QuestTypeDaily {
		bonusSource = quest.SourceDailyBonus
		bonusXP = quest.DailyBonusXP
		bonusTitle = "Daily Bonus"
	} else {
		bonusSource = quest.SourceWeeklyBonus
		bonusXP = quest.WeeklyBonusXP
		bonusTitle = "Weekly Bonus"
	}

	if total > 0 && incomplete == 0 {
		return insertBonusXPEventTx(ctx, db, userID, bonusSource, bonusTitle, periodStart, bonusXP)
	}
	return deleteBonusXPEventTx(ctx, db, userID, bonusSource, periodStart)
}

// insertBonusXPEventTx inserts a bonus XP event with quest_id = NULL (the default).
// It checks for existence first; the partial unique index uq_xp_bonus_period is the
// backstop against any concurrent double-insert.
func insertBonusXPEventTx(ctx context.Context, db qrm.DB, userID uuid.UUID, source, title string, periodStart time.Time, xp int) error {
	// Skip if the bonus row for this period already exists.
	checkStmt := postgres.SELECT(postgres.COUNT(table.UserXpEvents.ID).AS("cnt")).
		FROM(table.UserXpEvents).
		WHERE(
			table.UserXpEvents.UserID.EQ(postgres.UUID(userID)).
				AND(table.UserXpEvents.Source.EQ(postgres.String(source))).
				AND(table.UserXpEvents.PeriodStart.EQ(postgres.DateT(periodStart))).
				AND(table.UserXpEvents.QuestID.IS_NULL()),
		)
	var check struct {
		Cnt int64 `alias:"cnt"`
	}
	if err := checkStmt.QueryContext(ctx, db, &check); err != nil {
		return fmt.Errorf("insert bonus xp event: check existence: %w", err)
	}
	if check.Cnt > 0 {
		return nil // already granted — nothing to do
	}

	// Omit QuestID column so it defaults to NULL.
	insStmt := table.UserXpEvents.INSERT(
		table.UserXpEvents.UserID,
		table.UserXpEvents.QuestTitle,
		table.UserXpEvents.Source,
		table.UserXpEvents.PeriodStart,
		table.UserXpEvents.Xp,
	).VALUES(
		postgres.UUID(userID),
		title,
		source,
		postgres.DateT(periodStart),
		xp,
	)
	if _, err := insStmt.ExecContext(ctx, db); err != nil {
		return fmt.Errorf("insert bonus xp event: %w", err)
	}
	return nil
}

// deleteBonusXPEventTx removes the bonus XP row (quest_id IS NULL) for the given
// user/source/period. It is a no-op when the row doesn't exist.
func deleteBonusXPEventTx(ctx context.Context, db qrm.DB, userID uuid.UUID, source string, periodStart time.Time) error {
	stmt := table.UserXpEvents.DELETE().WHERE(
		table.UserXpEvents.UserID.EQ(postgres.UUID(userID)).
			AND(table.UserXpEvents.Source.EQ(postgres.String(source))).
			AND(table.UserXpEvents.PeriodStart.EQ(postgres.DateT(periodStart))).
			AND(table.UserXpEvents.QuestID.IS_NULL()),
	)
	if _, err := stmt.ExecContext(ctx, db); err != nil {
		return fmt.Errorf("delete bonus xp event: %w", err)
	}
	return nil
}

// ── Mappers ───────────────────────────────────────────────────────────────────

func toQuest(m model.QuestDefinitions) quest.Quest {
	return quest.Quest{
		ID:          m.ID,
		UserID:      m.UserID,
		Type:        quest.QuestType(m.Type),
		SourceType:  quest.SourceType(m.SourceType),
		Title:       m.Title,
		Description: m.Description,
		XPReward:    int(m.XpReward),
		TargetCount: int(m.TargetCount),
		IsActive:    m.IsActive,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

func toXPEvent(m model.UserXpEvents) *quest.XPEvent {
	return &quest.XPEvent{
		ID:          m.ID,
		QuestID:     m.QuestID,
		QuestTitle:  m.QuestTitle,
		Source:      m.Source,
		PeriodStart: m.PeriodStart,
		XP:          int(m.Xp),
		CreatedAt:   m.CreatedAt,
	}
}
