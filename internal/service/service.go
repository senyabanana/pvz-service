package service

import (
	"context"
	"time"

	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"github.com/senyabanana/pvz-service/internal/entity"
	"github.com/senyabanana/pvz-service/internal/repository"
)

//go:generate mockgen -source=service.go -destination=mocks/mock.go

type Authorization interface {
	RegisterUser(ctx context.Context, user *entity.User) error
	LoginUser(ctx context.Context, email, password string) (string, error)
}

type PVZOperations interface {
	CreatePVZ(ctx context.Context, city string) (*entity.PVZ, error)
	GetFullPVZInfo(ctx context.Context, startDate, endDate *time.Time, page, limit int) ([]entity.FullPVZInfo, error)
	GetAllPVZ(ctx context.Context) ([]entity.PVZ, error)
}

type ReceptionOperations interface {
	CreateReception(ctx context.Context, pvzID uuid.UUID) (*entity.Reception, error)
	CloseLastReception(ctx context.Context, pvzID uuid.UUID) (*entity.Reception, error)
}

type ProductOperations interface {
	AddProduct(ctx context.Context, pvzID uuid.UUID, productType entity.ProductType) (*entity.Product, error)
	DeleteLastProduct(ctx context.Context, pvzID uuid.UUID) error
}

type Service struct {
	Authorization
	PVZOperations
	ReceptionOperations
	ProductOperations
}

func NewService(repos *repository.Repository, trManager *manager.Manager, secretKey string, log *logrus.Logger) *Service {
	return &Service{
		Authorization:       NewUserService(repos, trManager, secretKey, log),
		PVZOperations:       NewPVZService(repos, repos, repos, trManager, log),
		ReceptionOperations: NewReceptionService(repos, repos, trManager, log),
		ProductOperations:   NewProductService(repos, repos, trManager, log),
	}
}
