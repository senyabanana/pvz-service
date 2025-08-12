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

func TestReceptionHandler_CreateReception(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockReceptionOperations(ctrl)
	mockLog := logrus.New()
	h := NewReceptionHandler(mockService, mockLog)
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		inputBody  string
		mock       func()
		wantStatus int
	}{
		{
			name:      "success",
			inputBody: `{"pvzId":"` + uuid.New().String() + `"}`,
			mock: func() {
				mockService.EXPECT().CreateReception(gomock.Any(), gomock.Any()).Return(&entity.Reception{
					ID:       uuid.New(),
					DateTime: time.Now(),
					PVZID:    uuid.New(),
					Status:   entity.StatusInProgress,
				}, nil)
			},
			wantStatus: http.StatusCreated,
		},
		{
			name:       "invalid JSON",
			inputBody:  `{"pvzId":123}`,
			mock:       func() {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid UUID format",
			inputBody:  `{"pvzId":"not-a-uuid"}`,
			mock:       func() {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:      "PVZ not found",
			inputBody: `{"pvzId":"` + uuid.New().String() + `"}`,
			mock: func() {
				mockService.EXPECT().CreateReception(gomock.Any(), gomock.Any()).Return(nil, entity.ErrPVZNotFound)
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:      "reception already exists",
			inputBody: `{"pvzId":"` + uuid.New().String() + `"}`,
			mock: func() {
				mockService.EXPECT().CreateReception(gomock.Any(), gomock.Any()).Return(nil, entity.ErrReceptionAlreadyExists)
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:      "internal error",
			inputBody: `{"pvzId":"` + uuid.New().String() + `"}`,
			mock: func() {
				mockService.EXPECT().CreateReception(gomock.Any(), gomock.Any()).Return(nil, errors.New("db error"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			req, _ := http.NewRequest(http.MethodPost, "/receptions", bytes.NewBufferString(tt.inputBody))
			req.Header.Set("Content-Type", "application/json")
			c.Request = req

			tt.mock()
			h.CreateReception(c)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestReceptionHandler_CloseLastReception(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockReceptionOperations(ctrl)
	mockLog := logrus.New()
	h := NewReceptionHandler(mockService, mockLog)
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
				mockService.EXPECT().CloseLastReception(gomock.Any(), gomock.Any()).Return(&entity.Reception{
					ID:       uuid.New(),
					DateTime: time.Now(),
					PVZID:    uuid.New(),
					Status:   entity.StatusClosed,
				}, nil)
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
				mockService.EXPECT().CloseLastReception(gomock.Any(), gomock.Any()).Return(nil, entity.ErrNoOpenReception)
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:  "already closed",
			param: validID,
			mock: func() {
				mockService.EXPECT().CloseLastReception(gomock.Any(), gomock.Any()).Return(nil, entity.ErrReceptionAlreadyClosed)
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:  "internal error",
			param: validID,
			mock: func() {
				mockService.EXPECT().CloseLastReception(gomock.Any(), gomock.Any()).Return(nil, errors.New("unexpected error"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			req, _ := http.NewRequest(http.MethodPost, "/pvz/"+tt.param+"/close_last_reception", nil)
			c.Params = []gin.Param{{Key: "pvzId", Value: tt.param}}
			c.Request = req

			tt.mock()
			h.CloseLastReception(c)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}
