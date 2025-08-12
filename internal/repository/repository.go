package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/senyabanana/pvz-service/internal/entity"
)

//go:generate mockgen -source=repository.go -destination=mocks/mock.go

type UserRepository interface {
	CreateUser(ctx context.Context, user *entity.User) error
	IsEmailExists(ctx context.Context, email string) (bool, error)
	GetUserByEmail(ctx context.Context, email string) (*entity.User, error)
}

type PVZRepository interface {
	CreatePVZ(ctx context.Context, pvz *entity.PVZ) error
	IsPVZExists(ctx context.Context, pvzID uuid.UUID) (bool, error)
	GetAllPVZ(ctx context.Context) ([]entity.PVZ, error)
}

type ReceptionRepository interface {
	CreateReception(ctx context.Context, reception *entity.Reception) error
	IsReceptionOpenExists(ctx context.Context, pvzID uuid.UUID) (bool, error)
	GetOpenReception(ctx context.Context, pvzID uuid.UUID) (*entity.Reception, error)
	CloseReceptionByID(ctx context.Context, receptionID uuid.UUID, closedAt time.Time) error
	GetReceptionsByPVZIDs(ctx context.Context, pvzIDs []uuid.UUID) ([]entity.Reception, error)
}

type ProductRepository interface {
	CreateProduct(ctx context.Context, product *entity.Product) error
	DeleteLastProduct(ctx context.Context, receptionID uuid.UUID) (*uuid.UUID, error)
	GetProductsByReceptionIDs(ctx context.Context, receptionIDs []uuid.UUID) ([]entity.Product, error)
}

type Repository struct {
	UserRepository
	PVZRepository
	ReceptionRepository
	ProductRepository
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{
		UserRepository:      NewUserPostgres(db),
		PVZRepository:       NewPVZPostgres(db),
		ReceptionRepository: NewReceptionPostgres(db),
		ProductRepository:   NewProductPostgres(db),
	}
}
