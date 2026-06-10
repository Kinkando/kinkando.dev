package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kinkando/personal-dashboard/internal/workout"
	"github.com/kinkando/personal-dashboard/pkg/event"
)

// ── Fakes ─────────────────────────────────────────────────────────────────────

// fakeRepo implements Repository. Only FinishSession has configurable behaviour;
// every other method returns zero values.
type fakeRepo struct {
	session    *workout.Session
	sessionErr error
}

func (f *fakeRepo) ListPresets(_ context.Context, _ uuid.UUID) ([]*workout.Preset, error) {
	return nil, nil
}
func (f *fakeRepo) GetPreset(_ context.Context, _, _ uuid.UUID) (*workout.Preset, error) {
	return nil, nil
}
func (f *fakeRepo) CreatePreset(_ context.Context, _ uuid.UUID, _ workout.CreatePresetInput) (*workout.Preset, error) {
	return nil, nil
}
func (f *fakeRepo) UpdatePreset(_ context.Context, _, _ uuid.UUID, _ workout.UpdatePresetInput) (*workout.Preset, error) {
	return nil, nil
}
func (f *fakeRepo) DeletePreset(_ context.Context, _, _ uuid.UUID) error { return nil }
func (f *fakeRepo) GetSchedule(_ context.Context, _ uuid.UUID) ([]*workout.ScheduleEntry, error) {
	return nil, nil
}
func (f *fakeRepo) SetSchedule(_ context.Context, _ uuid.UUID, _ []workout.ScheduleEntryInput) ([]*workout.ScheduleEntry, error) {
	return nil, nil
}
func (f *fakeRepo) ListSessions(_ context.Context, _ uuid.UUID, _, _ time.Time) ([]*workout.Session, error) {
	return nil, nil
}
func (f *fakeRepo) GetSession(_ context.Context, _, _ uuid.UUID) (*workout.Session, error) {
	return nil, nil
}
func (f *fakeRepo) GenerateSession(_ context.Context, _ uuid.UUID, _ string) (*workout.Session, error) {
	return nil, nil
}
func (f *fakeRepo) CreateSession(_ context.Context, _ uuid.UUID, _ workout.CreateSessionInput) (*workout.Session, error) {
	return nil, nil
}
func (f *fakeRepo) UpdateSession(_ context.Context, _, _ uuid.UUID, _ workout.UpdateSessionInput) (*workout.Session, error) {
	return nil, nil
}
func (f *fakeRepo) UpdateSessionExercise(_ context.Context, _, _, _ uuid.UUID, _ workout.UpdateSessionExerciseInput) (*workout.SessionExercise, error) {
	return nil, nil
}
func (f *fakeRepo) BulkUpdateSessionExercises(_ context.Context, _, _ uuid.UUID, _ []workout.BulkUpdateSessionExerciseItem) ([]workout.SessionExercise, error) {
	return nil, nil
}
func (f *fakeRepo) DeleteSession(_ context.Context, _, _ uuid.UUID) error { return nil }
func (f *fakeRepo) FinishSession(_ context.Context, _, _ uuid.UUID) (*workout.Session, error) {
	return f.session, f.sessionErr
}
func (f *fakeRepo) AddSessionExercise(_ context.Context, _, _ uuid.UUID, _ workout.AddSessionExerciseInput) (*workout.SessionExercise, error) {
	return nil, nil
}
func (f *fakeRepo) DeleteSessionExercise(_ context.Context, _, _, _ uuid.UUID) error { return nil }

// fakePublisher records every Publish call.
type fakePublisher struct {
	published []event.Event
}

func (f *fakePublisher) Publish(_ context.Context, e event.Event) {
	f.published = append(f.published, e)
}

// ── Tests ─────────────────────────────────────────────────────────────────────

func TestFinishSession_PublishesEventOnSuccess(t *testing.T) {
	sess := &workout.Session{ID: uuid.New()}
	repo := &fakeRepo{session: sess}
	pub := &fakePublisher{}
	svc := New(repo, pub)

	got, err := svc.FinishSession(context.Background(), uuid.New(), uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != sess {
		t.Error("returned session does not match repo result")
	}
	if len(pub.published) != 1 {
		t.Fatalf("published %d events, want 1", len(pub.published))
	}
	if pub.published[0].Type != event.WorkoutSessionFinished {
		t.Errorf("event type = %q, want %q", pub.published[0].Type, event.WorkoutSessionFinished)
	}
}

func TestFinishSession_NoEventOnRepoError(t *testing.T) {
	repo := &fakeRepo{sessionErr: errors.New("db unavailable")}
	pub := &fakePublisher{}
	svc := New(repo, pub)

	_, err := svc.FinishSession(context.Background(), uuid.New(), uuid.New())
	if err == nil {
		t.Fatal("expected error from repo, got nil")
	}
	if len(pub.published) != 0 {
		t.Errorf("published %d events on repo error, want 0", len(pub.published))
	}
}

func TestFinishSession_NilPublisherIsSafe(t *testing.T) {
	sess := &workout.Session{ID: uuid.New()}
	repo := &fakeRepo{session: sess}
	svc := New(repo, nil) // nil EventPublisher must not panic

	if _, err := svc.FinishSession(context.Background(), uuid.New(), uuid.New()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
