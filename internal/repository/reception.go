package repository

import (
	"context"
	"time"

	trmsqlx "github.com/avito-tech/go-transaction-manager/drivers/sqlx/v2"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/senyabanana/pvz-service/internal/entity"
)

type ReceptionPostgres struct {
	db     *sqlx.DB
	getter *trmsqlx.CtxGetter
}

func NewReceptionPostgres(db *sqlx.DB) *ReceptionPostgres {
	return &ReceptionPostgres{
		db:     db,
		getter: trmsqlx.DefaultCtxGetter,
	}
}

func (r *ReceptionPostgres) CreateReception(ctx context.Context, reception *entity.Reception) error {
	reception.ID = uuid.New()
	query := `INSERT INTO receptions (id, date_time, pvz_id, status, created_at) VALUES ($1, $2, $3, $4, $5)`
	_, err := r.getter.DefaultTrOrDB(ctx, r.db).
		ExecContext(ctx, query, reception.ID, reception.DateTime, reception.PVZID, reception.Status, reception.CreatedAt)

	return err
}

func (r *ReceptionPostgres) IsReceptionOpenExists(ctx context.Context, pvzID uuid.UUID) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM receptions WHERE pvz_id = $1 AND status = 'in_progress'`
	err := r.getter.DefaultTrOrDB(ctx, r.db).GetContext(ctx, &count, query, pvzID)

	return count > 0, err
}

func (r *ReceptionPostgres) GetOpenReception(ctx context.Context, pvzID uuid.UUID) (*entity.Reception, error) {
	var reception entity.Reception
	query := `
		SELECT id, date_time, pvz_id, status, created_at, closed_at
		FROM receptions WHERE pvz_id = $1 AND status = 'in_progress'
		ORDER BY created_at DESC
		LIMIT 1
		`
	err := r.getter.DefaultTrOrDB(ctx, r.db).GetContext(ctx, &reception, query, pvzID)
	if err != nil {
		return nil, err
	}

	return &reception, nil
}

func (r *ReceptionPostgres) CloseReceptionByID(ctx context.Context, receptionID uuid.UUID, closedAt time.Time) error {
	query := `UPDATE receptions SET status = 'close', closed_at = $2 WHERE id = $1 AND status = 'in_progress'`
	res, err := r.getter.DefaultTrOrDB(ctx, r.db).ExecContext(ctx, query, receptionID, closedAt)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return entity.ErrReceptionAlreadyClosed
	}

	return nil
}

func (r *ReceptionPostgres) GetReceptionsByPVZIDs(ctx context.Context, pvzIDs []uuid.UUID) ([]entity.Reception, error) {
	var receptions []entity.Reception

	if len(pvzIDs) == 0 {
		return nil, nil
	}

	query, args, err := sqlx.In(`
			SELECT id, date_time, pvz_id, status, created_at, closed_at
			FROM receptions
			WHERE pvz_id IN (?)`, pvzIDs)
	if err != nil {
		return nil, err
	}

	query = r.db.Rebind(query)
	err = r.getter.DefaultTrOrDB(ctx, r.db).SelectContext(ctx, &receptions, query, args...)
	if err != nil {
		return nil, err
	}

	return receptions, nil
}
