package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/kinkando/personal-dashboard/internal/user"
)

type Repository interface {
	GetOrCreate(ctx context.Context, firebaseUID, email string) (*user.User, error)
	GetIDByFirebaseUID(ctx context.Context, firebaseUID string) (uuid.UUID, error)
	GetByFirebaseUID(ctx context.Context, firebaseUID string) (*user.User, error)
	CreateLinkCode(ctx context.Context, userID uuid.UUID) (code string, expiresAt time.Time, err error)
	Unlink(ctx context.Context, userID uuid.UUID) error
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

func (s *Service) GetByFirebaseUID(ctx context.Context, firebaseUID string) (*user.User, error) {
	return s.repo.GetByFirebaseUID(ctx, firebaseUID)
}

func (s *Service) CreateLinkCode(ctx context.Context, userID uuid.UUID) (code string, expiresAt time.Time, err error) {
	return s.repo.CreateLinkCode(ctx, userID)
}

func (s *Service) Unlink(ctx context.Context, userID uuid.UUID) error {
	return s.repo.Unlink(ctx, userID)
}
