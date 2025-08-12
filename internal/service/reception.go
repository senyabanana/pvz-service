package service

import (
	"context"
	"errors"
	"time"

	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"github.com/senyabanana/pvz-service/internal/entity"
	"github.com/senyabanana/pvz-service/internal/infrastructure/monitoring"
	"github.com/senyabanana/pvz-service/internal/repository"
)

type ReceptionService struct {
	receptionRepo repository.ReceptionRepository
	pvzRepo       repository.PVZRepository
	trManager     *manager.Manager
	log           *logrus.Logger
}

func NewReceptionService(
	receptionRepo repository.ReceptionRepository, pvzRepo repository.PVZRepository, trManager *manager.Manager, log *logrus.Logger,
) *ReceptionService {
	return &ReceptionService{
		receptionRepo: receptionRepo,
		pvzRepo:       pvzRepo,
		trManager:     trManager,
		log:           log,
	}
}

func (s *ReceptionService) CreateReception(ctx context.Context, pvzID uuid.UUID) (*entity.Reception, error) {
	var result *entity.Reception

	err := s.trManager.Do(ctx, func(ctx context.Context) error {
		pvzExists, err := s.pvzRepo.IsPVZExists(ctx, pvzID)
		if err != nil {
			s.log.Errorf("failed to check pvz existence: %v", err)
			return err
		}

		if !pvzExists {
			s.log.Warnf("pvz not found: %s", pvzID)
			return entity.ErrPVZNotFound
		}

		openExists, err := s.receptionRepo.IsReceptionOpenExists(ctx, pvzID)
		if err != nil {
			s.log.Errorf("failed to check open reception: %v", err)
			return err
		}

		if openExists {
			s.log.Infof("reception already exists for pvzID=%s", pvzID)
			return entity.ErrReceptionAlreadyExists
		}

		reception := &entity.Reception{
			DateTime:  time.Now(),
			PVZID:     pvzID,
			Status:    entity.StatusInProgress,
			CreatedAt: time.Now(),
		}

		if err := s.receptionRepo.CreateReception(ctx, reception); err != nil {
			s.log.Errorf("failed to create reception: %v", err)
			return err
		}

		result = reception
		return nil
	})

	if err != nil {
		return nil, err
	}

	s.log.Infof("reception created: id=%s, pvz=%s", result.ID, result.PVZID)
	monitoring.CreatedReceptionsCounter.Inc()
	return result, nil
}

func (s *ReceptionService) CloseLastReception(ctx context.Context, pvzID uuid.UUID) (*entity.Reception, error) {
	var result *entity.Reception

	err := s.trManager.Do(ctx, func(ctx context.Context) error {
		reception, err := s.receptionRepo.GetOpenReception(ctx, pvzID)
		if err != nil {
			s.log.Warnf("no open reception to close for pvz: %s, err: %v", pvzID, err)
			return entity.ErrNoOpenReception
		}

		timeClose := time.Now()
		if err := s.receptionRepo.CloseReceptionByID(ctx, reception.ID, timeClose); err != nil {
			if errors.Is(err, entity.ErrReceptionAlreadyClosed) {
				s.log.Warnf("reception already closed: %s", reception.ID)
				return entity.ErrReceptionAlreadyClosed
			}

			s.log.Errorf("failed to close reception: %v", err)
			return err
		}

		reception.Status = entity.StatusClosed
		reception.ClosedAt = &timeClose
		result = reception

		s.log.Infof("reception closed: id=%s", reception.ID)
		return nil
	})

	return result, err
}

func filterReceptionsByDate(receptions []entity.Reception, start, end *time.Time) []entity.Reception {
	var filtered []entity.Reception
	for _, reception := range receptions {
		if (start == nil || !reception.DateTime.Before(*start)) && (end == nil || !reception.DateTime.After(*end)) {
			filtered = append(filtered, reception)
		}
	}

	return filtered
}

func groupReceptionsByPVZ(receptions []entity.Reception) map[uuid.UUID][]entity.Reception {
	result := make(map[uuid.UUID][]entity.Reception)
	for _, reception := range receptions {
		result[reception.PVZID] = append(result[reception.PVZID], reception)
	}

	return result
}

func extractReceptionIDs(receptions []entity.Reception) []uuid.UUID {
	ids := make([]uuid.UUID, 0, len(receptions))
	for _, reception := range receptions {
		ids = append(ids, reception.ID)
	}

	return ids
}
