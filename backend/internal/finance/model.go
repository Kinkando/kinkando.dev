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

type Record struct {
	ID        uuid.UUID  `json:"id"`
	UserID    uuid.UUID  `json:"user_id"`
	Type      RecordType `json:"type"`
	Amount    float64    `json:"amount"`
	Category  string     `json:"category"`
	Note      string     `json:"note"`
	Date      time.Time  `json:"date"`
	CreatedAt time.Time  `json:"created_at"`
}

type CreateRecordInput struct {
	Type     RecordType `json:"type"`
	Amount   float64    `json:"amount"`
	Category string     `json:"category"`
	Note     string     `json:"note"`
	Date     string     `json:"date"` // YYYY-MM-DD
}

type MonthlySummary struct {
	Month      string            `json:"month"`
	Income     float64           `json:"income"`
	Expense    float64           `json:"expense"`
	Net        float64           `json:"net"`
	Categories []CategorySummary `json:"categories"`
}

type CategorySummary struct {
	Category string     `json:"category"`
	Type     RecordType `json:"type"`
	Total    float64    `json:"total"`
}
