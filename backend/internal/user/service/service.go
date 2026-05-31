package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/kinkando/personal-dashboard/internal/user"
)

type Repository interface {
	GetOrCreate(ctx context.Context, firebaseUID, email string) (*user.User, error)
	GetIDByFirebaseUID(ctx context.Context, firebaseUID string) (uuid.UUID, error)
}

type Service struct {
	repo Repository
}

func New(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetOrCreate(ctx context.Context, firebaseUID, email string) (*user.User, error) {
	return s.repo.GetOrCreate(ctx, firebaseUID, email)
}

func (s *Service) GetIDByFirebaseUID(ctx context.Context, firebaseUID string) (uuid.UUID, error) {
	return s.repo.GetIDByFirebaseUID(ctx, firebaseUID)
}
