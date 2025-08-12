package service

import (
	"context"
	"time"

	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"
	"github.com/sirupsen/logrus"

	"github.com/senyabanana/pvz-service/internal/entity"
	"github.com/senyabanana/pvz-service/internal/infrastructure/jwtutil"
	"github.com/senyabanana/pvz-service/internal/infrastructure/security"
	"github.com/senyabanana/pvz-service/internal/repository"
)

type UserService struct {
	repo      repository.UserRepository
	trManager *manager.Manager
	JWTSecret string
	log       *logrus.Logger
}

func NewUserService(repo repository.UserRepository, trManager *manager.Manager, secretKey string, log *logrus.Logger) *UserService {
	return &UserService{
		repo:      repo,
		trManager: trManager,
		JWTSecret: secretKey,
		log:       log,
	}
}

func (s *UserService) RegisterUser(ctx context.Context, user *entity.User) error {
	if !entity.IsValidUserRole(user.Role) {
		s.log.Warnf("invalid user role during registration: %s", user.Role)
		return entity.ErrInvalidUserRole
	}

	s.log.Infof("attempt to register user: email=%s", user.Email)

	return s.trManager.Do(ctx, func(ctx context.Context) error {
		exist, err := s.repo.IsEmailExists(ctx, user.Email)
		if err != nil {
			s.log.Errorf("failed to check email existence: %v", err)
			return err
		}

		if exist {
			s.log.Warnf("registration blocked: email already exists: %s", user.Email)
			return entity.ErrEmailTaken
		}

		hash, err := security.GeneratePasswordHash(user.Password)
		if err != nil {
			s.log.Errorf("failed to hash password for email=%s: %v", user.Email, err)
			return err
		}

		user.Password = hash
		user.CreatedAt = time.Now()

		if err := s.repo.CreateUser(ctx, user); err != nil {
			s.log.Errorf("failed to create user: email=%s: %v", user.Email, err)
			return err
		}

		s.log.Infof("user registered successfully: id=%s, email=%s", user.ID.String(), user.Email)
		return nil
	})
}

func (s *UserService) LoginUser(ctx context.Context, email, password string) (string, error) {
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		s.log.Warnf("user not found: %s", err)
		return "", entity.ErrInvalidCredentials
	}

	if err := security.ComparePassword(password, user.Password); err != nil {
		s.log.Warnf("invalid password for user: %s", email)
		return "", entity.ErrInvalidCredentials
	}

	s.log.Infof("user logged in successfully: id=%s, email=%s", user.ID.String(), user.Email)

	token, err := jwtutil.GenerateToken(user.ID.String(), string(user.Role), s.JWTSecret, 2*time.Hour)
	if err != nil {
		s.log.Warnf("failed to generate JWT: %v", err)
		return "", err
	}

	return token, nil
}
