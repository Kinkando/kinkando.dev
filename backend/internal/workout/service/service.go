package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/kinkando/personal-dashboard/internal/workout"
)

type Repository interface {
	ListPresets(ctx context.Context, userID uuid.UUID) ([]*workout.Preset, error)
	GetPreset(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*workout.Preset, error)
	CreatePreset(ctx context.Context, userID uuid.UUID, in workout.CreatePresetInput) (*workout.Preset, error)
	UpdatePreset(ctx context.Context, id uuid.UUID, userID uuid.UUID, in workout.UpdatePresetInput) (*workout.Preset, error)
	DeletePreset(ctx context.Context, id uuid.UUID, userID uuid.UUID) error

	GetSchedule(ctx context.Context, userID uuid.UUID) ([]*workout.ScheduleEntry, error)
	SetSchedule(ctx context.Context, userID uuid.UUID, entries []workout.ScheduleEntryInput) ([]*workout.ScheduleEntry, error)

	ListSessions(ctx context.Context, userID uuid.UUID, from, to time.Time) ([]*workout.Session, error)
	GetSession(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*workout.Session, error)
	GenerateSession(ctx context.Context, userID uuid.UUID, date string) (*workout.Session, error)
	CreateSession(ctx context.Context, userID uuid.UUID, in workout.CreateSessionInput) (*workout.Session, error)
	UpdateSession(ctx context.Context, id uuid.UUID, userID uuid.UUID, in workout.UpdateSessionInput) (*workout.Session, error)
	UpdateSessionExercise(ctx context.Context, id uuid.UUID, sessionID uuid.UUID, userID uuid.UUID, in workout.UpdateSessionExerciseInput) (*workout.SessionExercise, error)
	BulkUpdateSessionExercises(ctx context.Context, sessionID uuid.UUID, userID uuid.UUID, items []workout.BulkUpdateSessionExerciseItem) ([]workout.SessionExercise, error)
	DeleteSession(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
	AddSessionExercise(ctx context.Context, sessionID uuid.UUID, userID uuid.UUID, in workout.AddSessionExerciseInput) (*workout.SessionExercise, error)
	DeleteSessionExercise(ctx context.Context, exID uuid.UUID, sessionID uuid.UUID, userID uuid.UUID) error
}

type Service struct {
	repo Repository
}

func New(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) ListPresets(ctx context.Context, userID uuid.UUID) ([]*workout.Preset, error) {
	return s.repo.ListPresets(ctx, userID)
}

func (s *Service) GetPreset(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*workout.Preset, error) {
	return s.repo.GetPreset(ctx, id, userID)
}

func (s *Service) CreatePreset(ctx context.Context, userID uuid.UUID, in workout.CreatePresetInput) (*workout.Preset, error) {
	return s.repo.CreatePreset(ctx, userID, in)
}

func (s *Service) UpdatePreset(ctx context.Context, id uuid.UUID, userID uuid.UUID, in workout.UpdatePresetInput) (*workout.Preset, error) {
	return s.repo.UpdatePreset(ctx, id, userID, in)
}

func (s *Service) DeletePreset(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	return s.repo.DeletePreset(ctx, id, userID)
}

func (s *Service) GetSchedule(ctx context.Context, userID uuid.UUID) ([]*workout.ScheduleEntry, error) {
	return s.repo.GetSchedule(ctx, userID)
}

func (s *Service) SetSchedule(ctx context.Context, userID uuid.UUID, entries []workout.ScheduleEntryInput) ([]*workout.ScheduleEntry, error) {
	return s.repo.SetSchedule(ctx, userID, entries)
}

func (s *Service) ListSessions(ctx context.Context, userID uuid.UUID, from, to time.Time) ([]*workout.Session, error) {
	return s.repo.ListSessions(ctx, userID, from, to)
}

func (s *Service) GetSession(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*workout.Session, error) {
	return s.repo.GetSession(ctx, id, userID)
}

func (s *Service) GenerateSession(ctx context.Context, userID uuid.UUID, date string) (*workout.Session, error) {
	return s.repo.GenerateSession(ctx, userID, date)
}

func (s *Service) CreateSession(ctx context.Context, userID uuid.UUID, in workout.CreateSessionInput) (*workout.Session, error) {
	return s.repo.CreateSession(ctx, userID, in)
}

func (s *Service) UpdateSession(ctx context.Context, id uuid.UUID, userID uuid.UUID, in workout.UpdateSessionInput) (*workout.Session, error) {
	return s.repo.UpdateSession(ctx, id, userID, in)
}

func (s *Service) UpdateSessionExercise(ctx context.Context, id uuid.UUID, sessionID uuid.UUID, userID uuid.UUID, in workout.UpdateSessionExerciseInput) (*workout.SessionExercise, error) {
	return s.repo.UpdateSessionExercise(ctx, id, sessionID, userID, in)
}

func (s *Service) BulkUpdateSessionExercises(ctx context.Context, sessionID uuid.UUID, userID uuid.UUID, items []workout.BulkUpdateSessionExerciseItem) ([]workout.SessionExercise, error) {
	return s.repo.BulkUpdateSessionExercises(ctx, sessionID, userID, items)
}

func (s *Service) DeleteSession(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	return s.repo.DeleteSession(ctx, id, userID)
}

func (s *Service) AddSessionExercise(ctx context.Context, sessionID uuid.UUID, userID uuid.UUID, in workout.AddSessionExerciseInput) (*workout.SessionExercise, error) {
	return s.repo.AddSessionExercise(ctx, sessionID, userID, in)
}

func (s *Service) DeleteSessionExercise(ctx context.Context, exID uuid.UUID, sessionID uuid.UUID, userID uuid.UUID) error {
	return s.repo.DeleteSessionExercise(ctx, exID, sessionID, userID)
}
