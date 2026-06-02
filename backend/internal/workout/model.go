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
	TypeRunning        Type = "running"
	TypeWalking        Type = "walking"
	TypeCardio         Type = "cardio"
	TypeMobility       Type = "mobility"
	TypeCustom         Type = "custom"
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
	Section         Section  `json:"section"          validate:"omitempty,oneof=warmup main cooldown"`
	Name            string   `json:"name"             validate:"required"`
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
	Name        string                `json:"name"        validate:"required"`
	Type        Type                  `json:"type"        validate:"required,oneof=weight_training body_weight running walking cardio mobility"`
	Description *string               `json:"description"`
	Exercises   []PresetExerciseInput `json:"exercises"   validate:"dive"`
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
	DayOfWeek int       `json:"day_of_week" validate:"min=0,max=6"`
	PresetID  uuid.UUID `json:"preset_id"   validate:"required"`
}

type SetScheduleInput struct {
	Entries []ScheduleEntryInput `json:"entries" validate:"dive"`
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
	CompletedAt     *time.Time        `json:"completed_at"`
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
	Type     *Type      `json:"type"      validate:"omitempty,oneof=weight_training body_weight running walking cardio mobility custom"` // required when preset_id is nil (quick start)
	Date     string     `json:"date"` // YYYY-MM-DD, optional
	Name     *string    `json:"name"` // overrides preset name if provided
}

// ── Session exercise DTOs ──────────────────────────────────────────────────────

type AddSessionExerciseInput struct {
	Section               Section `json:"section"       validate:"omitempty,oneof=warmup main cooldown"`
	Name                  string  `json:"name"          validate:"required"`
	TargetMuscles         *string `json:"target_muscles"`
	Instructions          *string `json:"instructions"`
	TargetSets            *int    `json:"target_sets"`
	TargetReps            *int    `json:"target_reps"`
	TargetDurationSeconds *int    `json:"target_duration_seconds"`
	RestSeconds           *int    `json:"rest_seconds"`
}

type UpdateSessionInput struct {
	Name            string  `json:"name"             validate:"required"`
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

type BulkUpdateSessionExerciseItem struct {
	ID                    uuid.UUID `json:"id"                     validate:"required"`
	ActualSets            *int      `json:"actual_sets"`
	ActualReps            *int      `json:"actual_reps"`
	ActualDurationSeconds *int      `json:"actual_duration_seconds"`
	WeightKg              *float64  `json:"weight_kg"`
	Completed             bool      `json:"completed"`
	Notes                 *string   `json:"notes"`
}

type BulkUpdateSessionExercisesInput struct {
	Items []BulkUpdateSessionExerciseItem `json:"items" validate:"required,min=1,dive"`
}
