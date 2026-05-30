package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/kinkando/personal-dashboard/internal/finance"
)

type Repository interface {
	Create(ctx context.Context, userID uuid.UUID, in finance.CreateRecordInput) (*finance.Record, error)
	List(ctx context.Context, userID uuid.UUID, month string) ([]*finance.Record, error)
	MonthlySummary(ctx context.Context, userID uuid.UUID, month string) (*finance.MonthlySummary, error)
	Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
}

type Service struct {
	repo Repository
}

func New(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateRecord(ctx context.Context, userID uuid.UUID, in finance.CreateRecordInput) (*finance.Record, error) {
	return s.repo.Create(ctx, userID, in)
}

func (s *Service) ListRecords(ctx context.Context, userID uuid.UUID, month string) ([]*finance.Record, error) {
	return s.repo.List(ctx, userID, month)
}

func (s *Service) MonthlySummary(ctx context.Context, userID uuid.UUID, month string) (*finance.MonthlySummary, error) {
	return s.repo.MonthlySummary(ctx, userID, month)
}

func (s *Service) DeleteRecord(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	return s.repo.Delete(ctx, id, userID)
}
