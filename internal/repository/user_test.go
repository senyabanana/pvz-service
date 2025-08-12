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
	"github.com/stretchr/testify/require"

	"github.com/senyabanana/pvz-service/internal/entity"
)

const testDriverName = "sqlmock"

func TestUserPostgres_CreateUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, testDriverName)
	repo := NewUserPostgres(sqlxDB)

	tests := []struct {
		name      string
		setupMock func()
		inputUser *entity.User
		wantErr   bool
	}{
		{
			name: "success",
			setupMock: func() {
				mock.ExpectExec(`INSERT INTO users`).
					WithArgs(sqlmock.AnyArg(), "test@example.com", "hashedpassword", entity.RoleClient, sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			inputUser: &entity.User{
				Email:    "test@example.com",
				Password: "hashedpassword",
				Role:     entity.RoleClient,
			},
			wantErr: false,
		},
		{
			name: "db error",
			setupMock: func() {
				mock.ExpectExec(`INSERT INTO users`).
					WithArgs(sqlmock.AnyArg(), "test@example.com", "hashedpassword", entity.RoleClient, sqlmock.AnyArg()).
					WillReturnError(errors.New("db failure"))
			},
			inputUser: &entity.User{
				Email:    "test@example.com",
				Password: "hashedpassword",
				Role:     entity.RoleClient,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			err := repo.CreateUser(context.Background(), tt.inputUser)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestUserPostgres_IsEmailExists(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, testDriverName)
	repo := NewUserPostgres(sqlxDB)

	tests := []struct {
		name      string
		setupMock func()
		email     string
		exists    bool
		wantErr   bool
	}{
		{
			name: "email exists",
			setupMock: func() {
				mock.ExpectQuery(`SELECT COUNT\(\*\) FROM users WHERE email = \$1`).
					WithArgs("exists@example.com").
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
			},
			email:   "exists@example.com",
			exists:  true,
			wantErr: false,
		},
		{
			name: "email does not exist",
			setupMock: func() {
				mock.ExpectQuery(`SELECT COUNT\(\*\) FROM users WHERE email = \$1`).
					WithArgs("notfound@example.com").
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
			},
			email:   "notfound@example.com",
			exists:  false,
			wantErr: false,
		},
		{
			name: "db error",
			setupMock: func() {
				mock.ExpectQuery(`SELECT COUNT\(\*\) FROM users WHERE email = \$1`).
					WithArgs("fail@example.com").
					WillReturnError(errors.New("query error"))
			},
			email:   "fail@example.com",
			exists:  false,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			exists, err := repo.IsEmailExists(context.Background(), tt.email)

			assert.Equal(t, tt.exists, exists)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestUserPostgres_GetUserByEmail(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, testDriverName)
	repo := NewUserPostgres(sqlxDB)

	now := time.Now()
	id := uuid.New()

	tests := []struct {
		name      string
		setupMock func()
		email     string
		expectErr bool
	}{
		{
			name: "success",
			setupMock: func() {
				mock.ExpectQuery(`SELECT id, email, password_hash, role, created_at FROM users WHERE email = \$1`).
					WithArgs("test@example.com").
					WillReturnRows(sqlmock.NewRows([]string{"id", "email", "password_hash", "role", "created_at"}).
						AddRow(id, "test@example.com", "hashed", entity.RoleClient, now))
			},
			email:     "test@example.com",
			expectErr: false,
		},
		{
			name: "query error",
			setupMock: func() {
				mock.ExpectQuery(`SELECT id, email, password_hash, role, created_at FROM users WHERE email = \$1`).
					WithArgs("fail@example.com").
					WillReturnError(errors.New("db error"))
			},
			email:     "fail@example.com",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			user, err := repo.GetUserByEmail(context.Background(), tt.email)
			if tt.expectErr {
				assert.Nil(t, user)
				assert.Error(t, err)
			} else {
				assert.NotNil(t, user)
				assert.Equal(t, tt.email, user.Email)
				assert.NoError(t, err)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
