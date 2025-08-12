package repository

import (
	"context"
	"database/sql"
	"errors"

	trmsqlx "github.com/avito-tech/go-transaction-manager/drivers/sqlx/v2"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/senyabanana/pvz-service/internal/entity"
)

type ProductPostgres struct {
	db     *sqlx.DB
	getter *trmsqlx.CtxGetter
}

func NewProductPostgres(db *sqlx.DB) *ProductPostgres {
	return &ProductPostgres{
		db:     db,
		getter: trmsqlx.DefaultCtxGetter,
	}
}

func (r *ProductPostgres) CreateProduct(ctx context.Context, product *entity.Product) error {
	product.ID = uuid.New()
	query := `INSERT INTO products (id, date_time, type, reception_id) VALUES ($1, $2, $3, $4)`
	_, err := r.getter.DefaultTrOrDB(ctx, r.db).
		ExecContext(ctx, query, product.ID, product.DateTime, product.Type, product.ReceptionID)

	return err
}

func (r *ProductPostgres) DeleteLastProduct(ctx context.Context, receptionID uuid.UUID) (*uuid.UUID, error) {
	var productID uuid.UUID
	query := `
		DELETE FROM products
		WHERE id = (
		    SELECT id FROM products
		    WHERE reception_id = $1
		    ORDER BY date_time DESC
		    LIMIT 1
		)
		RETURNING id
		`
	err := r.getter.DefaultTrOrDB(ctx, r.db).GetContext(ctx, &productID, query, receptionID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	return &productID, nil
}

func (r *ProductPostgres) GetProductsByReceptionIDs(ctx context.Context, receptionIDs []uuid.UUID) ([]entity.Product, error) {
	var products []entity.Product

	if len(receptionIDs) == 0 {
		return nil, nil
	}

	query, args, err := sqlx.In(`
			SELECT id, date_time, type, reception_id
			FROM products
			WHERE reception_id IN (?)`, receptionIDs)
	if err != nil {
		return nil, err
	}

	query = r.db.Rebind(query)
	err = r.getter.DefaultTrOrDB(ctx, r.db).SelectContext(ctx, &products, query, args...)
	if err != nil {
		return nil, err
	}

	return products, nil
}
