package handler

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/senyabanana/pvz-service/internal/entity"
	mocks "github.com/senyabanana/pvz-service/internal/service/mocks"
)

func TestProductHandler_AddProduct(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockProductOperations(ctrl)
	mockLog := logrus.New()
	h := NewProductHandler(mockService, mockLog)
	gin.SetMode(gin.TestMode)

	validID := uuid.New().String()

	tests := []struct {
		name       string
		input      string
		mock       func()
		wantStatus int
	}{
		{
			name:  "success",
			input: `{"pvzId":"` + validID + `", "type":"электроника"}`,
			mock: func() {
				mockService.EXPECT().AddProduct(gomock.Any(), gomock.Any(), entity.ProductElectronics).Return(&entity.Product{
					ID:          uuid.New(),
					DateTime:    time.Now(),
					Type:        entity.ProductElectronics,
					ReceptionID: uuid.New(),
				}, nil)
			},
			wantStatus: http.StatusCreated,
		},
		{
			name:       "invalid JSON",
			input:      `{"pvzId":123}`,
			mock:       func() {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid UUID format",
			input:      `{"pvzId":"not-a-uuid", "type":"электроника"}`,
			mock:       func() {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:  "no open reception",
			input: `{"pvzId":"` + validID + `", "type":"электроника"}`,
			mock: func() {
				mockService.EXPECT().AddProduct(gomock.Any(), gomock.Any(), entity.ProductElectronics).Return(nil, entity.ErrNoActiveReception)
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:  "invalid product type",
			input: `{"pvzId":"` + validID + `", "type":"invalid"}`,
			mock: func() {
				mockService.EXPECT().AddProduct(gomock.Any(), gomock.Any(), entity.ProductType("invalid")).Return(nil, entity.ErrInvalidProductType)
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:  "internal error",
			input: `{"pvzId":"` + validID + `", "type":"электроника"}`,
			mock: func() {
				mockService.EXPECT().AddProduct(gomock.Any(), gomock.Any(), entity.ProductElectronics).Return(nil, errors.New("db error"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			req, _ := http.NewRequest(http.MethodPost, "/products", bytes.NewBufferString(tt.input))
			req.Header.Set("Content-Type", "application/json")
			c.Request = req

			tt.mock()
			h.AddProduct(c)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestProductHandler_DeleteLastProduct(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockProductOperations(ctrl)
	mockLog := logrus.New()
	h := NewProductHandler(mockService, mockLog)
	gin.SetMode(gin.TestMode)

	validID := uuid.New().String()

	tests := []struct {
		name       string
		param      string
		mock       func()
		wantStatus int
	}{
		{
			name:  "success",
			param: validID,
			mock: func() {
				mockService.EXPECT().DeleteLastProduct(gomock.Any(), gomock.Any()).Return(nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "invalid UUID",
			param:      "not-a-uuid",
			mock:       func() {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:  "no open reception",
			param: validID,
			mock: func() {
				mockService.EXPECT().DeleteLastProduct(gomock.Any(), gomock.Any()).Return(entity.ErrNoOpenReception)
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:  "no products to delete",
			param: validID,
			mock: func() {
				mockService.EXPECT().DeleteLastProduct(gomock.Any(), gomock.Any()).Return(entity.ErrNoProductsToDelete)
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:  "internal error",
			param: validID,
			mock: func() {
				mockService.EXPECT().DeleteLastProduct(gomock.Any(), gomock.Any()).Return(errors.New("unexpected"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			req, _ := http.NewRequest(http.MethodPost, "/pvz/"+tt.param+"/delete_last_product", nil)
			c.Params = []gin.Param{{Key: "pvzId", Value: tt.param}}
			c.Request = req

			tt.mock()
			h.DeleteLastProduct(c)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}
