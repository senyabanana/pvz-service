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

type ProductService struct {
	productRepo   repository.ProductRepository
	receptionRepo repository.ReceptionRepository
	trManager     *manager.Manager
	log           *logrus.Logger
}

func NewProductService(
	productRepo repository.ProductRepository, receptionRepo repository.ReceptionRepository, trManager *manager.Manager, log *logrus.Logger,
) *ProductService {
	return &ProductService{
		receptionRepo: receptionRepo,
		productRepo:   productRepo,
		trManager:     trManager,
		log:           log,
	}
}

func (s *ProductService) AddProduct(ctx context.Context, pvzID uuid.UUID, productType entity.ProductType) (*entity.Product, error) {
	if !entity.IsValidProductType(productType) {
		s.log.Warnf("invalid product type: %s", productType)
		return nil, entity.ErrInvalidProductType
	}

	var result *entity.Product

	err := s.trManager.Do(ctx, func(ctx context.Context) error {
		reception, err := s.receptionRepo.GetOpenReception(ctx, pvzID)
		if err != nil {
			s.log.Warnf("no open reception for pvz: %s, err: %v", pvzID, err)
			return entity.ErrNoActiveReception
		}

		product := &entity.Product{
			DateTime:    time.Now(),
			Type:        productType,
			ReceptionID: reception.ID,
		}

		if err := s.productRepo.CreateProduct(ctx, product); err != nil {
			s.log.Errorf("failed to create product: %v", err)
			return err
		}

		result = product
		return nil
	})

	if err != nil {
		return nil, err
	}

	s.log.Infof("product added to reception: type=%s, pvz=%s", result.Type, pvzID)
	monitoring.AddedProductsCounter.Inc()
	return result, nil
}

func (s *ProductService) DeleteLastProduct(ctx context.Context, pvzID uuid.UUID) error {
	return s.trManager.Do(ctx, func(ctx context.Context) error {
		reception, err := s.receptionRepo.GetOpenReception(ctx, pvzID)
		if err != nil {
			s.log.Warnf("no open reception for pvz: %s, err: %v", pvzID, err)
			return entity.ErrNoOpenReception
		}

		productID, err := s.productRepo.DeleteLastProduct(ctx, reception.ID)
		if err != nil {
			s.log.Errorf("failed to delete last product: %v", err)
			return err
		}

		if productID == nil {
			s.log.Warnf("no products to delete for reception: %s", reception.ID)
			return entity.ErrNoProductsToDelete
		}

		s.log.Infof("product deleted: %s", *productID)
		return nil
	})
}

func groupProductsByReceptionID(products []entity.Product) map[uuid.UUID][]entity.Product {
	result := make(map[uuid.UUID][]entity.Product)
	for _, product := range products {
		result[product.ReceptionID] = append(result[product.ReceptionID], product)
	}

	return result
}
