// Package reminder implements the batch cron reminder job for medicines.
// It is invoked via POST /api/v1/cron/medicine-reminders (authenticated with
// a shared CRON_SECRET) and sends one digest notification per user covering:
//
//   - Low-stock alert   (stock_quantity <= low_stock_threshold)
//   - Refill reminder   (estimated days remaining <= 7)
//   - Dose reminder     (scheduled reminder_times slot falls in the past 30 min)
//   - Missed-dose alert (scheduled slot passed >2 h ago with no taken intake)
//
// All four item types are batched into a single notification.Message per user
// so the user receives at most one push per cron run.
package reminder

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/kinkando/personal-dashboard/internal/medicine"
	"github.com/kinkando/personal-dashboard/internal/notification"
	"go.uber.org/zap"
)

// cronInterval is the expected interval between cron runs.  A dose slot is
// considered "due now" when it falls inside (now-cronInterval, now].
const cronInterval = 30 * time.Minute

// missedGrace is how long after a scheduled slot we wait before treating a
// missing intake as a "missed" dose.
const missedGrace = 2 * time.Hour

// supplyDigestHour is the earliest Bangkok hour at which a supply-digest
// (low-stock + refill) notification may be sent for the day.
const supplyDigestHour = 9

// refillDaysThreshold is the estimated-days-remaining value at or below which
// a refill reminder fires (provided stock is still above the low-stock threshold).
const refillDaysThreshold = 7

// ── Repository interface ──────────────────────────────────────────────────────

// MedicineRepository is the narrow data-access interface the reminder service
// depends on. *repository.Repository satisfies it.
type MedicineRepository interface {
	ScanActiveMedicinesForReminders(ctx context.Context) ([]*medicine.Medicine, error)
	LogReminder(ctx context.Context, userID, medicineID uuid.UUID, reminderType, reminderKey string) (bool, error)
	ListIntakesInRange(ctx context.Context, medicineID uuid.UUID, from, to time.Time) ([]*medicine.MedicineIntake, error)
}

// Notifier fans out a notification to all enabled channels for a user.
// *notificationSvc.Service satisfies it.
type Notifier interface {
	Notify(ctx context.Context, userID uuid.UUID, msg notification.Message) *notification.DeliveryResult
}

// ── Result ────────────────────────────────────────────────────────────────────

// RunResult summarises the outcome of one cron run for observability.
type RunResult struct {
	UsersNotified int            `json:"users_notified"`
	ItemsByType   map[string]int `json:"items_by_type"`
}

// ── Service ───────────────────────────────────────────────────────────────────

type Service struct {
	medRepo MedicineRepository
	noti    Notifier
	log     *zap.Logger
}

func New(medRepo MedicineRepository, noti Notifier, log *zap.Logger) *Service {
	return &Service{medRepo: medRepo, noti: noti, log: log}
}

