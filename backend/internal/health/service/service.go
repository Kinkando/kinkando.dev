package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/kinkando/personal-dashboard/internal/health"
)

type Repository interface {
	GetProfile(ctx context.Context, userID uuid.UUID) (*health.Profile, error)
	UpsertProfile(ctx context.Context, userID uuid.UUID, in health.UpsertProfileInput) (*health.Profile, error)

	ListWeightLogs(ctx context.Context, userID uuid.UUID) ([]*health.WeightLog, error)
	CreateWeightLog(ctx context.Context, userID uuid.UUID, in health.CreateWeightInput) (*health.WeightLog, error)
	DeleteWeightLog(ctx context.Context, id uuid.UUID, userID uuid.UUID) error

	ListExercises(ctx context.Context, userID uuid.UUID) ([]*health.Exercise, error)
	CreateExercise(ctx context.Context, userID uuid.UUID, in health.CreateExerciseInput) (*health.Exercise, error)
	UpdateExercise(ctx context.Context, id uuid.UUID, userID uuid.UUID, in health.UpdateExerciseInput) (*health.Exercise, error)
	DeleteExercise(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
}

type Service struct {
	repo Repository
}

func New(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetProfile(ctx context.Context, userID uuid.UUID) (*health.Profile, error) {
	return s.repo.GetProfile(ctx, userID)
}

func (s *Service) UpsertProfile(ctx context.Context, userID uuid.UUID, in health.UpsertProfileInput) (*health.Profile, error) {
	return s.repo.UpsertProfile(ctx, userID, in)
}

func (s *Service) ListWeightLogs(ctx context.Context, userID uuid.UUID) ([]*health.WeightLog, error) {
	return s.repo.ListWeightLogs(ctx, userID)
}

func (s *Service) CreateWeightLog(ctx context.Context, userID uuid.UUID, in health.CreateWeightInput) (*health.WeightLog, error) {
	return s.repo.CreateWeightLog(ctx, userID, in)
}

func (s *Service) DeleteWeightLog(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	return s.repo.DeleteWeightLog(ctx, id, userID)
}

func (s *Service) ListExercises(ctx context.Context, userID uuid.UUID) ([]*health.Exercise, error) {
	return s.repo.ListExercises(ctx, userID)
}

func (s *Service) CreateExercise(ctx context.Context, userID uuid.UUID, in health.CreateExerciseInput) (*health.Exercise, error) {
	return s.repo.CreateExercise(ctx, userID, in)
}

func (s *Service) UpdateExercise(ctx context.Context, id uuid.UUID, userID uuid.UUID, in health.UpdateExerciseInput) (*health.Exercise, error) {
	return s.repo.UpdateExercise(ctx, id, userID, in)
}

func (s *Service) DeleteExercise(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	return s.repo.DeleteExercise(ctx, id, userID)
}
