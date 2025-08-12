package service

import (
	"context"
	"errors"
	"testing"
	"time"

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

func TestPVZService_CreatePVZ(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPVZRepo := mocks.NewMockPVZRepository(ctrl)
	mockLog := logrus.New()

	svc := NewPVZService(mockPVZRepo, nil, nil, nil, mockLog)

	tests := []struct {
		name    string
		city    string
		setup   func()
		wantErr error
	}{
		{
			name: "success",
			city: string(entity.CityMoscow),
			setup: func() {
				mockPVZRepo.EXPECT().CreatePVZ(gomock.Any(), gomock.Any()).Return(nil)
			},
			wantErr: nil,
		},
		{
			name:    "invalid city",
			city:    "Тверь",
			setup:   func() {},
			wantErr: entity.ErrInvalidCity,
		},
		{
			name: "repo error",
			city: string(entity.CityKazan),
			setup: func() {
				mockPVZRepo.EXPECT().CreatePVZ(gomock.Any(), gomock.Any()).Return(errors.New("insert failed"))
			},
			wantErr: errors.New("insert failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			_, err := svc.CreatePVZ(context.Background(), tt.city)
			if tt.wantErr != nil {
				assert.ErrorContains(t, err, tt.wantErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPVZService_GetFullPVZInfo(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPVZRepo := mocks.NewMockPVZRepository(ctrl)
	mockReceptionRepo := mocks.NewMockReceptionRepository(ctrl)
	mockProductRepo := mocks.NewMockProductRepository(ctrl)
	db, mock, _ := sqlmock.New()
	mockDB := sqlx.NewDb(db, "sqlmock")
	trxManager := manager.Must(trmsqlx.NewDefaultFactory(mockDB))
	mockLog := logrus.New()

	svc := NewPVZService(mockPVZRepo, mockReceptionRepo, mockProductRepo, trxManager, mockLog)

	now := time.Now()
	pvzID := uuid.New()
	receptionID := uuid.New()

	tests := []struct {
		name         string
		setup        func()
		wantErr      bool
		expectedSize int
	}{
		{
			name: "success",
			setup: func() {
				mock.ExpectBegin()
				mockPVZRepo.EXPECT().GetAllPVZ(gomock.Any()).Return([]entity.PVZ{
					{ID: pvzID, RegistrationDate: now, City: "Москва"},
				}, nil)

				mockReceptionRepo.EXPECT().
					GetReceptionsByPVZIDs(gomock.Any(), []uuid.UUID{pvzID}).
					Return([]entity.Reception{
						{ID: receptionID, PVZID: pvzID, CreatedAt: now, DateTime: now},
					}, nil)

				mockProductRepo.EXPECT().
					GetProductsByReceptionIDs(gomock.Any(), []uuid.UUID{receptionID}).
					Return([]entity.Product{
						{ID: uuid.New(), Type: entity.ProductClothing, ReceptionID: receptionID, DateTime: now},
					}, nil)
				mock.ExpectCommit()
			},
			wantErr:      false,
			expectedSize: 1,
		},
		{
			name: "fail on GetAllPVZ",
			setup: func() {
				mock.ExpectBegin()
				mockPVZRepo.EXPECT().GetAllPVZ(gomock.Any()).Return(nil, errors.New("db error"))
				mock.ExpectRollback()
			},
			wantErr: true,
		},
		{
			name: "fail on GetReceptionsByPVZIDs",
			setup: func() {
				mock.ExpectBegin()
				mockPVZRepo.EXPECT().GetAllPVZ(gomock.Any()).Return([]entity.PVZ{
					{ID: pvzID, RegistrationDate: now, City: "Казань"},
				}, nil)
				mockReceptionRepo.EXPECT().
					GetReceptionsByPVZIDs(gomock.Any(), []uuid.UUID{pvzID}).
					Return(nil, errors.New("fail get receptions"))
				mock.ExpectRollback()
			},
			wantErr: true,
		},
		{
			name: "fail on GetProductsByReceptionIDs",
			setup: func() {
				mock.ExpectBegin()
				mockPVZRepo.EXPECT().GetAllPVZ(gomock.Any()).Return([]entity.PVZ{
					{ID: pvzID, RegistrationDate: now, City: "Казань"},
				}, nil)

				mockReceptionRepo.EXPECT().
					GetReceptionsByPVZIDs(gomock.Any(), []uuid.UUID{pvzID}).
					Return([]entity.Reception{
						{ID: receptionID, PVZID: pvzID, CreatedAt: now, DateTime: now},
					}, nil)

				mockProductRepo.EXPECT().
					GetProductsByReceptionIDs(gomock.Any(), []uuid.UUID{receptionID}).
					Return(nil, errors.New("fail get products"))
				mock.ExpectRollback()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			resp, err := svc.GetFullPVZInfo(context.Background(), nil, nil, 1, 10)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, resp, tt.expectedSize)
			}
		})
	}
}

func TestPVZService_GetAllPVZ(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPVZRepo := mocks.NewMockPVZRepository(ctrl)
	mockLog := logrus.New()

	svc := NewPVZService(mockPVZRepo, nil, nil, nil, mockLog)

	now := time.Now()
	expectedPVZ := []entity.PVZ{
		{ID: uuid.New(), RegistrationDate: now, City: entity.CityMoscow},
	}

	tests := []struct {
		name     string
		setup    func()
		wantErr  error
		expected []entity.PVZ
	}{
		{
			name: "success",
			setup: func() {
				mockPVZRepo.EXPECT().
					GetAllPVZ(gomock.Any()).
					Return(expectedPVZ, nil)
			},
			wantErr:  nil,
			expected: expectedPVZ,
		},
		{
			name: "repo error",
			setup: func() {
				mockPVZRepo.EXPECT().
					GetAllPVZ(gomock.Any()).
					Return(nil, errors.New("db fail"))
			},
			wantErr:  errors.New("db fail"),
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()

			result, err := svc.GetAllPVZ(context.Background())
			if tt.wantErr != nil {
				assert.ErrorContains(t, err, tt.wantErr.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.expected, result)
		})
	}
}
