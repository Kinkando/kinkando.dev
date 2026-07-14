package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/kinkando/personal-dashboard/internal/finance"
)

type Repository interface {
	CreateCategory(ctx context.Context, userID uuid.UUID, in finance.CreateCategoryInput) (*finance.Category, error)
	ListCategories(ctx context.Context, userID uuid.UUID) ([]*finance.Category, error)
	GetCategory(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*finance.Category, error)
	UpdateCategory(ctx context.Context, id uuid.UUID, userID uuid.UUID, in finance.UpdateCategoryInput) (*finance.Category, error)
	DeleteCategory(ctx context.Context, id uuid.UUID, userID uuid.UUID) error

	Create(ctx context.Context, userID uuid.UUID, in finance.CreateRecordInput) (*finance.Record, error)
	List(ctx context.Context, userID uuid.UUID, month string) ([]*finance.Record, error)
	MonthlySummary(ctx context.Context, userID uuid.UUID, month string) (*finance.MonthlySummary, error)
	DistinctNotes(ctx context.Context, userID uuid.UUID) ([]string, error)
	Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
}

type Service struct {
	repo Repository
}

func New(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateCategory(ctx context.Context, userID uuid.UUID, in finance.CreateCategoryInput) (*finance.Category, error) {
	return s.repo.CreateCategory(ctx, userID, in)
}

func (s *Service) ListCategories(ctx context.Context, userID uuid.UUID) ([]*finance.Category, error) {
	return s.repo.ListCategories(ctx, userID)
}

func (s *Service) GetCategory(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*finance.Category, error) {
	return s.repo.GetCategory(ctx, id, userID)
}

func (s *Service) UpdateCategory(ctx context.Context, id uuid.UUID, userID uuid.UUID, in finance.UpdateCategoryInput) (*finance.Category, error) {
	return s.repo.UpdateCategory(ctx, id, userID, in)
}

func (s *Service) DeleteCategory(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	return s.repo.DeleteCategory(ctx, id, userID)
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

func (s *Service) DistinctNotes(ctx context.Context, userID uuid.UUID) ([]string, error) {
	return s.repo.DistinctNotes(ctx, userID)
}
