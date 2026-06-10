package reminder

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kinkando/personal-dashboard/internal/medicine"
	"github.com/kinkando/personal-dashboard/internal/notification"
	"go.uber.org/zap"
)

// ── Helpers ───────────────────────────────────────────────────────────────────

var bangkokLoc *time.Location

func init() {
	loc, err := time.LoadLocation("Asia/Bangkok")
	if err != nil {
		panic(err)
	}
	bangkokLoc = loc
}

// bkkAt returns a time.Time in Asia/Bangkok at the given components.
func bkkAt(year, month, day, hour, min int) time.Time {
	return time.Date(year, time.Month(month), day, hour, min, 0, 0, bangkokLoc)
}

// ptr returns a pointer to v.
func ptr[T any](v T) *T { return &v }

// ── Pure function tests ───────────────────────────────────────────────────────

func TestParseSlotTime(t *testing.T) {
	now := bkkAt(2026, 6, 10, 9, 0) // reference time: 09:00 BKK

	cases := []struct {
		name    string
		hhmm    string
		wantOK  bool
		wantH   int
		wantMin int
	}{
		{"valid 08:30", "08:30", true, 8, 30},
		{"valid 00:00", "00:00", true, 0, 0},
		{"valid 23:59", "23:59", true, 23, 59},
		{"invalid hour 25:00", "25:00", false, 0, 0},
		{"invalid minute 08:60", "08:60", false, 0, 0},
		{"missing colon", "0830", false, 0, 0},
		{"empty string", "", false, 0, 0},
		{"negative hour -1:00", "-1:00", false, 0, 0},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := parseSlotTime(now, tc.hhmm)
			if ok != tc.wantOK {
				t.Fatalf("parseSlotTime(%q) ok = %v, want %v", tc.hhmm, ok, tc.wantOK)
			}
			if !tc.wantOK {
				return
			}
			if got.Hour() != tc.wantH || got.Minute() != tc.wantMin {
				t.Errorf("parseSlotTime(%q) = %02d:%02d, want %02d:%02d",
					tc.hhmm, got.Hour(), got.Minute(), tc.wantH, tc.wantMin)
			}
			// Result must be on the same calendar day as now.
			if got.Year() != now.Year() || got.Month() != now.Month() || got.Day() != now.Day() {
				t.Errorf("parseSlotTime(%q) date = %s, want %s",
					tc.hhmm, got.Format(time.DateOnly), now.Format(time.DateOnly))
			}
		})
	}
}

func TestEstimatedDaysRemaining(t *testing.T) {
	cases := []struct {
		name      string
		med       *medicine.Medicine
		wantNil   bool
		wantDays  int
	}{
		{
			name: "daily: floor(10/2) = 5",
			med: &medicine.Medicine{
				DosageAmount:  2, FrequencyType: medicine.FrequencyTypeDaily,
				FrequencyValue: ptr(1), StockQuantity: 10,
			},
			wantDays: 5,
		},
		{
			name: "daily with freq=2: 10/(2*2) = 2",
			med: &medicine.Medicine{
				DosageAmount:  2, FrequencyType: medicine.FrequencyTypeDaily,
				FrequencyValue: ptr(2), StockQuantity: 10,
			},
			wantDays: 2,
		},
		{
			name: "daily nil freq_value defaults to 1: floor(7/1) = 7",
			med: &medicine.Medicine{
				DosageAmount: 1, FrequencyType: medicine.FrequencyTypeDaily,
				FrequencyValue: nil, StockQuantity: 7,
			},
			wantDays: 7,
		},
		{
			name: "weekly: floor(7 / (1*7/7)) = 7",
			med: &medicine.Medicine{
				DosageAmount:  1, FrequencyType: medicine.FrequencyTypeWeekly,
				FrequencyValue: ptr(7), StockQuantity: 7,
			},
			wantDays: 7,
		},
		{
			name:    "as_needed: nil",
			med:     &medicine.Medicine{DosageAmount: 1, FrequencyType: medicine.FrequencyTypeAsNeeded},
			wantNil: true,
		},
		{
			name:    "dosage_amount zero: nil",
			med:     &medicine.Medicine{DosageAmount: 0, FrequencyType: medicine.FrequencyTypeDaily, FrequencyValue: ptr(1)},
			wantNil: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := estimatedDaysRemaining(tc.med)
			if tc.wantNil {
				if got != nil {
					t.Errorf("estimatedDaysRemaining() = %d, want nil", *got)
				}
				return
			}
			if got == nil {
				t.Fatalf("estimatedDaysRemaining() = nil, want %d", tc.wantDays)
			}
			if *got != tc.wantDays {
				t.Errorf("estimatedDaysRemaining() = %d, want %d", *got, tc.wantDays)
			}
		})
	}
}

