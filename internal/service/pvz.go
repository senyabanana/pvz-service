package service

import (
	"context"
	"time"

	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"github.com/senyabanana/pvz-service/internal/entity"
	"github.com/senyabanana/pvz-service/internal/infrastructure/monitoring"
	"github.com/senyabanana/pvz-service/internal/repository"
)

type PVZService struct {
	pvzRepo       repository.PVZRepository
	receptionRepo repository.ReceptionRepository
	productRepo   repository.ProductRepository
	trManager     *manager.Manager
	log           *logrus.Logger
}

func NewPVZService(pvzRepo repository.PVZRepository, receptionRepo repository.ReceptionRepository, productRepo repository.ProductRepository, trManager *manager.Manager, log *logrus.Logger) *PVZService {
	return &PVZService{
		pvzRepo:       pvzRepo,
		receptionRepo: receptionRepo,
		productRepo:   productRepo,
		trManager:     trManager,
		log:           log,
	}
}

func (s *PVZService) CreatePVZ(ctx context.Context, city string) (*entity.PVZ, error) {
	if !entity.IsValidCity(city) {
		s.log.Warnf("attempt to create PVZ in unsupported city: %s", city)
		return nil, entity.ErrInvalidCity
	}

	pvz := &entity.PVZ{
		RegistrationDate: time.Now(),
		City:             entity.PVZCity(city),
	}

	if err := s.pvzRepo.CreatePVZ(ctx, pvz); err != nil {
		s.log.Errorf("failed to create PVZ: %v", err)
		return nil, err
	}

	s.log.Infof("PVZ created: id=%s, city=%s", pvz.ID, city)
	monitoring.CreatedPVZCounter.Inc()

	return pvz, nil
}

func (s *PVZService) GetFullPVZInfo(ctx context.Context, startDate, endDate *time.Time, page, limit int) ([]entity.FullPVZInfo, error) {
	var result []entity.FullPVZInfo

	s.log.Infof("get full PVZ info: page=%d, limit=%d, startDate=%v, endDate=%v", page, limit, startDate, endDate)

	err := s.trManager.Do(ctx, func(ctx context.Context) error {
		allPVZ, err := s.pvzRepo.GetAllPVZ(ctx)
		if err != nil {
			s.log.Errorf("failed to get PVZ list: %v", err)
			return err
		}

		paginated := paginatePVZ(allPVZ, page, limit)
		pvzIDs := extractPVZIDs(paginated)

		allReceptions, err := s.receptionRepo.GetReceptionsByPVZIDs(ctx, pvzIDs)
		if err != nil {
			s.log.Errorf("failed to get receptions: %v", err)
			return err
		}

		filtered := filterReceptionsByDate(allReceptions, startDate, endDate)
		receptionMap := groupReceptionsByPVZ(filtered)
		receptionIDs := extractReceptionIDs(filtered)

		allProducts, err := s.productRepo.GetProductsByReceptionIDs(ctx, receptionIDs)
		if err != nil {
			s.log.Errorf("failed to get products: %v", err)
			return err
		}

		productMap := groupProductsByReceptionID(allProducts)
		result = buildFullPVZInfo(paginated, receptionMap, productMap)

		s.log.Infof("successfully built full PVZ info, total %d pvz returned", len(result))

		return nil
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *PVZService) GetAllPVZ(ctx context.Context) ([]entity.PVZ, error) {
	s.log.Info("fetching all PVZ records")
	return s.pvzRepo.GetAllPVZ(ctx)
}

func paginatePVZ(pvz []entity.PVZ, page, limit int) []entity.PVZ {
	start := (page - 1) * limit
	if start >= len(pvz) {
		return nil
	}

	end := start + limit
	if end >= len(pvz) {
		end = len(pvz)
	}

	return pvz[start:end]
}

func extractPVZIDs(pvz []entity.PVZ) []uuid.UUID {
	ids := make([]uuid.UUID, 0, len(pvz))
	for _, p := range pvz {
		ids = append(ids, p.ID)
	}

	return ids
}

func buildFullPVZInfo(
	pvz []entity.PVZ, receptionMap map[uuid.UUID][]entity.Reception, productMap map[uuid.UUID][]entity.Product,
) []entity.FullPVZInfo {
	var result []entity.FullPVZInfo
	for _, p := range pvz {
		var receptions []entity.ReceptionWithProducts
		for _, r := range receptionMap[p.ID] {
			receptions = append(receptions, entity.ReceptionWithProducts{
				Reception: r,
				Products:  productMap[r.ID],
			})
		}

		result = append(result, entity.FullPVZInfo{
			PVZ:        p,
			Receptions: receptions,
		})
	}
	return result
}
