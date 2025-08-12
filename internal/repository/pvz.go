package repository

import (
	"context"

	trmsqlx "github.com/avito-tech/go-transaction-manager/drivers/sqlx/v2"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/senyabanana/pvz-service/internal/entity"
)

type PVZPostgres struct {
	db     *sqlx.DB
	getter *trmsqlx.CtxGetter
}

func NewPVZPostgres(db *sqlx.DB) *PVZPostgres {
	return &PVZPostgres{
		db:     db,
		getter: trmsqlx.DefaultCtxGetter,
	}
}

func (r *PVZPostgres) CreatePVZ(ctx context.Context, pvz *entity.PVZ) error {
	pvz.ID = uuid.New()
	pvz.RegistrationDate = pvz.RegistrationDate.UTC()
	query := `INSERT INTO pvz (id, registration_date, city) VALUES ($1, $2, $3)`
	_, err := r.getter.DefaultTrOrDB(ctx, r.db).
		ExecContext(ctx, query, pvz.ID, pvz.RegistrationDate, pvz.City)

	return err
}

func (r *PVZPostgres) IsPVZExists(ctx context.Context, pvzID uuid.UUID) (bool, error) {
	var exists bool
	query := `SELECT EXISTS (SELECT 1 FROM pvz WHERE id = $1)`
	err := r.getter.DefaultTrOrDB(ctx, r.db).GetContext(ctx, &exists, query, pvzID)

	return exists, err
}

func (r *PVZPostgres) GetAllPVZ(ctx context.Context) ([]entity.PVZ, error) {
	var allPVZ []entity.PVZ
	query := `SELECT id, registration_date, city FROM pvz ORDER BY registration_date DESC`
	err := r.getter.DefaultTrOrDB(ctx, r.db).SelectContext(ctx, &allPVZ, query)
	if err != nil {
		return nil, err
	}

	return allPVZ, nil
}
