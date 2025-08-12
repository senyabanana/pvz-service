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
	mocks "github.com/senyabanana/pvz-service/internal/repository/mocks"
)

func TestReceptionService_CreateReception(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockReceptionRepo := mocks.NewMockReceptionRepository(ctrl)
	mockPVZRepo := mocks.NewMockPVZRepository(ctrl)
	db, mock, _ := sqlmock.New()
	mockDB := sqlx.NewDb(db, testDriverName)
	mockTrManager := manager.Must(trmsqlx.NewDefaultFactory(mockDB))
	mockLog := logrus.New()

	svc := NewReceptionService(mockReceptionRepo, mockPVZRepo, mockTrManager, mockLog)

	pvzID := uuid.New()

	tests := []struct {
		name    string
		setup   func()
		wantErr error
	}{
		{
			name: "success",
			setup: func() {
				mock.ExpectBegin()
				mockPVZRepo.EXPECT().IsPVZExists(gomock.Any(), pvzID).Return(true, nil)
				mockReceptionRepo.EXPECT().IsReceptionOpenExists(gomock.Any(), pvzID).Return(false, nil)
				mockReceptionRepo.EXPECT().CreateReception(gomock.Any(), gomock.Any()).Return(nil)
				mock.ExpectCommit()
			},
			wantErr: nil,
		},
		{
			name: "pvz not found",
			setup: func() {
				mock.ExpectBegin()
				mockPVZRepo.EXPECT().IsPVZExists(gomock.Any(), pvzID).Return(false, nil)
				mock.ExpectRollback()
			},
			wantErr: entity.ErrPVZNotFound,
		},
		{
			name: "reception already exists",
			setup: func() {
				mock.ExpectBegin()
				mockPVZRepo.EXPECT().IsPVZExists(gomock.Any(), pvzID).Return(true, nil)
				mockReceptionRepo.EXPECT().IsReceptionOpenExists(gomock.Any(), pvzID).Return(true, nil)
				mock.ExpectRollback()
			},
			wantErr: entity.ErrReceptionAlreadyExists,
		},
		{
			name: "db error on create",
			setup: func() {
				mock.ExpectBegin()
				mockPVZRepo.EXPECT().IsPVZExists(gomock.Any(), pvzID).Return(true, nil)
				mockReceptionRepo.EXPECT().IsReceptionOpenExists(gomock.Any(), pvzID).Return(false, nil)
				mockReceptionRepo.EXPECT().CreateReception(gomock.Any(), gomock.Any()).Return(errors.New("create error"))
				mock.ExpectRollback()
			},
			wantErr: errors.New("create error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			_, err := svc.CreateReception(context.Background(), pvzID)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.ErrorContains(t, err, tt.wantErr.Error())
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestReceptionService_CloseLastReception(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockReceptionRepo := mocks.NewMockReceptionRepository(ctrl)
	mockPVZRepo := mocks.NewMockPVZRepository(ctrl)
	db, mock, _ := sqlmock.New()
	mockDB := sqlx.NewDb(db, testDriverName)
	mockTrManager := manager.Must(trmsqlx.NewDefaultFactory(mockDB))
	mockLog := logrus.New()

	svc := NewReceptionService(mockReceptionRepo, mockPVZRepo, mockTrManager, mockLog)

	pvzID := uuid.New()
	reception := &entity.Reception{
		ID:     uuid.New(),
		PVZID:  pvzID,
		Status: entity.StatusInProgress,
	}

	tests := []struct {
		name    string
		setup   func()
		wantErr error
	}{
		{
			name: "success",
			setup: func() {
				mock.ExpectBegin()
				mockReceptionRepo.EXPECT().GetOpenReception(gomock.Any(), pvzID).Return(reception, nil)
				mockReceptionRepo.EXPECT().CloseReceptionByID(gomock.Any(), reception.ID, gomock.Any()).Return(nil)
				mock.ExpectCommit()
			},
			wantErr: nil,
		},
		{
			name: "no open reception",
			setup: func() {
				mock.ExpectBegin()
				mockReceptionRepo.EXPECT().GetOpenReception(gomock.Any(), pvzID).Return(nil, errors.New("not found"))
				mock.ExpectRollback()
			},
			wantErr: entity.ErrNoOpenReception,
		},
		{
			name: "already closed",
			setup: func() {
				mock.ExpectBegin()
				mockReceptionRepo.EXPECT().GetOpenReception(gomock.Any(), pvzID).Return(reception, nil)
				mockReceptionRepo.EXPECT().CloseReceptionByID(gomock.Any(), reception.ID, gomock.Any()).
					Return(entity.ErrReceptionAlreadyClosed)
				mock.ExpectRollback()
			},
			wantErr: entity.ErrReceptionAlreadyClosed,
		},
		{
			name: "db error on close",
			setup: func() {
				mock.ExpectBegin()
				mockReceptionRepo.EXPECT().GetOpenReception(gomock.Any(), pvzID).Return(reception, nil)
				mockReceptionRepo.EXPECT().CloseReceptionByID(gomock.Any(), reception.ID, gomock.Any()).
					Return(errors.New("close error"))
				mock.ExpectRollback()
			},
			wantErr: errors.New("close error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			_, err := svc.CloseLastReception(context.Background(), pvzID)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.ErrorContains(t, err, tt.wantErr.Error())
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
