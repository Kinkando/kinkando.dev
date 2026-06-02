package medicine

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// ErrInsufficientStock is returned by Take when stock would go negative
// and allow_negative is false.
var ErrInsufficientStock = errors.New("insufficient stock")

// ── Enums ─────────────────────────────────────────────────────────────────────

type FrequencyType string

const (
	FrequencyTypeDaily    FrequencyType = "daily"
	FrequencyTypeWeekly   FrequencyType = "weekly"
	FrequencyTypeAsNeeded FrequencyType = "as_needed"
	FrequencyTypeCustom   FrequencyType = "custom"
)

type Timing string

const (
	TimingBeforeMeal      Timing = "before_meal"
	TimingAfterMeal       Timing = "after_meal"
	TimingBeforeBreakfast Timing = "before_breakfast"
	TimingBeforeBed       Timing = "before_bed"
	TimingAnytime         Timing = "anytime"
)

type IntakeStatus string

const (
	IntakeStatusTaken   IntakeStatus = "taken"
	IntakeStatusSkipped IntakeStatus = "skipped"
	IntakeStatusMissed  IntakeStatus = "missed"
)

type AdjustmentType string

const (
	AdjustmentTypeAdd        AdjustmentType = "add"
	AdjustmentTypeRemove     AdjustmentType = "remove"
	AdjustmentTypeCorrection AdjustmentType = "correction"
)

// ── DTOs ──────────────────────────────────────────────────────────────────────

type Medicine struct {
	ID                uuid.UUID  `json:"id"`
	UserID            uuid.UUID  `json:"user_id"`
	Name              string     `json:"name"`
	GenericName       *string    `json:"generic_name"`
	Description       *string    `json:"description"`
	StockQuantity     float64    `json:"stock_quantity"`
	StockUnit         string     `json:"stock_unit"`
	DosageAmount      float64    `json:"dosage_amount"`
	DosageUnit        *string    `json:"dosage_unit"`
	FrequencyType     FrequencyType `json:"frequency_type"`
	FrequencyValue    *int       `json:"frequency_value"`
	Timing            *Timing    `json:"timing"`
	StartDate         *time.Time `json:"start_date"`
	EndDate           *time.Time `json:"end_date"`
	LowStockThreshold float64    `json:"low_stock_threshold"`
	Note              *string    `json:"note"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
	ArchivedAt        *time.Time `json:"archived_at"`
}

type MedicineIntake struct {
	ID           uuid.UUID    `json:"id"`
	MedicineID   uuid.UUID    `json:"medicine_id"`
	UserID       uuid.UUID    `json:"user_id"`
	MedicineName string       `json:"medicine_name"`
	TakenAt      time.Time    `json:"taken_at"`
	QuantityTaken float64     `json:"quantity_taken"`
	StockBefore  float64      `json:"stock_before"`
	StockAfter   float64      `json:"stock_after"`
	Status       IntakeStatus `json:"status"`
	Note         *string      `json:"note"`
	CreatedAt    time.Time    `json:"created_at"`
}

type MedicineStockAdjustment struct {
	ID          uuid.UUID      `json:"id"`
	MedicineID  uuid.UUID      `json:"medicine_id"`
	UserID      uuid.UUID      `json:"user_id"`
	Type        AdjustmentType `json:"type"`
	Quantity    float64        `json:"quantity"`
	StockBefore float64        `json:"stock_before"`
	StockAfter  float64        `json:"stock_after"`
	Reason      *string        `json:"reason"`
	CreatedAt   time.Time      `json:"created_at"`
}

// ── Inputs ────────────────────────────────────────────────────────────────────

type CreateMedicineInput struct {
	Name              string        `json:"name"`
	GenericName       *string       `json:"generic_name"`
	Description       *string       `json:"description"`
	StockQuantity     float64       `json:"stock_quantity"`
	StockUnit         string        `json:"stock_unit"`
	DosageAmount      float64       `json:"dosage_amount"`
	DosageUnit        *string       `json:"dosage_unit"`
	FrequencyType     FrequencyType `json:"frequency_type"`
	FrequencyValue    *int          `json:"frequency_value"`
	Timing            *Timing       `json:"timing"`
	StartDate         string        `json:"start_date"` // YYYY-MM-DD, optional
	EndDate           string        `json:"end_date"`   // YYYY-MM-DD, optional
	LowStockThreshold *float64      `json:"low_stock_threshold"` // defaults to 7
	Note              *string       `json:"note"`
}

type UpdateMedicineInput = CreateMedicineInput

type TakeMedicineInput struct {
	QuantityTaken float64      `json:"quantity_taken"`
	Status        *IntakeStatus `json:"status"` // defaults to "taken"
	Note          *string      `json:"note"`
	TakenAt       *string      `json:"taken_at"`      // RFC3339, optional (defaults to now)
	AllowNegative bool         `json:"allow_negative"` // override negative-stock guard
}

type AdjustStockInput struct {
	Type     AdjustmentType `json:"type"`
	Quantity float64        `json:"quantity"`
	Reason   *string        `json:"reason"`
}

// ── List options ──────────────────────────────────────────────────────────────

type ListIntakeOpts struct {
	MedicineID *uuid.UUID
	Date       *time.Time // filter to a single calendar day (UTC)
	Limit      int        // 0 → default (50)
}

type ListAdjustmentOpts struct {
	MedicineID *uuid.UUID
	Date       *time.Time
	Limit      int
}
