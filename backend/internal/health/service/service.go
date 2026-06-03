package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/kinkando/personal-dashboard/internal/health"
	"github.com/kinkando/personal-dashboard/pkg/event"
	"github.com/kinkando/personal-dashboard/pkg/helper"
)

// EventPublisher is the narrow interface health depends on.
// *event.Bus satisfies it; health never imports the quest package.
type EventPublisher interface {
	Publish(ctx context.Context, e event.Event)
}

type Repository interface {
	GetProfile(ctx context.Context, userID uuid.UUID) (*health.Profile, error)
	UpsertProfile(ctx context.Context, userID uuid.UUID, in health.UpsertProfileInput) (*health.Profile, error)

	ListWeightLogs(ctx context.Context, userID uuid.UUID) ([]*health.WeightLog, error)
	CreateWeightLog(ctx context.Context, userID uuid.UUID, in health.CreateWeightInput) (*health.WeightLog, error)
	DeleteWeightLog(ctx context.Context, id uuid.UUID, userID uuid.UUID) error

	ListFoodLogs(ctx context.Context, userID uuid.UUID) ([]*health.FoodLog, error)
	CreateFoodLog(ctx context.Context, userID uuid.UUID, in health.CreateFoodInput) (*health.FoodLog, error)
	UpdateFoodLog(ctx context.Context, id uuid.UUID, userID uuid.UUID, in health.UpdateFoodInput) (*health.FoodLog, error)
	DeleteFoodLog(ctx context.Context, id uuid.UUID, userID uuid.UUID) error

	ListSleepLogs(ctx context.Context, userID uuid.UUID) ([]*health.SleepLog, error)
	CreateSleepLog(ctx context.Context, userID uuid.UUID, in health.CreateSleepInput) (*health.SleepLog, error)
	UpdateSleepLog(ctx context.Context, id uuid.UUID, userID uuid.UUID, in health.UpdateSleepInput) (*health.SleepLog, error)
	DeleteSleepLog(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
}

type Service struct {
	repo   Repository
	events EventPublisher // nil-safe; set via New
}

func New(repo Repository, events EventPublisher) *Service {
	return &Service{repo: repo, events: events}
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
	log, err := s.repo.CreateWeightLog(ctx, userID, in)
	if err != nil {
		return nil, err
	}
	// Publish only when the logged date is the local current day (Asia/Bangkok).
	// Logging a past or future date must not complete today's quest.
	if s.events != nil && log.LoggedAt.Equal(helper.Today()) {
		s.events.Publish(ctx, event.Event{Type: event.WeightLogged, UserID: userID})
	}
	return log, nil
}

func (s *Service) DeleteWeightLog(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	return s.repo.DeleteWeightLog(ctx, id, userID)
}

func (s *Service) ListFoodLogs(ctx context.Context, userID uuid.UUID) ([]*health.FoodLog, error) {
	return s.repo.ListFoodLogs(ctx, userID)
}

func (s *Service) CreateFoodLog(ctx context.Context, userID uuid.UUID, in health.CreateFoodInput) (*health.FoodLog, error) {
	return s.repo.CreateFoodLog(ctx, userID, in)
}

func (s *Service) UpdateFoodLog(ctx context.Context, id uuid.UUID, userID uuid.UUID, in health.UpdateFoodInput) (*health.FoodLog, error) {
	return s.repo.UpdateFoodLog(ctx, id, userID, in)
}

func (s *Service) DeleteFoodLog(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	return s.repo.DeleteFoodLog(ctx, id, userID)
}

func (s *Service) ListSleepLogs(ctx context.Context, userID uuid.UUID) ([]*health.SleepLog, error) {
	return s.repo.ListSleepLogs(ctx, userID)
}

func (s *Service) CreateSleepLog(ctx context.Context, userID uuid.UUID, in health.CreateSleepInput) (*health.SleepLog, error) {
	return s.repo.CreateSleepLog(ctx, userID, in)
}

func (s *Service) UpdateSleepLog(ctx context.Context, id uuid.UUID, userID uuid.UUID, in health.UpdateSleepInput) (*health.SleepLog, error) {
	return s.repo.UpdateSleepLog(ctx, id, userID, in)
}

func (s *Service) DeleteSleepLog(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	return s.repo.DeleteSleepLog(ctx, id, userID)
}
