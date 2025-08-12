package repository

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"

	"github.com/senyabanana/pvz-service/internal/entity"
)

func TestProductPostgres_CreateProduct(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, testDriverName)
	repo := NewProductPostgres(sqlxDB)

	tests := []struct {
		name    string
		setup   func()
		input   *entity.Product
		wantErr bool
	}{
		{
			name: "success",
			setup: func() {
				mock.ExpectExec(`INSERT INTO products`).
					WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), "электроника", sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			input: &entity.Product{
				Type:        entity.ProductElectronics,
				ReceptionID: uuid.New(),
				DateTime:    time.Now(),
			},
			wantErr: false,
		},
		{
			name: "db failure",
			setup: func() {
				mock.ExpectExec(`INSERT INTO products`).
					WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), "электроника", sqlmock.AnyArg()).
					WillReturnError(errors.New("db error"))
			},
			input: &entity.Product{
				Type:        entity.ProductElectronics,
				ReceptionID: uuid.New(),
				DateTime:    time.Now(),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			err := repo.CreateProduct(context.Background(), tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestProductPostgres_DeleteLastProduct(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, testDriverName)
	repo := NewProductPostgres(sqlxDB)

	productID := uuid.New()
	receptionID := uuid.New()

	tests := []struct {
		name     string
		setup    func()
		expected *uuid.UUID
		wantErr  bool
	}{
		{
			name: "success",
			setup: func() {
				mock.ExpectQuery(`DELETE FROM products`).
					WithArgs(receptionID).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(productID))
			},
			expected: &productID,
			wantErr:  false,
		},
		{
			name: "no rows",
			setup: func() {
				mock.ExpectQuery(`DELETE FROM products`).
					WithArgs(receptionID).
					WillReturnError(sql.ErrNoRows)
			},
			expected: nil,
			wantErr:  false,
		},
		{
			name: "db error",
			setup: func() {
				mock.ExpectQuery(`DELETE FROM products`).
					WithArgs(receptionID).
					WillReturnError(errors.New("db error"))
			},
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			id, err := repo.DeleteLastProduct(context.Background(), receptionID)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.Equal(t, tt.expected, id)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestProductPostgres_GetProductsByReceptionIDs(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, testDriverName)
	repo := NewProductPostgres(sqlxDB)

	id := uuid.New()
	receptionID := uuid.New()

	tests := []struct {
		name    string
		input   []uuid.UUID
		setup   func()
		wantErr bool
	}{
		{
			name:    "empty input",
			input:   []uuid.UUID{},
			setup:   func() {},
			wantErr: false,
		},
		{
			name:  "success",
			input: []uuid.UUID{receptionID},
			setup: func() {
				mock.ExpectQuery(`SELECT id, date_time, type, reception_id FROM products`).
					WithArgs(receptionID).
					WillReturnRows(sqlmock.NewRows([]string{"id", "date_time", "type", "reception_id"}).
						AddRow(id, time.Now(), "одежда", receptionID))
			},
			wantErr: false,
		},
		{
			name:  "db error",
			input: []uuid.UUID{receptionID},
			setup: func() {
				mock.ExpectQuery(`SELECT id, date_time, type, reception_id FROM products`).
					WithArgs(receptionID).
					WillReturnError(errors.New("db error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			_, err := repo.GetProductsByReceptionIDs(context.Background(), tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
