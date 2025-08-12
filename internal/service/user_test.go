package service

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	trmsqlx "github.com/avito-tech/go-transaction-manager/drivers/sqlx/v2"
	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/senyabanana/pvz-service/internal/entity"
	"github.com/senyabanana/pvz-service/internal/infrastructure/security"
	mocks "github.com/senyabanana/pvz-service/internal/repository/mocks"
)

const (
	testDriverName = "sqlmock"
	testJWTSecret  = "secret"
)

func TestUserService_RegisterUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockUserRepository(ctrl)
	db, mock, _ := sqlmock.New()
	mockDB := sqlx.NewDb(db, testDriverName)
	mockTrManager := manager.Must(trmsqlx.NewDefaultFactory(mockDB))
	mockLog := logrus.New()

	svc := NewUserService(mockRepo, mockTrManager, testJWTSecret, mockLog)

	tests := []struct {
		name      string
		inputUser *entity.User
		setup     func()
		wantErr   error
	}{
		{
			name: "success",
			inputUser: &entity.User{
				Email:    "test@example.com",
				Password: "password",
				Role:     entity.RoleClient,
			},
			setup: func() {
				mock.ExpectBegin()
				mockRepo.EXPECT().IsEmailExists(gomock.Any(), "test@example.com").Return(false, nil)
				mockRepo.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Return(nil)
				mock.ExpectCommit()
			},
			wantErr: nil,
		},
		{
			name: "email exists",
			inputUser: &entity.User{
				Email:    "exists@example.com",
				Password: "password",
				Role:     entity.RoleClient,
			},
			setup: func() {
				mock.ExpectBegin()
				mockRepo.EXPECT().IsEmailExists(gomock.Any(), "exists@example.com").Return(true, nil)
				mock.ExpectRollback()
			},
			wantErr: entity.ErrEmailTaken,
		},
		{
			name: "invalid role",
			inputUser: &entity.User{
				Email:    "bad@example.com",
				Password: "password",
				Role:     "invalid",
			},
			setup:   func() {},
			wantErr: entity.ErrInvalidUserRole,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			err := svc.RegisterUser(context.Background(), tt.inputUser)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}

			err = mock.ExpectationsWereMet()
			assert.NoError(t, err)
		})
	}
}

func TestUserService_LoginUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockUserRepository(ctrl)
	mockLog := logrus.New()

	svc := NewUserService(mockRepo, nil, testJWTSecret, mockLog)

	hashedPassword, _ := security.GeneratePasswordHash("correct-password")
	user := &entity.User{
		ID:       uuid.New(),
		Email:    "test@example.com",
		Password: hashedPassword,
		Role:     entity.RoleClient,
	}

	tests := []struct {
		name      string
		email     string
		password  string
		setup     func()
		wantErr   error
		expectJWT bool
	}{
		{
			name:     "success login",
			email:    "test@example.com",
			password: "correct-password",
			setup: func() {
				mockRepo.EXPECT().GetUserByEmail(gomock.Any(), "test@example.com").Return(user, nil)
			},
			wantErr:   nil,
			expectJWT: true,
		},
		{
			name:     "invalid credentials - wrong password",
			email:    "test@example.com",
			password: "wrong-password",
			setup: func() {
				mockRepo.EXPECT().GetUserByEmail(gomock.Any(), "test@example.com").Return(user, nil)
			},
			wantErr:   entity.ErrInvalidCredentials,
			expectJWT: false,
		},
		{
			name:     "invalid credentials - user not found",
			email:    "notfound@example.com",
			password: "irrelevant",
			setup: func() {
				mockRepo.EXPECT().GetUserByEmail(gomock.Any(), "notfound@example.com").Return(nil, errors.New("sql: no rows in result set"))
			},
			wantErr:   entity.ErrInvalidCredentials,
			expectJWT: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			token, err := svc.LoginUser(context.Background(), tt.email, tt.password)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Empty(t, token)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, token)
			}
		})
	}
}
