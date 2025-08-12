package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/senyabanana/pvz-service/internal/service"
)

type Authorization interface {
	DummyLogin(c *gin.Context)
	Register(c *gin.Context)
	Login(c *gin.Context)
}

type PVZOperations interface {
	CreatePVZ(c *gin.Context)
	GetFullInfoPVZ(c *gin.Context)
}

type ReceptionOperations interface {
	CreateReception(c *gin.Context)
	CloseLastReception(c *gin.Context)
}

type ProductOperations interface {
	AddProduct(c *gin.Context)
	DeleteLastProduct(c *gin.Context)
}

type Handler struct {
	Authorization
	PVZOperations
	ReceptionOperations
	ProductOperations
}

func NewHandler(services *service.Service, secretKey string, log *logrus.Logger) *Handler {
	return &Handler{
		Authorization:       NewAuthHandler(services, secretKey, log),
		PVZOperations:       NewPVZHandler(services, log),
		ReceptionOperations: NewReceptionHandler(services, log),
		ProductOperations:   NewProductHandler(services, log),
	}
}
