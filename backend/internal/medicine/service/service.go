package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/kinkando/personal-dashboard/internal/medicine"
	"github.com/kinkando/personal-dashboard/pkg/event"
)

// EventPublisher is the narrow interface medicine depends on.
// *event.Bus satisfies it; medicine never imports the quest package.
type EventPublisher interface {
	Publish(ctx context.Context, e event.Event)
}

type Repository interface {
	ListMedicines(ctx context.Context, userID uuid.UUID, includeArchived bool) ([]*medicine.Medicine, error)
	CreateMedicine(ctx context.Context, userID uuid.UUID, in medicine.CreateMedicineInput) (*medicine.Medicine, error)
	UpdateMedicine(ctx context.Context, id uuid.UUID, userID uuid.UUID, in medicine.UpdateMedicineInput) (*medicine.Medicine, error)
	SetArchived(ctx context.Context, id uuid.UUID, userID uuid.UUID, archived bool) (*medicine.Medicine, error)

	Take(ctx context.Context, userID uuid.UUID, medicineID uuid.UUID, in medicine.TakeMedicineInput) (*medicine.MedicineIntake, *medicine.Medicine, error)

	AdjustStock(ctx context.Context, userID uuid.UUID, medicineID uuid.UUID, in medicine.AdjustStockInput) (*medicine.MedicineStockAdjustment, *medicine.Medicine, error)

	ListIntakes(ctx context.Context, userID uuid.UUID, opts medicine.ListIntakeOpts) ([]*medicine.MedicineIntake, error)

	ListStockAdjustments(ctx context.Context, userID uuid.UUID, opts medicine.ListAdjustmentOpts) ([]*medicine.MedicineStockAdjustment, error)
}

type Service struct {
	repo   Repository
	events EventPublisher // nil-safe; set via New
}

func New(repo Repository, events EventPublisher) *Service {
	return &Service{repo: repo, events: events}
}

func (s *Service) ListMedicines(ctx context.Context, userID uuid.UUID, includeArchived bool) ([]*medicine.Medicine, error) {
	return s.repo.ListMedicines(ctx, userID, includeArchived)
}

func (s *Service) CreateMedicine(ctx context.Context, userID uuid.UUID, in medicine.CreateMedicineInput) (*medicine.Medicine, error) {
	return s.repo.CreateMedicine(ctx, userID, in)
}

func (s *Service) UpdateMedicine(ctx context.Context, id uuid.UUID, userID uuid.UUID, in medicine.UpdateMedicineInput) (*medicine.Medicine, error) {
	return s.repo.UpdateMedicine(ctx, id, userID, in)
}

func (s *Service) SetArchived(ctx context.Context, id uuid.UUID, userID uuid.UUID, archived bool) (*medicine.Medicine, error) {
	return s.repo.SetArchived(ctx, id, userID, archived)
}

func (s *Service) Take(ctx context.Context, userID uuid.UUID, medicineID uuid.UUID, in medicine.TakeMedicineInput) (*medicine.MedicineIntake, *medicine.Medicine, error) {
	intake, med, err := s.repo.Take(ctx, userID, medicineID, in)
	if err != nil {
		return nil, nil, err
	}
	// Publish event so quest subscribers can react without medicine importing quest.
	// Include the medicine's source_type so subscribers can route to the correct quest bucket.
	if s.events != nil {
		eventType := event.MedicineTaken
		if med.SourceType == medicine.SourceTypeSupplement {
			eventType = event.SupplementTaken
		}
		s.events.Publish(ctx, event.Event{
			Type:   eventType,
			UserID: userID,
		})
	}
	return intake, med, nil
}

func (s *Service) AdjustStock(ctx context.Context, userID uuid.UUID, medicineID uuid.UUID, in medicine.AdjustStockInput) (*medicine.MedicineStockAdjustment, *medicine.Medicine, error) {
	return s.repo.AdjustStock(ctx, userID, medicineID, in)
}

func (s *Service) ListIntakes(ctx context.Context, userID uuid.UUID, opts medicine.ListIntakeOpts) ([]*medicine.MedicineIntake, error) {
	return s.repo.ListIntakes(ctx, userID, opts)
}

func (s *Service) ListStockAdjustments(ctx context.Context, userID uuid.UUID, opts medicine.ListAdjustmentOpts) ([]*medicine.MedicineStockAdjustment, error) {
	return s.repo.ListStockAdjustments(ctx, userID, opts)
}
