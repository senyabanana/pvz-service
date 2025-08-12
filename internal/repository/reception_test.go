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

func TestReceptionPostgres_CreateReception(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, testDriverName)
	repo := NewReceptionPostgres(sqlxDB)

	now := time.Now()

	tests := []struct {
		name    string
		setup   func()
		input   *entity.Reception
		wantErr bool
	}{
		{
			name: "success",
			setup: func() {
				mock.ExpectExec(`INSERT INTO receptions`).
					WithArgs(sqlmock.AnyArg(), now, sqlmock.AnyArg(), entity.StatusInProgress, now).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			input: &entity.Reception{
				DateTime:  now,
				Status:    entity.StatusInProgress,
				CreatedAt: now,
				PVZID:     uuid.New(),
			},
			wantErr: false,
		},
		{
			name: "db error",
			setup: func() {
				mock.ExpectExec(`INSERT INTO receptions`).
					WithArgs(sqlmock.AnyArg(), now, sqlmock.AnyArg(), entity.StatusInProgress, now).
					WillReturnError(errors.New("insert error"))
			},
			input: &entity.Reception{
				DateTime:  now,
				Status:    entity.StatusInProgress,
				CreatedAt: now,
				PVZID:     uuid.New(),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			err := repo.CreateReception(context.Background(), tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestReceptionPostgres_IsReceptionOpenExists(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, testDriverName)
	repo := NewReceptionPostgres(sqlxDB)

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
				mock.ExpectQuery(`SELECT COUNT\(\*\) FROM receptions`).
					WithArgs(pvzID).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
			},
			expected: true,
			wantErr:  false,
		},
		{
			name: "not exists",
			setup: func() {
				mock.ExpectQuery(`SELECT COUNT\(\*\) FROM receptions`).
					WithArgs(pvzID).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
			},
			expected: false,
			wantErr:  false,
		},
		{
			name: "db error",
			setup: func() {
				mock.ExpectQuery(`SELECT COUNT\(\*\) FROM receptions`).
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
			result, err := repo.IsReceptionOpenExists(context.Background(), pvzID)
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

func TestReceptionPostgres_GetOpenReception(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, testDriverName)
	repo := NewReceptionPostgres(sqlxDB)

	pvzID := uuid.New()
	expected := entity.Reception{
		ID:        uuid.New(),
		DateTime:  time.Now(),
		PVZID:     pvzID,
		Status:    entity.StatusInProgress,
		CreatedAt: time.Now(),
	}

	tests := []struct {
		name      string
		setupMock func()
		wantErr   bool
	}{
		{
			name: "success",
			setupMock: func() {
				rows := sqlmock.NewRows([]string{"id", "date_time", "pvz_id", "status", "created_at", "closed_at"}).
					AddRow(expected.ID, expected.DateTime, expected.PVZID, expected.Status, expected.CreatedAt, nil)

				mock.ExpectQuery("SELECT id, date_time, pvz_id, status").
					WithArgs(pvzID).
					WillReturnRows(rows)
			},
			wantErr: false,
		},
		{
			name: "db error",
			setupMock: func() {
				mock.ExpectQuery("SELECT id, date_time, pvz_id, status").
					WithArgs(pvzID).
					WillReturnError(errors.New("db failure"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			_, err := repo.GetOpenReception(context.Background(), pvzID)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestReceptionPostgres_CloseReceptionByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, testDriverName)
	repo := NewReceptionPostgres(sqlxDB)

	receptionID := uuid.New()
	closedAt := time.Now()

	tests := []struct {
		name      string
		setupMock func()
		wantErr   error
	}{
		{
			name: "success",
			setupMock: func() {
				mock.ExpectExec("UPDATE receptions SET status = 'close'").
					WithArgs(receptionID, closedAt).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			wantErr: nil,
		},
		{
			name: "already closed",
			setupMock: func() {
				mock.ExpectExec("UPDATE receptions SET status = 'close'").
					WithArgs(receptionID, closedAt).
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
			wantErr: entity.ErrReceptionAlreadyClosed,
		},
		{
			name: "db error",
			setupMock: func() {
				mock.ExpectExec("UPDATE receptions SET status = 'close'").
					WithArgs(receptionID, closedAt).
					WillReturnError(errors.New("db error"))
			},
			wantErr: errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			err := repo.CloseReceptionByID(context.Background(), receptionID, closedAt)
			if tt.wantErr != nil {
				assert.Equal(t, tt.wantErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestReceptionPostgres_GetReceptionsByPVZIDs(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, testDriverName)
	repo := NewReceptionPostgres(sqlxDB)

	pvzID := uuid.New()
	expected := entity.Reception{
		ID:        uuid.New(),
		DateTime:  time.Now(),
		PVZID:     pvzID,
		Status:    entity.StatusInProgress,
		CreatedAt: time.Now(),
	}

	tests := []struct {
		name      string
		pvzIDs    []uuid.UUID
		setupMock func()
		wantLen   int
		wantErr   bool
	}{
		{
			name:      "empty input",
			pvzIDs:    nil,
			setupMock: func() {},
			wantLen:   0,
			wantErr:   false,
		},
		{
			name:   "success",
			pvzIDs: []uuid.UUID{pvzID},
			setupMock: func() {
				rows := sqlmock.NewRows([]string{"id", "date_time", "pvz_id", "status", "created_at", "closed_at"}).
					AddRow(expected.ID, expected.DateTime, expected.PVZID, expected.Status, expected.CreatedAt, sql.NullTime{})
				mock.ExpectQuery("SELECT id, date_time, pvz_id, status").
					WithArgs(sqlmock.AnyArg()).
					WillReturnRows(rows)
			},
			wantLen: 1,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			result, err := repo.GetReceptionsByPVZIDs(context.Background(), tt.pvzIDs)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, result, tt.wantLen)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
