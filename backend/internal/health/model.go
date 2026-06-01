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

type ExerciseType string

const (
	ExerciseTypeCardio      ExerciseType = "cardio"
	ExerciseTypeStrength    ExerciseType = "strength"
	ExerciseTypeFlexibility ExerciseType = "flexibility"
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
	Height *float64 `json:"height"`
	Age    *int     `json:"age"`
	Gender *Gender  `json:"gender"`
	Goal   *Goal    `json:"goal"`
}

type WeightLog struct {
	ID       uuid.UUID `json:"id"`
	Weight   float64   `json:"weight"` // kg
	LoggedAt time.Time `json:"logged_at"`
}

type CreateWeightInput struct {
	Weight   float64 `json:"weight"`
	LoggedAt string  `json:"logged_at"` // YYYY-MM-DD, optional
}

type Exercise struct {
	ID              uuid.UUID    `json:"id"`
	UserID          uuid.UUID    `json:"user_id"`
	Name            string       `json:"name"`
	Type            ExerciseType `json:"type"`
	DurationMinutes *int         `json:"duration_minutes"`
	Calories        *int         `json:"calories"`
	Notes           *string      `json:"notes"`
	PerformedAt     time.Time    `json:"performed_at"`
	CreatedAt       time.Time    `json:"created_at"`
}

type CreateExerciseInput struct {
	Name            string       `json:"name"`
	Type            ExerciseType `json:"type"`
	DurationMinutes *int         `json:"duration_minutes"`
	Calories        *int         `json:"calories"`
	Notes           *string      `json:"notes"`
	PerformedAt     string       `json:"performed_at"` // YYYY-MM-DD
}

type UpdateExerciseInput struct {
	Name            string       `json:"name"`
	Type            ExerciseType `json:"type"`
	DurationMinutes *int         `json:"duration_minutes"`
	Calories        *int         `json:"calories"`
	Notes           *string      `json:"notes"`
	PerformedAt     string       `json:"performed_at"` // YYYY-MM-DD
}