// Run executes the reminder batch job.  It is safe to call concurrently
// (idempotency is guaranteed by the medicine_reminder_log unique constraint).
func (s *Service) Run(ctx context.Context) (*RunResult, error) {
	now := time.Now() // Asia/Bangkok via time.Local set in main.go
	todayKey := now.Format("2006-01-02")

	meds, err := s.medRepo.ScanActiveMedicinesForReminders(ctx)
	if err != nil {
		return nil, fmt.Errorf("scan medicines: %w", err)
	}

	// Group medicines by user.
	byUser := make(map[uuid.UUID][]*medicine.Medicine)
	for _, m := range meds {
		byUser[m.UserID] = append(byUser[m.UserID], m)
	}

	result := &RunResult{ItemsByType: make(map[string]int)}

	for userID, userMeds := range byUser {
		var (
			lowStockLines []string
			refillLines   []string
			doseLines     []string
			missedLines   []string
		)

		for _, m := range userMeds {
			// ── Supply digest (once per day, only after supplyDigestHour) ──
			if now.Hour() >= supplyDigestHour {
				// Low stock
				if m.StockQuantity <= m.LowStockThreshold {
					logged, logErr := s.medRepo.LogReminder(ctx, userID, m.ID, "low_stock", todayKey)
					if logErr != nil {
						s.log.Warn("reminder: log low_stock", zap.String("medicine_id", m.ID.String()), zap.Error(logErr))
					}
					if logged {
						lowStockLines = append(lowStockLines, fmt.Sprintf("%s (%.0f left)", m.Name, m.StockQuantity))
						result.ItemsByType["low_stock"]++
					}
				}

				// Refill (days remaining <= threshold, but stock still above low_stock line)
				days := estimatedDaysRemaining(m)
				if days != nil && *days <= refillDaysThreshold && m.StockQuantity > m.LowStockThreshold {
					logged, logErr := s.medRepo.LogReminder(ctx, userID, m.ID, "refill", todayKey)
					if logErr != nil {
						s.log.Warn("reminder: log refill", zap.String("medicine_id", m.ID.String()), zap.Error(logErr))
					}
					if logged {
						refillLines = append(refillLines, fmt.Sprintf("%s (~%d days left)", m.Name, *days))
						result.ItemsByType["refill"]++
					}
				}
			}

			// ── Dose + missed reminders (only when reminders are enabled) ──
			if !m.ReminderEnabled || len(m.ReminderTimes) == 0 {
				continue
			}

			for _, timeStr := range m.ReminderTimes {
				slotTime, ok := parseSlotTime(now, timeStr)
				if !ok {
					s.log.Warn("reminder: invalid reminder_time format",
						zap.String("medicine_id", m.ID.String()),
						zap.String("time", timeStr))
					continue
				}

				slotKey := fmt.Sprintf("%s#%s", todayKey, timeStr)

				// Dose due: slot falls in (now - cronInterval, now]
				windowStart := now.Add(-cronInterval)
				if !slotTime.Before(windowStart) && !slotTime.After(now) {
					logged, logErr := s.medRepo.LogReminder(ctx, userID, m.ID, "dose", slotKey)
					if logErr != nil {
						s.log.Warn("reminder: log dose", zap.String("medicine_id", m.ID.String()), zap.Error(logErr))
					}
					if logged {
						unit := ""
						if m.DosageUnit != nil {
							unit = " " + *m.DosageUnit
						}
						doseLines = append(doseLines, fmt.Sprintf("%s (%.4g%s)", m.Name, m.DosageAmount, unit))
						result.ItemsByType["dose"]++
					}
				}

				// Missed: slot passed more than missedGrace ago AND no taken intake near that slot
				if now.Sub(slotTime) > missedGrace && slotTime.Before(now) {
					missedKey := "missed#" + slotKey
					// Check first whether we already logged a missed reminder for this slot
					// to avoid the intake scan on repeat runs.
					intakes, intakeErr := s.medRepo.ListIntakesInRange(
						ctx, m.ID,
						slotTime.Add(-30*time.Minute),
						slotTime.Add(90*time.Minute),
					)
					if intakeErr != nil {
						s.log.Warn("reminder: list intakes for missed check",
							zap.String("medicine_id", m.ID.String()), zap.Error(intakeErr))
						continue
					}
					hasTaken := false
					for _, intake := range intakes {
						if intake.Status == medicine.IntakeStatusTaken {
							hasTaken = true
							break
						}
					}
					if !hasTaken {
						logged, logErr := s.medRepo.LogReminder(ctx, userID, m.ID, "missed", missedKey)
						if logErr != nil {
							s.log.Warn("reminder: log missed", zap.String("medicine_id", m.ID.String()), zap.Error(logErr))
						}
						if logged {
							missedLines = append(missedLines, fmt.Sprintf("%s at %s", m.Name, timeStr))
							result.ItemsByType["missed"]++
						}
					}
				}
			}
		}

		// Build the combined notification body for this user.
		var parts []string
		if len(doseLines) > 0 {
			parts = append(parts, "Time to take: "+strings.Join(doseLines, ", ")+".")
		}
		if len(missedLines) > 0 {
			parts = append(parts, "Missed earlier: "+strings.Join(missedLines, ", ")+".")
		}
		if len(lowStockLines) > 0 {
			parts = append(parts, "Low stock: "+strings.Join(lowStockLines, ", ")+".")
		}
		if len(refillLines) > 0 {
			parts = append(parts, "Running out soon: "+strings.Join(refillLines, ", ")+".")
		}

		if len(parts) == 0 {
			continue
		}

		s.noti.Notify(ctx, userID, notification.Message{
			Title: "Medicine reminders",
			Body:  strings.Join(parts, " "),
		})
		result.UsersNotified++
	}

	s.log.Info("reminder run complete",
		zap.Int("users_notified", result.UsersNotified),
		zap.Any("items_by_type", result.ItemsByType),
	)
	return result, nil
}

// ── Helpers ───────────────────────────────────────────────────────────────────

// estimatedDaysRemaining ports the frontend estimatedDaysRemaining helper to Go.
// Returns nil when the daily-dose rate cannot be determined (as_needed, custom).
func estimatedDaysRemaining(m *medicine.Medicine) *int {
	if m.DosageAmount <= 0 {
		return nil
	}
	var dosesPerDay float64
	switch m.FrequencyType {
	case medicine.FrequencyTypeDaily:
		fv := 1
		if m.FrequencyValue != nil {
			fv = *m.FrequencyValue
		}
		dosesPerDay = m.DosageAmount * float64(fv)
	case medicine.FrequencyTypeWeekly:
		fv := 1
		if m.FrequencyValue != nil {
			fv = *m.FrequencyValue
		}
		dosesPerDay = (m.DosageAmount * float64(fv)) / 7.0
	default:
		return nil
	}
	if dosesPerDay <= 0 {
		return nil
	}
	days := int(math.Floor(m.StockQuantity / dosesPerDay))
	return &days
}

// parseSlotTime parses an "HH:MM" string and returns the corresponding
// time.Time on the same calendar day as `now` in the local timezone.
// Returns (zero, false) if the format is invalid.
func parseSlotTime(now time.Time, hhmm string) (time.Time, bool) {
	parts := strings.SplitN(hhmm, ":", 2)
	if len(parts) != 2 {
		return time.Time{}, false
	}
	var hour, min int
	if _, err := fmt.Sscanf(parts[0], "%d", &hour); err != nil || hour < 0 || hour > 23 {
		return time.Time{}, false
	}
	if _, err := fmt.Sscanf(parts[1], "%d", &min); err != nil || min < 0 || min > 59 {
		return time.Time{}, false
	}
	slot := time.Date(now.Year(), now.Month(), now.Day(), hour, min, 0, 0, now.Location())
	return slot, true
}
