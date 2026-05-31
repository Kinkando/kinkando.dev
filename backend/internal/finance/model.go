package finance

import (
	"time"

	"github.com/google/uuid"
)

type RecordType string

const (
	RecordTypeIncome  RecordType = "income"
	RecordTypeExpense RecordType = "expense"
)

type CategoryRef struct {
	ID    uuid.UUID `json:"id"`
	Name  string    `json:"name"`
	Icon  string    `json:"icon"`
	Color string    `json:"color"`
}

type Record struct {
	ID           uuid.UUID    `json:"id"`
	UserID       uuid.UUID    `json:"user_id"`
	Type         RecordType   `json:"type"`
	Amount       float64      `json:"amount"`
	CategoryID   *uuid.UUID   `json:"category_id"`
	Category     *CategoryRef `json:"category"`
	CategoryName string       `json:"category_name"`
	Note         string       `json:"note"`
	Date         time.Time    `json:"date"`
	CreatedAt    time.Time    `json:"created_at"`
}

type CreateRecordInput struct {
	Type       RecordType `json:"type"`
	Amount     float64    `json:"amount"`
	CategoryID uuid.UUID  `json:"category_id"`
	Note       string     `json:"note"`
	Date       string     `json:"date"` // YYYY-MM-DD
}

type MonthlySummary struct {
	Month      string            `json:"month"`
	Income     float64           `json:"income"`
	Expense    float64           `json:"expense"`
	Net        float64           `json:"net"`
	Categories []CategorySummary `json:"categories"`
}

type CategorySummary struct {
	CategoryID *uuid.UUID `json:"category_id"`
	Category   string     `json:"category"`
	Type       RecordType `json:"type"`
	Total      float64    `json:"total"`
	Icon       string     `json:"icon"`
	Color      string     `json:"color"`
}

type Category struct {
	ID        uuid.UUID  `json:"id"`
	UserID    uuid.UUID  `json:"user_id"`
	Name      string     `json:"name"`
	Type      RecordType `json:"type"`
	Icon      string     `json:"icon"`
	Color     string     `json:"color"`
	CreatedAt time.Time  `json:"created_at"`
}

type CreateCategoryInput struct {
	Name  string     `json:"name"`
	Type  RecordType `json:"type"`
	Icon  string     `json:"icon"`
	Color string     `json:"color"`
}

type UpdateCategoryInput struct {
	Name  string `json:"name"`
	Icon  string `json:"icon"`
	Color string `json:"color"`
}
