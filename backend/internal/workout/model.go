package workout

import (
	"time"

	"github.com/google/uuid"
)

// ── Enums ─────────────────────────────────────────────────────────────────────

type Type string

const (
	TypeWeightTraining Type = "weight_training"
	TypeBodyWeight     Type = "body_weight"
)

type Section string

const (
	SectionWarmup   Section = "warmup"
	SectionMain     Section = "main"
	SectionCooldown Section = "cooldown"
)

// ── Preset ────────────────────────────────────────────────────────────────────

type PresetExercise struct {
	ID              uuid.UUID `json:"id"`
	PresetID        uuid.UUID `json:"preset_id"`
	Section         Section   `json:"section"`
	OrderIndex      int       `json:"order_index"`
	Name            string    `json:"name"`
	TargetMuscles   *string   `json:"target_muscles"`
	Instructions    *string   `json:"instructions"`
	Sets            *int      `json:"sets"`
	Reps            *int      `json:"reps"`
	DurationSeconds *int      `json:"duration_seconds"`
	RestSeconds     *int      `json:"rest_seconds"`
	WeightKg        *float64  `json:"weight_kg"`
	Equipment       *string   `json:"equipment"`
	Notes           *string   `json:"notes"`
}

type Preset struct {
	ID          uuid.UUID        `json:"id"`
	UserID      uuid.UUID        `json:"user_id"`
	Name        string           `json:"name"`
	Type        Type             `json:"type"`
	Description *string          `json:"description"`
	Exercises   []PresetExercise `json:"exercises"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
}

// ── Preset DTOs ───────────────────────────────────────────────────────────────

type PresetExerciseInput struct {
	Section         Section  `json:"section"`
	Name            string   `json:"name"`
	TargetMuscles   *string  `json:"target_muscles"`
	Instructions    *string  `json:"instructions"`
	Sets            *int     `json:"sets"`
	Reps            *int     `json:"reps"`
	DurationSeconds *int     `json:"duration_seconds"`
	RestSeconds     *int     `json:"rest_seconds"`
	WeightKg        *float64 `json:"weight_kg"`
	Equipment       *string  `json:"equipment"`
	Notes           *string  `json:"notes"`
}

type CreatePresetInput struct {
	Name        string               `json:"name"`
	Type        Type                 `json:"type"`
	Description *string              `json:"description"`
	Exercises   []PresetExerciseInput `json:"exercises"`
}

type UpdatePresetInput = CreatePresetInput

// ── Schedule ──────────────────────────────────────────────────────────────────

type ScheduleEntry struct {
	ID         uuid.UUID `json:"id"`
	UserID     uuid.UUID `json:"user_id"`
	DayOfWeek  int       `json:"day_of_week"` // 0=Sun … 6=Sat
	PresetID   uuid.UUID `json:"preset_id"`
	PresetName string    `json:"preset_name"`
	PresetType Type      `json:"preset_type"`
	CreatedAt  time.Time `json:"created_at"`
}

type ScheduleEntryInput struct {
	DayOfWeek int       `json:"day_of_week"`
	PresetID  uuid.UUID `json:"preset_id"`
}

type SetScheduleInput struct {
	Entries []ScheduleEntryInput `json:"entries"`
}

// ── Session ───────────────────────────────────────────────────────────────────

type SessionExercise struct {
	ID                    uuid.UUID `json:"id"`
	SessionID             uuid.UUID `json:"session_id"`
	Section               Section   `json:"section"`
	OrderIndex            int       `json:"order_index"`
	Name                  string    `json:"name"`
	TargetMuscles         *string   `json:"target_muscles"`
	Instructions          *string   `json:"instructions"`
	TargetSets            *int      `json:"target_sets"`
	TargetReps            *int      `json:"target_reps"`
	TargetDurationSeconds *int      `json:"target_duration_seconds"`
	RestSeconds           *int      `json:"rest_seconds"`
	ActualSets            *int      `json:"actual_sets"`
	ActualReps            *int      `json:"actual_reps"`
	ActualDurationSeconds *int      `json:"actual_duration_seconds"`
	WeightKg              *float64  `json:"weight_kg"`
	Completed             bool      `json:"completed"`
	Notes                 *string   `json:"notes"`
}

type Session struct {
	ID              uuid.UUID         `json:"id"`
	UserID          uuid.UUID         `json:"user_id"`
	PresetID        *uuid.UUID        `json:"preset_id"`
	Name            string            `json:"name"`
	Type            Type              `json:"type"`
	PerformedAt     time.Time         `json:"performed_at"`
	DurationMinutes *int              `json:"duration_minutes"`
	Notes           *string           `json:"notes"`
	Exercises       []SessionExercise `json:"exercises"`
	CreatedAt       time.Time         `json:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at"`
}

// ── Session DTOs ──────────────────────────────────────────────────────────────

type GenerateSessionInput struct {
	Date string `json:"date"` // YYYY-MM-DD, optional (defaults to today)
}

type CreateSessionInput struct {
	PresetID *uuid.UUID `json:"preset_id"`
	Date     string     `json:"date"` // YYYY-MM-DD, optional
	Name     *string    `json:"name"` // overrides preset name if provided
}

type UpdateSessionInput struct {
	Name            string  `json:"name"`
	DurationMinutes *int    `json:"duration_minutes"`
	Notes           *string `json:"notes"`
}

type UpdateSessionExerciseInput struct {
	ActualSets            *int     `json:"actual_sets"`
	ActualReps            *int     `json:"actual_reps"`
	ActualDurationSeconds *int     `json:"actual_duration_seconds"`
	WeightKg              *float64 `json:"weight_kg"`
	Completed             bool     `json:"completed"`
	Notes                 *string  `json:"notes"`
}
