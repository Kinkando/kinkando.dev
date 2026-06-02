package health

import (
	"time"

	"github.com/google/uuid"
)

type Gender string

const (
	GenderMale   Gender = "male"
	GenderFemale Gender = "female"
	GenderOther  Gender = "other"
)

type Goal string

const (
	GoalLoseWeight Goal = "lose_weight"
	GoalMaintain   Goal = "maintain"
	GoalGainMuscle Goal = "gain_muscle"
)

type Profile struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	Height    *float64  `json:"height"` // cm
	Age       *int      `json:"age"`
	Gender    *Gender   `json:"gender"`
	Goal      *Goal     `json:"goal"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UpsertProfileInput struct {
	Height *float64 `json:"height" validate:"omitempty,gt=0"`
	Age    *int     `json:"age"    validate:"omitempty,gt=0"`
	Gender *Gender  `json:"gender" validate:"omitempty,oneof=male female other"`
	Goal   *Goal    `json:"goal"   validate:"omitempty,oneof=lose_weight maintain gain_muscle"`
}

type WeightLog struct {
	ID       uuid.UUID `json:"id"`
	Weight   float64   `json:"weight"` // kg
	LoggedAt time.Time `json:"logged_at"`
}

type CreateWeightInput struct {
	Weight   float64 `json:"weight"    validate:"gt=0"`
	LoggedAt string  `json:"logged_at"` // YYYY-MM-DD, optional
}

// ── Food ─────────────────────────────────────────────────────────────────────

type MealType string

const (
	MealTypeBreakfast MealType = "breakfast"
	MealTypeLunch     MealType = "lunch"
	MealTypeDinner    MealType = "dinner"
	MealTypeSnack     MealType = "snack"
)

type FoodLog struct {
	ID         uuid.UUID `json:"id"`
	UserID     uuid.UUID `json:"user_id"`
	Name       string    `json:"name"`
	MealType   MealType  `json:"meal_type"`
	Calories   *int      `json:"calories"`
	ProteinG   *float64  `json:"protein_g"`
	CarbsG     *float64  `json:"carbs_g"`
	FatG       *float64  `json:"fat_g"`
	Notes      *string   `json:"notes"`
	ConsumedAt time.Time `json:"consumed_at"`
	CreatedAt  time.Time `json:"created_at"`
}

type CreateFoodInput struct {
	Name       string   `json:"name"         validate:"required"`
	MealType   MealType `json:"meal_type"    validate:"required,oneof=breakfast lunch dinner snack"`
	Calories   *int     `json:"calories"     validate:"omitempty,min=0"`
	ProteinG   *float64 `json:"protein_g"`
	CarbsG     *float64 `json:"carbs_g"`
	FatG       *float64 `json:"fat_g"`
	Notes      *string  `json:"notes"`
	ConsumedAt string   `json:"consumed_at"` // YYYY-MM-DD, optional
}

type UpdateFoodInput = CreateFoodInput

// ── Sleep ─────────────────────────────────────────────────────────────────────

type SleepLog struct {
	ID              uuid.UUID `json:"id"`
	UserID          uuid.UUID `json:"user_id"`
	StartedAt       time.Time `json:"started_at"`       // bedtime
	EndedAt         time.Time `json:"ended_at"`         // wake time
	DurationMinutes int       `json:"duration_minutes"` // derived: (ended_at - started_at)
	Score           *int      `json:"score"`            // 0–100 (Samsung Health)
	Notes           *string   `json:"notes"`
	LoggedAt        time.Time `json:"logged_at"` // night-of date for grouping
	CreatedAt       time.Time `json:"created_at"`
}

type CreateSleepInput struct {
	StartedAt string  `json:"started_at" validate:"required"` // RFC3339
	EndedAt   string  `json:"ended_at"   validate:"required"` // RFC3339
	Score     *int    `json:"score"      validate:"omitempty,min=0,max=100"` // 0–100, optional
	Notes     *string `json:"notes"`
	LoggedAt  string  `json:"logged_at"` // YYYY-MM-DD, optional (defaults to started_at date)
}

type UpdateSleepInput = CreateSleepInput
