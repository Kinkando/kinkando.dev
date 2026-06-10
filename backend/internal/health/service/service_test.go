package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kinkando/personal-dashboard/internal/health"
	"github.com/kinkando/personal-dashboard/pkg/event"
)

// ── Fakes ─────────────────────────────────────────────────────────────────────

type fakeRepo struct {
	weightLog *health.WeightLog
	sleepLog  *health.SleepLog
}

func (f *fakeRepo) GetProfile(_ context.Context, _ uuid.UUID) (*health.Profile, error) {
	return nil, nil
}
func (f *fakeRepo) UpsertProfile(_ context.Context, _ uuid.UUID, _ health.UpsertProfileInput) (*health.Profile, error) {
	return nil, nil
}
func (f *fakeRepo) ListWeightLogs(_ context.Context, _ uuid.UUID, _, _ time.Time) ([]*health.WeightLog, error) {
	return nil, nil
}
func (f *fakeRepo) CreateWeightLog(_ context.Context, _ uuid.UUID, _ health.CreateWeightInput) (*health.WeightLog, error) {
	return f.weightLog, nil
}
func (f *fakeRepo) UpdateWeightLog(_ context.Context, _ uuid.UUID, _ uuid.UUID, _ health.UpdateWeightInput) (*health.WeightLog, error) {
	return nil, nil
}
func (f *fakeRepo) DeleteWeightLog(_ context.Context, _ uuid.UUID, _ uuid.UUID) error { return nil }
func (f *fakeRepo) ListFoodLogs(_ context.Context, _ uuid.UUID) ([]*health.FoodLog, error) {
	return nil, nil
}
func (f *fakeRepo) CreateFoodLog(_ context.Context, _ uuid.UUID, _ health.CreateFoodInput) (*health.FoodLog, error) {
	return nil, nil
}
func (f *fakeRepo) UpdateFoodLog(_ context.Context, _ uuid.UUID, _ uuid.UUID, _ health.UpdateFoodInput) (*health.FoodLog, error) {
	return nil, nil
}
func (f *fakeRepo) DeleteFoodLog(_ context.Context, _ uuid.UUID, _ uuid.UUID) error { return nil }
func (f *fakeRepo) ListSleepLogs(_ context.Context, _ uuid.UUID, _, _ time.Time) ([]*health.SleepLog, error) {
	return nil, nil
}
func (f *fakeRepo) CreateSleepLog(_ context.Context, _ uuid.UUID, _ health.CreateSleepInput) (*health.SleepLog, error) {
	return f.sleepLog, nil
}
func (f *fakeRepo) UpdateSleepLog(_ context.Context, _ uuid.UUID, _ uuid.UUID, _ health.UpdateSleepInput) (*health.SleepLog, error) {
	return nil, nil
}
func (f *fakeRepo) DeleteSleepLog(_ context.Context, _ uuid.UUID, _ uuid.UUID) error { return nil }

type fakeEvents struct {
	published []event.Event
}

func (f *fakeEvents) Publish(_ context.Context, e event.Event) {
	f.published = append(f.published, e)
}

// bangkokMidnightUTC returns midnight UTC for the Bangkok calendar day.
func bangkokMidnightUTC(year, month, day int) time.Time {
	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
}

// ── CreateWeightLog ───────────────────────────────────────────────────────────

func TestCreateWeightLog_PublishesEvent_WhenLoggedToday(t *testing.T) {
	today := bangkokMidnightUTC(2026, 6, 10)
	repo := &fakeRepo{weightLog: &health.WeightLog{LoggedAt: today}}
	ev := &fakeEvents{}

	svc := &Service{repo: repo, events: ev, now: func() time.Time { return today }}

	_, err := svc.CreateWeightLog(context.Background(), uuid.New(), health.CreateWeightInput{Weight: 70})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ev.published) != 1 {
		t.Fatalf("published %d events, want 1", len(ev.published))
	}
	if ev.published[0].Type != event.WeightLogged {
		t.Errorf("event type = %v, want WeightLogged", ev.published[0].Type)
	}
}

func TestCreateWeightLog_NoEvent_WhenLoggedPastDate(t *testing.T) {
	today := bangkokMidnightUTC(2026, 6, 10)
	yesterday := bangkokMidnightUTC(2026, 6, 9)
	repo := &fakeRepo{weightLog: &health.WeightLog{LoggedAt: yesterday}}
	ev := &fakeEvents{}

	svc := &Service{repo: repo, events: ev, now: func() time.Time { return today }}

	_, err := svc.CreateWeightLog(context.Background(), uuid.New(), health.CreateWeightInput{Weight: 70})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ev.published) != 0 {
		t.Errorf("published %d events for past date, want 0", len(ev.published))
	}
}

func TestCreateWeightLog_NoEvent_WhenEventsNil(t *testing.T) {
	today := bangkokMidnightUTC(2026, 6, 10)
	repo := &fakeRepo{weightLog: &health.WeightLog{LoggedAt: today}}

	svc := &Service{repo: repo, events: nil, now: func() time.Time { return today }}

	_, err := svc.CreateWeightLog(context.Background(), uuid.New(), health.CreateWeightInput{Weight: 70})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// No panic — events nil-guard works.
}

// ── CreateSleepLog ────────────────────────────────────────────────────────────

func TestCreateSleepLog_PublishesEvent_WhenLoggedToday(t *testing.T) {
	today := bangkokMidnightUTC(2026, 6, 10)
	repo := &fakeRepo{sleepLog: &health.SleepLog{LoggedAt: today}}
	ev := &fakeEvents{}

	svc := &Service{repo: repo, events: ev, now: func() time.Time { return today }}

	_, err := svc.CreateSleepLog(context.Background(), uuid.New(), health.CreateSleepInput{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ev.published) != 1 {
		t.Fatalf("published %d events, want 1", len(ev.published))
	}
	if ev.published[0].Type != event.SleepLogged {
		t.Errorf("event type = %v, want SleepLogged", ev.published[0].Type)
	}
}

func TestCreateSleepLog_NoEvent_WhenLoggedPastDate(t *testing.T) {
	today := bangkokMidnightUTC(2026, 6, 10)
	yesterday := bangkokMidnightUTC(2026, 6, 9)
	repo := &fakeRepo{sleepLog: &health.SleepLog{LoggedAt: yesterday}}
	ev := &fakeEvents{}

	svc := &Service{repo: repo, events: ev, now: func() time.Time { return today }}

	_, err := svc.CreateSleepLog(context.Background(), uuid.New(), health.CreateSleepInput{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ev.published) != 0 {
		t.Errorf("published %d events for past date, want 0", len(ev.published))
	}
}
