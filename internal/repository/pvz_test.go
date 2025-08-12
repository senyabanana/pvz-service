package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"

	"github.com/senyabanana/pvz-service/internal/entity"
)

func TestPVZPostgres_CreatePVZ(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, testDriverName)
	repo := NewPVZPostgres(sqlxDB)

	now := time.Now().UTC()

	tests := []struct {
		name    string
		setup   func()
		input   *entity.PVZ
		wantErr bool
	}{
		{
			name: "success",
			setup: func() {
				mock.ExpectExec(`INSERT INTO pvz`).
					WithArgs(sqlmock.AnyArg(), now, entity.CityMoscow).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			input: &entity.PVZ{
				RegistrationDate: now,
				City:             entity.CityMoscow,
			},
			wantErr: false,
		},
		{
			name: "db error",
			setup: func() {
				mock.ExpectExec(`INSERT INTO pvz`).
					WithArgs(sqlmock.AnyArg(), now, entity.CityMoscow).
					WillReturnError(errors.New("insert error"))
			},
			input: &entity.PVZ{
				RegistrationDate: now,
				City:             entity.CityMoscow,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			err := repo.CreatePVZ(context.Background(), tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestPVZPostgres_IsPVZExists(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, testDriverName)
	repo := NewPVZPostgres(sqlxDB)

	pvzID := uuid.New()

	tests := []struct {
		name     string
		setup    func()
		expected bool
		wantErr  bool
	}{
		{
			name: "exists",
			setup: func() {
				mock.ExpectQuery(`SELECT EXISTS`).
					WithArgs(pvzID).
					WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
			},
			expected: true,
			wantErr:  false,
		},
		{
			name: "not exists",
			setup: func() {
				mock.ExpectQuery(`SELECT EXISTS`).
					WithArgs(pvzID).
					WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
			},
			expected: false,
			wantErr:  false,
		},
		{
			name: "db error",
			setup: func() {
				mock.ExpectQuery(`SELECT EXISTS`).
					WithArgs(pvzID).
					WillReturnError(errors.New("query error"))
			},
			expected: false,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			result, err := repo.IsPVZExists(context.Background(), pvzID)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestPVZPostgres_GetAllPVZ(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, testDriverName)
	repo := NewPVZPostgres(sqlxDB)

	now := time.Now().UTC()

	tests := []struct {
		name    string
		setup   func()
		wantErr bool
	}{
		{
			name: "success",
			setup: func() {
				mock.ExpectQuery(`SELECT id, registration_date, city FROM pvz`).
					WillReturnRows(
						sqlmock.NewRows([]string{"id", "registration_date", "city"}).
							AddRow(uuid.New(), now, entity.CityMoscow),
					)
			},
			wantErr: false,
		},
		{
			name: "query error",
			setup: func() {
				mock.ExpectQuery(`SELECT id, registration_date, city FROM pvz`).
					WillReturnError(errors.New("query error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			_, err := repo.GetAllPVZ(context.Background())
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
