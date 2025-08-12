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

func TestProductService_AddProduct(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockReceptionRepo := mocks.NewMockReceptionRepository(ctrl)
	mockProductRepo := mocks.NewMockProductRepository(ctrl)
	db, mock, _ := sqlmock.New()
	mockDB := sqlx.NewDb(db, testDriverName)
	trManager := manager.Must(trmsqlx.NewDefaultFactory(mockDB))
	mockLog := logrus.New()

	svc := NewProductService(mockProductRepo, mockReceptionRepo, trManager, mockLog)

	validReception := &entity.Reception{
		ID: uuid.New(),
	}

	tests := []struct {
		name        string
		pvzID       uuid.UUID
		productType entity.ProductType
		setup       func()
		wantErr     error
	}{
		{
			name:        "success",
			pvzID:       uuid.New(),
			productType: entity.ProductClothing,
			setup: func() {
				mock.ExpectBegin()
				mockReceptionRepo.EXPECT().GetOpenReception(gomock.Any(), gomock.Any()).Return(validReception, nil)
				mockProductRepo.EXPECT().CreateProduct(gomock.Any(), gomock.Any()).Return(nil)
				mock.ExpectCommit()
			},
			wantErr: nil,
		},
		{
			name:        "invalid product type",
			pvzID:       uuid.New(),
			productType: "invalid",
			setup:       func() {},
			wantErr:     entity.ErrInvalidProductType,
		},
		{
			name:        "no open reception",
			pvzID:       uuid.New(),
			productType: entity.ProductShoes,
			setup: func() {
				mock.ExpectBegin()
				mockReceptionRepo.EXPECT().GetOpenReception(gomock.Any(), gomock.Any()).Return(nil, errors.New("not found"))
				mock.ExpectRollback()
			},
			wantErr: entity.ErrNoActiveReception,
		},
		{
			name:        "db error on create product",
			pvzID:       uuid.New(),
			productType: entity.ProductElectronics,
			setup: func() {
				mock.ExpectBegin()
				mockReceptionRepo.EXPECT().GetOpenReception(gomock.Any(), gomock.Any()).Return(validReception, nil)
				mockProductRepo.EXPECT().CreateProduct(gomock.Any(), gomock.Any()).Return(errors.New("insert error"))
				mock.ExpectRollback()
			},
			wantErr: errors.New("insert error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			_, err := svc.AddProduct(context.Background(), tt.pvzID, tt.productType)
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

func TestProductService_DeleteLastProduct(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockReceptionRepo := mocks.NewMockReceptionRepository(ctrl)
	mockProductRepo := mocks.NewMockProductRepository(ctrl)
	db, mock, _ := sqlmock.New()
	mockDB := sqlx.NewDb(db, testDriverName)
	trManager := manager.Must(trmsqlx.NewDefaultFactory(mockDB))
	mockLog := logrus.New()

	svc := NewProductService(mockProductRepo, mockReceptionRepo, trManager, mockLog)

	receptionID := uuid.New()
	reception := &entity.Reception{ID: receptionID}

	tests := []struct {
		name    string
		pvzID   uuid.UUID
		setup   func()
		wantErr error
	}{
		{
			name:  "success",
			pvzID: uuid.New(),
			setup: func() {
				mock.ExpectBegin()
				mockReceptionRepo.EXPECT().GetOpenReception(gomock.Any(), gomock.Any()).Return(reception, nil)
				mockProductRepo.EXPECT().DeleteLastProduct(gomock.Any(), receptionID).Return(&uuid.UUID{}, nil)
				mock.ExpectCommit()
			},
			wantErr: nil,
		},
		{
			name:  "no open reception",
			pvzID: uuid.New(),
			setup: func() {
				mock.ExpectBegin()
				mockReceptionRepo.EXPECT().GetOpenReception(gomock.Any(), gomock.Any()).Return(nil, errors.New("not found"))
				mock.ExpectRollback()
			},
			wantErr: entity.ErrNoOpenReception,
		},
		{
			name:  "no products to delete",
			pvzID: uuid.New(),
			setup: func() {
				mock.ExpectBegin()
				mockReceptionRepo.EXPECT().GetOpenReception(gomock.Any(), gomock.Any()).Return(reception, nil)
				mockProductRepo.EXPECT().DeleteLastProduct(gomock.Any(), receptionID).Return(nil, nil)
				mock.ExpectRollback()
			},
			wantErr: entity.ErrNoProductsToDelete,
		},
		{
			name:  "db error on delete",
			pvzID: uuid.New(),
			setup: func() {
				mock.ExpectBegin()
				mockReceptionRepo.EXPECT().GetOpenReception(gomock.Any(), gomock.Any()).Return(reception, nil)
				mockProductRepo.EXPECT().DeleteLastProduct(gomock.Any(), receptionID).Return(nil, errors.New("db error"))
				mock.ExpectRollback()
			},
			wantErr: errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			err := svc.DeleteLastProduct(context.Background(), tt.pvzID)
			if tt.wantErr != nil {
				assert.ErrorContains(t, err, tt.wantErr.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