// ── Fake implementations ──────────────────────────────────────────────────────

type fakeMedRepo struct {
	medicines  []*medicine.Medicine
	logFn      func(ctx context.Context, userID, medID uuid.UUID, rType, key string) (bool, error)
	intakesFn  func(ctx context.Context, medID uuid.UUID, from, to time.Time) ([]*medicine.MedicineIntake, error)
}

func (f *fakeMedRepo) ScanActiveMedicinesForReminders(_ context.Context) ([]*medicine.Medicine, error) {
	return f.medicines, nil
}
func (f *fakeMedRepo) LogReminder(ctx context.Context, userID, medID uuid.UUID, rType, key string) (bool, error) {
	if f.logFn != nil {
		return f.logFn(ctx, userID, medID, rType, key)
	}
	return true, nil // default: always newly logged
}
func (f *fakeMedRepo) ListIntakesInRange(ctx context.Context, medID uuid.UUID, from, to time.Time) ([]*medicine.MedicineIntake, error) {
	if f.intakesFn != nil {
		return f.intakesFn(ctx, medID, from, to)
	}
	return nil, nil
}

type fakeNotifier struct {
	calls []notification.Message
	users []uuid.UUID
}

func (f *fakeNotifier) Notify(_ context.Context, userID uuid.UUID, msg notification.Message) *notification.DeliveryResult {
	f.calls = append(f.calls, msg)
	f.users = append(f.users, userID)
	return &notification.DeliveryResult{Attempted: 1, Delivered: 1}
}

// ── Service tests ─────────────────────────────────────────────────────────────

