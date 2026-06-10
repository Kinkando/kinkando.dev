package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/kinkando/personal-dashboard/internal/medicine"
	"github.com/kinkando/personal-dashboard/pkg/event"
)

// ── Fakes ─────────────────────────────────────────────────────────────────────

type fakeRepo struct {
	intake *medicine.MedicineIntake
	med    *medicine.Medicine
}

func (f *fakeRepo) ListMedicines(_ context.Context, _ uuid.UUID, _ bool, _ *medicine.SourceType) ([]*medicine.Medicine, error) {
	return nil, nil
}
func (f *fakeRepo) CreateMedicine(_ context.Context, _ uuid.UUID, _ medicine.CreateMedicineInput) (*medicine.Medicine, error) {
	return nil, nil
}
func (f *fakeRepo) UpdateMedicine(_ context.Context, _ uuid.UUID, _ uuid.UUID, _ medicine.UpdateMedicineInput) (*medicine.Medicine, error) {
	return nil, nil
}
func (f *fakeRepo) SetArchived(_ context.Context, _ uuid.UUID, _ uuid.UUID, _ bool) (*medicine.Medicine, error) {
	return nil, nil
}
func (f *fakeRepo) Take(_ context.Context, _ uuid.UUID, _ uuid.UUID, _ medicine.TakeMedicineInput) (*medicine.MedicineIntake, *medicine.Medicine, error) {
	return f.intake, f.med, nil
}
func (f *fakeRepo) AdjustStock(_ context.Context, _ uuid.UUID, _ uuid.UUID, _ medicine.AdjustStockInput) (*medicine.MedicineStockAdjustment, *medicine.Medicine, error) {
	return nil, nil, nil
}
func (f *fakeRepo) ListIntakes(_ context.Context, _ uuid.UUID, _ medicine.ListIntakeOpts) ([]*medicine.MedicineIntake, error) {
	return nil, nil
}
func (f *fakeRepo) ListStockAdjustments(_ context.Context, _ uuid.UUID, _ medicine.ListAdjustmentOpts) ([]*medicine.MedicineStockAdjustment, error) {
	return nil, nil
}

type fakeEvents struct {
	published []event.Event
}

func (f *fakeEvents) Publish(_ context.Context, e event.Event) {
	f.published = append(f.published, e)
}

// ── Take: event dispatch by SourceType ───────────────────────────────────────

func TestTake_PublishesMedicineTaken_ForMedication(t *testing.T) {
	med := &medicine.Medicine{SourceType: medicine.SourceTypeMedication}
	repo := &fakeRepo{intake: &medicine.MedicineIntake{}, med: med}
	ev := &fakeEvents{}
	svc := &Service{repo: repo, events: ev}

	_, _, err := svc.Take(context.Background(), uuid.New(), uuid.New(), medicine.TakeMedicineInput{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ev.published) != 1 {
		t.Fatalf("published %d events, want 1", len(ev.published))
	}
	if ev.published[0].Type != event.MedicineTaken {
		t.Errorf("event type = %v, want MedicineTaken", ev.published[0].Type)
	}
}

func TestTake_PublishesSupplementTaken_ForSupplement(t *testing.T) {
	med := &medicine.Medicine{SourceType: medicine.SourceTypeSupplement}
	repo := &fakeRepo{intake: &medicine.MedicineIntake{}, med: med}
	ev := &fakeEvents{}
	svc := &Service{repo: repo, events: ev}

	_, _, err := svc.Take(context.Background(), uuid.New(), uuid.New(), medicine.TakeMedicineInput{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ev.published) != 1 {
		t.Fatalf("published %d events, want 1", len(ev.published))
	}
	if ev.published[0].Type != event.SupplementTaken {
		t.Errorf("event type = %v, want SupplementTaken", ev.published[0].Type)
	}
}

func TestTake_NoEvent_WhenEventsNil(t *testing.T) {
	med := &medicine.Medicine{SourceType: medicine.SourceTypeMedication}
	repo := &fakeRepo{intake: &medicine.MedicineIntake{}, med: med}
	svc := &Service{repo: repo, events: nil} // nil events

	_, _, err := svc.Take(context.Background(), uuid.New(), uuid.New(), medicine.TakeMedicineInput{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// No panic — nil events guard works.
}

func TestTake_UserIDIsSetOnPublishedEvent(t *testing.T) {
	userID := uuid.New()
	med := &medicine.Medicine{SourceType: medicine.SourceTypeMedication}
	repo := &fakeRepo{intake: &medicine.MedicineIntake{}, med: med}
	ev := &fakeEvents{}
	svc := &Service{repo: repo, events: ev}

	_, _, err := svc.Take(context.Background(), userID, uuid.New(), medicine.TakeMedicineInput{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ev.published[0].UserID != userID {
		t.Errorf("event UserID = %v, want %v", ev.published[0].UserID, userID)
	}
}
