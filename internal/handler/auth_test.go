package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/senyabanana/pvz-service/internal/dto"
	"github.com/senyabanana/pvz-service/internal/entity"
	mocks "github.com/senyabanana/pvz-service/internal/service/mocks"
)

const testSecretKey = "secret"

func TestAuthHandler_DummyLogin(t *testing.T) {
	mockLog := logrus.New()
	h := NewAuthHandler(nil, testSecretKey, mockLog)

	router := gin.New()
	router.POST("/dummyLogin", h.DummyLogin)

	tests := []struct {
		name         string
		requestBody  interface{}
		expectedCode int
	}{
		{
			name:         "success",
			requestBody:  dto.DummyLoginRequest{Role: "client"},
			expectedCode: http.StatusOK,
		},
		{
			name:         "invalid json",
			requestBody:  `{"role":123}`,
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "invalid role",
			requestBody:  dto.DummyLoginRequest{Role: "admin"},
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body []byte
			var err error

			switch v := tt.requestBody.(type) {
			case string:
				body = []byte(v)
			default:
				body, err = json.Marshal(v)
				assert.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodPost, "/dummyLogin", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)
			assert.Equal(t, tt.expectedCode, w.Code)
		})
	}
}

func TestAuthHandler_Register(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockAuthorization(ctrl)
	mockLog := logrus.New()
	h := NewAuthHandler(mockService, testSecretKey, mockLog)

	router := gin.New()
	router.POST("/register", h.Register)

	tests := []struct {
		name         string
		input        dto.RegisterRequest
		setup        func()
		expectedCode int
	}{
		{
			name: "success",
			input: dto.RegisterRequest{
				Email:    "user@example.com",
				Password: "123456",
				Role:     "client",
			},
			setup: func() {
				mockService.EXPECT().RegisterUser(gomock.Any(), gomock.Any()).Return(nil)
			},
			expectedCode: http.StatusCreated,
		},
		{
			name: "email taken",
			input: dto.RegisterRequest{
				Email:    "user@example.com",
				Password: "123456",
				Role:     "client",
			},
			setup: func() {
				mockService.EXPECT().RegisterUser(gomock.Any(), gomock.Any()).Return(entity.ErrEmailTaken)
			},
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "invalid role",
			input: dto.RegisterRequest{
				Email:    "user@example.com",
				Password: "123456",
				Role:     "unknown",
			},
			setup: func() {
				mockService.EXPECT().RegisterUser(gomock.Any(), gomock.Any()).Return(entity.ErrInvalidUserRole)
			},
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "internal error",
			input: dto.RegisterRequest{
				Email:    "user@example.com",
				Password: "123456",
				Role:     "client",
			},
			setup: func() {
				mockService.EXPECT().RegisterUser(gomock.Any(), gomock.Any()).Return(errors.New("db down"))
			},
			expectedCode: http.StatusInternalServerError,
		},
		{
			name: "invalid json payload for register",
			input: dto.RegisterRequest{
				Email:    "invalid",
				Password: "33333333",
				Role:     "client",
			},
			setup:        func() {},
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			body, _ := json.Marshal(tt.input)
			req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)
			assert.Equal(t, tt.expectedCode, w.Code)
		})
	}
}

func TestAuthHandler_Login(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockAuthorization(ctrl)
	mockLog := logrus.New()
	h := NewAuthHandler(mockService, testSecretKey, mockLog)

	router := gin.New()
	router.POST("/login", h.Login)

	tests := []struct {
		name         string
		input        dto.LoginRequest
		setup        func()
		expectedCode int
	}{
		{
			name: "success",
			input: dto.LoginRequest{
				Email:    "user@example.com",
				Password: "success_pass",
			},
			setup: func() {
				mockService.EXPECT().LoginUser(gomock.Any(), "user@example.com", "success_pass").Return("jwt-token", nil)
			},
			expectedCode: http.StatusOK,
		},
		{
			name: "invalid credentials",
			input: dto.LoginRequest{
				Email:    "user@example.com",
				Password: "wrong_pass",
			},
			setup: func() {
				mockService.EXPECT().LoginUser(gomock.Any(), "user@example.com", "wrong_pass").Return("", entity.ErrInvalidCredentials)
			},
			expectedCode: http.StatusUnauthorized,
		},
		{
			name: "internal error",
			input: dto.LoginRequest{
				Email:    "user@example.com",
				Password: "error_pass",
			},
			setup: func() {
				mockService.EXPECT().LoginUser(gomock.Any(), "user@example.com", "error_pass").Return("", errors.New("db down"))
			},
			expectedCode: http.StatusInternalServerError,
		},
		{
			name: "invalid json payload for register",
			input: dto.LoginRequest{
				Email:    "invalid",
				Password: "33333333",
			},
			setup:        func() {},
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			body, _ := json.Marshal(tt.input)
			req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)
			assert.Equal(t, tt.expectedCode, w.Code)
		})
	}
}
