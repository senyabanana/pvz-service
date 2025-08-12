package handler

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"github.com/senyabanana/pvz-service/internal/dto"
	"github.com/senyabanana/pvz-service/internal/entity"
	"github.com/senyabanana/pvz-service/internal/infrastructure/jwtutil"
	"github.com/senyabanana/pvz-service/internal/service"
)

type AuthHandler struct {
	service   service.Authorization
	JWTSecret string
	log       *logrus.Logger
}

func NewAuthHandler(service service.Authorization, secretKey string, log *logrus.Logger) *AuthHandler {
	return &AuthHandler{
		service:   service,
		JWTSecret: secretKey,
		log:       log,
	}
}

// DummyLogin godoc
// @Summary Dummy Login
// @Tags auth
// @Description Получение токена без регистрации (по роли)
// @Accept json
// @Produce json
// @Param input body dto.DummyLoginRequest true "User role"
// @Success 200 {object} dto.TokenResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /dummyLogin [post]
func (h *AuthHandler) DummyLogin(c *gin.Context) {
	var req dto.DummyLoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Warnf("invalid input: %v", err)
		dto.BadRequest(c, "role must be: client, employee, or moderator")
		return
	}

	role := entity.UserRole(req.Role)
	if !entity.IsValidUserRole(role) {
		h.log.Warnf("invalid dummy login role: %s", req.Role)
		dto.BadRequest(c, "invalid role")
		return
	}

	userID := uuid.New().String()
	token, err := jwtutil.GenerateToken(userID, req.Role, h.JWTSecret, 2*time.Hour)
	if err != nil {
		h.log.Errorf("failed to generate JWT: %v", err)
		dto.InternalError(c, "token generation error")
		return
	}

	c.JSON(http.StatusOK, dto.TokenResponse{Token: token})
}

// Register godoc
// @Summary Register User
// @Tags auth
// @Description Регистрация нового пользователя
// @Accept json
// @Produce json
// @Param input body dto.RegisterRequest true "User credentials"
// @Success 201 {object} dto.UserResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Warnf("invalid register input: %v", err)
		dto.BadRequest(c, "invalid email, password or role")
		return
	}

	user := &entity.User{
		Email:    req.Email,
		Password: req.Password,
		Role:     entity.UserRole(req.Role),
	}

	if err := h.service.RegisterUser(c.Request.Context(), user); err != nil {
		switch {
		case errors.Is(err, entity.ErrEmailTaken):
			dto.BadRequest(c, "email already taken")
			return
		case errors.Is(err, entity.ErrInvalidUserRole):
			dto.BadRequest(c, "invalid user role")
			return
		default:
			dto.InternalError(c, "failed to register user")
			return
		}
	}

	c.JSON(http.StatusCreated, dto.UserResponse{
		ID:    user.ID.String(),
		Email: user.Email,
		Role:  string(user.Role),
	})
}

// Login godoc
// @Summary Login User
// @Tags auth
// @Description Авторизация пользователя и получение токена
// @Accept json
// @Produce json
// @Param input body dto.LoginRequest true "Login credentials"
// @Success 200 {object} dto.TokenResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Warnf("invalid login input: %v", err)
		dto.BadRequest(c, "invalid email or password format")
		return
	}

	token, err := h.service.LoginUser(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, entity.ErrInvalidCredentials) {
			h.log.Infof("login failed: invalid credentials for email=%s", req.Email)
			dto.Unauthorized(c, "invalid email or password")
			return
		}

		h.log.Errorf("login error: %v", err)
		dto.InternalError(c, "login failed due to internal error")
		return
	}

	c.JSON(http.StatusOK, dto.TokenResponse{Token: token})
}