func TestMedicineReminder_SupplyDigest_HourGate(t *testing.T) {
	// Before supplyDigestHour (09:00 BKK): low-stock check must not fire.
	userID := uuid.New()
	medID := uuid.New()
	m := &medicine.Medicine{
		ID: medID, UserID: userID, Name: "Aspirin",
		DosageAmount: 1, FrequencyType: medicine.FrequencyTypeDaily, FrequencyValue: ptr(1),
		StockQuantity: 2, LowStockThreshold: 5, // low stock
		ReminderEnabled: false,
	}
	repo := &fakeMedRepo{medicines: []*medicine.Medicine{m}}
	noti := &fakeNotifier{}
	svc := &Service{
		medRepo: repo, noti: noti, log: zap.NewNop(),
		now: func() time.Time { return bkkAt(2026, 6, 10, 8, 0) }, // 08:00 < supplyDigestHour(9)
	}

	result, err := svc.Run(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.UsersNotified != 0 {
		t.Errorf("UsersNotified = %d, want 0 (before hour gate)", result.UsersNotified)
	}
	if len(noti.calls) != 0 {
		t.Errorf("Notify called %d times, want 0 (before hour gate)", len(noti.calls))
	}
}

func TestMedicineReminder_LowStock_FiresAfterHour(t *testing.T) {
	userID := uuid.New()
	medID := uuid.New()
	m := &medicine.Medicine{
		ID: medID, UserID: userID, Name: "Aspirin",
		DosageAmount: 1, FrequencyType: medicine.FrequencyTypeDaily, FrequencyValue: ptr(1),
		StockQuantity: 2, LowStockThreshold: 5, // low stock
		ReminderEnabled: false,
	}
	repo := &fakeMedRepo{medicines: []*medicine.Medicine{m}}
	noti := &fakeNotifier{}
	svc := &Service{
		medRepo: repo, noti: noti, log: zap.NewNop(),
		now: func() time.Time { return bkkAt(2026, 6, 10, 9, 30) }, // 09:30 >= supplyDigestHour(9)
	}

	result, err := svc.Run(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.UsersNotified != 1 {
		t.Errorf("UsersNotified = %d, want 1", result.UsersNotified)
	}
	if result.ItemsByType["low_stock"] != 1 {
		t.Errorf("low_stock count = %d, want 1", result.ItemsByType["low_stock"])
	}
	if len(noti.calls) != 1 {
		t.Fatalf("Notify calls = %d, want 1", len(noti.calls))
	}
	if !strings.Contains(noti.calls[0].Body, "Aspirin") {
		t.Errorf("notification body %q does not mention medicine name", noti.calls[0].Body)
	}
}

func TestMedicineReminder_DoseWindow_Fires(t *testing.T) {
	// Dose slot is 09:00; now is 09:15 (within the 30-min window). No intake logged.
	userID := uuid.New()
	medID := uuid.New()
	now := bkkAt(2026, 6, 10, 9, 15)
	m := &medicine.Medicine{
		ID: medID, UserID: userID, Name: "Vitamin",
		DosageAmount: 1, FrequencyType: medicine.FrequencyTypeDaily, FrequencyValue: ptr(1),
		StockQuantity: 30, LowStockThreshold: 5,
		ReminderEnabled: true, ReminderTimes: []string{"09:00"},
	}
	repo := &fakeMedRepo{medicines: []*medicine.Medicine{m}}
	noti := &fakeNotifier{}
	svc := &Service{
		medRepo: repo, noti: noti, log: zap.NewNop(),
		now: func() time.Time { return now },
	}

	result, err := svc.Run(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ItemsByType["dose"] != 1 {
		t.Errorf("dose count = %d, want 1", result.ItemsByType["dose"])
	}
}

func TestMedicineReminder_DoseWindow_SkippedWhenAlreadyTaken(t *testing.T) {
	// Dose slot is 09:00; intake was taken at 09:05 — within the near-slot window.
	userID := uuid.New()
	medID := uuid.New()
	now := bkkAt(2026, 6, 10, 9, 15)
	m := &medicine.Medicine{
		ID: medID, UserID: userID, Name: "Vitamin",
		DosageAmount: 1, FrequencyType: medicine.FrequencyTypeDaily, FrequencyValue: ptr(1),
		StockQuantity: 30, LowStockThreshold: 5,
		ReminderEnabled: true, ReminderTimes: []string{"09:00"},
	}
	takenAt := bkkAt(2026, 6, 10, 9, 5)
	repo := &fakeMedRepo{
		medicines: []*medicine.Medicine{m},
		intakesFn: func(_ context.Context, _ uuid.UUID, _, _ time.Time) ([]*medicine.MedicineIntake, error) {
			return []*medicine.MedicineIntake{
				{Status: medicine.IntakeStatusTaken, TakenAt: takenAt, QuantityTaken: 1},
			}, nil
		},
	}
	noti := &fakeNotifier{}
	svc := &Service{
		medRepo: repo, noti: noti, log: zap.NewNop(),
		now: func() time.Time { return now },
	}

	result, err := svc.Run(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ItemsByType["dose"] != 0 {
		t.Errorf("dose count = %d, want 0 (already taken)", result.ItemsByType["dose"])
	}
	if len(noti.calls) != 0 {
		t.Errorf("Notify calls = %d, want 0 (already taken)", len(noti.calls))
	}
}
