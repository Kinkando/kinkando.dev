package quest

import (
	"time"

	"github.com/google/uuid"
)

type QuestType string

const (
	QuestTypeDaily  QuestType = "daily"
	QuestTypeWeekly QuestType = "weekly"
)

type SourceType string

const (
	SourceTypeManual     SourceType = "manual"
	SourceTypeMedicine   SourceType = "medicine"
	SourceTypeWorkout    SourceType = "workout"
	SourceTypeSupplement SourceType = "supplement"
	SourceTypeWeight     SourceType = "weight"
	SourceTypeSleep      SourceType = "sleep"
)

// Bonus XP source labels — written into user_xp_events.source for all-set bonuses.
const (
	SourceDailyBonus  = "daily_bonus"
	SourceWeeklyBonus = "weekly_bonus"
)

// Bonus XP amounts awarded when every active quest of a type is complete for the period.
const (
	DailyBonusXP  = 50
	WeeklyBonusXP = 200
)

type Quest struct {
	ID          uuid.UUID  `json:"id"`
	UserID      uuid.UUID  `json:"user_id"`
	Type        QuestType  `json:"type"`
	SourceType  SourceType `json:"source_type"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	XPReward    int        `json:"xp_reward"`
	TargetCount int        `json:"target_count"`
	IsActive    bool       `json:"is_active"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type CreateQuestInput struct {
	Type        QuestType  `json:"type"         validate:"required,oneof=daily weekly"`
	SourceType  SourceType `json:"source_type"  validate:"omitempty,oneof=manual medicine workout supplement weight sleep"`
	Title       string     `json:"title"        validate:"required"`
	Description string     `json:"description"`
	XPReward    int        `json:"xp_reward"    validate:"min=0"`
	TargetCount int        `json:"target_count" validate:"min=1"`
}

type UpdateQuestInput struct {
	SourceType  SourceType `json:"source_type"  validate:"omitempty,oneof=manual medicine workout supplement weight sleep"`
	Title       string     `json:"title"        validate:"required"`
	Description string     `json:"description"`
	XPReward    int        `json:"xp_reward"    validate:"min=0"`
	TargetCount int        `json:"target_count" validate:"min=1"`
	IsActive    bool       `json:"is_active"`
}

type QuestStatus struct {
	Quest
	CurrentCount int  `json:"current_count"`
	Completed    bool `json:"completed"`
}

type XPSummary struct {
	TotalXP     int `json:"total_xp"`
	Level       int `json:"level"`
	XPIntoLevel int `json:"xp_into_level"`
	XPForLevel  int `json:"xp_for_level"`
	XPToNext    int `json:"xp_to_next"`
}

type Overview struct {
	Date         string         `json:"date"`
	WeekStart    string         `json:"week_start"`
	XP           XPSummary      `json:"xp"`
	Daily        []*QuestStatus `json:"daily"`
	Weekly       []*QuestStatus `json:"weekly"`
	DailyDone    int            `json:"daily_done"`
	DailyTotal   int            `json:"daily_total"`
	WeeklyDone   int            `json:"weekly_done"`
	WeeklyTotal  int            `json:"weekly_total"`
	DailyBonusXP int            `json:"daily_bonus_xp"`
	WeeklyBonusXP int           `json:"weekly_bonus_xp"`
}

type XPEvent struct {
	ID          uuid.UUID  `json:"id"`
	QuestID     *uuid.UUID `json:"quest_id"`
	QuestTitle  string     `json:"quest_title"`
	Source      string     `json:"source"`
	PeriodStart time.Time  `json:"period_start"`
	XP          int        `json:"xp"`
	CreatedAt   time.Time  `json:"created_at"`
}

// PeriodSnapshotResult summarises one execution of RecordPeriodResults.
type PeriodSnapshotResult struct {
	Total      int `json:"total"`      // active quests evaluated
	Completed  int `json:"completed"`  // quests that reached target_count
	Incomplete int `json:"incomplete"` // quests below target_count
	Inserted   int `json:"inserted"`   // rows actually written (0 when already idempotent)
}
