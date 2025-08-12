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

func TestPVZHandler_CreatePVZ(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockPVZOperations(ctrl)
	mockLog := logrus.New()
	h := NewPVZHandler(mockService, mockLog)

	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		input      string
		mock       func()
		wantStatus int
	}{
		{
			name:  "success",
			input: `{"city":"moscow"}`,
			mock: func() {
				mockService.EXPECT().
					CreatePVZ(gomock.Any(), "moscow").
					Return(&entity.PVZ{
						ID:               uuid.New(),
						RegistrationDate: time.Now(),
						City:             "moscow",
					}, nil)
			},
			wantStatus: http.StatusCreated,
		},
		{
			name:       "invalid JSON",
			input:      `{"city":123}`,
			mock:       func() {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:  "unsupported city",
			input: `{"city":"unknown"}`,
			mock: func() {
				mockService.EXPECT().
					CreatePVZ(gomock.Any(), "unknown").
					Return(nil, entity.ErrInvalidCity)
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:  "internal error",
			input: `{"city":"moscow"}`,
			mock: func() {
				mockService.EXPECT().
					CreatePVZ(gomock.Any(), "moscow").
					Return(nil, errors.New("db failure"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(rr)

			ctx.Request, _ = http.NewRequest(http.MethodPost, "/pvz", bytes.NewBufferString(tt.input))
			ctx.Request.Header.Set("Content-Type", "application/json")

			tt.mock()
			h.CreatePVZ(ctx)

			assert.Equal(t, tt.wantStatus, rr.Code)
		})
	}
}

func TestPVZHandler_GetFullInfoPVZ(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockPVZOperations(ctrl)
	mockLog := logrus.New()
	h := NewPVZHandler(mockService, mockLog)

	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		input      string
		mock       func()
		wantStatus int
	}{
		{
			name:  "success without params",
			input: "",
			mock: func() {
				mockService.EXPECT().
					GetFullPVZInfo(gomock.Any(), nil, nil, 1, 10).
					Return([]entity.FullPVZInfo{}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "invalid startDate",
			input:      "?startDate=invalid-date",
			mock:       func() {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:  "internal error",
			input: "",
			mock: func() {
				mockService.EXPECT().
					GetFullPVZInfo(gomock.Any(), nil, nil, 1, 10).
					Return(nil, assert.AnError)
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(rr)

			req, _ := http.NewRequest(http.MethodGet, "/pvz"+tt.input, nil)
			ctx.Request = req

			tt.mock()
			h.GetFullInfoPVZ(ctx)

			assert.Equal(t, tt.wantStatus, rr.Code)
		})
	}
}

func TestParseQueryTime(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockLog := logrus.New()

	tests := []struct {
		name       string
		input      string
		expectNil  bool
		statusCode int
	}{
		{
			name:      "empty string",
			input:     "",
			expectNil: true,
		},
		{
			name:      "valid RFC3339 date",
			input:     "2023-05-01T10:00:00Z",
			expectNil: false,
		},
		{
			name:       "invalid format",
			input:      "2023/05/01",
			expectNil:  true,
			statusCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			res := parseQueryTime(tt.input, "testField", c, mockLog)

			if tt.expectNil {
				assert.Nil(t, res)
			} else {
				assert.NotNil(t, res)
				assert.Equal(t, tt.input, res.Format(time.RFC3339))
			}

			if tt.statusCode != 0 {
				assert.Equal(t, tt.statusCode, w.Code)
			}
		})
	}
}

func TestConvertToResponse(t *testing.T) {
	id := uuid.New()
	now := time.Now()

	full := []entity.FullPVZInfo{
		{
			PVZ: entity.PVZ{
				ID:               id,
				RegistrationDate: now,
				City:             entity.CityMoscow,
			},
			Receptions: []entity.ReceptionWithProducts{
				{
					Reception: entity.Reception{
						ID:        uuid.New(),
						DateTime:  now,
						PVZID:     id,
						Status:    entity.StatusClosed,
						CreatedAt: now,
						ClosedAt:  &now,
					},
					Products: []entity.Product{
						{
							ID:          uuid.New(),
							DateTime:    now,
							Type:        entity.ProductElectronics,
							ReceptionID: id,
						},
					},
				},
			},
		},
	}

	resp := convertToResponse(full)
	assert.Len(t, resp, 1)
	assert.Equal(t, string(entity.CityMoscow), resp[0].PVZ.City)
	assert.Len(t, resp[0].Receptions, 1)
	assert.Len(t, resp[0].Receptions[0].Products, 1)
}
